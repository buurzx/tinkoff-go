# 🚀 Tinkoff Go Streaming & Advanced Orders Guide

This guide demonstrates the real-time streaming and advanced order capabilities implemented in the Tinkoff Go client.

## 📡 Real-Time Streaming

### Features Implemented

#### ✅ Market Data Streaming
- **Real-time Candles**: 1-minute, 5-minute, hourly, daily candles
- **Live Trades**: Every trade with price, volume, direction
- **Order Book Updates**: Bid/ask spreads with configurable depth (1-50 levels)
- **Last Price Updates**: Real-time price updates
- **Trading Status**: Instrument trading status changes

#### ✅ Order State Streaming
- **Order Updates**: Real-time order execution status
- **Trade Notifications**: When your orders are filled
- **Order Lifecycle**: New → Partial → Filled/Cancelled

### Quick Start - Streaming

```go
// Create streaming client
client, err := client.NewRealDemo(token)
if err != nil {
    log.Fatal(err)
}

// Start market data stream
stream, err := client.StartMarketDataStream()
if err != nil {
    log.Fatal(err)
}

// Subscribe to different data types
instruments := []string{"BBG004730N88"} // SBER

// Subscribe to 1-minute candles
client.SubscribeCandles(stream, instruments,
    investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE, false)

// Subscribe to order book (depth 10)
client.SubscribeOrderBook(stream, instruments, 10)

// Subscribe to trades
client.SubscribeTrades(stream, instruments)

// Subscribe to last prices
client.SubscribeLastPrices(stream, instruments)

// Handle incoming data
for {
    resp, err := stream.Recv()
    if err != nil {
        break
    }
    // Process real-time data
    processMarketData(resp)
}
```

### Example Output

->

# 🚀 Tinkoff Go Real-Time Streaming Guide

This comprehensive guide demonstrates the real-time streaming and advanced order capabilities implemented in the Tinkoff Go client.

## 📡 Real-Time Market Data Streaming

### Overview

The Tinkoff Go client provides full real-time market data streaming capabilities through gRPC bidirectional streams. You can subscribe to multiple data types simultaneously and receive live updates as they happen on the exchange.

### Supported Data Types

#### ✅ Market Data Streaming
- **📊 Real-time Candles**: 1-minute, 5-minute, 15-minute, hourly, daily candles
- **💰 Live Trades**: Every trade with price, volume, direction, and timing
- **📖 Order Book Updates**: Bid/ask spreads with configurable depth (1-50 levels)
- **💲 Last Price Updates**: Real-time price updates for instruments
- **📈 Trading Status**: Instrument trading status changes

#### ✅ Order State Streaming
- **📋 Order Updates**: Real-time order execution status changes
- **🔔 Trade Notifications**: Instant notifications when your orders are filled
- **📊 Order Lifecycle**: Complete tracking from New → Partial → Filled/Cancelled

### Quick Start Example

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/buurzx/tinkoff-go/client"
    investapi "github.com/buurzx/tinkoff-go/proto"
)

func main() {
    // Get token from environment
    token := os.Getenv("TINKOFF_TOKEN")
    if token == "" {
        log.Fatal("TINKOFF_TOKEN environment variable is required")
    }

    // Create real client (demo mode for testing)
    realClient, err := client.NewRealDemo(token)
    if err != nil {
        log.Fatalf("Failed to create real client: %v", err)
    }
    defer realClient.Close()

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Define instruments to monitor (FIGI codes)
    instruments := []string{
        "BBG004730N88", // SBER - Sberbank
        "BBG004730ZJ9", // GAZP - Gazprom
        "BBG006L8G4H1", // YNDX - Yandex
    }

    // Start market data streaming
    marketDataStream, err := realClient.StartMarketDataStream()
    if err != nil {
        log.Fatalf("Failed to start market data stream: %v", err)
    }

    // Subscribe to different data types
    err = realClient.SubscribeCandles(
        marketDataStream,
        instruments,
        investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE,
        false, // waitingClose - false for real-time updates
    )
    if err != nil {
        log.Printf("Failed to subscribe to candles: %v", err)
    }

    err = realClient.SubscribeOrderBook(marketDataStream, instruments, 10) // depth = 10
    if err != nil {
        log.Printf("Failed to subscribe to order books: %v", err)
    }

    err = realClient.SubscribeTrades(marketDataStream, instruments)
    if err != nil {
        log.Printf("Failed to subscribe to trades: %v", err)
    }

    err = realClient.SubscribeLastPrices(marketDataStream, instruments)
    if err != nil {
        log.Printf("Failed to subscribe to last prices: %v", err)
    }

    // Handle incoming data
    go handleMarketDataStream(marketDataStream)

    // Keep running
    select {
    case <-ctx.Done():
        log.Println("Context timeout reached")
    }
}

