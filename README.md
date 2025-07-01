# TinkoffGo

A comprehensive Go library for the Tinkoff Invest API, providing both mock implementations for testing and real API integration for production trading.

## Features

- 🔄 **Real API Integration**: Direct connection to Tinkoff Invest API
- 🧪 **Mock Implementation**: Perfect for testing and development
- 🛡️ **Type Safety**: Full type safety with generated protobuf types
- ⚡ **High Performance**: Native Go implementation with goroutines
- 🔐 **Secure**: TLS connections and proper authentication
- 📊 **Comprehensive**: Support for accounts, portfolios, orders, market data, and streaming
- 🎯 **Production Ready**: Error handling, retries, and proper resource management

## Installation

```bash
go get github.com/buurzx/tinkoff-go
```

## Quick Start

### Using Real API

```go
package main

import (
    "context"
    "log"

    "github.com/buurzx/tinkoff-go/client"
)

func main() {
    // Create real client (demo mode for testing)
    client, err := client.NewRealDemo("your_token_here")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Get accounts
    accounts, err := client.GetAccounts(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for _, account := range accounts {
        log.Printf("Account: %s (%s)", account.Name, account.Id)
    }
}
```

### Using Mock Implementation

```go
package main

import (
    "context"
    "log"

    "github.com/buurzx/tinkoff-go/client"
)

func main() {
    // Create mock client for testing
    client := client.NewMock()
    defer client.Close()

    // Works exactly the same as real client
    accounts, err := client.GetAccounts(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for _, account := range accounts {
        log.Printf("Mock Account: %s (%s)", account.Name, account.Id)
    }
}
```

## Getting Your API Token

1. Open [Tinkoff Invest](https://www.tbank.ru/invest/)
2. Go to Settings → API
3. Create a new token with appropriate permissions
4. **Important**: Start with demo mode for testing!

## API Methods

### Account Management
- `GetAccounts()` - Get all user accounts
- `GetUserInfo()` - Get user information and permissions

### Portfolio & Positions
- `GetPortfolio(accountID)` - Get portfolio summary
- `GetPositions(accountID)` - Get detailed positions

### Orders
- `GetOrders(accountID)` - Get active orders
- `PostOrder(request)` - Place new order
- `CancelOrder(accountID, orderID)` - Cancel order

### Market Data
- `GetInstrumentByFIGI(figi)` - Get instrument by FIGI
- `GetInstrumentByTicker(ticker, classCode)` - Get instrument by ticker
- `GetCandles(figi, from, to, interval)` - Get historical candles

## Examples

### Basic Connection Test
```bash
make run-connect
```

### Account Information
```bash
make run-accounts
```

### Market Data Streaming
```bash
make run-streaming
```

### Real API Demo (requires token)
```bash
TINKOFF_TOKEN=your_token make run-real-api
```

## Building

```bash
# Install dependencies
make deps

# Generate proto files (if needed)
make proto

# Build all examples
make examples

# Run tests
make test

# Format code
make fmt
```

## Project Structure

```
tinkoff-go/
├── client/           # Client implementations
│   ├── client.go     # Main client interface
│   ├── mock_client.go # Mock implementation
│   └── real_client.go # Real API implementation
├── config/           # Configuration management
├── types/            # Common types and utilities
├── proto/            # Generated protobuf files
├── examples/         # Example applications
│   ├── connect/      # Basic connection test
│   ├── accounts/     # Account management
│   ├── streaming/    # Market data streaming
│   └── real_api/     # Real API demo
└── internal/         # Internal utilities
```

## Safety & Best Practices

### Demo vs Production

**Always start with demo mode:**
```go
// Demo mode (safe for testing)
client, err := client.NewRealDemo(token)

// Production mode (real money!)
client, err := client.NewReal(token)
```

### Error Handling
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

accounts, err := client.GetAccounts(ctx)
if err != nil {
    log.Printf("Failed to get accounts: %v", err)
    return
}
```

### Resource Management
```go
client, err := client.NewReal(token)
if err != nil {
    log.Fatal(err)
}
defer client.Close() // Always close the client
```

## Advanced Usage

### Custom Configuration
```go
cfg, err := config.New(token, false) // false = production mode
if err != nil {
    log.Fatal(err)
}

client, err := client.NewRealWithConfig(cfg)
```

### Order Placement
```go
import investapi "github.com/buurzx/tinkoff-go/proto"

// Create order request
orderReq := &investapi.PostOrderRequest{
    Figi:      "BBG004730N88", // Sber FIGI
    Quantity:  1,
    Price:     &investapi.Quotation{Units: 250, Nano: 0}, // 250.00 RUB
    Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY,
    AccountId: accountID,
    OrderType: investapi.OrderType_ORDER_TYPE_LIMIT,
    OrderId:   "unique_order_id",
}

response, err := client.PostOrder(ctx, orderReq)
```

### Historical Data
```go
// Get daily candles for the last month
to := time.Now()
from := to.Add(-30 * 24 * time.Hour)

candles, err := client.GetCandles(ctx,
    "BBG004730N88", // Sber FIGI
    from, to,
    investapi.CandleInterval_CANDLE_INTERVAL_DAY)
```

## Testing

The library includes comprehensive tests and mock implementations:

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./client
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This library is for educational and development purposes. When using with real money:

- Always test in demo mode first
- Start with small amounts
- Understand the risks of algorithmic trading
- Ensure compliance with local regulations
- The authors are not responsible for any financial losses

## Support

- 📚 [Tinkoff Invest API Documentation](https://tinkoff.github.io/investAPI/)
- 🐛 [Report Issues](https://github.com/buurzx/tinkoff-go/issues)
- 💬 [Discussions](https://github.com/buurzx/tinkoff-go/discussions)

---

**⚠️ Warning**: This library allows real trading operations. Always use demo mode for testing and understand the risks before trading with real money.
