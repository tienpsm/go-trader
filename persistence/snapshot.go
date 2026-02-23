package persistence

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/tienpsm/go-trader/matching"
)

// snapshotMagic is written at the start of every snapshot file so that corrupt
// or foreign files are rejected quickly.
var snapshotMagic = [8]byte{'G', 'T', 'S', 'N', 'A', 'P', 0, 1}

// Snapshot is the full, self-contained state of the matching engine at a single
// point in time.  Symbols carry their order-book association implicitly: an
// order book exists for every Symbol in the snapshot.
type Snapshot struct {
	// Timestamp is the Unix nanosecond at which the snapshot was captured.
	Timestamp int64
	// Symbols is the ordered list of all active symbols.
	Symbols []matching.Symbol
	// Orders is the list of all active orders (with their current execution
	// state) across all order books.
	Orders []matching.Order
}

// Snapshotter manages snapshot files inside a directory.
type Snapshotter struct {
	dir string
}

// NewSnapshotter creates a Snapshotter that stores files in dir.
// dir is created if it does not exist.
func NewSnapshotter(dir string) (*Snapshotter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Snapshotter{dir: dir}, nil
}

// snapshotPath returns the full path for a snapshot with the given timestamp.
func (s *Snapshotter) snapshotPath(ts int64) string {
	return filepath.Join(s.dir, fmt.Sprintf("snapshot-%d.snap", ts))
}

// Save serialises snap and writes it to a zstd-compressed file.
// The file is written atomically: data is first flushed to a temp file and then
// renamed so that a crash mid-write never leaves a corrupt snapshot.
func (s *Snapshotter) Save(snap Snapshot) error {
	dst := s.snapshotPath(snap.Timestamp)
	tmp := dst + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	enc, err := zstd.NewWriter(f)
	if err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}

	if err := writeSnapshot(enc, snap); err != nil {
		_ = enc.Close()
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := enc.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// LoadLatest finds the most-recent snapshot in the directory and deserialises
// it.  It returns nil (with no error) when no snapshot exists yet.
func (s *Snapshotter) LoadLatest() (*Snapshot, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Collect snapshot timestamps so we can find the maximum.
	var timestamps []int64
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "snapshot-") || !strings.HasSuffix(name, ".snap") {
			continue
		}
		tsStr := strings.TrimPrefix(name, "snapshot-")
		tsStr = strings.TrimSuffix(tsStr, ".snap")
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			continue
		}
		timestamps = append(timestamps, ts)
	}
	if len(timestamps) == 0 {
		return nil, nil
	}

	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] > timestamps[j] })
	path := s.snapshotPath(timestamps[0])

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec, err := zstd.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	return readSnapshot(dec)
}

// TakeSnapshot captures the current state of mm and saves it to disk.
// It uses a simple copy-on-write approach: the caller-facing lock is held only
// for the brief clone operation; the actual I/O is performed without holding
// any locks from the matching engine.
func (s *Snapshotter) TakeSnapshot(mm *matching.MarketManager) error {
	snap := captureSnapshot(mm)
	return s.Save(snap)
}

// captureSnapshot clones the minimal state needed for recovery from mm.
// The matching engine should not be modified concurrently while this runs
// (a brief pause is sufficient; the actual disk write happens after).
func captureSnapshot(mm *matching.MarketManager) Snapshot {
	ts := time.Now().UnixNano()

	symbols := make([]matching.Symbol, 0, len(mm.Symbols()))
	for _, sym := range mm.Symbols() {
		symbols = append(symbols, *sym)
	}

	orders := make([]matching.Order, 0, len(mm.Orders()))
	for _, node := range mm.Orders() {
		orders = append(orders, node.Order)
	}

	return Snapshot{
		Timestamp: ts,
		Symbols:   symbols,
		Orders:    orders,
	}
}

// ─── Binary snapshot wire format ────────────────────────────────────────────
//
// All integers are big-endian.
//
//	 8 bytes – magic
//	 8 bytes – Timestamp (int64)
//	 4 bytes – number of symbols (uint32)
//	   per symbol:
//	     4 bytes – ID (uint32)
//	     1 byte  – name length (uint8)
//	     N bytes – name (UTF-8)
//	 4 bytes – number of orders (uint32)
//	   per order: 87 bytes (orderWireSize)

