// ITCH Analyzer - A CLI tool for analyzing NASDAQ ITCH data files
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tienpsm/go-trader/itch"
)

type options struct {
	verbose   bool
	benchmark bool
	validate  bool
}

// AnalyzerHandler collects detailed statistics and optionally prints messages
type AnalyzerHandler struct {
	itch.StatsHandler
	verbose            bool
	buyOrders          int
	sellOrders         int
	buyVolume          uint64
	sellVolume         uint64
	minPrice           uint32
	maxPrice           uint32
	priceInitialized   bool
}

func (h *AnalyzerHandler) OnAddOrder(msg itch.AddOrderMessage) error {
	// Call parent to update stats
	h.StatsHandler.OnAddOrder(msg)
	
	// Track buy/sell orders and volume
	if msg.BuySellIndicator == 'B' {
		h.buyOrders++
		h.buyVolume += uint64(msg.Shares)
	} else {
		h.sellOrders++
		h.sellVolume += uint64(msg.Shares)
	}
	
	// Track price range
	if !h.priceInitialized || msg.Price < h.minPrice {
		h.minPrice = msg.Price
		h.priceInitialized = true
	}
	if msg.Price > h.maxPrice {
		h.maxPrice = msg.Price
	}
	
	if h.verbose {
		stock := string(msg.Stock[:])
		side := "BUY "
		if msg.BuySellIndicator == 'S' {
			side = "SELL"
		}
		fmt.Printf("AddOrder: Ref=%d %s %d shares of %s @ $%.2f\n",
			msg.OrderReferenceNumber, side, msg.Shares, stock, float64(msg.Price)/10000.0)
	}
	
	return nil
}

func (h *AnalyzerHandler) OnAddOrderMPID(msg itch.AddOrderMPIDMessage) error {
	// Call parent to update stats
	h.StatsHandler.OnAddOrderMPID(msg)
	
	// Track buy/sell orders and volume
	if msg.BuySellIndicator == 'B' {
		h.buyOrders++
		h.buyVolume += uint64(msg.Shares)
	} else {
		h.sellOrders++
		h.sellVolume += uint64(msg.Shares)
	}
	
	// Track price range
	if !h.priceInitialized || msg.Price < h.minPrice {
		h.minPrice = msg.Price
		h.priceInitialized = true
	}
	if msg.Price > h.maxPrice {
		h.maxPrice = msg.Price
	}
	
	if h.verbose {
		stock := string(msg.Stock[:])
		side := "BUY "
		if msg.BuySellIndicator == 'S' {
			side = "SELL"
		}
		fmt.Printf("AddOrderMPID: Ref=%d %s %d shares of %s @ $%.2f\n",
			msg.OrderReferenceNumber, side, msg.Shares, stock, float64(msg.Price)/10000.0)
	}
	
	return nil
}

func (h *AnalyzerHandler) OnSystemEvent(msg itch.SystemEventMessage) error {
	// Call parent to update stats
	h.StatsHandler.OnSystemEvent(msg)
	
	if h.verbose {
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
		fmt.Printf("SystemEvent: %s (code: %c)\n", eventName, msg.EventCode)
	}
	
	return nil
}

func (h *AnalyzerHandler) OnOrderExecuted(msg itch.OrderExecutedMessage) error {
	// Call parent to update stats
	h.StatsHandler.OnOrderExecuted(msg)
	
	if h.verbose {
		fmt.Printf("OrderExecuted: Ref=%d, %d shares, Match=%d\n",
			msg.OrderReferenceNumber, msg.ExecutedShares, msg.MatchNumber)
	}
	
	return nil
}

func (h *AnalyzerHandler) OnTrade(msg itch.TradeMessage) error {
	// Call parent to update stats
	h.StatsHandler.OnTrade(msg)
	
	if h.verbose {
		stock := string(msg.Stock[:])
		side := "BUY "
		if msg.BuySellIndicator == 'S' {
			side = "SELL"
		}
		fmt.Printf("Trade: %s %d shares of %s @ $%.2f (Match=%d)\n",
			side, msg.Shares, stock, float64(msg.Price)/10000.0, msg.MatchNumber)
	}
	
	return nil
}

