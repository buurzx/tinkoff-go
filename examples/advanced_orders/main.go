package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/buurzx/tinkoff-go/client"
	investapi "github.com/buurzx/tinkoff-go/proto"
)

func main() {
	// Get token from environment
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		log.Fatal("TINKOFF_TOKEN environment variable is required")
	}

	// Create real client (demo mode for safety)
	realClient, err := client.NewRealDemo(token)
	if err != nil {
		log.Fatalf("Failed to create real client: %v", err)
	}
	defer realClient.Close()

	log.Println("🚀 Tinkoff Go Advanced Orders Demo")
	log.Println("==================================")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Get accounts
	accounts, err := realClient.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	if len(accounts) == 0 {
		log.Fatal("No accounts found")
	}

	selectedAccount := accounts[0]
	log.Printf("Using account: %s (%s)", selectedAccount.Name, selectedAccount.Id)

	// Get SBER instrument info
	sberInstrument, err := realClient.GetInstrumentByTicker(ctx, "SBER", "TQBR")
	if err != nil {
		log.Fatalf("Failed to get SBER instrument: %v", err)
	}

	log.Printf("Instrument: %s (%s) - %s", sberInstrument.Name, sberInstrument.Figi, sberInstrument.Currency)

	// Demonstrate advanced order functionality
	demonstrateAdvancedOrders(ctx, realClient, selectedAccount.Id, sberInstrument.Figi)
}

