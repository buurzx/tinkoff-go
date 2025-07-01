package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	// Create real client (demo mode)
	realClient, err := client.NewRealDemo(token)
	if err != nil {
		log.Fatalf("Failed to create real client: %v", err)
	}
	defer realClient.Close()

	log.Println("🚀 Tinkoff Go Real-Time Streaming Demo")
	log.Println("=====================================")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get accounts first
	accounts, err := realClient.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	if len(accounts) == 0 {
		log.Fatal("No accounts found")
	}

	selectedAccount := accounts[0]
	log.Printf("Using account: %s (%s)", selectedAccount.Name, selectedAccount.Id)

	// Define instruments to monitor
	instruments := []string{
		"BBG004730N88", // SBER
		"BBG004730ZJ9", // GAZP
		"BBG006L8G4H1", // YNDX
	}

	var wg sync.WaitGroup

	// Start market data streaming
	log.Println("\n📡 Starting market data streaming...")
	marketDataStream, err := realClient.StartMarketDataStream()
	if err != nil {
		log.Fatalf("Failed to start market data stream: %v", err)
	}

	// Subscribe to different data types
	log.Println("📊 Subscribing to candles...")
	err = realClient.SubscribeCandles(
		marketDataStream,
		instruments,
		investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE,
		false,
	)
	if err != nil {
		log.Printf("Failed to subscribe to candles: %v", err)
	}

	log.Println("📖 Subscribing to order books...")
	err = realClient.SubscribeOrderBook(marketDataStream, instruments, 10)
	if err != nil {
		log.Printf("Failed to subscribe to order books: %v", err)
	}

	log.Println("💰 Subscribing to trades...")
	err = realClient.SubscribeTrades(marketDataStream, instruments)
	if err != nil {
		log.Printf("Failed to subscribe to trades: %v", err)
	}

	log.Println("💲 Subscribing to last prices...")
	err = realClient.SubscribeLastPrices(marketDataStream, instruments)
	if err != nil {
		log.Printf("Failed to subscribe to last prices: %v", err)
	}

	// Start market data handler
	wg.Add(1)
	go handleMarketDataStream(marketDataStream, &wg)

	// Start order streaming
	log.Println("\n📋 Starting order streaming...")
	orderStream, err := realClient.StartOrderStream([]string{selectedAccount.Id})
	if err != nil {
		log.Printf("Failed to start order stream: %v", err)
	} else {
		// Start order handler
		wg.Add(1)
		go handleOrderStream(orderStream, &wg)
	}

	log.Println("\n✅ All streams started successfully!")
	log.Println("📊 Monitoring real-time data...")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("\n🛑 Shutting down...")
	case <-ctx.Done():
		log.Println("\n⏰ Timeout reached, shutting down...")
	}

	// Close streams
	if marketDataStream != nil {
		marketDataStream.CloseSend()
	}
	if orderStream != nil {
		orderStream.CloseSend()
	}

	// Wait for handlers to finish
	wg.Wait()
	log.Println("✅ Shutdown complete")
}

// handleMarketDataStream processes real-time market data
func handleMarketDataStream(stream investapi.MarketDataStreamService_MarketDataStreamClient, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Println("📡 Market data stream handler stopped")

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("📡 Market data stream closed by server")
				return
			}
			log.Printf("❌ Market data stream error: %v", err)
			return
		}

		processMarketDataResponse(resp)
	}
}

