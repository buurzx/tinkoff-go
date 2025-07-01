# Real API Integration Guide

This guide explains how to use the TinkoffGo library with the actual Tinkoff Invest API to make real trading operations.

## ⚠️ Important Safety Notice

**ALWAYS START WITH DEMO MODE!** The library provides both demo and production modes:

- **Demo Mode**: Safe for testing, uses demo environment
- **Production Mode**: Real money, real trades, real consequences

## Getting Started

### 1. Get Your API Token

1. Visit [Tinkoff Invest](https://www.tbank.ru/invest/)
2. Navigate to Settings → API
3. Create a new token with required permissions
4. **Keep your token secure and never commit it to version control**

### 2. Basic Setup

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/buurzx/tinkoff-go/client"
)

func main() {
    token := os.Getenv("TINKOFF_TOKEN")
    if token == "" {
        log.Fatal("TINKOFF_TOKEN environment variable is required")
    }

    // Start with demo mode for testing
    client, err := client.NewRealDemo(token)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Your trading logic here...
}
```

### 3. Environment Variables

Set your token as an environment variable:

```bash
export TINKOFF_TOKEN="your_actual_token_here"
```

## API Methods Reference

### Account Management

```go
// Get user information
userInfo, err := client.GetUserInfo(ctx)
if err != nil {
    log.Printf("Error: %v", err)
} else {
    fmt.Printf("User ID: %s, Premium: %v\n", userInfo.UserId, userInfo.PremStatus)
}

// Get all accounts
accounts, err := client.GetAccounts(ctx)
if err != nil {
    log.Printf("Error: %v", err)
} else {
    for _, account := range accounts {
        fmt.Printf("Account: %s (%s) - %s\n",
            account.Name, account.Id, account.Type.String())
    }
}
```

### Portfolio & Positions

```go
// Get portfolio summary
portfolio, err := client.GetPortfolio(ctx, accountID)
if err != nil {
    log.Printf("Error: %v", err)
} else {
    fmt.Printf("Total value: %s %s\n",
        formatMoney(portfolio.TotalAmountPortfolio),
        portfolio.TotalAmountPortfolio.Currency)
}

// Get detailed positions
positions, err := client.GetPositions(ctx, accountID)
if err != nil {
    log.Printf("Error: %v", err)
} else {
    fmt.Printf("Securities: %d, Money: %d\n",
        len(positions.Securities), len(positions.Money))
}
```

### Market Data

```go
import investapi "github.com/buurzx/tinkoff-go/proto"

// Get instrument by ticker
instrument, err := client.GetInstrumentByTicker(ctx, "SBER", "TQBR")
if err != nil {
    log.Printf("Error: %v", err)
} else {
    fmt.Printf("Name: %s, FIGI: %s\n", instrument.Name, instrument.Figi)
}

// Get instrument by FIGI
instrument, err := client.GetInstrumentByFIGI(ctx, "BBG004730N88")

// Get historical candles
to := time.Now()
from := to.Add(-30 * 24 * time.Hour) // Last 30 days

candles, err := client.GetCandles(ctx,
    "BBG004730N88", // Sber FIGI
    from, to,
    investapi.CandleInterval_CANDLE_INTERVAL_DAY)
```

### Order Management

```go
// Get active orders
orders, err := client.GetOrders(ctx, accountID)
if err != nil {
    log.Printf("Error: %v", err)
} else {
    fmt.Printf("Active orders: %d\n", len(orders.Orders))
}

// Place a limit order (BE VERY CAREFUL!)
orderReq := &investapi.PostOrderRequest{
    Figi:      "BBG004730N88", // Sber FIGI
    Quantity:  1,
    Price:     &investapi.Quotation{Units: 250, Nano: 0}, // 250.00 RUB
    Direction: investapi.OrderDirection_ORDER_DIRECTION_BUY,
    AccountId: accountID,
    OrderType: investapi.OrderType_ORDER_TYPE_LIMIT,
    OrderId:   "unique_order_id_" + time.Now().Format("20060102150405"),
}

response, err := client.PostOrder(ctx, orderReq)
if err != nil {
    log.Printf("Order failed: %v", err)
} else {
    fmt.Printf("Order placed: %s\n", response.OrderId)
}

// Cancel an order
cancelResp, err := client.CancelOrder(ctx, accountID, orderID)
```

## Utility Functions

### Money and Quotation Formatting

```go
// Helper function to format MoneyValue
func formatMoney(money *investapi.MoneyValue) string {
    if money == nil {
        return "0"
    }
    decimal := float64(money.Units) + float64(money.Nano)/1_000_000_000
    return fmt.Sprintf("%.2f", decimal)
}

