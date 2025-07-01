# 🚀 TinkoffGo - Comprehensive Go Client for Tinkoff Invest API

A powerful, production-ready Go library for the Tinkoff Invest API with full real-time streaming, advanced order management, and comprehensive market data access.

[![Go Version](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![API Version](https://img.shields.io/badge/Tinkoff%20API-v2-orange.svg)](https://tinkoff.github.io/investAPI/)

## ✨ Key Features

- 🔄 **Real API Integration**: Direct connection to Tinkoff Invest API with gRPC
- 📡 **Real-Time Streaming**: Live market data, order updates, and portfolio changes
- 🛡️ **Type Safety**: Full type safety with generated protobuf types
- ⚡ **High Performance**: Native Go implementation with goroutines and channels
- 🔐 **Secure**: TLS connections, proper authentication, and demo mode
- 🧪 **Mock Implementation**: Perfect for testing and development
- 📊 **Comprehensive Coverage**: All API endpoints including advanced orders
- 🎯 **Production Ready**: Error handling, retries, reconnection logic
- 📈 **Advanced Orders**: Stop orders, conditional orders, order replacement
- 💰 **Portfolio Management**: Real-time positions, P&L tracking, risk management

## 🚀 Quick Start

### Installation

```bash
go get github.com/buurzx/tinkoff-go
```

### Basic Usage - Real API

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/buurzx/tinkoff-go/client"
)

func main() {
    // Get your token from https://www.tbank.ru/invest/
    token := os.Getenv("TINKOFF_TOKEN")

    // Create real client in demo mode (safe for testing)
    client, err := client.NewRealDemo(token)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Get accounts
    accounts, err := client.GetAccounts(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, account := range accounts {
        log.Printf("Account: %s (%s)", account.Name, account.Id)

        // Get portfolio
        portfolio, err := client.GetPortfolio(ctx, account.Id)
        if err != nil {
            log.Printf("Failed to get portfolio: %v", err)
            continue
        }

        log.Printf("Total value: %.2f %s",
            float64(portfolio.TotalAmountShares.Units),
            portfolio.TotalAmountShares.Currency)
    }
}
```

### Real-Time Market Data Streaming

```go
// Start real-time streaming
stream, err := client.StartMarketDataStream()
if err != nil {
    log.Fatal(err)
}

// Subscribe to real-time data
instruments := []string{"BBG004730N88"} // SBER
client.SubscribeCandles(stream, instruments,
    investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE, false)
client.SubscribeTrades(stream, instruments)
client.SubscribeOrderBook(stream, instruments, 10)

// Process real-time updates
for {
    resp, err := stream.Recv()
    if err != nil {
        break
    }
    // Handle real-time market data...
}
```

## 📡 Real-Time Streaming Capabilities

### Market Data Streaming
- **📊 Real-time Candles**: All intervals from 1-minute to daily
- **💰 Live Trades**: Every trade with price, volume, direction
- **📖 Order Book**: Bid/ask spreads with configurable depth (1-50 levels)
- **💲 Last Prices**: Real-time price updates
- **📈 Trading Status**: Market status changes

### Order & Portfolio Streaming
- **📋 Order Updates**: Real-time order execution status
- **🔔 Trade Notifications**: Instant fill notifications
- **📊 Portfolio Changes**: Live P&L and position updates

## 🛠️ Complete API Coverage

### Account Management
- `GetAccounts()` - Get all user accounts
- `GetUserInfo()` - User information and permissions

### Portfolio & Positions
- `GetPortfolio(accountID)` - Portfolio summary with P&L
- `GetPositions(accountID)` - Detailed positions and metrics

### Order Management
- `GetOrders(accountID)` - Active orders
- `PostOrder(request)` - Place market/limit orders
- `CancelOrder(accountID, orderID)` - Cancel orders
- `ReplaceOrder(...)` - Replace existing orders

### Advanced Orders
- `PostStopOrder(request)` - Place stop-loss/take-profit orders
- `GetStopOrders(accountID)` - Get stop orders
- `CancelStopOrder(accountID, stopOrderID)` - Cancel stop orders

### Market Data
- `GetInstrumentByFIGI(figi)` - Instrument details by FIGI
- `GetInstrumentByTicker(ticker, classCode)` - Find by ticker
- `GetCandles(figi, from, to, interval)` - Historical candles
- `GetOrderPrice(...)` - Calculate order execution price
- `GetMaxLots(...)` - Maximum available lots for trading

### Real-Time Streaming
- `StartMarketDataStream()` - Market data streaming
- `StartOrderStream(accountIDs)` - Order state streaming
- `SubscribeCandles()` - Real-time candles
- `SubscribeTrades()` - Live trades
- `SubscribeOrderBook()` - Order book updates
- `SubscribeLastPrices()` - Price updates

## 📚 Examples & Guides

### Available Examples

```bash
# Basic connection and API test
make run-connect

# Account and portfolio management
make run-accounts

# Mock streaming (for development)
make run-streaming

# Real API demonstration
TINKOFF_TOKEN=your_token make run-real-api

# Real-time market data streaming
TINKOFF_TOKEN=your_token make run-real-streaming

# Advanced orders (stop-loss, take-profit)
TINKOFF_TOKEN=your_token make run-advanced-orders
```

### Comprehensive Guides

- 📡 **[Streaming Guide](STREAMING_GUIDE.md)** - Complete real-time streaming tutorial
- 📖 **[Real API Guide](REAL_API_GUIDE.md)** - Production API usage patterns

## 🔧 Development & Building

### Prerequisites

```bash
# Install protobuf compiler
# macOS: brew install protobuf
# Ubuntu: apt-get install protobuf-compiler

# Set up development environment
make dev-setup
```

### Build Commands

```bash
# Install dependencies and generate proto files
make deps proto

# Build all examples
make examples

# Run tests with coverage
make test

# Format and lint code
make fmt vet lint

# Build for multiple platforms
make release
```

### Proto Files Management

```bash
# Update proto files from Tinkoff repository
make proto-update

# Regenerate Go code from proto files
make proto

# Clean generated files
make proto-clean
```

## 🏗️ Project Structure

```
tinkoff-go/
├── client/                 # Client implementations
│   ├── client.go          # Main client interface
│   ├── real_client.go     # Real API implementation
│   └── client_test.go     # Comprehensive tests
├── config/                # Configuration management
│   └── config.go          # API endpoints and settings
├── types/                 # Common types and utilities
│   ├── common.go          # Type definitions
│   └── common_test.go     # Type tests
├── proto/                 # Generated protobuf files
│   ├── *.proto           # Official Tinkoff API definitions
│   └── *.pb.go           # Generated Go code
├── examples/              # Example applications
│   ├── connect/          # Basic connection test
│   ├── accounts/         # Account management
│   ├── streaming/        # Mock streaming demo
│   ├── real_api/         # Real API demo
│   ├── real_streaming/   # Real-time streaming
│   └── advanced_orders/  # Advanced order management
├── internal/             # Internal utilities
├── STREAMING_GUIDE.md    # Comprehensive streaming guide
├── REAL_API_GUIDE.md     # Real API usage guide
└── Makefile              # Build automation
```

## 🛡️ Safety & Best Practices

### Demo vs Production Mode

**⚠️ Always start with demo mode for testing:**

```go
// Demo mode - safe for testing, no real money
client, err := client.NewRealDemo(token)

// Production mode - real trading with real money!
client, err := client.NewReal(token)
```

### Proper Error Handling

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

accounts, err := client.GetAccounts(ctx)
if err != nil {
    log.Printf("API call failed: %v", err)
    // Handle error appropriately
    return
}
```

### Resource Management

```go
client, err := client.NewRealDemo(token)
if err != nil {
    log.Fatal(err)
}
defer client.Close() // Always close the client

// For streaming
stream, err := client.StartMarketDataStream()
if err != nil {
    log.Fatal(err)
}
defer stream.CloseSend() // Close streams properly
```

## 💼 Advanced Usage Examples

### Order Placement with Error Handling

```go
import investapi "github.com/buurzx/tinkoff-go/proto"

orderReq := &investapi.PostOrderRequest{
    Figi:      "BBG004730N88", // Sberbank FIGI
    Quantity:  1,
    Price:     &investapi.Quotation{Units: 250, Nano: 0}, // 250.00 RUB
    Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY,
    AccountId: accountID,
    OrderType: investapi.OrderType_ORDER_TYPE_LIMIT,
    OrderId:   uuid.New().String(), // Unique order ID
}

response, err := client.PostOrder(ctx, orderReq)
if err != nil {
    log.Printf("Order failed: %v", err)
    return
}

log.Printf("Order placed: %s", response.OrderId)
```

### Stop-Loss Order

```go
stopOrderReq := &investapi.PostStopOrderRequest{
    Figi:           "BBG004730N88",
    Quantity:       1,
    Price:          &investapi.Quotation{Units: 240, Nano: 0}, // Stop price
    StopPrice:      &investapi.Quotation{Units: 245, Nano: 0}, // Trigger price
    Direction:      investapi.StopOrderDirection_STOP_ORDER_DIRECTION_SELL,
    AccountId:      accountID,
    ExpirationType: investapi.StopOrderExpirationType_STOP_ORDER_EXPIRATION_TYPE_GOOD_TILL_CANCEL,
    StopOrderType:  investapi.StopOrderType_STOP_ORDER_TYPE_STOP_LOSS,
    ExpireDate:     timestamppb.New(time.Now().Add(24 * time.Hour)),
}

response, err := client.PostStopOrder(ctx, stopOrderReq)
```

### Portfolio Monitoring with Real-Time Updates

```go
// Get current positions
positions, err := client.GetPositions(ctx, accountID)
if err != nil {
    log.Fatal(err)
}

// Extract FIGIs for streaming
var figis []string
for _, position := range positions.Securities {
    figis = append(figis, position.Figi)
}

// Start streaming for portfolio instruments
stream, err := client.StartMarketDataStream()
if err != nil {
    log.Fatal(err)
}

// Subscribe to real-time updates
client.SubscribeLastPrices(stream, figis)
client.SubscribeTrades(stream, figis)

// Monitor portfolio in real-time
for {
    resp, err := stream.Recv()
    if err != nil {
        log.Printf("Stream error: %v", err)
        break
    }
    // Process portfolio updates...
}
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v ./client
go test -v ./types

# Benchmark tests
go test -bench=. ./...
```

### Using Mock Client for Testing

```go
func TestTradingStrategy(t *testing.T) {
    // Use mock client for unit tests
    client := client.NewMock()
    defer client.Close()

    // Test your trading logic with predictable mock data
    accounts, err := client.GetAccounts(context.Background())
    assert.NoError(t, err)
    assert.Len(t, accounts, 1) // Mock returns 1 account
}
```

## 🔗 Getting Your API Token

1. Open [Tinkoff Invest](https://www.tbank.ru/invest/)
2. Go to Settings → Data for developers → API
3. Create a new token with appropriate permissions
4. **Important**: Start with demo/sandbox mode for testing!
5. Set environment variable: `export TINKOFF_TOKEN="your_token_here"`

## 📈 Performance & Production

### Optimizations Implemented

- **Connection Pooling**: Efficient gRPC connection management
- **Concurrent Processing**: Goroutines for parallel API calls
- **Smart Retries**: Exponential backoff for failed requests
- **Memory Efficient**: Streaming data processing without buffering
- **Type Safety**: Compile-time safety with generated protobuf types

### Production Checklist

- ✅ Use proper error handling and timeouts
- ✅ Implement reconnection logic for streams
- ✅ Monitor API rate limits and quotas
- ✅ Log important events (but never tokens!)
- ✅ Use context cancellation for graceful shutdowns
- ✅ Test thoroughly in demo mode first
- ✅ Implement proper security practices

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Ensure all tests pass (`make test`)
5. Format code (`make fmt`)
6. Submit a pull request

### Development Workflow

```bash
# Set up development environment
make dev-setup

# Make changes and test
make fmt vet test

# Update proto files if needed
make proto-update proto

# Build and test examples
make examples
make run-connect
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ⚠️ Important Disclaimers

- **Educational Purpose**: This library is primarily for educational and development purposes
- **Financial Risk**: Trading involves significant financial risk - never trade with money you can't afford to lose
- **Demo First**: Always test thoroughly in demo mode before using real money
- **No Guarantees**: The authors provide no guarantees about the library's performance or reliability
- **Compliance**: Ensure compliance with your local financial regulations
- **Not Financial Advice**: This library does not provide financial advice

## 🆘 Support & Resources

- 📚 **[Official Tinkoff Invest API Docs](https://tinkoff.github.io/investAPI/)**
- 📡 **[Streaming Guide](STREAMING_GUIDE.md)** - Comprehensive streaming tutorial
- 📖 **[Real API Guide](REAL_API_GUIDE.md)** - Production usage patterns
- 🐛 **[Report Issues](https://github.com/buurzx/tinkoff-go/issues)**
- 💬 **[Discussions](https://github.com/buurzx/tinkoff-go/discussions)**

## 🌟 Features Roadmap

- [ ] WebSocket streaming support
- [ ] Advanced analytics and indicators
- [ ] Backtesting framework integration
- [ ] Options and futures support
- [ ] Multi-account management
- [ ] Risk management tools
- [ ] Performance metrics and reporting

---

**⚠️ Final Warning**: This library enables real trading operations with real money. Always use demo mode for testing and fully understand the risks before live trading. The authors are not responsible for any financial losses.

**🚀 Happy Trading!** - Built with ❤️ for the Go and algorithmic trading community.