// processMarketDataResponse processes individual market data messages
func processMarketDataResponse(resp *investapi.MarketDataResponse) {
	switch payload := resp.Payload.(type) {
	case *investapi.MarketDataResponse_Candle:
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
		spreadPercent := 0.0
		if bestBid > 0 {
			spreadPercent = (spread / bestBid) * 100
		}

		log.Printf("📖 ORDER BOOK %s: Bid=%.4f Ask=%.4f Spread=%.4f (%.3f%%) Depth=%d/%d",
			getInstrumentName(orderBook.Figi),
			bestBid,
			bestAsk,
			spread,
			spreadPercent,
			len(orderBook.Bids),
			len(orderBook.Asks))

	case *investapi.MarketDataResponse_LastPrice:
		lastPrice := payload.LastPrice
		log.Printf("💲 LAST PRICE %s: %.4f [%s]",
			getInstrumentName(lastPrice.Figi),
			quotationToFloat(lastPrice.Price),
			lastPrice.Time.AsTime().Format("15:04:05"))

	case *investapi.MarketDataResponse_TradingStatus:
		status := payload.TradingStatus
		log.Printf("📈 TRADING STATUS %s: %s",
			getInstrumentName(status.Figi),
			status.TradingStatus.String())

	case *investapi.MarketDataResponse_Ping:
		log.Printf("🏓 Market data ping received")

	case *investapi.MarketDataResponse_SubscribeCandlesResponse:
		log.Printf("✅ Candles subscription confirmed: %s", payload.SubscribeCandlesResponse.TrackingId)

	case *investapi.MarketDataResponse_SubscribeOrderBookResponse:
		log.Printf("✅ Order book subscription confirmed: %s", payload.SubscribeOrderBookResponse.TrackingId)

	case *investapi.MarketDataResponse_SubscribeTradesResponse:
		log.Printf("✅ Trades subscription confirmed: %s", payload.SubscribeTradesResponse.TrackingId)

	case *investapi.MarketDataResponse_SubscribeLastPriceResponse:
		log.Printf("✅ Last price subscription confirmed: %s", payload.SubscribeLastPriceResponse.TrackingId)

	default:
		log.Printf("🤷 Unknown market data response type: %T", payload)
	}
}

// handleOrderStream processes real-time order updates
func handleOrderStream(stream investapi.OrdersStreamService_OrderStateStreamClient, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Println("📡 Order stream handler stopped")

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("📡 Order stream closed by server")
				return
			}
			log.Printf("❌ Order stream error: %v", err)
			return
		}

		processOrderStreamResponse(resp)
	}
}

// processOrderStreamResponse processes individual order messages
func processOrderStreamResponse(resp *investapi.OrderStateStreamResponse) {
	switch payload := resp.Payload.(type) {
	case *investapi.OrderStateStreamResponse_OrderState_:
		orderState := payload.OrderState
		direction := "BUY"
		if orderState.Direction == investapi.OrderDirection_ORDER_DIRECTION_SELL {
			direction = "SELL"
		}

		log.Printf("📋 ORDER %s: %s %s %s x%d @ %.4f - %s",
			orderState.OrderId,
			direction,
			orderState.Ticker,
			orderState.OrderType.String(),
			orderState.LotsRequested,
			moneyValueToFloat(orderState.InitialOrderPrice),
			orderState.ExecutionReportStatus.String())

	case *investapi.OrderStateStreamResponse_Ping:
		log.Printf("🏓 Order stream ping received")

	case *investapi.OrderStateStreamResponse_Subscription:
		log.Printf("✅ Order stream subscription confirmed: %s", payload.Subscription.TrackingId)

	default:
		log.Printf("🤷 Unknown order stream response type: %T", payload)
	}
}

// Helper functions
func quotationToFloat(q *investapi.Quotation) float64 {
	if q == nil {
		return 0.0
	}
	return float64(q.Units) + float64(q.Nano)/1e9
}

func moneyValueToFloat(m *investapi.MoneyValue) float64 {
	if m == nil {
		return 0.0
	}
	return float64(m.Units) + float64(m.Nano)/1e9
}

func getInstrumentName(figi string) string {
	switch figi {
	case "BBG004730N88":
		return "SBER"
	case "BBG004730ZJ9":
		return "GAZP"
	case "BBG006L8G4H1":
		return "YNDX"
	default:
		return figi
	}
}
