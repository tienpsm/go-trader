package matching

import (
	"fmt"
	"math"
)

// OrderSide represents the side of an order (buy or sell)
type OrderSide uint8

const (
	// OrderSideBuy represents a buy order
	OrderSideBuy OrderSide = iota
	// OrderSideSell represents a sell order
	OrderSideSell
)

// String returns the string representation of an OrderSide
func (s OrderSide) String() string {
	switch s {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// OrderType represents the type of an order
type OrderType uint8

const (
	// OrderTypeMarket is a market order executed at the best available price
	OrderTypeMarket OrderType = iota
	// OrderTypeLimit is a limit order executed at a specific price or better
	OrderTypeLimit
	// OrderTypeStop is a stop order that becomes a market order when triggered
	OrderTypeStop
	// OrderTypeStopLimit is a stop order that becomes a limit order when triggered
	OrderTypeStopLimit
	// OrderTypeTrailingStop is a trailing stop order with a dynamic stop price
	OrderTypeTrailingStop
	// OrderTypeTrailingStopLimit is a trailing stop-limit order
	OrderTypeTrailingStopLimit
)

// String returns the string representation of an OrderType
func (t OrderType) String() string {
	switch t {
	case OrderTypeMarket:
		return "MARKET"
	case OrderTypeLimit:
		return "LIMIT"
	case OrderTypeStop:
		return "STOP"
	case OrderTypeStopLimit:
		return "STOP_LIMIT"
	case OrderTypeTrailingStop:
		return "TRAILING_STOP"
	case OrderTypeTrailingStopLimit:
		return "TRAILING_STOP_LIMIT"
	default:
		return "UNKNOWN"
	}
}

// OrderTimeInForce represents the time-in-force for an order
type OrderTimeInForce uint8

const (
	// OrderTimeInForceGTC is Good-Till-Cancelled
	OrderTimeInForceGTC OrderTimeInForce = iota
	// OrderTimeInForceIOC is Immediate-Or-Cancel
	OrderTimeInForceIOC
	// OrderTimeInForceFOK is Fill-Or-Kill
	OrderTimeInForceFOK
	// OrderTimeInForceAON is All-Or-None
	OrderTimeInForceAON
)

// String returns the string representation of an OrderTimeInForce
func (tif OrderTimeInForce) String() string {
	switch tif {
	case OrderTimeInForceGTC:
		return "GTC"
	case OrderTimeInForceIOC:
		return "IOC"
	case OrderTimeInForceFOK:
		return "FOK"
	case OrderTimeInForceAON:
		return "AON"
	default:
		return "UNKNOWN"
	}
}

// MaxVisibleQuantity is the default value for max visible quantity (no limit)
const MaxVisibleQuantity = math.MaxUint64

// MaxSlippage is the default value for slippage (no limit)
const MaxSlippage = math.MaxUint64

// Order represents a trading order
type Order struct {
	// ID is the unique identifier for the order
	ID uint64
	// SymbolID is the symbol this order is for
	SymbolID uint32
	// Type is the order type
	Type OrderType
	// Side is the order side (buy/sell)
	Side OrderSide
	// Price is the order price (for limit orders)
	Price uint64
	// StopPrice is the stop price (for stop orders)
	StopPrice uint64

	// Quantity is the total order quantity
	Quantity uint64
	// ExecutedQuantity is the quantity that has been executed
	ExecutedQuantity uint64
	// LeavesQuantity is the remaining quantity to be executed
	LeavesQuantity uint64

	// TimeInForce specifies how long the order remains active
	TimeInForce OrderTimeInForce

	// MaxVisibleQuantity allows for iceberg/hidden orders
	// >= LeavesQuantity: Regular order
	// == 0: Hidden order
	// < LeavesQuantity: Iceberg order
	MaxVisibleQuantity uint64

	// Slippage protects market orders from executing at unfavorable prices
	Slippage uint64

	// TrailingDistance is the distance from market for trailing stop orders
	// Positive value: absolute distance
	// Negative value: percentage distance (0.01% precision, -10000 = 100%)
	TrailingDistance int64

	// TrailingStep is the step value for trailing stop updates
	TrailingStep int64
}

// NewOrder creates a new order with default values
func NewOrder(id uint64, symbolID uint32, orderType OrderType, side OrderSide, price, stopPrice, quantity uint64) *Order {
	return &Order{
		ID:                 id,
		SymbolID:           symbolID,
		Type:               orderType,
		Side:               side,
		Price:              price,
		StopPrice:          stopPrice,
		Quantity:           quantity,
		ExecutedQuantity:   0,
		LeavesQuantity:     quantity,
		TimeInForce:        OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage:           MaxSlippage,
		TrailingDistance:   0,
		TrailingStep:       0,
	}
}

// NewLimitOrder creates a new limit order
func NewLimitOrder(id uint64, symbolID uint32, side OrderSide, price, quantity uint64) *Order {
	return NewOrder(id, symbolID, OrderTypeLimit, side, price, 0, quantity)
}

// NewMarketOrder creates a new market order
func NewMarketOrder(id uint64, symbolID uint32, side OrderSide, quantity uint64) *Order {
	return NewOrder(id, symbolID, OrderTypeMarket, side, 0, 0, quantity)
}

// NewStopOrder creates a new stop order
func NewStopOrder(id uint64, symbolID uint32, side OrderSide, stopPrice, quantity uint64) *Order {
	return NewOrder(id, symbolID, OrderTypeStop, side, 0, stopPrice, quantity)
}

// NewStopLimitOrder creates a new stop-limit order
func NewStopLimitOrder(id uint64, symbolID uint32, side OrderSide, price, stopPrice, quantity uint64) *Order {
	return NewOrder(id, symbolID, OrderTypeStopLimit, side, price, stopPrice, quantity)
}

// IsMarket returns true if this is a market order
func (o *Order) IsMarket() bool {
	return o.Type == OrderTypeMarket
}

// IsLimit returns true if this is a limit order
func (o *Order) IsLimit() bool {
	return o.Type == OrderTypeLimit
}

// IsStop returns true if this is a stop order
func (o *Order) IsStop() bool {
	return o.Type == OrderTypeStop
}

// IsStopLimit returns true if this is a stop-limit order
func (o *Order) IsStopLimit() bool {
	return o.Type == OrderTypeStopLimit
}

// IsTrailingStop returns true if this is a trailing stop order
func (o *Order) IsTrailingStop() bool {
	return o.Type == OrderTypeTrailingStop
}

// IsTrailingStopLimit returns true if this is a trailing stop-limit order
func (o *Order) IsTrailingStopLimit() bool {
	return o.Type == OrderTypeTrailingStopLimit
}

// IsBuy returns true if this is a buy order
func (o *Order) IsBuy() bool {
	return o.Side == OrderSideBuy
}

// IsSell returns true if this is a sell order
func (o *Order) IsSell() bool {
	return o.Side == OrderSideSell
}

// IsGTC returns true if this is a Good-Till-Cancelled order
func (o *Order) IsGTC() bool {
	return o.TimeInForce == OrderTimeInForceGTC
}

// IsIOC returns true if this is an Immediate-Or-Cancel order
func (o *Order) IsIOC() bool {
	return o.TimeInForce == OrderTimeInForceIOC
}

// IsFOK returns true if this is a Fill-Or-Kill order
func (o *Order) IsFOK() bool {
	return o.TimeInForce == OrderTimeInForceFOK
}

// IsAON returns true if this is an All-Or-None order
func (o *Order) IsAON() bool {
	return o.TimeInForce == OrderTimeInForceAON
}

// HiddenQuantity returns the hidden quantity for iceberg orders
func (o *Order) HiddenQuantity() uint64 {
	if o.LeavesQuantity > o.MaxVisibleQuantity {
		return o.LeavesQuantity - o.MaxVisibleQuantity
	}
	return 0
}

// VisibleQuantity returns the visible quantity for iceberg orders
func (o *Order) VisibleQuantity() uint64 {
	if o.LeavesQuantity < o.MaxVisibleQuantity {
		return o.LeavesQuantity
	}
	return o.MaxVisibleQuantity
}

// IsHidden returns true if this is a hidden order
func (o *Order) IsHidden() bool {
	return o.MaxVisibleQuantity == 0
}

// IsIceberg returns true if this is an iceberg order
func (o *Order) IsIceberg() bool {
	return o.MaxVisibleQuantity > 0 && o.MaxVisibleQuantity < o.LeavesQuantity
}

// String returns the string representation of an Order
func (o *Order) String() string {
	return fmt.Sprintf(
		"Order(ID=%d, Symbol=%d, Type=%s, Side=%s, Price=%d, StopPrice=%d, "+
			"Quantity=%d, Executed=%d, Leaves=%d, TIF=%s)",
		o.ID, o.SymbolID, o.Type, o.Side, o.Price, o.StopPrice,
		o.Quantity, o.ExecutedQuantity, o.LeavesQuantity, o.TimeInForce,
	)
}

// OrderNode is an Order with linked list pointers for use in price levels
type OrderNode struct {
	Order
	// Next points to the next order in the level
	Next *OrderNode
	// Prev points to the previous order in the level
	Prev *OrderNode
	// Level points to the price level containing this order
	Level *LevelNode
}

// NewOrderNode creates a new OrderNode from an Order
func NewOrderNode(order Order) *OrderNode {
	return &OrderNode{
		Order: order,
		Next:  nil,
		Prev:  nil,
		Level: nil,
	}
}