func writeSnapshot(w io.Writer, snap Snapshot) error {
	// Magic
	if _, err := w.Write(snapshotMagic[:]); err != nil {
		return err
	}

	// Timestamp
	var buf8 [8]byte
	binary.BigEndian.PutUint64(buf8[:], uint64(snap.Timestamp))
	if _, err := w.Write(buf8[:]); err != nil {
		return err
	}

	// Symbols
	var buf4 [4]byte
	binary.BigEndian.PutUint32(buf4[:], uint32(len(snap.Symbols)))
	if _, err := w.Write(buf4[:]); err != nil {
		return err
	}
	for _, sym := range snap.Symbols {
		binary.BigEndian.PutUint32(buf4[:], sym.ID)
		if _, err := w.Write(buf4[:]); err != nil {
			return err
		}
		name := sym.Name
		if len(name) > 255 {
			name = name[:255]
		}
		if _, err := w.Write([]byte{uint8(len(name))}); err != nil {
			return err
		}
		if len(name) > 0 {
			if _, err := w.Write([]byte(name)); err != nil {
				return err
			}
		}
	}

	// Orders
	binary.BigEndian.PutUint32(buf4[:], uint32(len(snap.Orders)))
	if _, err := w.Write(buf4[:]); err != nil {
		return err
	}
	orderBuf := make([]byte, orderWireSize)
	for _, o := range snap.Orders {
		marshalOrder(orderBuf, o)
		if _, err := w.Write(orderBuf); err != nil {
			return err
		}
	}
	return nil
}

func readSnapshot(r io.Reader) (*Snapshot, error) {
	// Magic
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, fmt.Errorf("persistence: reading snapshot magic: %w", err)
	}
	if magic != snapshotMagic {
		return nil, fmt.Errorf("persistence: invalid snapshot magic")
	}

	// Timestamp
	var buf8 [8]byte
	if _, err := io.ReadFull(r, buf8[:]); err != nil {
		return nil, fmt.Errorf("persistence: reading snapshot timestamp: %w", err)
	}
	snap := &Snapshot{
		Timestamp: int64(binary.BigEndian.Uint64(buf8[:])),
	}

	// Symbols
	var buf4 [4]byte
	if _, err := io.ReadFull(r, buf4[:]); err != nil {
		return nil, fmt.Errorf("persistence: reading symbol count: %w", err)
	}
	symCount := binary.BigEndian.Uint32(buf4[:])
	snap.Symbols = make([]matching.Symbol, 0, symCount)
	for i := uint32(0); i < symCount; i++ {
		if _, err := io.ReadFull(r, buf4[:]); err != nil {
			return nil, fmt.Errorf("persistence: reading symbol ID: %w", err)
		}
		id := binary.BigEndian.Uint32(buf4[:])

		var lenBuf [1]byte
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			return nil, fmt.Errorf("persistence: reading symbol name length: %w", err)
		}
		nameLen := int(lenBuf[0])
		nameBuf := make([]byte, nameLen)
		if nameLen > 0 {
			if _, err := io.ReadFull(r, nameBuf); err != nil {
				return nil, fmt.Errorf("persistence: reading symbol name: %w", err)
			}
		}
		snap.Symbols = append(snap.Symbols, matching.Symbol{ID: id, Name: string(nameBuf)})
	}

	// Orders
	if _, err := io.ReadFull(r, buf4[:]); err != nil {
		return nil, fmt.Errorf("persistence: reading order count: %w", err)
	}
	orderCount := binary.BigEndian.Uint32(buf4[:])
	snap.Orders = make([]matching.Order, 0, orderCount)
	orderBuf := make([]byte, orderWireSize)
	for i := uint32(0); i < orderCount; i++ {
		if _, err := io.ReadFull(r, orderBuf); err != nil {
			return nil, fmt.Errorf("persistence: reading order: %w", err)
		}
		snap.Orders = append(snap.Orders, unmarshalOrder(orderBuf))
	}

	return snap, nil
}