func handleMarketDataStream(stream investapi.MarketDataStreamService_MarketDataStreamClient) {
    for {
        resp, err := stream.Recv()
        if err != nil {
            log.Printf("Stream error: %v", err)
            return
        }
        processMarketDataResponse(resp)
    }
}
```

## 📊 Market Data Processing

### Processing Different Data Types

```go
func processMarketDataResponse(resp *investapi.MarketDataResponse) {
    switch payload := resp.Payload.(type) {

    case *investapi.MarketDataResponse_Candle:
        // Handle real-time candle updates
        candle := payload.Candle
        log.Printf("📊 CANDLE %s [%s]: O=%.4f H=%.4f L=%.4f C=%.4f V=%d",
            getInstrumentName(candle.Figi),
            candle.Time.AsTime().Format("15:04:05"),
            quotationToFloat(candle.Open),
            quotationToFloat(candle.High),
            quotationToFloat(candle.Low),
            quotationToFloat(candle.Close),
            candle.Volume)

    case *investapi.MarketDataResponse_Trade:
        // Handle individual trade updates
        trade := payload.Trade
        direction := "🟢 BUY "
        if trade.Direction == investapi.TradeDirection_TRADE_DIRECTION_SELL {
            direction = "🔴 SELL"
        }

        size := "small"
        if trade.Quantity >= 100 {
            size = "medium"
        }
        if trade.Quantity >= 1000 {
            size = "large"
        }

        log.Printf("💰 TRADE %s [%s]: %s %.4f x%d (%s)",
            getInstrumentName(trade.Figi),
            trade.Time.AsTime().Format("15:04:05"),
            direction,
            quotationToFloat(trade.Price),
            trade.Quantity,
            size)

    case *investapi.MarketDataResponse_Orderbook:
        // Handle order book updates
        orderBook := payload.Orderbook
        bestBid := 0.0
        bestAsk := 0.0

        if len(orderBook.Bids) > 0 {
            bestBid = quotationToFloat(orderBook.Bids[0].Price)
        }
        if len(orderBook.Asks) > 0 {
            bestAsk = quotationToFloat(orderBook.Asks[0].Price)
        }

        spread := bestAsk - bestBid
        spreadPercent := (spread / bestBid) * 100

        log.Printf("📖 ORDER_BOOK %s: Bid=%.4f Ask=%.4f Spread=%.4f (%.3f%%) Depth=%d/%d",
            getInstrumentName(orderBook.Figi),
            bestBid,
            bestAsk,
            spread,
            spreadPercent,
            len(orderBook.Bids),
            len(orderBook.Asks))

    case *investapi.MarketDataResponse_LastPrice:
        // Handle last price updates
        lastPrice := payload.LastPrice
        log.Printf("💲 LAST_PRICE %s [%s]: %.4f",
            getInstrumentName(lastPrice.Figi),
            lastPrice.Time.AsTime().Format("15:04:05"),
            quotationToFloat(lastPrice.Price))

    case *investapi.MarketDataResponse_TradingStatus:
        // Handle trading status changes
        status := payload.TradingStatus
        log.Printf("📈 TRADING_STATUS %s: %s",
            getInstrumentName(status.Figi),
            status.TradingStatus.String())
    }
}
```

## 📋 Order State Streaming

### Setting Up Order Streaming

```go
// Get accounts first
accounts, err := realClient.GetAccounts(ctx)
if err != nil {
    log.Fatalf("Failed to get accounts: %v", err)
}

selectedAccount := accounts[0]

// Start order streaming for specific accounts
orderStream, err := realClient.StartOrderStream([]string{selectedAccount.Id})
if err != nil {
    log.Printf("Failed to start order stream: %v", err)
    return
}