func printStats(handler *AnalyzerHandler, elapsed time.Duration) {
	separator := "================================================================================"
	divider := "--------------------------------------------------------------------------------"
	
	fmt.Println("\n" + separator)
	fmt.Println("                     ITCH File Analysis Results")
	fmt.Println(separator)
	
	// Message counts
	fmt.Println("\n📊 Message Statistics:")
	fmt.Println(divider)
	fmt.Printf("  Total Messages:         %d\n", handler.Stats.TotalMessages)
	fmt.Printf("  System Events:          %d\n", handler.Stats.SystemEvents)
	fmt.Printf("  Stock Directories:      %d\n", handler.Stats.StockDirectories)
	fmt.Printf("  Stock Trading Actions:  %d\n", handler.Stats.StockTradingActions)
	fmt.Printf("  Add Orders:             %d\n", handler.Stats.AddOrders)
	fmt.Printf("  Add Orders (MPID):      %d\n", handler.Stats.AddOrderMPID)
	fmt.Printf("  Order Executed:         %d\n", handler.Stats.OrderExecuted)
	fmt.Printf("  Order Executed (Price): %d\n", handler.Stats.OrderExecutedPrice)
	fmt.Printf("  Order Cancels:          %d\n", handler.Stats.OrderCancels)
	fmt.Printf("  Order Deletes:          %d\n", handler.Stats.OrderDeletes)
	fmt.Printf("  Order Replaces:         %d\n", handler.Stats.OrderReplaces)
	fmt.Printf("  Trades:                 %d\n", handler.Stats.Trades)
	fmt.Printf("  Cross Trades:           %d\n", handler.Stats.CrossTrades)
	fmt.Printf("  Broken Trades:          %d\n", handler.Stats.BrokenTrades)
	fmt.Printf("  NOII:                   %d\n", handler.Stats.NOII)
	fmt.Printf("  RPII:                   %d\n", handler.Stats.RPII)
	
	// Order statistics
	fmt.Println("\n📈 Order Statistics:")
	fmt.Println(divider)
	totalOrders := handler.buyOrders + handler.sellOrders
	if totalOrders > 0 {
		buyPercent := float64(handler.buyOrders) * 100.0 / float64(totalOrders)
		sellPercent := float64(handler.sellOrders) * 100.0 / float64(totalOrders)
		fmt.Printf("  Buy Orders:             %d (%.2f%%)\n", handler.buyOrders, buyPercent)
		fmt.Printf("  Sell Orders:            %d (%.2f%%)\n", handler.sellOrders, sellPercent)
		fmt.Printf("  Buy Volume:             %d shares\n", handler.buyVolume)
		fmt.Printf("  Sell Volume:            %d shares\n", handler.sellVolume)
		
		if handler.priceInitialized {
			fmt.Printf("  Price Range:            $%.2f - $%.2f\n",
				float64(handler.minPrice)/10000.0, float64(handler.maxPrice)/10000.0)
		}
	} else {
		fmt.Println("  No orders found")
	}
	
	// Performance metrics
	fmt.Println("\n⚡ Performance Metrics:")
	fmt.Println(divider)
	fmt.Printf("  Parsing Time:           %.2f seconds\n", elapsed.Seconds())
	
	if elapsed.Seconds() > 0 {
		msgPerSec := float64(handler.Stats.TotalMessages) / elapsed.Seconds()
		fmt.Printf("  Throughput:             %.2f messages/second\n", msgPerSec)
		fmt.Printf("                          %.2f K messages/second\n", msgPerSec/1000.0)
	}
	
	fmt.Println(separator)
}

func analyzeFile(filename string, opts options) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}
	
	// Get file size
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	
	fmt.Printf("Analyzing ITCH file: %s\n", filename)
	fmt.Printf("File size: %.2f MB\n", float64(fileSize)/(1024*1024))
	fmt.Println()
	
	// Create handler
	handler := &AnalyzerHandler{
		verbose: opts.verbose,
	}
	
	// Parse file
	startTime := time.Now()
	bytesRead, err := itch.ParseFile(filename, handler)
	elapsed := time.Since(startTime)
	
	if err != nil {
		return fmt.Errorf("error parsing file: %v", err)
	}
	
	if opts.validate {
		// Validation mode - just check for errors
		if handler.Stats.UnknownMessages > 0 {
			fmt.Printf("❌ Validation FAILED: %d unknown messages\n", handler.Stats.UnknownMessages)
			return fmt.Errorf("file contains unknown messages")
		}
		fmt.Printf("✅ Validation PASSED: Successfully parsed %d messages\n", handler.Stats.TotalMessages)
		fmt.Printf("   Parsed %d bytes in %.2f seconds\n", bytesRead, elapsed.Seconds())
		return nil
	}
	
	// Print statistics
	printStats(handler, elapsed)
	
	if bytesRead > 0 && elapsed.Seconds() > 0 {
		mbPerSec := float64(bytesRead) / (1024 * 1024) / elapsed.Seconds()
		fmt.Printf("  Data Rate:              %.2f MB/s\n", mbPerSec)
	}
	
	// Run benchmark if requested
	if opts.benchmark {
		divider := "--------------------------------------------------------------------------------"
		fmt.Println("\n🏃 Running Benchmark...")
		fmt.Println(divider)
		
		// Run multiple iterations
		iterations := 5
		var totalTime time.Duration
		
		for i := 0; i < iterations; i++ {
			benchHandler := &itch.DefaultHandler{}
			start := time.Now()
			_, err := itch.ParseFile(filename, benchHandler)
			if err != nil {
				return err
			}
			iterTime := time.Since(start)
			totalTime += iterTime
			fmt.Printf("  Iteration %d: %.2f seconds\n", i+1, iterTime.Seconds())
		}
		
		avgTime := totalTime / time.Duration(iterations)
		fmt.Printf("\n  Average Time:           %.2f seconds\n", avgTime.Seconds())
		
		if avgTime.Seconds() > 0 {
			avgMsgPerSec := float64(handler.Stats.TotalMessages) / avgTime.Seconds()
			avgMBPerSec := float64(bytesRead) / (1024 * 1024) / avgTime.Seconds()
			fmt.Printf("  Avg Throughput:         %.2f messages/second\n", avgMsgPerSec)
			fmt.Printf("                          %.2f MB/s\n", avgMBPerSec)
		}
	}
	
	return nil
}

func main() {
	// Define flags
	var opts options
	flag.BoolVar(&opts.verbose, "verbose", false, "Print detailed message information")
	flag.BoolVar(&opts.benchmark, "benchmark", false, "Run benchmark test (5 iterations)")
	flag.BoolVar(&opts.validate, "validate", false, "Validate file format only")
	
	// Parse flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <itch-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "ITCH Analyzer - Analyze NASDAQ ITCH data files\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s sample.itch                    # Basic analysis\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --verbose sample.itch          # Show detailed messages\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --benchmark sample.itch        # Run performance benchmark\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --validate sample.itch         # Validate file format\n", os.Args[0])
	}
	
	flag.Parse()
	
	// Check arguments
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	
	filename := flag.Arg(0)
	
	// Analyze file
	if err := analyzeFile(filename, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
