package matching

// MarketManager is used to manage the market with symbols, orders and order books.
// Automatic order matching can be enabled with EnableMatching() or manually performed with Match().
// Not thread-safe.
type MarketManager struct {
	// handler is the market event handler
	handler MarketHandler

	// symbols is the list of all symbols
	symbols map[uint32]*Symbol
	// orderBooks is the list of all order books
	orderBooks map[uint32]*OrderBook
	// orders is the map of all orders by ID
	orders map[uint64]*OrderNode

	// matching indicates if automatic matching is enabled
	matching bool
}

// NewMarketManager creates a new market manager
func NewMarketManager() *MarketManager {
	return &MarketManager{
		handler:    &DefaultMarketHandler{},
		symbols:    make(map[uint32]*Symbol),
		orderBooks: make(map[uint32]*OrderBook),
		orders:     make(map[uint64]*OrderNode),
		matching:   false,
	}
}

// NewMarketManagerWithHandler creates a new market manager with a custom handler
func NewMarketManagerWithHandler(handler MarketHandler) *MarketManager {
	return &MarketManager{
		handler:    handler,
		symbols:    make(map[uint32]*Symbol),
		orderBooks: make(map[uint32]*OrderBook),
		orders:     make(map[uint64]*OrderNode),
		matching:   false,
	}
}

// Symbols returns all symbols
func (m *MarketManager) Symbols() map[uint32]*Symbol {
	return m.symbols
}

// OrderBooks returns all order books
func (m *MarketManager) OrderBooks() map[uint32]*OrderBook {
	return m.orderBooks
}

// Orders returns all orders
func (m *MarketManager) Orders() map[uint64]*OrderNode {
	return m.orders
}

// GetSymbol returns a symbol by ID
func (m *MarketManager) GetSymbol(id uint32) *Symbol {
	return m.symbols[id]
}

// GetOrderBook returns an order book by symbol ID
func (m *MarketManager) GetOrderBook(id uint32) *OrderBook {
	return m.orderBooks[id]
}

// GetOrder returns an order by ID
func (m *MarketManager) GetOrder(id uint64) *OrderNode {
	return m.orders[id]
}

// IsMatchingEnabled returns true if automatic matching is enabled
func (m *MarketManager) IsMatchingEnabled() bool {
	return m.matching
}

// EnableMatching enables automatic order matching
func (m *MarketManager) EnableMatching() {
	m.matching = true
}

// DisableMatching disables automatic order matching
func (m *MarketManager) DisableMatching() {
	m.matching = false
}

// AddSymbol adds a new symbol
func (m *MarketManager) AddSymbol(symbol Symbol) ErrorCode {
	if _, exists := m.symbols[symbol.ID]; exists {
		return ErrorSymbolDuplicate
	}

	m.symbols[symbol.ID] = &symbol
	m.handler.OnAddSymbol(symbol)
	return ErrorOK
}

// DeleteSymbol deletes a symbol
func (m *MarketManager) DeleteSymbol(id uint32) ErrorCode {
	symbol, exists := m.symbols[id]
	if !exists {
		return ErrorSymbolNotFound
	}

	// Delete associated order book first
	if ob := m.orderBooks[id]; ob != nil {
		m.DeleteOrderBook(id)
	}

	delete(m.symbols, id)
	m.handler.OnDeleteSymbol(*symbol)
	return ErrorOK
}

// AddOrderBook adds a new order book for a symbol
func (m *MarketManager) AddOrderBook(symbol Symbol) ErrorCode {
	if _, exists := m.orderBooks[symbol.ID]; exists {
		return ErrorOrderBookDuplicate
	}

	// Create the order book
	ob := NewOrderBook(m, symbol)
	m.orderBooks[symbol.ID] = ob
	m.handler.OnAddOrderBook(ob)
	return ErrorOK
}