func demonstrateAdvancedOrders(ctx context.Context, client *client.RealClient, accountID, instrumentID string) {
	log.Println("\n📊 Advanced Order Features Demo")
	log.Println("===============================")

	// 1. Get maximum available lots
	log.Println("\n1️⃣ Getting maximum available lots...")
	price := 250.0 // Example price for SBER
	maxLots, err := client.GetMaxLots(ctx, accountID, instrumentID, &price)
	if err != nil {
		log.Printf("❌ Failed to get max lots: %v", err)
	} else {
		log.Printf("   💰 Buy limits:")
		log.Printf("     Available money: %.2f %s",
			quotationToFloat(maxLots.BuyLimits.BuyMoneyAmount), maxLots.Currency)
		log.Printf("     Max lots: %d", maxLots.BuyLimits.BuyMaxLots)
		log.Printf("     Max market lots: %d", maxLots.BuyLimits.BuyMaxMarketLots)
		log.Printf("   📈 Sell limits:")
		log.Printf("     Max lots: %d", maxLots.SellLimits.SellMaxLots)
	}

	// 2. Get order price estimation
	log.Println("\n2️⃣ Getting order price estimation...")
	orderPrice, err := client.GetOrderPrice(ctx, accountID, instrumentID, price, investapi.OrderDirection_ORDER_DIRECTION_BUY, 1)
	if err != nil {
		log.Printf("❌ Failed to get order price: %v", err)
	} else {
		log.Printf("   💵 Order cost estimation for 1 lot:")
		log.Printf("     Total amount: %.2f %s",
			moneyValueToFloat(orderPrice.TotalOrderAmount), orderPrice.TotalOrderAmount.Currency)
		log.Printf("     Initial amount: %.2f %s",
			moneyValueToFloat(orderPrice.InitialOrderAmount), orderPrice.InitialOrderAmount.Currency)
		log.Printf("     Commission: %.2f %s",
			moneyValueToFloat(orderPrice.ExecutedCommission), orderPrice.ExecutedCommission.Currency)
		log.Printf("     Lots requested: %d", orderPrice.LotsRequested)
	}

	// 3. Demonstrate different order types
	log.Println("\n3️⃣ Order Types Examples (Demo Mode - Not Actually Executed)")
	log.Println("============================================================")

	// Market Order Example
	log.Println("\n📈 Market Order Example:")
	marketOrderReq := &investapi.PostOrderRequest{
		InstrumentId: instrumentID,
		Quantity:     1,
		Direction:    investapi.OrderDirection_ORDER_DIRECTION_BUY,
		OrderType:    investapi.OrderType_ORDER_TYPE_MARKET,
		AccountId:    accountID,
		OrderId:      uuid.New().String(),
	}
	log.Printf("   Order: BUY 1 lot of %s at MARKET price", instrumentID)
	log.Printf("   Request: %+v", marketOrderReq)

	// Limit Order Example
	log.Println("\n📊 Limit Order Example:")
	limitOrderReq := &investapi.PostOrderRequest{
		InstrumentId: instrumentID,
		Quantity:     2,
		Price:        floatToQuotation(245.0),
		Direction:    investapi.OrderDirection_ORDER_DIRECTION_BUY,
		OrderType:    investapi.OrderType_ORDER_TYPE_LIMIT,
		AccountId:    accountID,
		OrderId:      uuid.New().String(),
		TimeInForce:  investapi.TimeInForceType_TIME_IN_FORCE_DAY,
	}
	log.Printf("   Order: BUY 2 lots of %s at 245.00 LIMIT (Day order)", instrumentID)
	log.Printf("   Request: %+v", limitOrderReq)

	// Fill or Kill Order Example
	log.Println("\n⚡ Fill or Kill Order Example:")
	fokOrderReq := &investapi.PostOrderRequest{
		InstrumentId: instrumentID,
		Quantity:     1,
		Price:        floatToQuotation(248.0),
		Direction:    investapi.OrderDirection_ORDER_DIRECTION_BUY,
		OrderType:    investapi.OrderType_ORDER_TYPE_LIMIT,
		AccountId:    accountID,
		OrderId:      uuid.New().String(),
		TimeInForce:  investapi.TimeInForceType_TIME_IN_FORCE_FILL_OR_KILL,
	}
	log.Printf("   Order: BUY 1 lot of %s at 248.00 LIMIT (Fill or Kill)", instrumentID)
	log.Printf("   Request: %+v", fokOrderReq)

	// 4. Stop Orders Examples
	log.Println("\n4️⃣ Stop Orders Examples (Demo Mode - Not Actually Executed)")
	log.Println("===========================================================")

	// Stop Loss Order
	log.Println("\n🛑 Stop Loss Order Example:")
	stopLossReq := &investapi.PostStopOrderRequest{
		InstrumentId:      instrumentID,
		Quantity:          1,
		StopPrice:         floatToQuotation(240.0),
		Direction:         investapi.StopOrderDirection_STOP_ORDER_DIRECTION_SELL,
		StopOrderType:     investapi.StopOrderType_STOP_ORDER_TYPE_STOP_LOSS,
		ExpirationType:    investapi.StopOrderExpirationType_STOP_ORDER_EXPIRATION_TYPE_GOOD_TILL_CANCEL,
		AccountId:         accountID,
		OrderId:           uuid.New().String(),
		ExchangeOrderType: investapi.ExchangeOrderType_EXCHANGE_ORDER_TYPE_MARKET,
	}
	log.Printf("   Order: SELL 1 lot of %s when price drops to 240.00 (Stop Loss)", instrumentID)
	log.Printf("   Request: %+v", stopLossReq)

	// Take Profit Order
	log.Println("\n🎯 Take Profit Order Example:")
	takeProfitReq := &investapi.PostStopOrderRequest{
		InstrumentId:      instrumentID,
		Quantity:          1,
		StopPrice:         floatToQuotation(260.0),
		Direction:         investapi.StopOrderDirection_STOP_ORDER_DIRECTION_SELL,
		StopOrderType:     investapi.StopOrderType_STOP_ORDER_TYPE_TAKE_PROFIT,
		ExpirationType:    investapi.StopOrderExpirationType_STOP_ORDER_EXPIRATION_TYPE_GOOD_TILL_CANCEL,
		AccountId:         accountID,
		OrderId:           uuid.New().String(),
		ExchangeOrderType: investapi.ExchangeOrderType_EXCHANGE_ORDER_TYPE_MARKET,
		TakeProfitType:    investapi.TakeProfitType_TAKE_PROFIT_TYPE_REGULAR,
	}
	log.Printf("   Order: SELL 1 lot of %s when price rises to 260.00 (Take Profit)", instrumentID)
	log.Printf("   Request: %+v", takeProfitReq)

	// Stop Limit Order
	log.Println("\n📋 Stop Limit Order Example:")
	stopLimitReq := &investapi.PostStopOrderRequest{
		InstrumentId:      instrumentID,
		Quantity:          1,
		Price:             floatToQuotation(238.0), // Limit price
		StopPrice:         floatToQuotation(240.0), // Stop price
		Direction:         investapi.StopOrderDirection_STOP_ORDER_DIRECTION_SELL,
		StopOrderType:     investapi.StopOrderType_STOP_ORDER_TYPE_STOP_LIMIT,
		ExpirationType:    investapi.StopOrderExpirationType_STOP_ORDER_EXPIRATION_TYPE_GOOD_TILL_CANCEL,
		AccountId:         accountID,
		OrderId:           uuid.New().String(),
		ExchangeOrderType: investapi.ExchangeOrderType_EXCHANGE_ORDER_TYPE_LIMIT,
	}
	log.Printf("   Order: SELL 1 lot of %s at 238.00 LIMIT when price drops to 240.00 (Stop Limit)", instrumentID)
	log.Printf("   Request: %+v", stopLimitReq)

	// Trailing Stop Order
	log.Println("\n🎢 Trailing Stop Order Example:")
	trailingStopReq := &investapi.PostStopOrderRequest{
		InstrumentId:      instrumentID,
		Quantity:          1,
		StopPrice:         floatToQuotation(245.0),
		Direction:         investapi.StopOrderDirection_STOP_ORDER_DIRECTION_SELL,
		StopOrderType:     investapi.StopOrderType_STOP_ORDER_TYPE_TAKE_PROFIT,
		ExpirationType:    investapi.StopOrderExpirationType_STOP_ORDER_EXPIRATION_TYPE_GOOD_TILL_CANCEL,
		AccountId:         accountID,
		OrderId:           uuid.New().String(),
		ExchangeOrderType: investapi.ExchangeOrderType_EXCHANGE_ORDER_TYPE_MARKET,
		TakeProfitType:    investapi.TakeProfitType_TAKE_PROFIT_TYPE_TRAILING,
		TrailingData: &investapi.PostStopOrderRequest_TrailingData{
			Indent:     floatToQuotation(5.0), // 5 rubles trailing distance
			IndentType: investapi.TrailingValueType_TRAILING_VALUE_ABSOLUTE,
			Spread:     floatToQuotation(1.0), // 1 ruble protective spread
			SpreadType: investapi.TrailingValueType_TRAILING_VALUE_ABSOLUTE,
		},
	}
	log.Printf("   Order: SELL 1 lot of %s with trailing stop (5 rubles indent, 1 ruble spread)", instrumentID)
	log.Printf("   Request: %+v", trailingStopReq)

	// 5. Get existing stop orders
	log.Println("\n5️⃣ Getting existing stop orders...")
	stopOrders, err := client.GetStopOrders(ctx, accountID, investapi.StopOrderStatusOption_STOP_ORDER_STATUS_ACTIVE)
	if err != nil {
		log.Printf("❌ Failed to get stop orders: %v", err)
	} else {
		log.Printf("   📋 Active stop orders: %d", len(stopOrders.StopOrders))
		for i, order := range stopOrders.StopOrders {
			if i >= 3 { // Show only first 3
				break
			}
			direction := "BUY"
			if order.Direction == investapi.StopOrderDirection_STOP_ORDER_DIRECTION_SELL {
				direction = "SELL"
			}
			log.Printf("     [%d] %s: %s %d lots at stop %.2f",
				i+1, order.StopOrderId, direction, order.LotsRequested,
				moneyValueToFloat(order.StopPrice))
		}
	}

	log.Println("\n✅ Advanced orders demo completed!")
	log.Println("\n💡 Key Features Demonstrated:")
	log.Println("   • Market, Limit, Best Price orders")
	log.Println("   • Time in Force: Day, Fill or Kill, Fill and Kill")
	log.Println("   • Stop Loss, Take Profit, Stop Limit orders")
	log.Println("   • Trailing Stop orders with customizable parameters")
	log.Println("   • Order cost estimation and position limits")
	log.Println("   • Order replacement and management")

	log.Println("\n⚠️  Important Notes:")
	log.Println("   • This demo runs in DEMO mode for safety")
	log.Println("   • For production trading, use client.NewReal(token)")
	log.Println("   • Always test with small amounts first")
	log.Println("   • Understand the risks before live trading")
}

// Helper functions
func floatToQuotation(value float64) *investapi.Quotation {
	units := int64(value)
	nano := int32((value - float64(units)) * 1e9)

	return &investapi.Quotation{
		Units: units,
		Nano:  nano,
	}
}

func moneyValueToFloat(m *investapi.MoneyValue) float64 {
	if m == nil {
		return 0.0
	}
	return float64(m.Units) + float64(m.Nano)/1e9
}

func quotationToFloat(q *investapi.Quotation) float64 {
	if q == nil {
		return 0.0
	}
	return float64(q.Units) + float64(q.Nano)/1e9
}
