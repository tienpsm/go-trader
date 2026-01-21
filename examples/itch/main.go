// Example demonstrating the NASDAQ ITCH protocol parser
package main

import (
	"encoding/binary"
	"fmt"

	"github.com/tienpsm/go-trader/itch"
)

// StatsHandler collects statistics from ITCH messages
type StatsHandler struct {
	itch.DefaultHandler
	
	SystemEvents  int
	StockCount    int
	AddOrders     int
	Executions    int
	Cancellations int
	Deletions     int
	Trades        int
	
	// Track order volume
	TotalBuyShares  uint64
	TotalSellShares uint64
}

func (h *StatsHandler) OnSystemEvent(msg itch.SystemEventMessage) error {
	h.SystemEvents++
	eventName := "Unknown"
	switch msg.EventCode {
	case 'O':
		eventName = "Start of Messages"
	case 'S':
		eventName = "Start of System Hours"
	case 'Q':
		eventName = "Start of Market Hours"
	case 'M':
		eventName = "End of Market Hours"
	case 'E':
		eventName = "End of System Hours"
	case 'C':
		eventName = "End of Messages"
	}
	fmt.Printf("📅 System Event: %s (code: %c)\n", eventName, msg.EventCode)
	return nil
}

func (h *StatsHandler) OnStockDirectory(msg itch.StockDirectoryMessage) error {
	h.StockCount++
	stock := string(msg.Stock[:])
	fmt.Printf("📊 Stock Directory: %s (Locate: %d)\n", stock, msg.StockLocate)
	return nil
}

func (h *StatsHandler) OnAddOrder(msg itch.AddOrderMessage) error {
	h.AddOrders++
	if msg.BuySellIndicator == 'B' {
		h.TotalBuyShares += uint64(msg.Shares)
	} else {
		h.TotalSellShares += uint64(msg.Shares)
	}
	
	side := "BUY "
	if msg.BuySellIndicator == 'S' {
		side = "SELL"
	}
	stock := string(msg.Stock[:])
	fmt.Printf("➕ Add Order: Ref=%d %s %d shares of %s @ %d\n",
		msg.OrderReferenceNumber, side, msg.Shares, stock, msg.Price)
	return nil
}

func (h *StatsHandler) OnOrderExecuted(msg itch.OrderExecutedMessage) error {
	h.Executions++
	fmt.Printf("✅ Order Executed: Ref=%d, %d shares, Match=%d\n",
		msg.OrderReferenceNumber, msg.ExecutedShares, msg.MatchNumber)
	return nil
}

func (h *StatsHandler) OnOrderCancel(msg itch.OrderCancelMessage) error {
	h.Cancellations++
	fmt.Printf("⚠️  Order Cancel: Ref=%d, %d shares canceled\n",
		msg.OrderReferenceNumber, msg.CanceledShares)
	return nil
}

func (h *StatsHandler) OnOrderDelete(msg itch.OrderDeleteMessage) error {
	h.Deletions++
	fmt.Printf("❌ Order Delete: Ref=%d\n", msg.OrderReferenceNumber)
	return nil
}

func (h *StatsHandler) OnTrade(msg itch.TradeMessage) error {
	h.Trades++
	side := "BUY "
	if msg.BuySellIndicator == 'S' {
		side = "SELL"
	}
	stock := string(msg.Stock[:])
	fmt.Printf("💰 Trade: %s %d shares of %s @ %d (Match=%d)\n",
		side, msg.Shares, stock, msg.Price, msg.MatchNumber)
	return nil
}

func (h *StatsHandler) PrintStats() {
	fmt.Println("\n===========================================")
	fmt.Println("              ITCH Statistics")
	fmt.Println("===========================================")
	fmt.Printf("System Events:    %d\n", h.SystemEvents)
	fmt.Printf("Stocks:           %d\n", h.StockCount)
	fmt.Printf("Add Orders:       %d\n", h.AddOrders)
	fmt.Printf("  Buy Volume:     %d shares\n", h.TotalBuyShares)
	fmt.Printf("  Sell Volume:    %d shares\n", h.TotalSellShares)
	fmt.Printf("Executions:       %d\n", h.Executions)
	fmt.Printf("Cancellations:    %d\n", h.Cancellations)
	fmt.Printf("Deletions:        %d\n", h.Deletions)
	fmt.Printf("Trades:           %d\n", h.Trades)
	fmt.Println("===========================================")
}

