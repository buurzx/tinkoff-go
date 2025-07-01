package main

import (
	"context"
	"log"
	"os"

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

	log.Println("=== Tinkoff Go Client Accounts Example ===")

	ctx := context.Background()

	// Get all accounts
	accounts, err := c.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	log.Printf("\nFound %d accounts:", len(accounts))

	for i, account := range accounts {
		log.Printf("\n--- Account %d ---", i+1)
		log.Printf("ID: %s", account.ID)
		log.Printf("Name: %s", account.Name)
		log.Printf("Type: %s", account.Type)
		log.Printf("Status: %s", account.Status)
		log.Printf("Opened: %s", account.OpenedDate.Format("2006-01-02 15:04:05"))

		if !account.ClosedDate.IsZero() {
			log.Printf("Closed: %s", account.ClosedDate.Format("2006-01-02 15:04:05"))
		}

		// Get portfolio for this account
		err := getAccountPortfolio(ctx, c, account.ID)
		if err != nil {
			log.Printf("Failed to get portfolio for account %s: %v", account.ID, err)
		}

		// Get positions for this account
		err = getAccountPositions(ctx, c, account.ID)
		if err != nil {
			log.Printf("Failed to get positions for account %s: %v", account.ID, err)
		}

		// Get orders for this account
		err = getAccountOrders(ctx, c, account.ID)
		if err != nil {
			log.Printf("Failed to get orders for account %s: %v", account.ID, err)
		}
	}
}

// getAccountPortfolio retrieves and displays portfolio information
func getAccountPortfolio(ctx context.Context, c *client.Client, accountID string) error {
	log.Printf("\n  💰 Portfolio Summary:")

	// TODO: Implement actual portfolio retrieval when proto files are available
	// For now, show mock data
	portfolio := &types.Portfolio{
		TotalAmountShares:     types.NewMoneyValue(1500000.50, "rub"),
		TotalAmountBonds:      types.NewMoneyValue(500000.00, "rub"),
		TotalAmountEtf:        types.NewMoneyValue(250000.75, "rub"),
		TotalAmountCurrencies: types.NewMoneyValue(100000.00, "rub"),
		TotalAmountFutures:    types.NewMoneyValue(0.00, "rub"),
		ExpectedYield:         types.NewQuotation(125000.25),
	}

	log.Printf("    Shares: %s", portfolio.TotalAmountShares.String())
	log.Printf("    Bonds: %s", portfolio.TotalAmountBonds.String())
	log.Printf("    ETFs: %s", portfolio.TotalAmountEtf.String())
	log.Printf("    Currencies: %s", portfolio.TotalAmountCurrencies.String())
	log.Printf("    Futures: %s", portfolio.TotalAmountFutures.String())
	log.Printf("    Expected Yield: %.2f RUB", portfolio.ExpectedYield.ToFloat())

	totalValue := portfolio.TotalAmountShares.ToFloat() +
		portfolio.TotalAmountBonds.ToFloat() +
		portfolio.TotalAmountEtf.ToFloat() +
		portfolio.TotalAmountCurrencies.ToFloat() +
		portfolio.TotalAmountFutures.ToFloat()

	log.Printf("    Total Portfolio Value: %.2f RUB", totalValue)

	return nil
}

// getAccountPositions retrieves and displays position information
func getAccountPositions(ctx context.Context, c *client.Client, accountID string) error {
	log.Printf("\n  📈 Positions:")

	// TODO: Implement actual positions retrieval when proto files are available
	// For now, show mock data
	positions := []*types.Position{
		{
			FIGI:                 "BBG004730N88", // SBER
			InstrumentType:       "share",
			Quantity:             types.NewQuotation(100),
			AveragePositionPrice: types.NewMoneyValue(250.50, "rub"),
			ExpectedYield:        types.NewQuotation(2500.00),
			CurrentPrice:         types.NewQuotation(275.75),
			QuantityLots:         types.NewQuotation(10),
		},
		{
			FIGI:                 "BBG004730ZJ9", // GAZP
			InstrumentType:       "share",
			Quantity:             types.NewQuotation(50),
			AveragePositionPrice: types.NewMoneyValue(180.25, "rub"),
			ExpectedYield:        types.NewQuotation(-1250.00),
			CurrentPrice:         types.NewQuotation(155.00),
			QuantityLots:         types.NewQuotation(5),
		},
	}

	if len(positions) == 0 {
		log.Printf("    No positions found")
		return nil
	}

	for i, pos := range positions {
		log.Printf("    [%d] FIGI: %s", i+1, pos.FIGI)
		log.Printf("        Type: %s", pos.InstrumentType)
		log.Printf("        Quantity: %.0f (%.0f lots)", pos.Quantity.ToFloat(), pos.QuantityLots.ToFloat())
		log.Printf("        Avg Price: %s", pos.AveragePositionPrice.String())
		log.Printf("        Current Price: %.4f", pos.CurrentPrice.ToFloat())
		log.Printf("        P&L: %.2f RUB", pos.ExpectedYield.ToFloat())

		// Calculate percentage change
		avgPrice := pos.AveragePositionPrice.ToFloat()
		currentPrice := pos.CurrentPrice.ToFloat()
		if avgPrice > 0 {
			changePercent := ((currentPrice - avgPrice) / avgPrice) * 100
			log.Printf("        Change: %.2f%%", changePercent)
		}
		log.Println()
	}

	return nil
}

// getAccountOrders retrieves and displays order information
func getAccountOrders(ctx context.Context, c *client.Client, accountID string) error {
	log.Printf("\n  📋 Active Orders:")

	// TODO: Implement actual orders retrieval when proto files are available
	// For now, show mock data
	orders := []struct {
		ID        string
		FIGI      string
		Direction types.OrderDirection
		Type      types.OrderType
		State     types.OrderState
		Price     *types.Quotation
		Quantity  int64
	}{
		{
			ID:        "order-1",
			FIGI:      "BBG004730N88",
			Direction: types.OrderDirectionBuy,
			Type:      types.OrderTypeLimit,
			State:     types.OrderStateNew,
			Price:     types.NewQuotation(240.00),
			Quantity:  10,
		},
		{
			ID:        "order-2",
			FIGI:      "BBG004730ZJ9",
			Direction: types.OrderDirectionSell,
			Type:      types.OrderTypeLimit,
			State:     types.OrderStatePartiallyFill,
			Price:     types.NewQuotation(160.00),
			Quantity:  25,
		},
	}

	if len(orders) == 0 {
		log.Printf("    No active orders")
		return nil
	}

	for i, order := range orders {
		direction := "BUY"
		if order.Direction == types.OrderDirectionSell {
			direction = "SELL"
		}

		orderType := "LIMIT"
		if order.Type == types.OrderTypeMarket {
			orderType = "MARKET"
		}

		state := "NEW"
		switch order.State {
		case types.OrderStateFill:
			state = "FILLED"
		case types.OrderStatePartiallyFill:
			state = "PARTIALLY FILLED"
		case types.OrderStateCancelled:
			state = "CANCELLED"
		case types.OrderStateRejected:
			state = "REJECTED"
		}

		log.Printf("    [%d] %s", i+1, order.ID)
		log.Printf("        FIGI: %s", order.FIGI)
		log.Printf("        %s %s @ %.4f x%d", direction, orderType, order.Price.ToFloat(), order.Quantity)
		log.Printf("        State: %s", state)
		log.Println()
	}

	return nil
}