// Helper function to format Quotation
func formatQuotation(quotation *investapi.Quotation) string {
    if quotation == nil {
        return "0"
    }
    decimal := float64(quotation.Units) + float64(quotation.Nano)/1_000_000_000
    return fmt.Sprintf("%.4f", decimal)
}
```

## Error Handling Best Practices

```go
import (
    "context"
    "errors"
    "time"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func handleAPICall() {
    // Always use timeouts
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    accounts, err := client.GetAccounts(ctx)
    if err != nil {
        // Check for specific gRPC errors
        if st, ok := status.FromError(err); ok {
            switch st.Code() {
            case codes.Unauthenticated:
                log.Println("Invalid token or authentication failed")
            case codes.PermissionDenied:
                log.Println("Insufficient permissions")
            case codes.ResourceExhausted:
                log.Println("Rate limit exceeded")
            case codes.DeadlineExceeded:
                log.Println("Request timeout")
            default:
                log.Printf("API error: %v", err)
            }
        } else {
            log.Printf("Network or other error: %v", err)
        }
        return
    }

    // Process successful response
    for _, account := range accounts {
        log.Printf("Account: %s", account.Name)
    }
}
```

## Configuration Options

### Demo vs Production Mode

```go
// Demo mode (safe)
demoClient, err := client.NewRealDemo(token)

// Production mode (real money!)
prodClient, err := client.NewReal(token)

// Custom configuration
cfg, err := config.New(token, true) // true = demo mode
if err != nil {
    log.Fatal(err)
}
client, err := client.NewRealWithConfig(cfg)
```

### Custom Server URLs

```go
cfg := &config.Config{
    Token:     token,
    IsDemo:    true,
    ServerURL: "custom-server:443", // Usually not needed
}
client, err := client.NewRealWithConfig(cfg)
```

## Running the Examples

### Basic API Test
```bash
TINKOFF_TOKEN=your_token make run-real-api
```

### Build and Run Manually
```bash
make example-real-api
TINKOFF_TOKEN=your_token ./bin/example-real-api
```

## Common Issues and Solutions

### 1. Authentication Errors
```
Error: rpc error: code = Unauthenticated desc = ...
```
**Solution**: Check your token, ensure it's valid and has proper permissions.

### 2. Permission Denied
```
Error: rpc error: code = PermissionDenied desc = ...
```
**Solution**: Your token may not have the required permissions for the operation.

### 3. Rate Limiting
```
Error: rpc error: code = ResourceExhausted desc = ...
```
**Solution**: Implement exponential backoff and respect rate limits.

### 4. Network Timeouts
```
Error: context deadline exceeded
```
**Solution**: Increase timeout duration or check network connectivity.

## Security Best Practices

1. **Never hardcode tokens** in source code
2. **Use environment variables** for tokens
3. **Start with demo mode** for all testing
4. **Implement proper error handling**
5. **Use timeouts** for all API calls
6. **Log operations** for audit trails
7. **Validate all inputs** before sending orders

## Trading Safety Guidelines

1. **Always test in demo mode first**
2. **Start with small amounts** when going live
3. **Implement position limits** and risk management
4. **Use stop-losses** and take-profits
5. **Monitor your positions** regularly
6. **Have an emergency stop mechanism**
7. **Understand the market hours** and trading rules

## Monitoring and Logging

```go
import "log"

// Log all API calls
log.Printf("Requesting accounts for user")
accounts, err := client.GetAccounts(ctx)
if err != nil {
    log.Printf("GetAccounts failed: %v", err)
    return
}
log.Printf("Retrieved %d accounts", len(accounts))

// Log trading operations
log.Printf("Placing order: %s %d shares of %s at %s",
    direction, quantity, figi, price)
```

## Production Deployment

### Environment Setup
```bash
# Production environment variables
export TINKOFF_TOKEN="prod_token_here"
export ENVIRONMENT="production"
export LOG_LEVEL="info"
```

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o trading-bot ./cmd/bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/trading-bot .
CMD ["./trading-bot"]
```

## Legal and Compliance

- Ensure compliance with local financial regulations
- Understand tax implications of algorithmic trading
- Keep detailed records of all transactions
- Consider consulting with financial and legal advisors
- Be aware of market manipulation rules

## Support and Resources

- [Tinkoff Invest API Documentation](https://tinkoff.github.io/investAPI/)
- [API Rate Limits](https://tinkoff.github.io/investAPI/limits/)
- [Error Codes Reference](https://tinkoff.github.io/investAPI/errors/)
- [Trading Rules and Regulations](https://www.tbank.ru/invest/)

---

**⚠️ Final Warning**: This library enables real trading with real money. The authors are not responsible for any financial losses. Always understand the risks and start with demo mode and small amounts.
