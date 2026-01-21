package matching

import "fmt"

// OrderBook represents an order book for a single symbol
type OrderBook struct {
	// manager is the market manager that owns this order book
	manager *MarketManager
	// symbol is the trading symbol for this order book
	symbol Symbol

	// bestBid is the best (highest) bid price level
	bestBid *LevelNode
	// bestAsk is the best (lowest) ask price level
	bestAsk *LevelNode
	// bids is the AVL tree of bid price levels
	bids *AVLTree
	// asks is the AVL tree of ask price levels
	asks *AVLTree

	// Stop order levels
	bestBuyStop      *LevelNode
	bestSellStop     *LevelNode
	buyStopLevels    *AVLTree
	sellStopLevels   *AVLTree

	// Trailing stop order levels
	bestTrailingBuyStop   *LevelNode
	bestTrailingSellStop  *LevelNode
	trailingBuyStopLevels *AVLTree
	trailingSellStopLevels *AVLTree

	// Last executed prices
	lastBidPrice   uint64
	lastAskPrice   uint64
	matchingPrice  uint64
}

// NewOrderBook creates a new order book for a symbol
func NewOrderBook(manager *MarketManager, symbol Symbol) *OrderBook {
	return &OrderBook{
		manager:               manager,
		symbol:                symbol,
		bestBid:               nil,
		bestAsk:               nil,
		bids:                  NewAVLTree(true),  // Descending for bids (highest first)
		asks:                  NewAVLTree(false), // Ascending for asks (lowest first)
		bestBuyStop:           nil,
		bestSellStop:          nil,
		buyStopLevels:         NewAVLTree(false), // Ascending
		sellStopLevels:        NewAVLTree(true),  // Descending
		bestTrailingBuyStop:   nil,
		bestTrailingSellStop:  nil,
		trailingBuyStopLevels: NewAVLTree(false),
		trailingSellStopLevels: NewAVLTree(true),
		lastBidPrice:          0,
		lastAskPrice:          0,
		matchingPrice:         0,
	}
}

// Symbol returns the symbol for this order book
func (ob *OrderBook) Symbol() Symbol {
	return ob.symbol
}

// Empty returns true if the order book is empty
func (ob *OrderBook) Empty() bool {
	return ob.Size() == 0
}

// Size returns the total number of price levels in the order book
func (ob *OrderBook) Size() int {
	return ob.bids.Size() + ob.asks.Size() +
		ob.buyStopLevels.Size() + ob.sellStopLevels.Size() +
		ob.trailingBuyStopLevels.Size() + ob.trailingSellStopLevels.Size()
}

// BestBid returns the best bid price level
func (ob *OrderBook) BestBid() *LevelNode {
	return ob.bestBid
}

// BestAsk returns the best ask price level
func (ob *OrderBook) BestAsk() *LevelNode {
	return ob.bestAsk
}

// Bids returns the bid levels tree
func (ob *OrderBook) Bids() *AVLTree {
	return ob.bids
}

// Asks returns the ask levels tree
func (ob *OrderBook) Asks() *AVLTree {
	return ob.asks
}

// GetBid returns the bid level at the given price
func (ob *OrderBook) GetBid(price uint64) *LevelNode {
	return ob.bids.Find(price)
}

// GetAsk returns the ask level at the given price
func (ob *OrderBook) GetAsk(price uint64) *LevelNode {
	return ob.asks.Find(price)
}

// BestBuyStop returns the best buy stop level
func (ob *OrderBook) BestBuyStop() *LevelNode {
	return ob.bestBuyStop
}

// BestSellStop returns the best sell stop level
func (ob *OrderBook) BestSellStop() *LevelNode {
	return ob.bestSellStop
}

// GetBuyStopLevel returns the buy stop level at the given price
func (ob *OrderBook) GetBuyStopLevel(price uint64) *LevelNode {
	return ob.buyStopLevels.Find(price)
}

// GetSellStopLevel returns the sell stop level at the given price
func (ob *OrderBook) GetSellStopLevel(price uint64) *LevelNode {
	return ob.sellStopLevels.Find(price)
}

