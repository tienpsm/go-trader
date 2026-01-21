# go-trader

[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/tienpsm/go-trader.svg)](https://pkg.go.dev/github.com/tienpsm/go-trader)

Go Trader is a high-performance trading platform library written in Go. It is a port of the [CppTrader](https://github.com/chronoxor/CppTrader) C++ library.

## Features

- **Ultra-fast Matching Engine** - High-performance order matching with price-time priority
- **Order Book Processor** - Efficient order book management with AVL tree-based price levels
- **NASDAQ ITCH Handler** - Full ITCH 5.0 protocol parser for market data
- **Multiple Order Types** - Support for Market, Limit, Stop, Stop-Limit, Trailing Stop orders
- **Time-in-Force Options** - GTC, IOC, FOK, AON order types
- **Iceberg/Hidden Orders** - Support for hidden and iceberg order functionality
- **Event-Driven Architecture** - Custom handlers for all market events

## Installation

```bash
go get github.com/tienpsm/go-trader
```

## Quick Start

### Basic Matching Engine Usage

```go
package main

import (
    "fmt"
    "github.com/tienpsm/go-trader/matching"
)

func main() {
    // Create a market manager
    manager := matching.NewMarketManager()
    manager.EnableMatching()

    // Add a symbol and order book
    symbol := matching.NewSymbol(1, "AAPL")
    manager.AddSymbol(symbol)
    manager.AddOrderBook(symbol)

    // Add a sell order
    sellOrder := matching.Order{
        ID:                 1,
        SymbolID:           1,
        Type:               matching.OrderTypeLimit,
        Side:               matching.OrderSideSell,
        Price:              15000, // $150.00 (prices in cents)
        Quantity:           100,
        LeavesQuantity:     100,
        MaxVisibleQuantity: matching.MaxVisibleQuantity,
        Slippage:           matching.MaxSlippage,
    }
    manager.AddOrder(sellOrder)

    // Add a matching buy order
    buyOrder := matching.Order{
        ID:                 2,
        SymbolID:           1,
        Type:               matching.OrderTypeLimit,
        Side:               matching.OrderSideBuy,
        Price:              15000,
        Quantity:           50,
        LeavesQuantity:     50,
        MaxVisibleQuantity: matching.MaxVisibleQuantity,
        Slippage:           matching.MaxSlippage,
    }
    manager.AddOrder(buyOrder)

    // Check order status
    if remaining := manager.GetOrder(1); remaining != nil {
        fmt.Printf("Sell order remaining: %d shares\n", remaining.LeavesQuantity)
    }
    if manager.GetOrder(2) == nil {
        fmt.Println("Buy order fully executed")
    }
}
```

### Custom Market Handler

```go
package main

import (
    "fmt"
    "github.com/tienpsm/go-trader/matching"
)

// Custom handler to track executions
type MyHandler struct {
    matching.DefaultMarketHandler
}

func (h *MyHandler) OnExecuteOrder(order matching.Order, price, quantity uint64) {
    fmt.Printf("EXECUTION: Order %d executed %d shares @ $%.2f\n",
        order.ID, quantity, float64(price)/100)
}

func (h *MyHandler) OnAddOrder(order matching.Order) {
    fmt.Printf("ORDER ADDED: %s %d shares @ $%.2f\n",
        order.Side, order.Quantity, float64(order.Price)/100)
}

func main() {
    handler := &MyHandler{}
    manager := matching.NewMarketManagerWithHandler(handler)
    manager.EnableMatching()
    // ... add symbols, order books, and orders
}
```

### ITCH Protocol Parser

```go
package main

import (
    "fmt"
    "github.com/tienpsm/go-trader/itch"
)

// Custom ITCH handler
type MyITCHHandler struct {
    itch.DefaultHandler
    orderCount int
}

func (h *MyITCHHandler) OnAddOrder(msg itch.AddOrderMessage) error {
    h.orderCount++
    side := "BUY"
    if msg.BuySellIndicator == 'S' {
        side = "SELL"
    }
    fmt.Printf("Add Order: %s %d shares @ %d\n", side, msg.Shares, msg.Price)
    return nil
}

func (h *MyITCHHandler) OnOrderExecuted(msg itch.OrderExecutedMessage) error {
    fmt.Printf("Order %d executed %d shares\n", msg.OrderReferenceNumber, msg.ExecutedShares)
    return nil
}

func main() {
    handler := &MyITCHHandler{}
    parser := itch.NewParser(handler)

    // Parse ITCH data from file or network
    data := []byte{...} // ITCH message bytes
    consumed, messageCount, err := parser.ParseAll(data)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Parsed %d messages (%d bytes)\n", messageCount, consumed)
}
```

## Package Structure

```
go-trader/
├── matching/           # Order matching engine
│   ├── order.go       # Order types and structures
│   ├── level.go       # Price level management
│   ├── orderbook.go   # Order book implementation
│   ├── market_manager.go  # Main matching engine
│   ├── handler.go     # Market event handler interface
│   ├── avltree.go     # AVL tree for price levels
│   ├── symbol.go      # Trading symbol
│   ├── errors.go      # Error codes
│   └── update.go      # Update types
├── itch/              # NASDAQ ITCH protocol handler
│   └── handler.go     # ITCH message parser
└── README.md
```

## Order Types

| Type | Description |
|------|-------------|
| `OrderTypeMarket` | Execute immediately at best available price |
| `OrderTypeLimit` | Execute at specified price or better |
| `OrderTypeStop` | Becomes market order when stop price is reached |
| `OrderTypeStopLimit` | Becomes limit order when stop price is reached |
| `OrderTypeTrailingStop` | Stop order with dynamic stop price |
| `OrderTypeTrailingStopLimit` | Trailing stop that becomes limit order |

## Time-in-Force Options

| Type | Description |
|------|-------------|
| `OrderTimeInForceGTC` | Good-Till-Cancelled |
| `OrderTimeInForceIOC` | Immediate-Or-Cancel |
| `OrderTimeInForceFOK` | Fill-Or-Kill |
| `OrderTimeInForceAON` | All-Or-None |

## ITCH Message Types Supported

- System Event
- Stock Directory
- Stock Trading Action
- Reg SHO Restriction
- Market Participant Position
- MWCB Decline Level
- MWCB Status
- IPO Quoting Period Update
- Add Order (with and without MPID)
- Order Executed (with and without price)
- Order Cancel
- Order Delete
- Order Replace
- Trade
- Cross Trade
- Broken Trade
- NOII (Net Order Imbalance Indicator)
- RPII (Retail Price Improvement Indicator)

## Performance

The matching engine is designed for high-performance trading applications:

- O(log n) order insertion and deletion using AVL trees
- O(1) access to best bid/ask prices
- Efficient price-time priority matching
- Memory-efficient order storage

## Testing

```bash
go test ./... -v
```

## Benchmarking

```bash
go test ./... -bench=.
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original C++ implementation: [CppTrader](https://github.com/chronoxor/CppTrader) by Ivan Shynkarenka
- NASDAQ ITCH protocol specification