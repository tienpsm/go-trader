package itch

import (
	"os"
	"testing"
)

// BenchmarkParser_Parse benchmarks parsing individual message types
func BenchmarkParser_Parse(b *testing.B) {
	b.Run("SystemEvent", func(b *testing.B) {
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
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parser.Parse(data)
		}
		b.SetBytes(int64(len(data)))
	})

	b.Run("AddOrder", func(b *testing.B) {
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
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parser.Parse(data)
		}
		b.SetBytes(int64(len(data)))
	})

	b.Run("OrderExecuted", func(b *testing.B) {
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
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parser.Parse(data)
		}
		b.SetBytes(int64(len(data)))
	})

	b.Run("StockDirectory", func(b *testing.B) {
		handler := &DefaultHandler{}
		parser := NewParser(handler)
		data := make([]byte, 39)
		data[0] = 'R'
		data[1], data[2] = 0, 1
		data[3], data[4] = 0, 2
		data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
		copy(data[11:19], []byte("AAPL    "))
		
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parser.Parse(data)
		}
		b.SetBytes(int64(len(data)))
	})

	b.Run("Trade", func(b *testing.B) {
		handler := &DefaultHandler{}
		parser := NewParser(handler)
		data := make([]byte, 44)
		data[0] = 'P'
		data[1], data[2] = 0, 1
		data[3], data[4] = 0, 2
		data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
		data[11], data[12], data[13], data[14], data[15], data[16], data[17], data[18] = 0, 0, 0, 0, 0, 0, 0, 1
		data[19] = 'B'
		data[20], data[21], data[22], data[23] = 0, 0, 0, 100
		copy(data[24:32], []byte("AAPL    "))
		data[32], data[33], data[34], data[35] = 0, 0, 39, 16
		
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parser.Parse(data)
		}
		b.SetBytes(int64(len(data)))
	})

	b.Run("NOII", func(b *testing.B) {
		handler := &DefaultHandler{}
		parser := NewParser(handler)
		data := make([]byte, 50)
		data[0] = 'I'
		data[1], data[2] = 0, 1
		data[3], data[4] = 0, 2
		data[5], data[6], data[7], data[8], data[9], data[10] = 0, 0, 0, 0, 0, 100
		
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parser.Parse(data)
		}
		b.SetBytes(int64(len(data)))
	})
}

// BenchmarkParser_ParseAll benchmarks parsing a buffer of multiple messages
func BenchmarkParser_ParseAll(b *testing.B) {
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
	
	// Trade (44 bytes)
	trade := make([]byte, 44)
	trade[0] = 'P'
	trade[19] = 'B'
	copy(trade[24:32], []byte("AAPL    "))
	data = append(data, trade...)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parser.ParseAll(data)
	}
	b.SetBytes(int64(len(data)))
}

// BenchmarkParser_ParseFile benchmarks parsing the entire sample.itch file
func BenchmarkParser_ParseFile(b *testing.B) {
	filename := "../testdata/sample.itch"
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		b.Skipf("Sample file not found: %s", filename)
		return
	}
	
	// Get file size for throughput calculation
	fileInfo, err := os.Stat(filename)
	if err != nil {
		b.Fatal(err)
	}
	fileSize := fileInfo.Size()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		handler := &DefaultHandler{}
		_, err := ParseFile(filename, handler)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(fileSize)
}

// BenchmarkParser_ParseFileWithStats benchmarks parsing with statistics collection
func BenchmarkParser_ParseFileWithStats(b *testing.B) {
	filename := "../testdata/sample.itch"
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		b.Skipf("Sample file not found: %s", filename)
		return
	}
	
	// Get file size for throughput calculation
	fileInfo, err := os.Stat(filename)
	if err != nil {
		b.Fatal(err)
	}
	fileSize := fileInfo.Size()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		handler := &StatsHandler{}
		_, err := ParseFile(filename, handler)
		if err != nil {
			b.Fatal(err)
		}
		// Verify we got the expected number of messages
		if handler.Stats.TotalMessages != 1563071 {
			b.Fatalf("Expected 1563071 messages, got %d", handler.Stats.TotalMessages)
		}
	}
	b.SetBytes(fileSize)
}
