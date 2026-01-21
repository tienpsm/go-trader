package itch

import (
	"testing"
)

func BenchmarkParseSystemEvent(b *testing.B) {
	handler := &DefaultHandler{}
	parser := NewParser(handler)
	
	data := []byte{
		'S',        // Type
		0, 1,       // StockLocate
		0, 2,       // TrackingNumber
		0, 0, 0, 0, 0, 100, // Timestamp
		'O',        // EventCode
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(data)
	}
}

func BenchmarkParseAddOrder(b *testing.B) {
	handler := &DefaultHandler{}
	parser := NewParser(handler)
	
	data := make([]byte, 36)
	data[0] = 'A'
	data[1], data[2] = 0, 1
	data[3], data[4] = 0, 2
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18] = 0, 0, 0, 0, 0, 0, 0, 1
	data[19] = 'B'
	data[20], data[21], data[22], data[23] = 0, 0, 0, 100
	copy(data[24:32], []byte("AAPL    "))
	data[32], data[33], data[34], data[35] = 0, 0, 39, 16
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(data)
	}
}

func BenchmarkParseOrderExecuted(b *testing.B) {
	handler := &DefaultHandler{}
	parser := NewParser(handler)
	
	data := make([]byte, 31)
	data[0] = 'E'
	data[1], data[2] = 0, 1
	data[3], data[4] = 0, 2
	data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
	data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18] = 0, 0, 0, 0, 0, 0, 0, 1
	data[19], data[20], data[21], data[22] = 0, 0, 0, 50
	data[23], data[24], data[25], data[26], data[27], data[28], data[29], data[30] = 0, 0, 0, 0, 0, 0, 0, 1
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(data)
	}
}

func BenchmarkParseAllMessages(b *testing.B) {
	handler := &DefaultHandler{}
	parser := NewParser(handler)
	
	// Create a mix of messages
	var data []byte
	
	// System event (12 bytes)
	sysEvent := []byte{'S', 0, 1, 0, 2, 0, 0, 0, 0, 0, 100, 'O'}
	data = append(data, sysEvent...)
	
	// Add order (36 bytes)
	addOrder := make([]byte, 36)
	addOrder[0] = 'A'
	addOrder[19] = 'B'
	copy(addOrder[24:32], []byte("AAPL    "))
	data = append(data, addOrder...)
	
	// Order executed (31 bytes)
	orderExec := make([]byte, 31)
	orderExec[0] = 'E'
	data = append(data, orderExec...)
	
	// Order delete (19 bytes)
	orderDel := make([]byte, 19)
	orderDel[0] = 'D'
	data = append(data, orderDel...)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ParseAll(data)
	}
}

func BenchmarkReadUint48BE(b *testing.B) {
	data := []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x64}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readUint48BE(data)
	}
}

func BenchmarkReadUint64BE(b *testing.B) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readUint64BE(data)
	}
}