// BestTrailingBuyStop returns the best trailing buy stop level
func (ob *OrderBook) BestTrailingBuyStop() *LevelNode {
	return ob.bestTrailingBuyStop
}

// BestTrailingSellStop returns the best trailing sell stop level
func (ob *OrderBook) BestTrailingSellStop() *LevelNode {
	return ob.bestTrailingSellStop
}

// GetTrailingBuyStopLevel returns the trailing buy stop level at the given price
func (ob *OrderBook) GetTrailingBuyStopLevel(price uint64) *LevelNode {
	return ob.trailingBuyStopLevels.Find(price)
}

// GetTrailingSellStopLevel returns the trailing sell stop level at the given price
func (ob *OrderBook) GetTrailingSellStopLevel(price uint64) *LevelNode {
	return ob.trailingSellStopLevels.Find(price)
}

// LastBidPrice returns the last executed bid price
func (ob *OrderBook) LastBidPrice() uint64 {
	return ob.lastBidPrice
}

// LastAskPrice returns the last executed ask price
func (ob *OrderBook) LastAskPrice() uint64 {
	return ob.lastAskPrice
}

// MatchingPrice returns the current matching price
func (ob *OrderBook) MatchingPrice() uint64 {
	return ob.matchingPrice
}

// AddLevel adds a new price level to the order book
func (ob *OrderBook) AddLevel(order *OrderNode) *LevelNode {
	var level *LevelNode

	if order.IsTrailingStop() || order.IsTrailingStopLimit() {
		// Trailing stop orders
		if order.IsBuy() {
			level = NewLevelNode(LevelTypeBid, order.StopPrice)
			ob.trailingBuyStopLevels.Insert(level)
			if ob.bestTrailingBuyStop == nil || order.StopPrice < ob.bestTrailingBuyStop.Price {
				ob.bestTrailingBuyStop = level
			}
		} else {
			level = NewLevelNode(LevelTypeAsk, order.StopPrice)
			ob.trailingSellStopLevels.Insert(level)
			if ob.bestTrailingSellStop == nil || order.StopPrice > ob.bestTrailingSellStop.Price {
				ob.bestTrailingSellStop = level
			}
		}
	} else if order.IsStop() || order.IsStopLimit() {
		// Stop orders
		if order.IsBuy() {
			level = NewLevelNode(LevelTypeBid, order.StopPrice)
			ob.buyStopLevels.Insert(level)
			if ob.bestBuyStop == nil || order.StopPrice < ob.bestBuyStop.Price {
				ob.bestBuyStop = level
			}
		} else {
			level = NewLevelNode(LevelTypeAsk, order.StopPrice)
			ob.sellStopLevels.Insert(level)
			if ob.bestSellStop == nil || order.StopPrice > ob.bestSellStop.Price {
				ob.bestSellStop = level
			}
		}
	} else {
		// Limit orders (bids and asks)
		if order.IsBuy() {
			level = NewLevelNode(LevelTypeBid, order.Price)
			ob.bids.Insert(level)
			if ob.bestBid == nil || order.Price > ob.bestBid.Price {
				ob.bestBid = level
			}
		} else {
			level = NewLevelNode(LevelTypeAsk, order.Price)
			ob.asks.Insert(level)
			if ob.bestAsk == nil || order.Price < ob.bestAsk.Price {
				ob.bestAsk = level
			}
		}
	}

	return level
}

