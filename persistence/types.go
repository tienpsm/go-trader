// Package persistence provides Write-Ahead Logging (journal), snapshotting, and
// recovery for the matching engine's order book state.
//
// Architecture overview:
//
//	Manager                 – top-level facade; wraps MarketManager
//	  ├── Journal           – append-only binary WAL with batch-flush
//	  ├── Snapshotter       – zstd-compressed periodic snapshots
//	  └── Recover()         – load latest snapshot + replay journal on startup
package persistence

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tienpsm/go-trader/matching"
)

// EventType identifies the kind of event stored in the journal.
type EventType uint8

const (
	// EventNewOrder is written when a new order is submitted.
	EventNewOrder EventType = iota + 1
	// EventCancelOrder is written when an order is cancelled.
	EventCancelOrder
)

// MatchingEvent is the unit persisted to the journal.
// It carries a nanosecond-precision timestamp so recovery can skip events that
// are already reflected in a snapshot.
type MatchingEvent struct {
	// Type distinguishes new-order from cancel-order events.
	Type EventType
	// Timestamp is Unix nanoseconds at the time the event was accepted.
	Timestamp int64
	// Order is the full order state (for EventNewOrder).
	Order matching.Order
	// OrderID is used for EventCancelOrder.
	OrderID uint64
}

// orderWireSize is the fixed byte size of a serialised matching.Order.
// Layout (all big-endian):
//
//	 8 – ID
//	 4 – SymbolID
//	 1 – Type
//	 1 – Side
//	 8 – Price
//	 8 – StopPrice
//	 8 – Quantity
//	 8 – ExecutedQuantity
//	 8 – LeavesQuantity
//	 1 – TimeInForce
//	 8 – MaxVisibleQuantity
//	 8 – Slippage
//	 8 – TrailingDistance
//	 8 – TrailingStep
//
// Total: 87 bytes
const orderWireSize = 87

// eventHeaderSize = 1 (EventType) + 8 (Timestamp) = 9 bytes.
// A full NewOrder record is eventHeaderSize + orderWireSize = 96 bytes.
// A CancelOrder record is eventHeaderSize + 8 (OrderID) = 17 bytes.

// marshalOrder writes o into buf (must be at least orderWireSize bytes).
func marshalOrder(buf []byte, o matching.Order) {
	binary.BigEndian.PutUint64(buf[0:8], o.ID)
	binary.BigEndian.PutUint32(buf[8:12], o.SymbolID)
	buf[12] = uint8(o.Type)
	buf[13] = uint8(o.Side)
	binary.BigEndian.PutUint64(buf[14:22], o.Price)
	binary.BigEndian.PutUint64(buf[22:30], o.StopPrice)
	binary.BigEndian.PutUint64(buf[30:38], o.Quantity)
	binary.BigEndian.PutUint64(buf[38:46], o.ExecutedQuantity)
	binary.BigEndian.PutUint64(buf[46:54], o.LeavesQuantity)
	buf[54] = uint8(o.TimeInForce)
	binary.BigEndian.PutUint64(buf[55:63], o.MaxVisibleQuantity)
	binary.BigEndian.PutUint64(buf[63:71], o.Slippage)
	binary.BigEndian.PutUint64(buf[71:79], uint64(o.TrailingDistance))
	binary.BigEndian.PutUint64(buf[79:87], uint64(o.TrailingStep))
}

// unmarshalOrder reads an order from buf (must be at least orderWireSize bytes).
func unmarshalOrder(buf []byte) matching.Order {
	return matching.Order{
		ID:                 binary.BigEndian.Uint64(buf[0:8]),
		SymbolID:           binary.BigEndian.Uint32(buf[8:12]),
		Type:               matching.OrderType(buf[12]),
		Side:               matching.OrderSide(buf[13]),
		Price:              binary.BigEndian.Uint64(buf[14:22]),
		StopPrice:          binary.BigEndian.Uint64(buf[22:30]),
		Quantity:           binary.BigEndian.Uint64(buf[30:38]),
		ExecutedQuantity:   binary.BigEndian.Uint64(buf[38:46]),
		LeavesQuantity:     binary.BigEndian.Uint64(buf[46:54]),
		TimeInForce:        matching.OrderTimeInForce(buf[54]),
		MaxVisibleQuantity: binary.BigEndian.Uint64(buf[55:63]),
		Slippage:           binary.BigEndian.Uint64(buf[63:71]),
		TrailingDistance:   int64(binary.BigEndian.Uint64(buf[71:79])),
		TrailingStep:       int64(binary.BigEndian.Uint64(buf[79:87])),
	}
}

// encodeEvent encodes a MatchingEvent into a length-prefixed binary record.
//
// Record wire format:
//
//	4 bytes – payload length (big-endian uint32)
//	1 byte  – EventType
//	8 bytes – Timestamp (int64 big-endian)
//	N bytes – event-specific payload
//	             EventNewOrder:    87 bytes (order)
//	             EventCancelOrder:  8 bytes (order ID)
func encodeEvent(e MatchingEvent) ([]byte, error) {
	var payloadSize int
	switch e.Type {
	case EventNewOrder:
		payloadSize = 1 + 8 + orderWireSize
	case EventCancelOrder:
		payloadSize = 1 + 8 + 8
	default:
		return nil, fmt.Errorf("persistence: unknown EventType %d", e.Type)
	}

	record := make([]byte, 4+payloadSize)
	binary.BigEndian.PutUint32(record[0:4], uint32(payloadSize))
	record[4] = uint8(e.Type)
	binary.BigEndian.PutUint64(record[5:13], uint64(e.Timestamp))

	switch e.Type {
	case EventNewOrder:
		marshalOrder(record[13:], e.Order)
	case EventCancelOrder:
		binary.BigEndian.PutUint64(record[13:21], e.OrderID)
	}
	return record, nil
}

// decodeEvent reads one length-prefixed record from r and returns the decoded event.
func decodeEvent(r io.Reader) (MatchingEvent, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return MatchingEvent{}, err
	}
	payloadLen := binary.BigEndian.Uint32(lenBuf[:])
	if payloadLen < 9 { // minimum: 1 (type) + 8 (timestamp)
		return MatchingEvent{}, fmt.Errorf("persistence: invalid record length %d", payloadLen)
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return MatchingEvent{}, fmt.Errorf("persistence: reading record payload: %w", err)
	}

	e := MatchingEvent{
		Type:      EventType(payload[0]),
		Timestamp: int64(binary.BigEndian.Uint64(payload[1:9])),
	}
	switch e.Type {
	case EventNewOrder:
		if len(payload) < 9+orderWireSize {
			return MatchingEvent{}, fmt.Errorf("persistence: short NewOrder payload (%d bytes)", len(payload))
		}
		e.Order = unmarshalOrder(payload[9:])
	case EventCancelOrder:
		if len(payload) < 17 {
			return MatchingEvent{}, fmt.Errorf("persistence: short CancelOrder payload (%d bytes)", len(payload))
		}
		e.OrderID = binary.BigEndian.Uint64(payload[9:17])
	default:
		return MatchingEvent{}, fmt.Errorf("persistence: unknown EventType %d", e.Type)
	}
	return e, nil
}
