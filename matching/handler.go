package matching

// MarketHandler is an interface for handling market events
// Custom implementations can be used to monitor changes in the market:
// - Add/Remove/Modify symbols
// - Add/Remove/Modify orders
// - Order executions
// - Order book updates
type MarketHandler interface {
	// Symbol handlers
	OnAddSymbol(symbol Symbol)
	OnDeleteSymbol(symbol Symbol)

	// Order book handlers
	OnAddOrderBook(orderBook *OrderBook)
	OnUpdateOrderBook(orderBook *OrderBook, top bool)
	OnDeleteOrderBook(orderBook *OrderBook)

	// Price level handlers
	OnAddLevel(orderBook *OrderBook, level Level, top bool)
	OnUpdateLevel(orderBook *OrderBook, level Level, top bool)
	OnDeleteLevel(orderBook *OrderBook, level Level, top bool)

	// Order handlers
	OnAddOrder(order Order)
	OnUpdateOrder(order Order)
	OnDeleteOrder(order Order)

	// Order execution handlers
	OnExecuteOrder(order Order, price, quantity uint64)
}

// DefaultMarketHandler is a no-op implementation of MarketHandler
type DefaultMarketHandler struct{}

// OnAddSymbol is called when a symbol is added
func (h *DefaultMarketHandler) OnAddSymbol(symbol Symbol) {}

// OnDeleteSymbol is called when a symbol is deleted
func (h *DefaultMarketHandler) OnDeleteSymbol(symbol Symbol) {}

// OnAddOrderBook is called when an order book is added
func (h *DefaultMarketHandler) OnAddOrderBook(orderBook *OrderBook) {}

// OnUpdateOrderBook is called when an order book is updated
func (h *DefaultMarketHandler) OnUpdateOrderBook(orderBook *OrderBook, top bool) {}

// OnDeleteOrderBook is called when an order book is deleted
func (h *DefaultMarketHandler) OnDeleteOrderBook(orderBook *OrderBook) {}

// OnAddLevel is called when a price level is added
func (h *DefaultMarketHandler) OnAddLevel(orderBook *OrderBook, level Level, top bool) {}

// OnUpdateLevel is called when a price level is updated
func (h *DefaultMarketHandler) OnUpdateLevel(orderBook *OrderBook, level Level, top bool) {}

// OnDeleteLevel is called when a price level is deleted
func (h *DefaultMarketHandler) OnDeleteLevel(orderBook *OrderBook, level Level, top bool) {}

// OnAddOrder is called when an order is added
func (h *DefaultMarketHandler) OnAddOrder(order Order) {}

// OnUpdateOrder is called when an order is updated
func (h *DefaultMarketHandler) OnUpdateOrder(order Order) {}

// OnDeleteOrder is called when an order is deleted
func (h *DefaultMarketHandler) OnDeleteOrder(order Order) {}

// OnExecuteOrder is called when an order is executed
func (h *DefaultMarketHandler) OnExecuteOrder(order Order, price, quantity uint64) {}
