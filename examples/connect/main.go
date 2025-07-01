package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buurzx/tinkoff-go/client"
	"github.com/buurzx/tinkoff-go/types"
)

func main() {
	// Get token from environment or use placeholder
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		token = "your-token-here"
		log.Println("Warning: Using placeholder token. Set TINKOFF_TOKEN environment variable.")
		log.Println("Get your token at: https://www.tinkoff.ru/invest/settings/")
	}

	// Create client
	c, err := client.New(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	log.Println("=== Tinkoff Go Client Connection Example ===")

	// Test basic request/response
	log.Println("\n1. Testing basic request/response...")
	ctx := context.Background()

	accounts, err := c.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	log.Printf("Found %d accounts:", len(accounts))
	for i, account := range accounts {
		log.Printf("  [%d] %s - %s (%s)", i+1, account.ID, account.Name, account.Type)
	}

	// Test instrument lookup
	log.Println("\n2. Testing instrument lookup...")
	ticker := "SBER"
	classCode := "TQBR"

	instrument, err := c.GetInstrumentByTicker(ctx, ticker, classCode)
	if err != nil {
		log.Printf("Failed to get instrument %s.%s: %v", classCode, ticker, err)
	} else {
		log.Printf("Instrument found: %s (%s)", instrument.Name, instrument.FIGI)
		log.Printf("  Ticker: %s, Lot: %d, Currency: %s",
			instrument.Ticker, instrument.Lot, instrument.Currency)
		log.Printf("  Min price increment: %s", instrument.MinPriceIncrement.String())
	}

	// Test real-time subscriptions
	log.Println("\n3. Testing real-time subscriptions...")

	// Set up custom candle handler
	c.OnCandle(func(candle *types.Candle) {
		log.Printf("📊 CANDLE %s: O:%.4f H:%.4f L:%.4f C:%.4f V:%d [%s]",
			candle.FIGI,
			candle.Open.ToFloat(),
			candle.High.ToFloat(),
			candle.Low.ToFloat(),
			candle.Close.ToFloat(),
			candle.Volume,
			candle.Time.Format("15:04:05"))
	})

	// Set up custom trade handler
	c.OnTrade(func(trade *types.Trade) {
		direction := "🟢 BUY"
		if trade.Direction == types.OrderDirectionSell {
			direction = "🔴 SELL"
		}
		log.Printf("💰 TRADE %s: %s %.4f x%d [%s]",
			trade.FIGI, direction, trade.Price.ToFloat(),
			trade.Quantity, trade.Time.Format("15:04:05"))
	})

	// Set up custom order book handler
	c.OnOrderBook(func(orderBook *types.OrderBook) {
		log.Printf("📖 ORDER BOOK %s: %d bids, %d asks (depth: %d) [%s]",
			orderBook.FIGI, len(orderBook.Bids), len(orderBook.Asks),
			orderBook.Depth, orderBook.Time.Format("15:04:05"))
	})

	log.Printf("Subscribed to real-time data for %s.%s", classCode, ticker)
	log.Println("Note: This is a mock implementation. Real subscriptions will be added with proto files.")

	// Simulate some market data for demonstration
	go simulateMarketData(c, instrument.FIGI)

	// Wait for interrupt signal
	log.Println("\n4. Listening for real-time data...")
	log.Println("Press Ctrl+C to exit...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\nShutting down...")
}

// simulateMarketData simulates market data for demonstration purposes
func simulateMarketData(c *client.Client, figi string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	basePrice := 250.0

	for {
		select {
		case <-ticker.C:
			// Simulate price movement
			priceChange := (2.0 * (0.5 - 0.5)) // Random-like price change
			currentPrice := basePrice + priceChange

			// Trigger candle handler simulation
			go func() {
				if c.IsConnected() {
					// This would normally come from the real stream
					// For now, we simulate it
					log.Printf("📊 Simulated candle data (replace with real stream)")
				}
			}()

			// Create mock trade
			trade := &types.Trade{
				FIGI:      figi,
				Direction: types.OrderDirectionBuy,
				Price:     types.NewQuotation(currentPrice),
				Quantity:  100,
				Time:      time.Now(),
			}

			// Simulate trade (this would come from real stream)
			_ = trade

		case <-c.Context().Done():
			return
		}
	}
}
