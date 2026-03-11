package itch

import (
	"os"
	"testing"
)

// TestHandler tracks messages received
type TestHandler struct {
	DefaultHandler
	systemEvents     []SystemEventMessage
	stockDirectories []StockDirectoryMessage
	addOrders        []AddOrderMessage
	orderExecuted    []OrderExecutedMessage
	orderDeleted     []OrderDeleteMessage
	unknownMessages  int
}

func (h *TestHandler) OnSystemEvent(msg SystemEventMessage) error {
	h.systemEvents = append(h.systemEvents, msg)
	return nil
}

func (h *TestHandler) OnStockDirectory(msg StockDirectoryMessage) error {
	h.stockDirectories = append(h.stockDirectories, msg)
	return nil
}

func (h *TestHandler) OnAddOrder(msg AddOrderMessage) error {
	h.addOrders = append(h.addOrders, msg)
	return nil
}

func (h *TestHandler) OnOrderExecuted(msg OrderExecutedMessage) error {
	h.orderExecuted = append(h.orderExecuted, msg)
	return nil
}

func (h *TestHandler) OnOrderDelete(msg OrderDeleteMessage) error {
	h.orderDeleted = append(h.orderDeleted, msg)
	return nil
}

func (h *TestHandler) OnUnknownMessage(msgType byte, data []byte) error {
	h.unknownMessages++
	return nil
}

func TestParser_SystemEvent(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// System event message (12 bytes)
	// Type (1) + StockLocate (2) + TrackingNumber (2) + Timestamp (6) + EventCode (1)
	data := []byte{
		'S',        // Type
		0, 1,       // StockLocate
		0, 2,       // TrackingNumber
		0, 0, 0, 0, 0, 100, // Timestamp (6 bytes)
		'O',        // EventCode (market open)
	}
	
	consumed, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if consumed != 12 {
		t.Errorf("Expected 12 bytes consumed, got %d", consumed)
	}
	if len(handler.systemEvents) != 1 {
		t.Fatalf("Expected 1 system event, got %d", len(handler.systemEvents))
	}
	
	msg := handler.systemEvents[0]
	if msg.StockLocate != 1 {
		t.Errorf("Expected StockLocate 1, got %d", msg.StockLocate)
	}
	if msg.EventCode != 'O' {
		t.Errorf("Expected EventCode 'O', got %c", msg.EventCode)
	}
}

func TestParser_AddOrder(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// Add order message (36 bytes)
	data := make([]byte, 36)
	data[0] = 'A'                    // Type
	data[1], data[2] = 0, 1          // StockLocate
	data[3], data[4] = 0, 2          // TrackingNumber
	// Timestamp (6 bytes) - bytes 5-10
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	// OrderReferenceNumber (8 bytes) - bytes 11-18
	data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18] = 0, 0, 0, 0, 0, 0, 0, 1
	data[19] = 'B'                   // BuySellIndicator (Buy)
	// Shares (4 bytes) - bytes 20-23
	data[20], data[21], data[22], data[23] = 0, 0, 0, 100
	// Stock (8 bytes) - bytes 24-31
	copy(data[24:32], []byte("AAPL    "))
	// Price (4 bytes) - bytes 32-35
	data[32], data[33], data[34], data[35] = 0, 0, 39, 16 // 10000
	
	consumed, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if consumed != 36 {
		t.Errorf("Expected 36 bytes consumed, got %d", consumed)
	}
	if len(handler.addOrders) != 1 {
		t.Fatalf("Expected 1 add order, got %d", len(handler.addOrders))
	}
	
	msg := handler.addOrders[0]
	if msg.OrderReferenceNumber != 1 {
		t.Errorf("Expected OrderReferenceNumber 1, got %d", msg.OrderReferenceNumber)
	}
	if msg.BuySellIndicator != 'B' {
		t.Errorf("Expected BuySellIndicator 'B', got %c", msg.BuySellIndicator)
	}
	if msg.Shares != 100 {
		t.Errorf("Expected Shares 100, got %d", msg.Shares)
	}
	if msg.Price != 10000 {
		t.Errorf("Expected Price 10000, got %d", msg.Price)
	}
}