// DeleteOrderBook deletes an order book
func (m *MarketManager) DeleteOrderBook(id uint32) ErrorCode {
	ob, exists := m.orderBooks[id]
	if !exists {
		return ErrorOrderBookNotFound
	}

	// Cancel all orders in the order book
	ordersToDelete := make([]*OrderNode, 0)
	for _, order := range m.orders {
		if order.SymbolID == id {
			ordersToDelete = append(ordersToDelete, order)
		}
	}
	for _, order := range ordersToDelete {
		m.DeleteOrder(order.ID)
	}

	delete(m.orderBooks, id)
	m.handler.OnDeleteOrderBook(ob)
	return ErrorOK
}

// AddOrder adds a new order
func (m *MarketManager) AddOrder(order Order) ErrorCode {
	// Validate order
	if err := m.validateOrder(order); err != ErrorOK {
		return err
	}

	// Check for duplicate order
	if _, exists := m.orders[order.ID]; exists {
		return ErrorOrderDuplicate
	}

	// Get the order book
	ob, exists := m.orderBooks[order.SymbolID]
	if !exists {
		return ErrorOrderBookNotFound
	}

	// Create order node
	orderNode := NewOrderNode(order)
	m.orders[order.ID] = orderNode

	// Add order to the order book
	ob.AddOrder(orderNode)
	m.handler.OnAddOrder(order)

	// Update order book
	m.updateLevel(ob, orderNode, UpdateAdd)

	// Match if enabled
	if m.matching {
		m.match(ob)
	}

	return ErrorOK
}

// ReduceOrder reduces the quantity of an order
func (m *MarketManager) ReduceOrder(id uint64, quantity uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	if quantity == 0 {
		return ErrorOrderQuantityInvalid
	}

	// If reducing by more than leaves quantity, just cancel
	if quantity >= orderNode.LeavesQuantity {
		return m.DeleteOrder(id)
	}

	ob := m.orderBooks[orderNode.SymbolID]

	// Calculate hidden and visible reduction
	oldHidden := orderNode.HiddenQuantity()
	oldVisible := orderNode.VisibleQuantity()

	orderNode.LeavesQuantity -= quantity

	newHidden := orderNode.HiddenQuantity()
	newVisible := orderNode.VisibleQuantity()

	hiddenReduction := oldHidden - newHidden
	visibleReduction := oldVisible - newVisible

	// Update level
	ob.ReduceOrder(orderNode, quantity, hiddenReduction, visibleReduction)

	m.handler.OnUpdateOrder(orderNode.Order)
	m.updateLevel(ob, orderNode, UpdateUpdate)

	return ErrorOK
}

// ModifyOrder modifies an existing order
func (m *MarketManager) ModifyOrder(id uint64, newPrice, newQuantity uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	if newQuantity == 0 {
		return ErrorOrderQuantityInvalid
	}

	ob := m.orderBooks[orderNode.SymbolID]

	// Remove from old level
	m.updateLevel(ob, orderNode, UpdateDelete)
	ob.DeleteOrder(orderNode)

	// Update order
	orderNode.Price = newPrice
	orderNode.Quantity = newQuantity
	orderNode.LeavesQuantity = newQuantity
	orderNode.ExecutedQuantity = 0

	// Add to new level
	ob.AddOrder(orderNode)
	m.handler.OnUpdateOrder(orderNode.Order)
	m.updateLevel(ob, orderNode, UpdateAdd)

	// Match if enabled
	if m.matching {
		m.match(ob)
	}

	return ErrorOK
}

// MitigateOrder mitigates an order (in-flight mitigation)
func (m *MarketManager) MitigateOrder(id uint64, newPrice, newQuantity uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	ob := m.orderBooks[orderNode.SymbolID]

	if newQuantity <= orderNode.ExecutedQuantity {
		// Cancel the order
		return m.DeleteOrder(id)
	}

	// Remove from old level
	m.updateLevel(ob, orderNode, UpdateDelete)
	ob.DeleteOrder(orderNode)

	// Update order
	orderNode.Price = newPrice
	orderNode.Quantity = newQuantity
	orderNode.LeavesQuantity = newQuantity - orderNode.ExecutedQuantity

	// Add to new level
	ob.AddOrder(orderNode)
	m.handler.OnUpdateOrder(orderNode.Order)
	m.updateLevel(ob, orderNode, UpdateAdd)

	// Match if enabled
	if m.matching {
		m.match(ob)
	}

	return ErrorOK
}

