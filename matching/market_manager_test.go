package matching

import (
	"testing"
)

func TestMarketManager_AddSymbol(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	err := manager.AddSymbol(symbol)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	// Check symbol exists
	s := manager.GetSymbol(1)
	if s == nil {
		t.Error("Expected symbol to exist")
	}
	if s.Name != "AAPL" {
		t.Errorf("Expected AAPL, got %s", s.Name)
	}
	
	// Duplicate symbol
	err = manager.AddSymbol(symbol)
	if err != ErrorSymbolDuplicate {
		t.Errorf("Expected ErrorSymbolDuplicate, got %s", err)
	}
}

func TestMarketManager_DeleteSymbol(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	
	err := manager.DeleteSymbol(1)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	// Check symbol is gone
	s := manager.GetSymbol(1)
	if s != nil {
		t.Error("Expected symbol to be deleted")
	}
	
	// Delete non-existent
	err = manager.DeleteSymbol(1)
	if err != ErrorSymbolNotFound {
		t.Errorf("Expected ErrorSymbolNotFound, got %s", err)
	}
}

func TestMarketManager_AddOrderBook(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	
	err := manager.AddOrderBook(symbol)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	// Check order book exists
	ob := manager.GetOrderBook(1)
	if ob == nil {
		t.Error("Expected order book to exist")
	}
	
	// Duplicate order book
	err = manager.AddOrderBook(symbol)
	if err != ErrorOrderBookDuplicate {
		t.Errorf("Expected ErrorOrderBookDuplicate, got %s", err)
	}
}

func TestMarketManager_AddOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	err := manager.AddOrder(order)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	// Check order exists
	o := manager.GetOrder(1)
	if o == nil {
		t.Error("Expected order to exist")
	}
	if o.Price != 10000 {
		t.Errorf("Expected price 10000, got %d", o.Price)
	}
	
	// Duplicate order
	err = manager.AddOrder(order)
	if err != ErrorOrderDuplicate {
		t.Errorf("Expected ErrorOrderDuplicate, got %s", err)
	}
}

func TestMarketManager_AddOrder_InvalidOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	// Invalid ID
	order := Order{
		ID:       0,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
	}
	err := manager.AddOrder(order)
	if err != ErrorOrderIDInvalid {
		t.Errorf("Expected ErrorOrderIDInvalid, got %s", err)
	}
	
	// Invalid quantity
	order.ID = 1
	order.Quantity = 0
	err = manager.AddOrder(order)
	if err != ErrorOrderQuantityInvalid {
		t.Errorf("Expected ErrorOrderQuantityInvalid, got %s", err)
	}
	
	// Invalid limit price
	order.Quantity = 100
	order.Price = 0
	err = manager.AddOrder(order)
	if err != ErrorOrderParameterInvalid {
		t.Errorf("Expected ErrorOrderParameterInvalid, got %s", err)
	}
}

func TestMarketManager_OrderBookNotFound(t *testing.T) {
	manager := NewMarketManager()
	
	order := Order{
		ID:       1,
		SymbolID: 999, // Non-existent symbol
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
	}
	
	err := manager.AddOrder(order)
	if err != ErrorOrderBookNotFound {
		t.Errorf("Expected ErrorOrderBookNotFound, got %s", err)
	}
}

func TestMarketManager_DeleteOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	manager.AddOrder(order)
	
	err := manager.DeleteOrder(1)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	// Check order is gone
	o := manager.GetOrder(1)
	if o != nil {
		t.Error("Expected order to be deleted")
	}
	
	// Delete non-existent
	err = manager.DeleteOrder(1)
	if err != ErrorOrderNotFound {
		t.Errorf("Expected ErrorOrderNotFound, got %s", err)
	}
}

func TestMarketManager_ReduceOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	manager.AddOrder(order)
	
	err := manager.ReduceOrder(1, 30)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	o := manager.GetOrder(1)
	if o.LeavesQuantity != 70 {
		t.Errorf("Expected leaves quantity 70, got %d", o.LeavesQuantity)
	}
}

