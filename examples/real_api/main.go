package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/buurzx/tinkoff-go/client"
	investapi "github.com/buurzx/tinkoff-go/proto"
)

func main() {
	// Get token from environment variable
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		log.Fatal("TINKOFF_TOKEN environment variable is required")
	}

	// Create real client (demo mode)
	// For production trading, use client.NewReal(token) instead
	realClient, err := client.NewRealDemo(token)
	if err != nil {
		log.Fatalf("Failed to create real client: %v", err)
	}
	defer realClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🚀 Tinkoff Go Real API Demo")
	fmt.Println("===========================")

	// 1. Get user information
	fmt.Println("\n1. Getting user information...")
	userInfo, err := realClient.GetUserInfo(ctx)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
	} else {
		fmt.Printf("   User ID: %s\n", userInfo.UserId)
		fmt.Printf("   Premium status: %v\n", userInfo.PremStatus)
		fmt.Printf("   Qualified status: %v\n", userInfo.QualStatus)
		fmt.Printf("   Tariff: %s\n", userInfo.Tariff)
	}

	// 2. Get accounts
	fmt.Println("\n2. Getting accounts...")
	accounts, err := realClient.GetAccounts(ctx)
	if err != nil {
		log.Printf("Failed to get accounts: %v", err)
		return
	}

	fmt.Printf("   Found %d accounts:\n", len(accounts))
	var selectedAccount *investapi.Account
	for i, account := range accounts {
		fmt.Printf("   [%d] %s (%s) - %s\n", i+1, account.Name, account.Id, account.Type.String())
		if selectedAccount == nil && account.Status == investapi.AccountStatus_ACCOUNT_STATUS_OPEN {
			selectedAccount = account
		}
	}

	if selectedAccount == nil {
		log.Println("No open accounts found")
		return
	}

	fmt.Printf("   Using account: %s (%s)\n", selectedAccount.Name, selectedAccount.Id)

	// 3. Get portfolio
	fmt.Println("\n3. Getting portfolio...")
	portfolio, err := realClient.GetPortfolio(ctx, selectedAccount.Id)
	if err != nil {
		log.Printf("Failed to get portfolio: %v", err)
	} else {
		fmt.Printf("   Total portfolio value: %s %s\n",
			formatMoney(portfolio.TotalAmountPortfolio),
			portfolio.TotalAmountPortfolio.Currency)
		fmt.Printf("   Positions: %d\n", len(portfolio.Positions))

		if len(portfolio.Positions) > 0 {
			fmt.Println("   Top 3 positions:")
			for i, position := range portfolio.Positions {
				if i >= 3 {
					break
				}
				fmt.Printf("     - %s: %s shares, current price: %s %s\n",
					position.Figi,
					formatQuotation(position.Quantity),
					formatMoney(position.CurrentPrice),
					position.CurrentPrice.Currency)
			}
		}
	}

	// 4. Get positions
	fmt.Println("\n4. Getting positions...")
	positions, err := realClient.GetPositions(ctx, selectedAccount.Id)
	if err != nil {
		log.Printf("Failed to get positions: %v", err)
	} else {
		fmt.Printf("   Securities: %d\n", len(positions.Securities))
		fmt.Printf("   Money positions: %d\n", len(positions.Money))
		fmt.Printf("   Futures: %d\n", len(positions.Futures))
		fmt.Printf("   Options: %d\n", len(positions.Options))
	}

	// 5. Get orders
	fmt.Println("\n5. Getting active orders...")
	orders, err := realClient.GetOrders(ctx, selectedAccount.Id)
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
	} else {
		fmt.Printf("   Active orders: %d\n", len(orders.Orders))
		for _, order := range orders.Orders {
			fmt.Printf("     - Order %s: %s %s at %s %s\n",
				order.OrderId,
				order.Direction.String(),
				order.Figi,
				formatMoney(order.InitialOrderPrice),
				order.InitialOrderPrice.Currency)
		}
	}

	// 6. Look up a popular instrument (Sber)
	fmt.Println("\n6. Looking up instrument (Sber)...")
	instrument, err := realClient.GetInstrumentByTicker(ctx, "SBER", "TQBR")
	if err != nil {
		log.Printf("Failed to get instrument: %v", err)
	} else {
		fmt.Printf("   Name: %s\n", instrument.Name)
		fmt.Printf("   FIGI: %s\n", instrument.Figi)
		fmt.Printf("   Currency: %s\n", instrument.Currency)
		fmt.Printf("   Lot size: %d\n", instrument.Lot)
		fmt.Printf("   Trading status: %s\n", instrument.TradingStatus.String())

		// 7. Get historical candles for this instrument
		fmt.Println("\n7. Getting historical candles...")
		to := time.Now()
		from := to.Add(-7 * 24 * time.Hour) // Last 7 days

		candles, err := realClient.GetCandles(ctx, instrument.Figi, from, to, investapi.CandleInterval_CANDLE_INTERVAL_DAY)
		if err != nil {
			log.Printf("Failed to get candles: %v", err)
		} else {
			fmt.Printf("   Retrieved %d daily candles:\n", len(candles.Candles))
			for i, candle := range candles.Candles {
				if i >= 3 { // Show only first 3
					break
				}
				fmt.Printf("     %s: O=%s H=%s L=%s C=%s V=%d\n",
					candle.Time.AsTime().Format("2006-01-02"),
					formatQuotation(candle.Open),
					formatQuotation(candle.High),
					formatQuotation(candle.Low),
					formatQuotation(candle.Close),
					candle.Volume)
			}
		}
	}

	fmt.Println("\n✅ Real API demo completed successfully!")
	fmt.Println("\nNote: This was run in DEMO mode. To use with real trading:")
	fmt.Println("1. Use client.NewReal(token) instead of client.NewRealDemo(token)")
	fmt.Println("2. Ensure you have proper permissions and understand the risks")
	fmt.Println("3. Start with small amounts for testing")
}

// Helper function to format MoneyValue
func formatMoney(money *investapi.MoneyValue) string {
	if money == nil {
		return "0"
	}

	// Convert nano to decimal
	decimal := float64(money.Units) + float64(money.Nano)/1_000_000_000
	return fmt.Sprintf("%.2f", decimal)
}

// Helper function to format Quotation
func formatQuotation(quotation *investapi.Quotation) string {
	if quotation == nil {
		return "0"
	}

	// Convert nano to decimal
	decimal := float64(quotation.Units) + float64(quotation.Nano)/1_000_000_000
	return fmt.Sprintf("%.4f", decimal)
}