func TestParser_OrderExecuted(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// Order executed message (31 bytes)
	data := make([]byte, 31)
	data[0] = 'E'                    // Type
	data[1], data[2] = 0, 1          // StockLocate
	data[3], data[4] = 0, 2          // TrackingNumber
	// Timestamp (6 bytes)
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	// OrderReferenceNumber (8 bytes)
	data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18] = 0, 0, 0, 0, 0, 0, 0, 1
	// ExecutedShares (4 bytes)
	data[19], data[20], data[21], data[22] = 0, 0, 0, 50
	// MatchNumber (8 bytes)
	data[23], data[24], data[25], data[26], data[27], data[28], data[29], data[30] = 0, 0, 0, 0, 0, 0, 0, 1
	
	consumed, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if consumed != 31 {
		t.Errorf("Expected 31 bytes consumed, got %d", consumed)
	}
	if len(handler.orderExecuted) != 1 {
		t.Fatalf("Expected 1 order executed, got %d", len(handler.orderExecuted))
	}
	
	msg := handler.orderExecuted[0]
	if msg.OrderReferenceNumber != 1 {
		t.Errorf("Expected OrderReferenceNumber 1, got %d", msg.OrderReferenceNumber)
	}
	if msg.ExecutedShares != 50 {
		t.Errorf("Expected ExecutedShares 50, got %d", msg.ExecutedShares)
	}
}

func TestParser_OrderDelete(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// Order delete message (19 bytes)
	data := make([]byte, 19)
	data[0] = 'D'                    // Type
	data[1], data[2] = 0, 1          // StockLocate
	data[3], data[4] = 0, 2          // TrackingNumber
	// Timestamp (6 bytes)
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	// OrderReferenceNumber (8 bytes)
	data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18] = 0, 0, 0, 0, 0, 0, 0, 1
	
	consumed, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if consumed != 19 {
		t.Errorf("Expected 19 bytes consumed, got %d", consumed)
	}
	if len(handler.orderDeleted) != 1 {
		t.Fatalf("Expected 1 order deleted, got %d", len(handler.orderDeleted))
	}
	
	msg := handler.orderDeleted[0]
	if msg.OrderReferenceNumber != 1 {
		t.Errorf("Expected OrderReferenceNumber 1, got %d", msg.OrderReferenceNumber)
	}
}

func TestParser_UnknownMessage(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// Unknown message type
	data := []byte{'Z', 1, 2, 3, 4, 5}
	
	_, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if handler.unknownMessages != 1 {
		t.Errorf("Expected 1 unknown message, got %d", handler.unknownMessages)
	}
}

func TestParser_InsufficientData(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// System event needs 12 bytes, give it only 5
	data := []byte{'S', 0, 1, 0, 2}
	
	consumed, err := parser.Parse(data)
	if err != ErrInsufficientData {
		t.Errorf("Expected ErrInsufficientData, got %v", err)
	}
	if consumed != 0 {
		t.Errorf("Expected 0 bytes consumed, got %d", consumed)
	}
}

func TestParser_ParseAll(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// Create two system event messages
	data := make([]byte, 24)
	
	// First message
	data[0] = 'S'
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	data[11] = 'O'
	
	// Second message
	data[12] = 'S'
	data[17], data[18], data[19], data[20], data[21], data[22] = 0, 0, 0, 0, 0, 200
	data[23] = 'C'
	
	consumed, count, err := parser.ParseAll(data)
	if err != nil {
		t.Fatalf("ParseAll error: %v", err)
	}
	if consumed != 24 {
		t.Errorf("Expected 24 bytes consumed, got %d", consumed)
	}
	if count != 2 {
		t.Errorf("Expected 2 messages, got %d", count)
	}
	if len(handler.systemEvents) != 2 {
		t.Errorf("Expected 2 system events, got %d", len(handler.systemEvents))
	}
}

func TestParser_ParseAll_Partial(t *testing.T) {
	handler := &TestHandler{}
	parser := NewParser(handler)
	
	// One complete message + partial second message
	data := make([]byte, 17)
	
	// First message (complete)
	data[0] = 'S'
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	data[11] = 'O'
	
	// Partial second message (only 5 bytes)
	data[12] = 'S'
	data[13] = 0
	data[14] = 1
	data[15] = 0
	data[16] = 2
	
	consumed, count, err := parser.ParseAll(data)
	if err != nil {
		t.Fatalf("ParseAll error: %v", err)
	}
	if consumed != 12 {
		t.Errorf("Expected 12 bytes consumed, got %d", consumed)
	}
	if count != 1 {
		t.Errorf("Expected 1 message, got %d", count)
	}
}

