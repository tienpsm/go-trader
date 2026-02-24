package persistence

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tienpsm/go-trader/matching"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func newManager(t *testing.T) *matching.MarketManager {
	t.Helper()
	mm := matching.NewMarketManager()
	mm.EnableMatching()
	sym := matching.NewSymbol(1, "AAPL")
	if code := mm.AddSymbol(sym); code != matching.ErrorOK {
		t.Fatalf("AddSymbol: %s", code)
	}
	if code := mm.AddOrderBook(sym); code != matching.ErrorOK {
		t.Fatalf("AddOrderBook: %s", code)
	}
	return mm
}

func newLimitOrder(id uint64, side matching.OrderSide, price, qty uint64) matching.Order {
	return matching.Order{
		ID:                 id,
		SymbolID:           1,
		Type:               matching.OrderTypeLimit,
		Side:               side,
		Price:              price,
		Quantity:           qty,
		LeavesQuantity:     qty,
		MaxVisibleQuantity: matching.MaxVisibleQuantity,
		Slippage:           matching.MaxSlippage,
	}
}

// ─── encoding round-trip ──────────────────────────────────────────────────────

func TestEncodeDecodeNewOrder(t *testing.T) {
	orig := MatchingEvent{
		Type:      EventNewOrder,
		Timestamp: 1234567890,
		Order:     newLimitOrder(42, matching.OrderSideBuy, 10000, 100),
	}

	data, err := encodeEvent(orig)
	if err != nil {
		t.Fatalf("encodeEvent: %v", err)
	}

	// Wrap in a reader so we can call decodeEvent.
	r := newByteReader(data)
	got, err := decodeEvent(r)
	if err != nil {
		t.Fatalf("decodeEvent: %v", err)
	}

	if got.Type != orig.Type {
		t.Errorf("Type: got %d, want %d", got.Type, orig.Type)
	}
	if got.Timestamp != orig.Timestamp {
		t.Errorf("Timestamp: got %d, want %d", got.Timestamp, orig.Timestamp)
	}
	if got.Order != orig.Order {
		t.Errorf("Order: got %+v, want %+v", got.Order, orig.Order)
	}
}

func TestEncodeDecodeCancelOrder(t *testing.T) {
	orig := MatchingEvent{
		Type:      EventCancelOrder,
		Timestamp: 9876543210,
		OrderID:   77,
	}

	data, err := encodeEvent(orig)
	if err != nil {
		t.Fatalf("encodeEvent: %v", err)
	}
	r := newByteReader(data)
	got, err := decodeEvent(r)
	if err != nil {
		t.Fatalf("decodeEvent: %v", err)
	}

	if got.Type != orig.Type {
		t.Errorf("Type: got %d, want %d", got.Type, orig.Type)
	}
	if got.OrderID != orig.OrderID {
		t.Errorf("OrderID: got %d, want %d", got.OrderID, orig.OrderID)
	}
}

// ─── journal ─────────────────────────────────────────────────────────────────

func TestJournal_AppendAndReadAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.journal")

	j, err := OpenJournal(path)
	if err != nil {
		t.Fatalf("OpenJournal: %v", err)
	}

	events := []MatchingEvent{
		{Type: EventNewOrder, Timestamp: 1, Order: newLimitOrder(1, matching.OrderSideBuy, 100, 10)},
		{Type: EventNewOrder, Timestamp: 2, Order: newLimitOrder(2, matching.OrderSideSell, 105, 5)},
		{Type: EventCancelOrder, Timestamp: 3, OrderID: 1},
	}
	for _, e := range events {
		if err := j.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	if err := j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != len(events) {
		t.Fatalf("ReadAll: got %d events, want %d", len(got), len(events))
	}
	for i, e := range events {
		if got[i].Type != e.Type {
			t.Errorf("[%d] Type: got %d, want %d", i, got[i].Type, e.Type)
		}
		if got[i].Timestamp != e.Timestamp {
			t.Errorf("[%d] Timestamp: got %d, want %d", i, got[i].Timestamp, e.Timestamp)
		}
	}
}

func TestJournal_ReadAllMissing(t *testing.T) {
	// ReadAll on a non-existent file should return nil, nil.
	events, err := ReadAll("/tmp/this-file-should-not-exist-go-trader-test.journal")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if events != nil {
		t.Fatalf("expected nil events for missing file, got: %v", events)
	}
}

