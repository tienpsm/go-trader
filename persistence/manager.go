package persistence

import (
	"fmt"
	"sync"
	"time"

	"github.com/tienpsm/go-trader/matching"
)

// Manager is the top-level persistence facade.
//
// It wraps a matching.MarketManager and ensures that every order submission or
// cancellation is journalled before being forwarded to the engine.  It also
// provides TakeSnapshot for periodic checkpointing.
//
// Manager is safe for concurrent use: a single mutex serialises writes so that
// journal events and engine calls are always in the correct order.
type Manager struct {
	mu          sync.Mutex
	mm          *matching.MarketManager
	journal     *Journal
	snapshotter *Snapshotter
}

// NewManager opens (or creates) the journal at journalPath, initialises the
// snapshotter in snapshotDir, and returns a ready-to-use Manager.
//
// Call Recover separately before NewManager if you need to restore state from
// a previous run.
func NewManager(
	mm *matching.MarketManager,
	journalPath string,
	snapshotDir string,
) (*Manager, error) {
	j, err := OpenJournal(journalPath)
	if err != nil {
		return nil, fmt.Errorf("persistence: opening journal: %w", err)
	}

	sp, err := NewSnapshotter(snapshotDir)
	if err != nil {
		_ = j.Close()
		return nil, fmt.Errorf("persistence: opening snapshotter: %w", err)
	}

	return &Manager{
		mm:          mm,
		journal:     j,
		snapshotter: sp,
	}, nil
}

// AddOrder journals the order and then submits it to the matching engine.
// The journal write happens under the same lock as the engine call so that no
// engine state change can occur without a prior journal entry.
func (m *Manager) AddOrder(order matching.Order) error {
	event := MatchingEvent{
		Type:      EventNewOrder,
		Timestamp: time.Now().UnixNano(),
		Order:     order,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.journal.Append(event); err != nil {
		return fmt.Errorf("persistence: journalling NewOrder: %w", err)
	}
	if code := m.mm.AddOrder(order); code != matching.ErrorOK {
		return fmt.Errorf("persistence: AddOrder: %w", code.Error())
	}
	return nil
}

// CancelOrder journals the cancellation and then removes the order from the
// matching engine.
func (m *Manager) CancelOrder(orderID uint64) error {
	event := MatchingEvent{
		Type:      EventCancelOrder,
		Timestamp: time.Now().UnixNano(),
		OrderID:   orderID,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.journal.Append(event); err != nil {
		return fmt.Errorf("persistence: journalling CancelOrder: %w", err)
	}
	if code := m.mm.DeleteOrder(orderID); code != matching.ErrorOK {
		return fmt.Errorf("persistence: CancelOrder: %w", code.Error())
	}
	return nil
}

// TakeSnapshot captures the current engine state in a background goroutine.
//
// Copy-on-Write approach:
//  1. Acquire the manager lock briefly to clone (capture) the snapshot data.
//  2. Release the lock so the engine can continue matching.
//  3. Write the snapshot file in a background goroutine.
//
// errCh receives exactly one value when the background goroutine finishes.
// Callers that do not care about the result may pass nil for errCh.
func (m *Manager) TakeSnapshot(errCh chan<- error) {
	// ── Phase 1: clone under lock (microseconds) ──────────────────────────────
	m.mu.Lock()
	snap := captureSnapshot(m.mm)
	m.mu.Unlock()

	// ── Phase 2: write to disk in the background ──────────────────────────────
	go func() {
		err := m.snapshotter.Save(snap)
		if errCh != nil {
			errCh <- err
		}
	}()
}

// MarketManager returns the underlying MarketManager.
// Callers that need direct (non-persisted) access to the engine can use this,
// but note that operations performed directly on the MarketManager are not
// journalled.
func (m *Manager) MarketManager() *matching.MarketManager {
	return m.mm
}

// Close flushes the journal and releases all resources.
func (m *Manager) Close() error {
	return m.journal.Close()
}