// ReplaceOrder replaces an existing order with a new one
func (m *MarketManager) ReplaceOrder(id uint64, newID uint64, newPrice, newQuantity uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	if newQuantity == 0 {
		return ErrorOrderQuantityInvalid
	}

	if _, exists := m.orders[newID]; exists {
		return ErrorOrderDuplicate
	}

	ob := m.orderBooks[orderNode.SymbolID]

	// Remove old order
	m.updateLevel(ob, orderNode, UpdateDelete)
	ob.DeleteOrder(orderNode)
	delete(m.orders, id)
	m.handler.OnDeleteOrder(orderNode.Order)

	// Create new order
	newOrder := Order{
		ID:                 newID,
		SymbolID:           orderNode.SymbolID,
		Type:               orderNode.Type,
		Side:               orderNode.Side,
		Price:              newPrice,
		StopPrice:          orderNode.StopPrice,
		Quantity:           newQuantity,
		ExecutedQuantity:   0,
		LeavesQuantity:     newQuantity,
		TimeInForce:        orderNode.TimeInForce,
		MaxVisibleQuantity: orderNode.MaxVisibleQuantity,
		Slippage:           orderNode.Slippage,
		TrailingDistance:   orderNode.TrailingDistance,
		TrailingStep:       orderNode.TrailingStep,
	}

	newOrderNode := NewOrderNode(newOrder)
	m.orders[newID] = newOrderNode

	// Add new order
	ob.AddOrder(newOrderNode)
	m.handler.OnAddOrder(newOrder)
	m.updateLevel(ob, newOrderNode, UpdateAdd)

	// Match if enabled
	if m.matching {
		m.match(ob)
	}

	return ErrorOK
}

// DeleteOrder deletes an order
func (m *MarketManager) DeleteOrder(id uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	ob := m.orderBooks[orderNode.SymbolID]

	// Remove from order book
	m.updateLevel(ob, orderNode, UpdateDelete)
	ob.DeleteOrder(orderNode)
	delete(m.orders, id)
	m.handler.OnDeleteOrder(orderNode.Order)

	return ErrorOK
}

// ExecuteOrder executes a trade between two orders
func (m *MarketManager) ExecuteOrder(id uint64, quantity uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	if quantity == 0 || quantity > orderNode.LeavesQuantity {
		return ErrorOrderQuantityInvalid
	}

	return m.executeOrder(orderNode, orderNode.Price, quantity)
}

// ExecuteOrderWithPrice executes a trade at a specific price
func (m *MarketManager) ExecuteOrderWithPrice(id uint64, price, quantity uint64) ErrorCode {
	orderNode, exists := m.orders[id]
	if !exists {
		return ErrorOrderNotFound
	}

	if quantity == 0 || quantity > orderNode.LeavesQuantity {
		return ErrorOrderQuantityInvalid
	}

	return m.executeOrder(orderNode, price, quantity)
}

// executeOrder executes an order
func (m *MarketManager) executeOrder(orderNode *OrderNode, price, quantity uint64) ErrorCode {
	ob := m.orderBooks[orderNode.SymbolID]

	// Calculate hidden and visible reduction
	oldHidden := orderNode.HiddenQuantity()
	oldVisible := orderNode.VisibleQuantity()

	// Update order
	orderNode.ExecutedQuantity += quantity
	orderNode.LeavesQuantity -= quantity

	newHidden := orderNode.HiddenQuantity()
	newVisible := orderNode.VisibleQuantity()

	hiddenReduction := oldHidden - newHidden
	visibleReduction := oldVisible - newVisible

	// Update level
	ob.ReduceOrder(orderNode, quantity, hiddenReduction, visibleReduction)

	// Notify execution
	m.handler.OnExecuteOrder(orderNode.Order, price, quantity)

	// Check if order is complete
	if orderNode.LeavesQuantity == 0 {
		m.updateLevel(ob, orderNode, UpdateDelete)
		ob.DeleteOrder(orderNode)
		delete(m.orders, orderNode.ID)
		m.handler.OnDeleteOrder(orderNode.Order)
	} else {
		m.handler.OnUpdateOrder(orderNode.Order)
		m.updateLevel(ob, orderNode, UpdateUpdate)
	}

	return ErrorOK
}