// Helper to create a mock ITCH message for demonstration
func createSystemEvent(eventCode byte) []byte {
	data := make([]byte, 12)
	data[0] = 'S'                    // Message type
	binary.BigEndian.PutUint16(data[1:3], 0) // Stock locate
	binary.BigEndian.PutUint16(data[3:5], 0) // Tracking number
	// Timestamp (6 bytes) - simplified
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 1, 0
	data[11] = eventCode
	return data
}

func createAddOrder(ref uint64, buySell byte, shares uint32, stock string, price uint32) []byte {
	data := make([]byte, 36)
	data[0] = 'A' // Message type
	binary.BigEndian.PutUint16(data[1:3], 1) // Stock locate
	binary.BigEndian.PutUint16(data[3:5], 0) // Tracking number
	// Timestamp (6 bytes)
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 2, 0
	binary.BigEndian.PutUint64(data[11:19], ref) // Order reference
	data[19] = buySell
	binary.BigEndian.PutUint32(data[20:24], shares)
	copy(data[24:32], []byte(stock+"        ")[:8])
	binary.BigEndian.PutUint32(data[32:36], price)
	return data
}

func createOrderExecuted(ref uint64, shares uint32, match uint64) []byte {
	data := make([]byte, 31)
	data[0] = 'E' // Message type
	binary.BigEndian.PutUint16(data[1:3], 1) // Stock locate
	binary.BigEndian.PutUint16(data[3:5], 0) // Tracking number
	// Timestamp (6 bytes)
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 3, 0
	binary.BigEndian.PutUint64(data[11:19], ref)
	binary.BigEndian.PutUint32(data[19:23], shares)
	binary.BigEndian.PutUint64(data[23:31], match)
	return data
}

func createOrderDelete(ref uint64) []byte {
	data := make([]byte, 19)
	data[0] = 'D' // Message type
	binary.BigEndian.PutUint16(data[1:3], 1) // Stock locate
	binary.BigEndian.PutUint16(data[3:5], 0) // Tracking number
	// Timestamp (6 bytes)
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 4, 0
	binary.BigEndian.PutUint64(data[11:19], ref)
	return data
}

func main() {
	fmt.Println("===========================================")
	fmt.Println("    Go Trader - ITCH Protocol Demo")
	fmt.Println("===========================================")
	fmt.Println()
	
	// Create handler and parser
	handler := &StatsHandler{}
	parser := itch.NewParser(handler)
	
	// Simulate ITCH message stream
	fmt.Println("Simulating ITCH message stream...")
	fmt.Println()
	
	// Create sample messages
	var messages []byte
	
	// System event: Start of trading
	messages = append(messages, createSystemEvent('O')...)
	messages = append(messages, createSystemEvent('S')...)
	messages = append(messages, createSystemEvent('Q')...)
	
	// Add some orders
	messages = append(messages, createAddOrder(1001, 'B', 100, "AAPL", 15000)...)
	messages = append(messages, createAddOrder(1002, 'S', 200, "AAPL", 15100)...)
	messages = append(messages, createAddOrder(1003, 'B', 150, "GOOGL", 280050)...)
	messages = append(messages, createAddOrder(1004, 'S', 75, "MSFT", 37500)...)
	messages = append(messages, createAddOrder(1005, 'B', 300, "TSLA", 25000)...)
	
	// Execute some orders
	messages = append(messages, createOrderExecuted(1001, 50, 1)...)
	messages = append(messages, createOrderExecuted(1002, 100, 2)...)
	messages = append(messages, createOrderExecuted(1003, 150, 3)...)
	
	// Delete remaining orders
	messages = append(messages, createOrderDelete(1001)...)
	messages = append(messages, createOrderDelete(1004)...)
	
	// End of trading
	messages = append(messages, createSystemEvent('M')...)
	messages = append(messages, createSystemEvent('E')...)
	messages = append(messages, createSystemEvent('C')...)
	
	// Parse all messages
	consumed, count, err := parser.ParseAll(messages)
	if err != nil {
		fmt.Printf("Error parsing messages: %v\n", err)
		return
	}
	
	fmt.Printf("\n✓ Parsed %d messages (%d bytes)\n", count, consumed)
	
	// Print statistics
	handler.PrintStats()
}