func TestMarketManager_ReduceOrder_Cancel(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	manager.AddOrder(order)
	
	// Reduce by more than leaves quantity should cancel
	err := manager.ReduceOrder(1, 200)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	o := manager.GetOrder(1)
	if o != nil {
		t.Error("Expected order to be canceled")
	}
}

func TestMarketManager_ModifyOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	manager.AddOrder(order)
	
	err := manager.ModifyOrder(1, 10500, 150)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	o := manager.GetOrder(1)
	if o.Price != 10500 {
		t.Errorf("Expected price 10500, got %d", o.Price)
	}
	if o.Quantity != 150 {
		t.Errorf("Expected quantity 150, got %d", o.Quantity)
	}
	if o.LeavesQuantity != 150 {
		t.Errorf("Expected leaves quantity 150, got %d", o.LeavesQuantity)
	}
}

func TestMarketManager_ReplaceOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	manager.AddOrder(order)
	
	err := manager.ReplaceOrder(1, 2, 10500, 150)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	// Old order should be gone
	o := manager.GetOrder(1)
	if o != nil {
		t.Error("Expected old order to be deleted")
	}
	
	// New order should exist
	o = manager.GetOrder(2)
	if o == nil {
		t.Error("Expected new order to exist")
	}
	if o.Price != 10500 {
		t.Errorf("Expected price 10500, got %d", o.Price)
	}
}

