// Example demonstrating the order matching engine
package main

import (
	"fmt"

	"github.com/tienpsm/go-trader/matching"
)

// TradeLogger logs all market events
type TradeLogger struct {
	matching.DefaultMarketHandler
}

func (h *TradeLogger) OnAddSymbol(symbol matching.Symbol) {
	fmt.Printf("📊 Symbol added: %s (ID: %d)\n", symbol.Name, symbol.ID)
}

func (h *TradeLogger) OnAddOrderBook(ob *matching.OrderBook) {
	fmt.Printf("📚 Order book created for: %s\n", ob.Symbol().Name)
}

func (h *TradeLogger) OnAddOrder(order matching.Order) {
	side := "BUY "
	if order.Side == matching.OrderSideSell {
		side = "SELL"
	}
	fmt.Printf("➕ Order added: ID=%d %s %d @ $%.2f\n",
		order.ID, side, order.Quantity, float64(order.Price)/100)
}

func (h *TradeLogger) OnExecuteOrder(order matching.Order, price, quantity uint64) {
	side := "BUY "
	if order.Side == matching.OrderSideSell {
		side = "SELL"
	}
	fmt.Printf("✅ EXECUTION: Order %d (%s) - %d shares @ $%.2f\n",
		order.ID, side, quantity, float64(price)/100)
}

func (h *TradeLogger) OnDeleteOrder(order matching.Order) {
	if order.LeavesQuantity == 0 {
		fmt.Printf("🏁 Order %d fully filled\n", order.ID)
	} else {
		fmt.Printf("❌ Order %d canceled (remaining: %d)\n", order.ID, order.LeavesQuantity)
	}
}

func (h *TradeLogger) OnUpdateLevel(ob *matching.OrderBook, level matching.Level, top bool) {
	if top {
		levelType := "BID"
		if level.Type == matching.LevelTypeAsk {
			levelType = "ASK"
		}
		fmt.Printf("📈 Top of book update: %s $%.2f x %d (%d orders)\n",
			levelType, float64(level.Price)/100, level.TotalVolume, level.Orders)
	}
}

func main() {
	fmt.Println("===========================================")
	fmt.Println("    Go Trader - Order Matching Engine")
	fmt.Println("===========================================")
	fmt.Println()

	// Create market manager with custom handler
	handler := &TradeLogger{}
	manager := matching.NewMarketManagerWithHandler(handler)
	manager.EnableMatching()

	// Add symbols
	appl := matching.NewSymbol(1, "AAPL")
	manager.AddSymbol(appl)
	manager.AddOrderBook(appl)

	googl := matching.NewSymbol(2, "GOOGL")
	manager.AddSymbol(googl)
	manager.AddOrderBook(googl)

	fmt.Println("\n--- Scenario 1: Simple Match ---")
	
	// Add sell order at $150.00
	manager.AddOrder(matching.Order{
		ID:                 1,
		SymbolID:           1,
		Type:               matching.OrderTypeLimit,
		Side:               matching.OrderSideSell,
		Price:              15000,
		Quantity:           100,
		LeavesQuantity:     100,
		MaxVisibleQuantity: matching.MaxVisibleQuantity,
		Slippage:           matching.MaxSlippage,
	})

	// Add matching buy order at $150.00
	manager.AddOrder(matching.Order{
		ID:                 2,
		SymbolID:           1,
		Type:               matching.OrderTypeLimit,
		Side:               matching.OrderSideBuy,
		Price:              15000,
		Quantity:           100,
		LeavesQuantity:     100,
		MaxVisibleQuantity: matching.MaxVisibleQuantity,
		Slippage:           matching.MaxSlippage,
	})

	fmt.Println("\n--- Scenario 2: Partial Fill ---")
	
	// Add large sell order
	manager.AddOrder(matching.Order{
		ID:                 3,
		SymbolID:           1,
		Type:               matching.OrderTypeLimit,
		Side:               matching.OrderSideSell,
		Price:              15100, // $151.00
		Quantity:           500,
		LeavesQuantity:     500,
		MaxVisibleQuantity: matching.MaxVisibleQuantity,
		Slippage:           matching.MaxSlippage,
	})

	// Add smaller buy order at higher price (crosses spread)
	manager.AddOrder(matching.Order{
		ID:                 4,
		SymbolID:           1,
		Type:               matching.OrderTypeLimit,
		Side:               matching.OrderSideBuy,
		Price:              15200, // $152.00 (aggressive)
		Quantity:           200,
		LeavesQuantity:     200,
		MaxVisibleQuantity: matching.MaxVisibleQuantity,
		Slippage:           matching.MaxSlippage,
	})

	// Check remaining orders
	fmt.Println("\n--- Order Status ---")
	if order := manager.GetOrder(3); order != nil {
		fmt.Printf("Order 3: %d remaining (executed: %d)\n",
			order.LeavesQuantity, order.ExecutedQuantity)
	}

	fmt.Println("\n--- Scenario 3: Building Order Book ---")
	
	// Add multiple buy orders at different prices
	for i := 5; i <= 8; i++ {
		price := uint64(14500 + (8-i)*100) // $145.00, $146.00, $147.00, $148.00
		manager.AddOrder(matching.Order{
			ID:                 uint64(i),
			SymbolID:           1,
			Type:               matching.OrderTypeLimit,
			Side:               matching.OrderSideBuy,
			Price:              price,
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: matching.MaxVisibleQuantity,
			Slippage:           matching.MaxSlippage,
		})
	}

	// Add sell orders at different prices
	for i := 9; i <= 12; i++ {
		price := uint64(15100 + (i-9)*100) // $151.00, $152.00, $153.00, $154.00
		manager.AddOrder(matching.Order{
			ID:                 uint64(i),
			SymbolID:           1,
			Type:               matching.OrderTypeLimit,
			Side:               matching.OrderSideSell,
			Price:              price,
			Quantity:           100,
			LeavesQuantity:     100,
			MaxVisibleQuantity: matching.MaxVisibleQuantity,
			Slippage:           matching.MaxSlippage,
		})
	}

	// Print order book state
	ob := manager.GetOrderBook(1)
	fmt.Println("\n--- AAPL Order Book ---")
	
	if bestBid := ob.BestBid(); bestBid != nil {
		fmt.Printf("Best Bid: $%.2f x %d\n", float64(bestBid.Price)/100, bestBid.TotalVolume)
	}
	if bestAsk := ob.BestAsk(); bestAsk != nil {
		fmt.Printf("Best Ask: $%.2f x %d\n", float64(bestAsk.Price)/100, bestAsk.TotalVolume)
	}
	fmt.Printf("Spread: $%.2f\n", float64(ob.GetSpread())/100)
	fmt.Printf("Mid Price: $%.2f\n", float64(ob.GetMidPrice())/100)

	fmt.Println("\n--- Scenario 4: Order Modification ---")
	
	// Modify order 5's price
	fmt.Println("Modifying order 5...")
	manager.ModifyOrder(5, 14900, 150) // $149.00, 150 shares

	if order := manager.GetOrder(5); order != nil {
		fmt.Printf("Order 5 after modification: %d @ $%.2f\n",
			order.Quantity, float64(order.Price)/100)
	}

	fmt.Println("\n--- Scenario 5: Order Cancellation ---")
	
	manager.DeleteOrder(6)

	fmt.Println("\n===========================================")
	fmt.Println("    Demo Complete!")
	fmt.Println("===========================================")
}