// Handle order updates
go handleOrderStream(orderStream)
```

### Processing Order Updates

```go
func handleOrderStream(stream investapi.OrdersStreamService_OrderStateStreamClient) {
    for {
        resp, err := stream.Recv()
        if err != nil {
            log.Printf("Order stream error: %v", err)
            return
        }
        processOrderStreamResponse(resp)
    }
}

func processOrderStreamResponse(resp *investapi.OrderStateStreamResponse) {
    switch payload := resp.Payload.(type) {

    case *investapi.OrderStateStreamResponse_OrderState:
        // Handle order state changes
        order := payload.OrderState

        status := "UNKNOWN"
        switch order.ExecutionReportStatus {
        case investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW:
            status = "🆕 NEW"
        case investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL:
            status = "✅ FILLED"
        case investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_PARTIALLYFILL:
            status = "🔄 PARTIAL"
        case investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_CANCELLED:
            status = "❌ CANCELLED"
        case investapi.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_REJECTED:
            status = "🚫 REJECTED"
        }

        direction := "BUY"
        if order.Direction == investapi.OrderDirection_ORDER_DIRECTION_SELL {
            direction = "SELL"
        }

        log.Printf("📋 ORDER %s: %s %s %d@%.4f -> %s (Filled: %d)",
            order.OrderId,
            direction,
            getInstrumentName(order.Figi),
            order.LotsRequested,
            quotationToFloat(order.InitialOrderPrice),
            status,
            order.LotsExecuted)

    case *investapi.OrderStateStreamResponse_Ping:
        // Handle ping messages (keep-alive)
        log.Println("📋 Order stream ping received")
    }
}
```

## 🔧 Utility Functions

### Data Conversion Helpers

```go
// Convert Quotation to float64
func quotationToFloat(q *investapi.Quotation) float64 {
    if q == nil {
        return 0
    }
    return float64(q.Units) + float64(q.Nano)/1_000_000_000
}

// Convert MoneyValue to float64
func moneyValueToFloat(m *investapi.MoneyValue) float64 {
    if m == nil {
        return 0
    }
    return float64(m.Units) + float64(m.Nano)/1_000_000_000
}

// Get instrument name by FIGI (you can extend this with a cache)
func getInstrumentName(figi string) string {
    // This is a simple mapping - in production, you'd want to cache instrument data
    instrumentNames := map[string]string{
        "BBG004730N88": "SBER",
        "BBG004730ZJ9": "GAZP",
        "BBG006L8G4H1": "YNDX",
        "BBG004731354": "ROSN",
        "BBG004730RP0": "TATN",
    }

    if name, exists := instrumentNames[figi]; exists {
        return name
    }
    return figi[:8] // Return first 8 chars if not found
}
```

## 🚀 Advanced Streaming Patterns

### Multi-Instrument Portfolio Monitoring

```go
func monitorPortfolio(client *client.RealClient, accountID string) {
    ctx := context.Background()

    // Get current positions
    positions, err := client.GetPositions(ctx, accountID)
    if err != nil {
        log.Printf("Failed to get positions: %v", err)
        return
    }

    // Extract FIGIs from positions
    var figis []string
    for _, position := range positions.Securities {
        figis = append(figis, position.Figi)
    }

    if len(figis) == 0 {
        log.Println("No positions found")
        return
    }

    // Start streaming for all position instruments
    stream, err := client.StartMarketDataStream()
    if err != nil {
        log.Printf("Failed to start stream: %v", err)
        return
    }

    // Subscribe to all data types for portfolio instruments
    client.SubscribeCandles(stream, figis, investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE, false)
    client.SubscribeLastPrices(stream, figis)
    client.SubscribeTrades(stream, figis)

    log.Printf("📊 Monitoring %d portfolio instruments", len(figis))

    // Process updates
    for {
        resp, err := stream.Recv()
        if err != nil {
            log.Printf("Stream error: %v", err)
            return
        }
        processMarketDataResponse(resp)
    }
}
```

### High-Frequency Trade Monitoring

```go
func monitorHighFrequencyTrades(client *client.RealClient, instruments []string) {
    stream, err := client.StartMarketDataStream()
    if err != nil {
        log.Printf("Failed to start stream: %v", err)
        return
    }

    // Subscribe only to trades for minimal latency
    err = client.SubscribeTrades(stream, instruments)
    if err != nil {
        log.Printf("Failed to subscribe to trades: %v", err)
        return
    }

    tradeCount := make(map[string]int)
    volumeSum := make(map[string]int64)

    for {
        resp, err := stream.Recv()
        if err != nil {
            log.Printf("Stream error: %v", err)
            return
        }

        if trade, ok := resp.Payload.(*investapi.MarketDataResponse_Trade); ok {
            figi := trade.Trade.Figi
            tradeCount[figi]++
            volumeSum[figi] += trade.Trade.Quantity

            // Log high-volume trades
            if trade.Trade.Quantity >= 1000 {
                log.Printf("🔥 HIGH VOLUME TRADE %s: %.4f x%d (Total: %d trades, %d volume)",
                    getInstrumentName(figi),
                    quotationToFloat(trade.Trade.Price),
                    trade.Trade.Quantity,
                    tradeCount[figi],
                    volumeSum[figi])
            }
        }
    }
}
```

## 📋 Complete Working Example

See `examples/real_streaming/main.go` for a complete working example that demonstrates:

- ✅ Multi-instrument streaming setup
- ✅ All data type subscriptions (candles, trades, order books, last prices)
- ✅ Order state streaming
- ✅ Proper error handling and graceful shutdown
- ✅ Signal handling (Ctrl+C)
- ✅ Concurrent processing with goroutines
- ✅ Data formatting and logging

### Running the Example

```bash
# Set your token (get from Tinkoff Invest account)
export TINKOFF_TOKEN="your_token_here"