func TestMarketManager_Matching(t *testing.T) {
	manager := NewMarketManager()
	manager.EnableMatching()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	// Add a sell order
	sellOrder := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideSell,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(sellOrder)
	
	// Add a matching buy order
	buyOrder := Order{
		ID:       2,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 50,
		LeavesQuantity: 50,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(buyOrder)
	
	// Buy order should be completely filled and deleted
	o := manager.GetOrder(2)
	if o != nil {
		t.Error("Expected buy order to be filled and deleted")
	}
	
	// Sell order should be partially filled
	o = manager.GetOrder(1)
	if o == nil {
		t.Error("Expected sell order to exist")
	}
	if o.LeavesQuantity != 50 {
		t.Errorf("Expected leaves quantity 50, got %d", o.LeavesQuantity)
	}
	if o.ExecutedQuantity != 50 {
		t.Errorf("Expected executed quantity 50, got %d", o.ExecutedQuantity)
	}
}

func TestMarketManager_Matching_FullFill(t *testing.T) {
	manager := NewMarketManager()
	manager.EnableMatching()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	// Add a sell order
	sellOrder := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideSell,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(sellOrder)
	
	// Add a matching buy order with same quantity
	buyOrder := Order{
		ID:       2,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(buyOrder)
	
	// Both orders should be completely filled
	if manager.GetOrder(1) != nil {
		t.Error("Expected sell order to be filled and deleted")
	}
	if manager.GetOrder(2) != nil {
		t.Error("Expected buy order to be filled and deleted")
	}
}

func TestMarketManager_NoMatching_PriceDifference(t *testing.T) {
	manager := NewMarketManager()
	manager.EnableMatching()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	// Add a sell order at 10000
	sellOrder := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideSell,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(sellOrder)
	
	// Add a buy order at 9500 (below ask)
	buyOrder := Order{
		ID:       2,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    9500,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(buyOrder)
	
	// Both orders should remain
	if manager.GetOrder(1) == nil {
		t.Error("Expected sell order to remain")
	}
	if manager.GetOrder(2) == nil {
		t.Error("Expected buy order to remain")
	}
}

func TestMarketManager_ExecuteOrder(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	
	manager.AddOrder(order)
	
	err := manager.ExecuteOrder(1, 50)
	if err != ErrorOK {
		t.Errorf("Expected ErrorOK, got %s", err)
	}
	
	o := manager.GetOrder(1)
	if o.LeavesQuantity != 50 {
		t.Errorf("Expected leaves quantity 50, got %d", o.LeavesQuantity)
	}
	if o.ExecutedQuantity != 50 {
		t.Errorf("Expected executed quantity 50, got %d", o.ExecutedQuantity)
	}
}

// Test with custom handler
type testMarketHandler struct {
	DefaultMarketHandler
	executions []struct {
		orderID  uint64
		price    uint64
		quantity uint64
	}
}

func (h *testMarketHandler) OnExecuteOrder(order Order, price, quantity uint64) {
	h.executions = append(h.executions, struct {
		orderID  uint64
		price    uint64
		quantity uint64
	}{order.ID, price, quantity})
}

func TestMarketManager_CustomHandler(t *testing.T) {
	handler := &testMarketHandler{}
	manager := NewMarketManagerWithHandler(handler)
	manager.EnableMatching()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	// Add a sell order
	sellOrder := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideSell,
		Price:    10000,
		Quantity: 100,
		LeavesQuantity: 100,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(sellOrder)
	
	// Add a matching buy order
	buyOrder := Order{
		ID:       2,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 50,
		LeavesQuantity: 50,
		TimeInForce: OrderTimeInForceGTC,
		MaxVisibleQuantity: MaxVisibleQuantity,
		Slippage: MaxSlippage,
	}
	manager.AddOrder(buyOrder)
	
	// Should have 2 executions (one for each side)
	if len(handler.executions) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(handler.executions))
	}
}

func TestOrderBook_BestBidAsk(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	ob := manager.GetOrderBook(1)
	
	// Initially empty
	if ob.BestBid() != nil {
		t.Error("Expected no best bid")
	}
	if ob.BestAsk() != nil {
		t.Error("Expected no best ask")
	}
	
	// Add bid orders
	manager.AddOrder(Order{
		ID: 1, SymbolID: 1, Type: OrderTypeLimit, Side: OrderSideBuy,
		Price: 9900, Quantity: 100, LeavesQuantity: 100,
		MaxVisibleQuantity: MaxVisibleQuantity, Slippage: MaxSlippage,
	})
	manager.AddOrder(Order{
		ID: 2, SymbolID: 1, Type: OrderTypeLimit, Side: OrderSideBuy,
		Price: 10000, Quantity: 100, LeavesQuantity: 100,
		MaxVisibleQuantity: MaxVisibleQuantity, Slippage: MaxSlippage,
	})
	
	if ob.BestBid().Price != 10000 {
		t.Errorf("Expected best bid 10000, got %d", ob.BestBid().Price)
	}
	
	// Add ask orders
	manager.AddOrder(Order{
		ID: 3, SymbolID: 1, Type: OrderTypeLimit, Side: OrderSideSell,
		Price: 10100, Quantity: 100, LeavesQuantity: 100,
		MaxVisibleQuantity: MaxVisibleQuantity, Slippage: MaxSlippage,
	})
	manager.AddOrder(Order{
		ID: 4, SymbolID: 1, Type: OrderTypeLimit, Side: OrderSideSell,
		Price: 10200, Quantity: 100, LeavesQuantity: 100,
		MaxVisibleQuantity: MaxVisibleQuantity, Slippage: MaxSlippage,
	})
	
	if ob.BestAsk().Price != 10100 {
		t.Errorf("Expected best ask 10100, got %d", ob.BestAsk().Price)
	}
}

func TestOrderBook_Spread(t *testing.T) {
	manager := NewMarketManager()
	
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)
	
	ob := manager.GetOrderBook(1)
	
	// No spread when empty
	if ob.GetSpread() != 0 {
		t.Errorf("Expected spread 0, got %d", ob.GetSpread())
	}
	
	// Add bid and ask
	manager.AddOrder(Order{
		ID: 1, SymbolID: 1, Type: OrderTypeLimit, Side: OrderSideBuy,
		Price: 10000, Quantity: 100, LeavesQuantity: 100,
		MaxVisibleQuantity: MaxVisibleQuantity, Slippage: MaxSlippage,
	})
	manager.AddOrder(Order{
		ID: 2, SymbolID: 1, Type: OrderTypeLimit, Side: OrderSideSell,
		Price: 10100, Quantity: 100, LeavesQuantity: 100,
		MaxVisibleQuantity: MaxVisibleQuantity, Slippage: MaxSlippage,
	})
	
	if ob.GetSpread() != 100 {
		t.Errorf("Expected spread 100, got %d", ob.GetSpread())
	}
	
	if ob.GetMidPrice() != 10050 {
		t.Errorf("Expected mid price 10050, got %d", ob.GetMidPrice())
	}
}
