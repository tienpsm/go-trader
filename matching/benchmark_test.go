package matching

import (
	"testing"
)

func BenchmarkAddOrder(b *testing.B) {
	manager := NewMarketManager()
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := Order{
			ID:                 uint64(i + 1),
			SymbolID:           1,
			Type:               OrderTypeLimit,
			Side:               OrderSideBuy,
			Price:              uint64(10000 + i%100),
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: MaxVisibleQuantity,
			Slippage:           MaxSlippage,
		}
		manager.AddOrder(order)
	}
}

func BenchmarkAddAndMatchOrders(b *testing.B) {
	manager := NewMarketManager()
	manager.EnableMatching()
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add a sell order
		sellOrder := Order{
			ID:                 uint64(i*2 + 1),
			SymbolID:           1,
			Type:               OrderTypeLimit,
			Side:               OrderSideSell,
			Price:              10000,
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: MaxVisibleQuantity,
			Slippage:           MaxSlippage,
		}
		manager.AddOrder(sellOrder)

		// Add a matching buy order
		buyOrder := Order{
			ID:                 uint64(i*2 + 2),
			SymbolID:           1,
			Type:               OrderTypeLimit,
			Side:               OrderSideBuy,
			Price:              10000,
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: MaxVisibleQuantity,
			Slippage:           MaxSlippage,
		}
		manager.AddOrder(buyOrder)
	}
}

func BenchmarkOrderBookLookup(b *testing.B) {
	manager := NewMarketManager()
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)

	// Pre-populate order book
	for i := 0; i < 1000; i++ {
		order := Order{
			ID:                 uint64(i + 1),
			SymbolID:           1,
			Type:               OrderTypeLimit,
			Side:               OrderSideBuy,
			Price:              uint64(10000 + i),
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: MaxVisibleQuantity,
			Slippage:           MaxSlippage,
		}
		manager.AddOrder(order)
	}

	ob := manager.GetOrderBook(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ob.GetBid(uint64(10000 + i%1000))
	}
}

func BenchmarkAVLTreeInsert(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := NewAVLTree(false)
		for j := 0; j < 100; j++ {
			level := NewLevelNode(LevelTypeBid, uint64(j*10))
			tree.Insert(level)
		}
	}
}

func BenchmarkAVLTreeFind(b *testing.B) {
	tree := NewAVLTree(false)
	for i := 0; i < 1000; i++ {
		level := NewLevelNode(LevelTypeBid, uint64(i*10))
		tree.Insert(level)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tree.Find(uint64((i % 1000) * 10))
	}
}

func BenchmarkOrderListOperations(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list := &OrderList{}
		
		// Add 10 orders
		orders := make([]*OrderNode, 10)
		for j := 0; j < 10; j++ {
			orders[j] = NewOrderNode(Order{ID: uint64(j)})
			list.PushBack(orders[j])
		}
		
		// Remove middle orders
		for j := 2; j < 8; j++ {
			list.Remove(orders[j])
		}
	}
}

func BenchmarkModifyOrder(b *testing.B) {
	manager := NewMarketManager()
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)

	// Pre-populate with orders
	for i := 0; i < 1000; i++ {
		order := Order{
			ID:                 uint64(i + 1),
			SymbolID:           1,
			Type:               OrderTypeLimit,
			Side:               OrderSideBuy,
			Price:              uint64(10000 + i),
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: MaxVisibleQuantity,
			Slippage:           MaxSlippage,
		}
		manager.AddOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderID := uint64((i % 1000) + 1)
		newPrice := uint64(10000 + (i % 100))
		manager.ModifyOrder(orderID, newPrice, 150)
	}
}

func BenchmarkDeleteOrder(b *testing.B) {
	manager := NewMarketManager()
	symbol := NewSymbol(1, "AAPL")
	manager.AddSymbol(symbol)
	manager.AddOrderBook(symbol)

	// Pre-populate with orders
	for i := 0; i < b.N; i++ {
		order := Order{
			ID:                 uint64(i + 1),
			SymbolID:           1,
			Type:               OrderTypeLimit,
			Side:               OrderSideBuy,
			Price:              uint64(10000 + i%100),
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: MaxVisibleQuantity,
			Slippage:           MaxSlippage,
		}
		manager.AddOrder(order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.DeleteOrder(uint64(i + 1))
	}
}