// DeleteLevel removes a price level from the order book
func (ob *OrderBook) DeleteLevel(order *OrderNode) {
	level := order.Level

	if order.IsTrailingStop() || order.IsTrailingStopLimit() {
		// Trailing stop orders
		if order.IsBuy() {
			ob.trailingBuyStopLevels.Remove(level)
			if ob.bestTrailingBuyStop == level {
				ob.bestTrailingBuyStop = ob.trailingBuyStopLevels.First()
			}
		} else {
			ob.trailingSellStopLevels.Remove(level)
			if ob.bestTrailingSellStop == level {
				ob.bestTrailingSellStop = ob.trailingSellStopLevels.First()
			}
		}
	} else if order.IsStop() || order.IsStopLimit() {
		// Stop orders
		if order.IsBuy() {
			ob.buyStopLevels.Remove(level)
			if ob.bestBuyStop == level {
				ob.bestBuyStop = ob.buyStopLevels.First()
			}
		} else {
			ob.sellStopLevels.Remove(level)
			if ob.bestSellStop == level {
				ob.bestSellStop = ob.sellStopLevels.First()
			}
		}
	} else {
		// Limit orders
		if order.IsBuy() {
			ob.bids.Remove(level)
			if ob.bestBid == level {
				ob.bestBid = ob.bids.First()
			}
		} else {
			ob.asks.Remove(level)
			if ob.bestAsk == level {
				ob.bestAsk = ob.asks.First()
			}
		}
	}
}

// AddOrder adds an order to the order book
func (ob *OrderBook) AddOrder(order *OrderNode) {
	// Find or create the price level
	var level *LevelNode
	var price uint64

	if order.IsTrailingStop() || order.IsTrailingStopLimit() {
		price = order.StopPrice
		if order.IsBuy() {
			level = ob.trailingBuyStopLevels.Find(price)
		} else {
			level = ob.trailingSellStopLevels.Find(price)
		}
	} else if order.IsStop() || order.IsStopLimit() {
		price = order.StopPrice
		if order.IsBuy() {
			level = ob.buyStopLevels.Find(price)
		} else {
			level = ob.sellStopLevels.Find(price)
		}
	} else {
		price = order.Price
		if order.IsBuy() {
			level = ob.bids.Find(price)
		} else {
			level = ob.asks.Find(price)
		}
	}

	// Create new level if it doesn't exist
	if level == nil {
		level = ob.AddLevel(order)
	}

	// Add order to the level
	level.OrderList.PushBack(order)
	order.Level = level

	// Update level statistics
	level.TotalVolume += order.LeavesQuantity
	level.HiddenVolume += order.HiddenQuantity()
	level.VisibleVolume += order.VisibleQuantity()
	level.Orders++
}

// ReduceOrder reduces the quantity of an order
func (ob *OrderBook) ReduceOrder(order *OrderNode, quantity uint64, hidden, visible uint64) {
	level := order.Level
	level.TotalVolume -= quantity
	level.HiddenVolume -= hidden
	level.VisibleVolume -= visible
}

// DeleteOrder removes an order from the order book
func (ob *OrderBook) DeleteOrder(order *OrderNode) {
	level := order.Level

	// Remove order from level
	level.OrderList.Remove(order)
	level.TotalVolume -= order.LeavesQuantity
	level.HiddenVolume -= order.HiddenQuantity()
	level.VisibleVolume -= order.VisibleQuantity()
	level.Orders--

	// Remove level if empty
	if level.OrderList.Empty() {
		ob.DeleteLevel(order)
	}

	order.Level = nil
}

// String returns a string representation of the order book
func (ob *OrderBook) String() string {
	return fmt.Sprintf("OrderBook(Symbol=%s, Bids=%d, Asks=%d)",
		ob.symbol.Name, ob.bids.Size(), ob.asks.Size())
}

// GetSpread returns the bid-ask spread (ask - bid), or 0 if there's no spread
func (ob *OrderBook) GetSpread() uint64 {
	if ob.bestBid == nil || ob.bestAsk == nil {
		return 0
	}
	if ob.bestAsk.Price > ob.bestBid.Price {
		return ob.bestAsk.Price - ob.bestBid.Price
	}
	return 0
}

// GetMidPrice returns the mid price ((best bid + best ask) / 2)
func (ob *OrderBook) GetMidPrice() uint64 {
	if ob.bestBid == nil || ob.bestAsk == nil {
		return 0
	}
	return (ob.bestBid.Price + ob.bestAsk.Price) / 2
}
