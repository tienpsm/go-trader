package persistence

import (
	"fmt"

	"github.com/tienpsm/go-trader/matching"
)

// Recover restores a MarketManager to its last known state by:
//  1. Loading the most recent snapshot from dir (if any).
//  2. Replaying every journal event whose timestamp is strictly greater than
//     the snapshot timestamp.
//
// mm must be a freshly created, empty MarketManager.
// journalPath is the path to the journal file.
// snapshotDir is the directory that Snapshotter writes snapshots into.
//
// If neither a snapshot nor a journal exists the function is a no-op.
func Recover(mm *matching.MarketManager, journalPath, snapshotDir string) error {
	sp, err := NewSnapshotter(snapshotDir)
	if err != nil {
		return fmt.Errorf("persistence: opening snapshot dir: %w", err)
	}

	// ── 1. Load snapshot ──────────────────────────────────────────────────────
	snap, err := sp.LoadLatest()
	if err != nil {
		return fmt.Errorf("persistence: loading snapshot: %w", err)
	}

	var snapshotTS int64
	if snap != nil {
		if err := applySnapshot(mm, snap); err != nil {
			return fmt.Errorf("persistence: applying snapshot: %w", err)
		}
		snapshotTS = snap.Timestamp
	}

	// ── 2. Replay journal ─────────────────────────────────────────────────────
	events, err := ReadAll(journalPath)
	if err != nil {
		return fmt.Errorf("persistence: reading journal: %w", err)
	}

	for _, e := range events {
		// Skip events already covered by the snapshot.
		if e.Timestamp <= snapshotTS {
			continue
		}
		if err := applyEvent(mm, e); err != nil {
			return fmt.Errorf("persistence: replaying event at ts=%d: %w", e.Timestamp, err)
		}
	}

	return nil
}

// applySnapshot restores symbols and orders from snap into mm.
// Symbols are added first (which implicitly creates their order books), then
// all orders are restored via RestoreOrder so that partial fills are preserved.
func applySnapshot(mm *matching.MarketManager, snap *Snapshot) error {
	for _, sym := range snap.Symbols {
		if code := mm.AddSymbol(sym); code != matching.ErrorOK && code != matching.ErrorSymbolDuplicate {
			return fmt.Errorf("AddSymbol(%d): %s", sym.ID, code)
		}
		if code := mm.AddOrderBook(sym); code != matching.ErrorOK && code != matching.ErrorOrderBookDuplicate {
			return fmt.Errorf("AddOrderBook(%d): %s", sym.ID, code)
		}
	}

	for _, o := range snap.Orders {
		if code := mm.RestoreOrder(o); code != matching.ErrorOK && code != matching.ErrorOrderDuplicate {
			return fmt.Errorf("RestoreOrder(%d): %s", o.ID, code)
		}
	}
	return nil
}

// applyEvent replays a single journal event against mm.
func applyEvent(mm *matching.MarketManager, e MatchingEvent) error {
	switch e.Type {
	case EventNewOrder:
		code := mm.AddOrder(e.Order)
		if code != matching.ErrorOK && code != matching.ErrorOrderDuplicate {
			return fmt.Errorf("AddOrder(%d): %s", e.Order.ID, code)
		}
	case EventCancelOrder:
		code := mm.DeleteOrder(e.OrderID)
		if code != matching.ErrorOK && code != matching.ErrorOrderNotFound {
			return fmt.Errorf("DeleteOrder(%d): %s", e.OrderID, code)
		}
	default:
		return fmt.Errorf("unknown event type %d", e.Type)
	}
	return nil
}