// Match performs order matching for an order book
func (m *MarketManager) Match(symbolID uint32) ErrorCode {
	ob, exists := m.orderBooks[symbolID]
	if !exists {
		return ErrorOrderBookNotFound
	}

	m.match(ob)
	return ErrorOK
}

// match performs matching for an order book
func (m *MarketManager) match(ob *OrderBook) {
	// Match limit orders
	for {
		if ob.bestBid == nil || ob.bestAsk == nil {
			break
		}
		if ob.bestBid.Price < ob.bestAsk.Price {
			break
		}

		// Get the orders at the best levels
		bidOrder := ob.bestBid.OrderList.Front()
		askOrder := ob.bestAsk.OrderList.Front()

		if bidOrder == nil || askOrder == nil {
			break
		}

		// Determine execution quantity
		quantity := bidOrder.LeavesQuantity
		if askOrder.LeavesQuantity < quantity {
			quantity = askOrder.LeavesQuantity
		}

		// Determine execution price (price-time priority: earlier order's price)
		price := askOrder.Price

		// Execute both sides
		m.executeOrder(bidOrder, price, quantity)
		m.executeOrder(askOrder, price, quantity)
	}

	// TODO: Stop order activation
	// When market price moves through stop prices, stop orders should be activated:
	// - Buy stop orders activate when ask price >= stop price
	// - Sell stop orders activate when bid price <= stop price
	// This is left as a future enhancement as it requires additional price tracking.

	// TODO: Trailing stop order activation
	// Trailing stops need to track the market and update stop prices accordingly.
	// This is left as a future enhancement as it requires price monitoring.
}

// validateOrder validates an order
func (m *MarketManager) validateOrder(order Order) ErrorCode {
	if order.ID == 0 {
		return ErrorOrderIDInvalid
	}

	if order.Quantity == 0 {
		return ErrorOrderQuantityInvalid
	}

	// Validate order type specific requirements
	switch order.Type {
	case OrderTypeLimit:
		if order.Price == 0 {
			return ErrorOrderParameterInvalid
		}
	case OrderTypeStop:
		if order.StopPrice == 0 {
			return ErrorOrderParameterInvalid
		}
	case OrderTypeStopLimit:
		if order.Price == 0 || order.StopPrice == 0 {
			return ErrorOrderParameterInvalid
		}
	case OrderTypeTrailingStop:
		if order.TrailingDistance == 0 {
			return ErrorOrderParameterInvalid
		}
	case OrderTypeTrailingStopLimit:
		if order.Price == 0 || order.TrailingDistance == 0 {
			return ErrorOrderParameterInvalid
		}
	}

	return ErrorOK
}

// updateLevel notifies the handler about level updates
func (m *MarketManager) updateLevel(ob *OrderBook, order *OrderNode, updateType UpdateType) {
	if order.Level == nil {
		return
	}

	level := order.Level.Level
	top := false

	// Determine if this is a top-of-book update
	if order.IsBuy() {
		if ob.bestBid == order.Level {
			top = true
		}
	} else {
		if ob.bestAsk == order.Level {
			top = true
		}
	}

	switch updateType {
	case UpdateAdd:
		m.handler.OnAddLevel(ob, level, top)
	case UpdateUpdate:
		m.handler.OnUpdateLevel(ob, level, top)
	case UpdateDelete:
		m.handler.OnDeleteLevel(ob, level, top)
	}

	m.handler.OnUpdateOrderBook(ob, top)
}
