package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buurzx/tinkoff-go/client"
	"github.com/buurzx/tinkoff-go/types"
)

func main() {
	// Get token from environment
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		log.Fatal("TINKOFF_TOKEN environment variable is required")
	}

	// Create client
	c, err := client.New(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	log.Println("=== Tinkoff Go Client Streaming Example ===")

	// Define instruments to subscribe to
	instruments := []struct {
		ticker    string
		classCode string
		figi      string
	}{
		{"SBER", "TQBR", "BBG004730N88"},
		{"GAZP", "TQBR", "BBG004730ZJ9"},
		{"YNDX", "TQBR", "BBG006L8G4H1"},
	}

	// Set up event handlers
	setupEventHandlers(c)

	// Subscribe to market data for each instrument
	log.Println("\nSubscribing to market data streams...")
	for _, inst := range instruments {
		log.Printf("Subscribing to %s.%s (%s)", inst.classCode, inst.ticker, inst.figi)

		// In a real implementation, we would call:
		// err := c.SubscribeCandles(ctx, inst.figi, types.CandleInterval1Min)
		// err := c.SubscribeTrades(ctx, inst.figi)
		// err := c.SubscribeOrderBook(ctx, inst.figi, 10)

		// For now, just log the subscription intent
		log.Printf("  📊 Subscribed to 1-minute candles for %s", inst.ticker)
		log.Printf("  💰 Subscribed to trades for %s", inst.ticker)
		log.Printf("  📖 Subscribed to order book (depth 10) for %s", inst.ticker)
	}

	// Start market data simulation
	log.Println("\nStarting market data simulation...")
	log.Println("Note: This is mock data. Real streaming will be implemented with proto files.")

	for _, inst := range instruments {
		go simulateMarketDataStream(c, inst.figi, inst.ticker)
	}

	// Wait for interrupt signal
	log.Println("\nListening for real-time market data...")
	log.Println("Press Ctrl+C to exit...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\nShutting down...")
}

// setupEventHandlers configures all event handlers
func setupEventHandlers(c *client.Client) {
	// Candle handler with detailed logging
	c.OnCandle(func(candle *types.Candle) {
		log.Printf("📊 CANDLE [%s] %s: O:%.4f H:%.4f L:%.4f C:%.4f V:%d Complete:%t",
			candle.Time.Format("15:04:05"),
			candle.FIGI,
			candle.Open.ToFloat(),
			candle.High.ToFloat(),
			candle.Low.ToFloat(),
			candle.Close.ToFloat(),
			candle.Volume,
			candle.IsComplete)
	})

	// Trade handler with direction and size analysis
	c.OnTrade(func(trade *types.Trade) {
		direction := "🟢 BUY "
		if trade.Direction == types.OrderDirectionSell {
			direction = "🔴 SELL"
		}

		// Categorize trade size
		size := "small"
		if trade.Quantity >= 1000 {
			size = "medium"
		}
		if trade.Quantity >= 10000 {
			size = "large"
		}

		log.Printf("💰 TRADE [%s] %s: %s %.4f x%d (%s)",
			trade.Time.Format("15:04:05"),
			trade.FIGI,
			direction,
			trade.Price.ToFloat(),
			trade.Quantity,
			size)
	})

	// Order book handler with spread analysis
	c.OnOrderBook(func(orderBook *types.OrderBook) {
		bestBid := 0.0
		bestAsk := 0.0

		if len(orderBook.Bids) > 0 {
			bestBid = orderBook.Bids[0].Price.ToFloat()
		}
		if len(orderBook.Asks) > 0 {
			bestAsk = orderBook.Asks[0].Price.ToFloat()
		}

		spread := bestAsk - bestBid
		spreadPercent := 0.0
		if bestBid > 0 {
			spreadPercent = (spread / bestBid) * 100
		}

		log.Printf("📖 ORDER BOOK [%s] %s: Bid:%.4f Ask:%.4f Spread:%.4f (%.3f%%) Depth:%d/%d",
			orderBook.Time.Format("15:04:05"),
			orderBook.FIGI,
			bestBid,
			bestAsk,
			spread,
			spreadPercent,
			len(orderBook.Bids),
			len(orderBook.Asks))
	})
}

// simulateMarketDataStream simulates real-time market data for a given instrument
func simulateMarketDataStream(c *client.Client, figi, ticker string) {
	candleTicker := time.NewTicker(60 * time.Second)   // 1-minute candles
	tradeTicker := time.NewTicker(2 * time.Second)     // Trades every 2 seconds
	orderBookTicker := time.NewTicker(5 * time.Second) // Order book updates every 5 seconds

	defer candleTicker.Stop()
	defer tradeTicker.Stop()
	defer orderBookTicker.Stop()

	for {
		select {
		case <-candleTicker.C:
			if c.IsConnected() {
				log.Printf("📊 [%s] Simulated candle update", ticker)
			}

		case <-tradeTicker.C:
			if c.IsConnected() {
				log.Printf("💰 [%s] Simulated trade update", ticker)
			}

		case <-orderBookTicker.C:
			if c.IsConnected() {
				log.Printf("📖 [%s] Simulated order book update", ticker)
			}

		case <-c.Context().Done():
			return
		}
	}
}