func TestJournal_FlushTimer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "flush.journal")

	j, err := OpenJournal(path)
	if err != nil {
		t.Fatalf("OpenJournal: %v", err)
	}
	defer j.Close()

	e := MatchingEvent{Type: EventNewOrder, Timestamp: time.Now().UnixNano(),
		Order: newLimitOrder(99, matching.OrderSideBuy, 50, 10)}
	if err := j.Append(e); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// The flush timer fires every 10 ms; wait long enough to let it trigger.
	time.Sleep(50 * time.Millisecond)

	// File must be non-empty.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty journal file after flush timer")
	}
}

// ─── snapshot ────────────────────────────────────────────────────────────────

func TestSnapshot_SaveAndLoadLatest(t *testing.T) {
	dir := t.TempDir()
	sp, err := NewSnapshotter(dir)
	if err != nil {
		t.Fatalf("NewSnapshotter: %v", err)
	}

	snap := Snapshot{
		Timestamp: 42000000000,
		Symbols:   []matching.Symbol{{ID: 1, Name: "AAPL"}, {ID: 2, Name: "GOOGL"}},
		Orders: []matching.Order{
			newLimitOrder(1, matching.OrderSideBuy, 10000, 100),
			newLimitOrder(2, matching.OrderSideSell, 10100, 50),
		},
	}
	snap.Orders[0].ExecutedQuantity = 30
	snap.Orders[0].LeavesQuantity = 70

	if err := sp.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := sp.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest: %v", err)
	}
	if got == nil {
		t.Fatal("LoadLatest returned nil")
	}

	if got.Timestamp != snap.Timestamp {
		t.Errorf("Timestamp: got %d, want %d", got.Timestamp, snap.Timestamp)
	}
	if len(got.Symbols) != len(snap.Symbols) {
		t.Errorf("Symbols len: got %d, want %d", len(got.Symbols), len(snap.Symbols))
	}
	if len(got.Orders) != len(snap.Orders) {
		t.Errorf("Orders len: got %d, want %d", len(got.Orders), len(snap.Orders))
	}
	// Verify partial-fill state is preserved.
	if got.Orders[0].LeavesQuantity != 70 {
		t.Errorf("LeavesQuantity: got %d, want 70", got.Orders[0].LeavesQuantity)
	}
}

func TestSnapshotter_LoadLatest_NoSnapshots(t *testing.T) {
	dir := t.TempDir()
	sp, err := NewSnapshotter(dir)
	if err != nil {
		t.Fatalf("NewSnapshotter: %v", err)
	}
	snap, err := sp.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest: %v", err)
	}
	if snap != nil {
		t.Errorf("expected nil snapshot, got %+v", snap)
	}
}

func TestSnapshotter_LoadLatest_PicksMostRecent(t *testing.T) {
	dir := t.TempDir()
	sp, err := NewSnapshotter(dir)
	if err != nil {
		t.Fatalf("NewSnapshotter: %v", err)
	}

	// Write two snapshots; the second has a higher timestamp.
	for _, ts := range []int64{100, 200} {
		s := Snapshot{
			Timestamp: ts,
			Symbols:   []matching.Symbol{{ID: 1, Name: "SYM"}},
		}
		if err := sp.Save(s); err != nil {
			t.Fatalf("Save ts=%d: %v", ts, err)
		}
	}

	got, err := sp.LoadLatest()
	if err != nil {
		t.Fatalf("LoadLatest: %v", err)
	}
	if got.Timestamp != 200 {
		t.Errorf("expected ts=200, got %d", got.Timestamp)
	}
}

// ─── recovery ────────────────────────────────────────────────────────────────

func TestRecover_FromScratch(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "test.journal")

	mm := newManager(t)

	// Recovery with no data should be a no-op.
	if err := Recover(mm, journalPath, filepath.Join(dir, "snapshots")); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	if len(mm.Orders()) != 0 {
		t.Errorf("expected 0 orders, got %d", len(mm.Orders()))
	}
}

func TestRecover_JournalOnly(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "test.journal")
	snapshotDir := filepath.Join(dir, "snapshots")

	// Write two orders to the journal, then cancel one.
	j, err := OpenJournal(journalPath)
	if err != nil {
		t.Fatalf("OpenJournal: %v", err)
	}
	orders := []matching.Order{
		newLimitOrder(1, matching.OrderSideBuy, 10000, 100),
		newLimitOrder(2, matching.OrderSideSell, 10500, 50),
	}
	for i, o := range orders {
		e := MatchingEvent{Type: EventNewOrder, Timestamp: int64(i + 1), Order: o}
		if err := j.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	// Cancel order 1.
	if err := j.Append(MatchingEvent{Type: EventCancelOrder, Timestamp: 3, OrderID: 1}); err != nil {
		t.Fatalf("Append cancel: %v", err)
	}
	if err := j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Recover into a fresh manager.
	mm := newManager(t)
	if err := Recover(mm, journalPath, snapshotDir); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	// Only order 2 should survive.
	if mm.GetOrder(1) != nil {
		t.Error("order 1 should have been cancelled")
	}
	if mm.GetOrder(2) == nil {
		t.Error("order 2 should exist")
	}
}