# Run the streaming example
make run-real-streaming

# Or run directly
go run examples/real_streaming/main.go
```

### Expected Output

```
🚀 Tinkoff Go Real-Time Streaming Demo
=====================================
Using account: Тинькофф Брокерский счет (account_id_here)

📡 Starting market data streaming...
📊 Subscribing to candles...
📊 Subscribed to candles for 3 instruments
📖 Subscribing to order books...
📖 Subscribed to order book for 3 instruments
💰 Subscribing to trades...
💰 Subscribed to trades for 3 instruments
💲 Subscribing to last prices...
💲 Subscribed to last prices for 3 instruments

📋 Starting order streaming...
🚀 Order stream started for 1 accounts

✅ All streams started successfully!
📊 Monitoring real-time data...
Press Ctrl+C to stop...

📊 CANDLE SBER [14:32:00]: O=285.5000 H=285.7000 L=285.4000 C=285.6000 V=1250
💰 TRADE GAZP [14:32:15]: 🟢 BUY  187.2400 x150 (medium)
📖 ORDER_BOOK YNDX: Bid=2847.0000 Ask=2848.0000 Spread=1.0000 (0.035%) Depth=10/10
💲 LAST_PRICE SBER [14:32:30]: 285.6500
```

## 🔧 Configuration & Best Practices

### Environment Setup

```bash
# Required: Your Tinkoff Invest API token
export TINKOFF_TOKEN="your_token_here"

# Optional: Use demo environment for testing
# (Demo mode is used by default in examples)
```

### Performance Tips

1. **Selective Subscriptions**: Only subscribe to data you actually need
2. **Proper Context Management**: Use contexts with timeouts for long-running streams
3. **Concurrent Processing**: Use goroutines to handle multiple streams
4. **Error Handling**: Implement proper reconnection logic for production use
5. **Rate Limiting**: Be aware of API rate limits and implement backoff strategies

### Production Considerations

- **Reconnection Logic**: Implement automatic reconnection on stream failures
- **Data Persistence**: Store important data updates to database
- **Monitoring**: Add metrics and health checks for stream status
- **Security**: Never log tokens or sensitive account information
- **Resource Management**: Properly close streams and connections

## 🛠️ Troubleshooting

### Common Issues

1. **Authentication Errors**: Verify your token is correct and has proper permissions
2. **Stream Disconnections**: Implement reconnection logic with exponential backoff
3. **High Memory Usage**: Process data immediately instead of buffering large amounts
4. **Missing Data**: Check if instruments are trading and markets are open

### Debug Mode

Enable debug logging to see detailed gRPC communication:

```go
// Add to your client initialization
log.SetLevel(log.DebugLevel)
```

This comprehensive guide covers all aspects of real-time streaming with the Tinkoff Go client. The implementation provides enterprise-grade streaming capabilities with proper error handling, concurrent processing, and production-ready patterns.
