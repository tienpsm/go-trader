package matching

import "fmt"

// LevelType represents the type of price level (bid or ask)
type LevelType uint8

const (
	// LevelTypeBid represents a bid (buy) price level
	LevelTypeBid LevelType = iota
	// LevelTypeAsk represents an ask (sell) price level
	LevelTypeAsk
)

// String returns the string representation of a LevelType
func (lt LevelType) String() string {
	switch lt {
	case LevelTypeBid:
		return "BID"
	case LevelTypeAsk:
		return "ASK"
	default:
		return "UNKNOWN"
	}
}

// Level represents a price level in the order book
type Level struct {
	// Type is the level type (bid or ask)
	Type LevelType
	// Price is the price of this level
	Price uint64
	// TotalVolume is the total volume at this price level
	TotalVolume uint64
	// HiddenVolume is the hidden volume at this price level
	HiddenVolume uint64
	// VisibleVolume is the visible volume at this price level
	VisibleVolume uint64
	// Orders is the number of orders at this price level
	Orders uint64
}

// NewLevel creates a new price level
func NewLevel(levelType LevelType, price uint64) Level {
	return Level{
		Type:          levelType,
		Price:         price,
		TotalVolume:   0,
		HiddenVolume:  0,
		VisibleVolume: 0,
		Orders:        0,
	}
}

// IsBid returns true if this is a bid level
func (l *Level) IsBid() bool {
	return l.Type == LevelTypeBid
}

// IsAsk returns true if this is an ask level
func (l *Level) IsAsk() bool {
	return l.Type == LevelTypeAsk
}

// String returns the string representation of a Level
func (l *Level) String() string {
	return fmt.Sprintf(
		"Level(Type=%s, Price=%d, Volume=%d, Hidden=%d, Visible=%d, Orders=%d)",
		l.Type, l.Price, l.TotalVolume, l.HiddenVolume, l.VisibleVolume, l.Orders,
	)
}

// LevelNode is a price level with additional data structures for the order book
type LevelNode struct {
	Level
	// OrderList is the list of orders at this price level
	OrderList OrderList
	// Parent is the parent node in the AVL tree
	Parent *LevelNode
	// Left is the left child in the AVL tree
	Left *LevelNode
	// Right is the right child in the AVL tree
	Right *LevelNode
	// Balance is the AVL tree balance factor
	Balance int
}

// NewLevelNode creates a new level node
func NewLevelNode(levelType LevelType, price uint64) *LevelNode {
	return &LevelNode{
		Level:   NewLevel(levelType, price),
		Parent:  nil,
		Left:    nil,
		Right:   nil,
		Balance: 0,
	}
}

// LevelUpdate represents an update to a price level
type LevelUpdate struct {
	// Type is the type of update (add, update, delete)
	Type UpdateType
	// Update is the level data
	Update Level
	// Top indicates if this is the top of the book
	Top bool
}

// NewLevelUpdate creates a new level update
func NewLevelUpdate(updateType UpdateType, level Level, top bool) LevelUpdate {
	return LevelUpdate{
		Type:   updateType,
		Update: level,
		Top:    top,
	}
}

// String returns the string representation of a LevelUpdate
func (lu *LevelUpdate) String() string {
	return fmt.Sprintf("LevelUpdate(Type=%s, Level=%s, Top=%v)", lu.Type, lu.Update.String(), lu.Top)
}

// OrderList is a doubly-linked list of orders
type OrderList struct {
	// Head is the first order in the list
	Head *OrderNode
	// Tail is the last order in the list
	Tail *OrderNode
	// Size is the number of orders in the list
	Size uint64
}

// PushBack adds an order to the end of the list
func (ol *OrderList) PushBack(order *OrderNode) {
	order.Next = nil
	order.Prev = ol.Tail
	if ol.Tail != nil {
		ol.Tail.Next = order
	} else {
		ol.Head = order
	}
	ol.Tail = order
	ol.Size++
}

// Remove removes an order from the list
func (ol *OrderList) Remove(order *OrderNode) {
	if order.Prev != nil {
		order.Prev.Next = order.Next
	} else {
		ol.Head = order.Next
	}
	if order.Next != nil {
		order.Next.Prev = order.Prev
	} else {
		ol.Tail = order.Prev
	}
	order.Next = nil
	order.Prev = nil
	ol.Size--
}

// Front returns the first order in the list
func (ol *OrderList) Front() *OrderNode {
	return ol.Head
}

// Empty returns true if the list is empty
func (ol *OrderList) Empty() bool {
	return ol.Size == 0
}