func TestMessageTypes(t *testing.T) {
	// Test message type constants
	if MessageTypeSystemEvent != 'S' {
		t.Errorf("Expected 'S', got %c", MessageTypeSystemEvent)
	}
	if MessageTypeAddOrder != 'A' {
		t.Errorf("Expected 'A', got %c", MessageTypeAddOrder)
	}
	if MessageTypeOrderExecuted != 'E' {
		t.Errorf("Expected 'E', got %c", MessageTypeOrderExecuted)
	}
	if MessageTypeOrderDelete != 'D' {
		t.Errorf("Expected 'D', got %c", MessageTypeOrderDelete)
	}
	if MessageTypeTrade != 'P' {
		t.Errorf("Expected 'P', got %c", MessageTypeTrade)
	}
}

func TestAddOrderMessage_String(t *testing.T) {
	msg := AddOrderMessage{
		Type:                 'A',
		OrderReferenceNumber: 12345,
		BuySellIndicator:     'B',
		Shares:               100,
		Price:                10000,
	}
	copy(msg.Stock[:], []byte("AAPL    "))
	
	str := msg.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}

func TestSystemEventMessage_String(t *testing.T) {
	msg := SystemEventMessage{
		Type:      'S',
		EventCode: 'O',
		Timestamp: 12345,
	}
	
	str := msg.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}

func TestDefaultHandler(t *testing.T) {
	handler := &DefaultHandler{}
	
	// All methods should return nil (no-op)
	if err := handler.OnSystemEvent(SystemEventMessage{}); err != nil {
		t.Errorf("OnSystemEvent should return nil, got %v", err)
	}
	if err := handler.OnAddOrder(AddOrderMessage{}); err != nil {
		t.Errorf("OnAddOrder should return nil, got %v", err)
	}
	if err := handler.OnOrderExecuted(OrderExecutedMessage{}); err != nil {
		t.Errorf("OnOrderExecuted should return nil, got %v", err)
	}
	if err := handler.OnOrderDelete(OrderDeleteMessage{}); err != nil {
		t.Errorf("OnOrderDelete should return nil, got %v", err)
	}
	if err := handler.OnUnknownMessage('Z', []byte{}); err != nil {
		t.Errorf("OnUnknownMessage should return nil, got %v", err)
	}
}

func TestReadUint48BE(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x64} // 100 in 6 bytes big-endian
	result := readUint48BE(data)
	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}
	
	data = []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00} // 0x01000000 = 16777216
	result = readUint48BE(data)
	if result != 16777216 {
		t.Errorf("Expected 16777216, got %d", result)
	}
}

func TestParser_SampleFile(t *testing.T) {
	// Test with the sample.itch file from CppTrader
	filename := "../testdata/sample.itch"
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s. Run download script first.", filename)
		return
	}
	
	// Use stats handler to collect message statistics
	handler := &StatsHandler{}
	
	// Parse the file
	bytesRead, err := ParseFile(filename, handler)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	
	// Verify we read data
	if bytesRead == 0 {
		t.Fatal("No bytes read from file")
	}
	
	// Verify total message count matches expected (1,563,071 from CppTrader)
	expectedMessages := 1563071
	if handler.Stats.TotalMessages != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, handler.Stats.TotalMessages)
	}
	
	// Print statistics
	t.Logf("Parsed %d bytes", bytesRead)
	t.Logf("Total Messages: %d", handler.Stats.TotalMessages)
	t.Logf("  System Events: %d", handler.Stats.SystemEvents)
	t.Logf("  Stock Directories: %d", handler.Stats.StockDirectories)
	t.Logf("  Stock Trading Actions: %d", handler.Stats.StockTradingActions)
	t.Logf("  Add Orders: %d", handler.Stats.AddOrders)
	t.Logf("  Add Orders (MPID): %d", handler.Stats.AddOrderMPID)
	t.Logf("  Order Executed: %d", handler.Stats.OrderExecuted)
	t.Logf("  Order Executed (Price): %d", handler.Stats.OrderExecutedPrice)
	t.Logf("  Order Cancels: %d", handler.Stats.OrderCancels)
	t.Logf("  Order Deletes: %d", handler.Stats.OrderDeletes)
	t.Logf("  Order Replaces: %d", handler.Stats.OrderReplaces)
	t.Logf("  Trades: %d", handler.Stats.Trades)
	t.Logf("  Cross Trades: %d", handler.Stats.CrossTrades)
	t.Logf("  Broken Trades: %d", handler.Stats.BrokenTrades)
	t.Logf("  NOII: %d", handler.Stats.NOII)
	t.Logf("  RPII: %d", handler.Stats.RPII)
	
	// Verify no unknown messages
	if handler.Stats.UnknownMessages > 0 {
		t.Errorf("Encountered %d unknown messages", handler.Stats.UnknownMessages)
	}
}