func TestRecover_SnapshotAndJournal(t *testing.T) {
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "test.journal")
	snapshotDir := filepath.Join(dir, "snapshots")

	// Snapshot contains order 1 (partially filled).
	sp, err := NewSnapshotter(snapshotDir)
	if err != nil {
		t.Fatalf("NewSnapshotter: %v", err)
	}
	o1 := newLimitOrder(1, matching.OrderSideBuy, 10000, 100)
	o1.ExecutedQuantity = 40
	o1.LeavesQuantity = 60
	snap := Snapshot{
		Timestamp: 1000,
		Symbols:   []matching.Symbol{{ID: 1, Name: "AAPL"}},
		Orders:    []matching.Order{o1},
	}
	if err := sp.Save(snap); err != nil {
		t.Fatalf("Save snapshot: %v", err)
	}

	// Journal: one event before the snapshot (should be skipped) and one after.
	j, err := OpenJournal(journalPath)
	if err != nil {
		t.Fatalf("OpenJournal: %v", err)
	}
	// ts=500 < snapshotTS=1000, must be skipped.
	_ = j.Append(MatchingEvent{
		Type: EventNewOrder, Timestamp: 500,
		Order: newLimitOrder(99, matching.OrderSideSell, 9999, 10),
	})
	// ts=2000 > snapshotTS=1000, must be applied.
	_ = j.Append(MatchingEvent{
		Type: EventNewOrder, Timestamp: 2000,
		Order: newLimitOrder(2, matching.OrderSideSell, 11000, 20),
	})
	if err := j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Recover into a fresh manager.
	mm := newManager(t)
	if err := Recover(mm, journalPath, snapshotDir); err != nil {
		t.Fatalf("Recover: %v", err)
	}

	// Order 1 from snapshot with preserved execution state.
	node1 := mm.GetOrder(1)
	if node1 == nil {
		t.Fatal("order 1 should exist after recovery")
	}
	if node1.LeavesQuantity != 60 {
		t.Errorf("LeavesQuantity: got %d, want 60", node1.LeavesQuantity)
	}

	// Order 99 was before the snapshot and must NOT exist.
	if mm.GetOrder(99) != nil {
		t.Error("order 99 should be skipped (before snapshot)")
	}

	// Order 2 was after the snapshot and must exist.
	if mm.GetOrder(2) == nil {
		t.Error("order 2 should exist after recovery")
	}
}

// ─── manager ─────────────────────────────────────────────────────────────────

func TestManager_AddAndCancel(t *testing.T) {
	dir := t.TempDir()
	mm := newManager(t)

	mgr, err := NewManager(mm, filepath.Join(dir, "test.journal"), filepath.Join(dir, "snapshots"))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer mgr.Close()

	o := newLimitOrder(10, matching.OrderSideBuy, 5000, 50)
	if err := mgr.AddOrder(o); err != nil {
		t.Fatalf("AddOrder: %v", err)
	}
	if mm.GetOrder(10) == nil {
		t.Error("order 10 should exist in engine")
	}

	if err := mgr.CancelOrder(10); err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if mm.GetOrder(10) != nil {
		t.Error("order 10 should have been deleted")
	}
}

func TestManager_TakeSnapshot(t *testing.T) {
	dir := t.TempDir()
	mm := newManager(t)

	mgr, err := NewManager(mm, filepath.Join(dir, "test.journal"), filepath.Join(dir, "snapshots"))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer mgr.Close()

	o := newLimitOrder(5, matching.OrderSideBuy, 8000, 200)
	if err := mgr.AddOrder(o); err != nil {
		t.Fatalf("AddOrder: %v", err)
	}

	errCh := make(chan error, 1)
	mgr.TakeSnapshot(errCh)
	if err := <-errCh; err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}

	// Verify a snapshot file was created.
	entries, _ := os.ReadDir(filepath.Join(dir, "snapshots"))
	if len(entries) == 0 {
		t.Error("expected at least one snapshot file")
	}
}

// ─── internal helper ─────────────────────────────────────────────────────────

// newByteReader wraps a byte slice in an io.Reader for decodeEvent.
type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader { return &byteReader{data: data} }

func (b *byteReader) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}
