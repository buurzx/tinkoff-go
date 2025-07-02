package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/buurzx/tinkoff-go/client"
	investapi "github.com/buurzx/tinkoff-go/proto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <search_query> [instrument_type]")
		fmt.Println("Example: go run main.go \"CNY RUB\"")
		fmt.Println("Example: go run main.go \"SBER\" share")
		fmt.Println("Example: go run main.go \"USD\" currency")
		fmt.Println("\nAvailable instrument types:")
		fmt.Println("  share - Акции")
		fmt.Println("  bond - Облигации")
		fmt.Println("  etf - ETF")
		fmt.Println("  currency - Валюты")
		fmt.Println("  futures - Фьючерсы")
		fmt.Println("  option - Опционы")
		os.Exit(1)
	}

	query := os.Args[1]
	var instrumentType *investapi.InstrumentType

	if len(os.Args) > 2 {
		typeStr := strings.ToLower(os.Args[2])
		switch typeStr {
		case "share", "stock":
			t := investapi.InstrumentType_INSTRUMENT_TYPE_SHARE
			instrumentType = &t
		case "bond":
			t := investapi.InstrumentType_INSTRUMENT_TYPE_BOND
			instrumentType = &t
		case "etf":
			t := investapi.InstrumentType_INSTRUMENT_TYPE_ETF
			instrumentType = &t
		case "currency":
			t := investapi.InstrumentType_INSTRUMENT_TYPE_CURRENCY
			instrumentType = &t
		case "futures", "future":
			t := investapi.InstrumentType_INSTRUMENT_TYPE_FUTURES
			instrumentType = &t
		case "option":
			t := investapi.InstrumentType_INSTRUMENT_TYPE_OPTION
			instrumentType = &t
		default:
			fmt.Printf("Unknown instrument type: %s\n", typeStr)
			os.Exit(1)
		}
	}

	// Get token from environment
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		log.Fatal("TINKOFF_TOKEN environment variable is required")
	}

	// Create client (using demo environment by default for safety)
	realClient, err := client.NewRealDemo(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer realClient.Close()

	ctx := context.Background()

	// Search for instruments
	fmt.Printf("Searching for instruments matching: '%s'\n", query)
	if instrumentType != nil {
		fmt.Printf("Filtering by type: %s\n", instrumentType.String())
	}
	fmt.Println(strings.Repeat("-", 80))

	instruments, err := realClient.FindInstrument(ctx, query, instrumentType, true) // Only tradeable instruments
	if err != nil {
		log.Fatalf("Failed to search instruments: %v", err)
	}

	if len(instruments) == 0 {
		fmt.Println("No instruments found matching your query.")
		fmt.Println("\nTips:")
		fmt.Println("- Try using partial names (e.g., 'CNY' instead of 'CNYRUB')")
		fmt.Println("- Try different instrument types")
		fmt.Println("- Check spelling and try synonyms")
		return
	}

	fmt.Printf("Found %d instrument(s):\n\n", len(instruments))

	for i, inst := range instruments {
		fmt.Printf("%d. %s (%s)\n", i+1, inst.Name, inst.InstrumentType)
		fmt.Printf("   Ticker: %s\n", inst.Ticker)
		fmt.Printf("   Class Code: %s\n", inst.ClassCode)
		fmt.Printf("   FIGI: %s\n", inst.Figi)
		fmt.Printf("   Full Ticker: %s.%s\n", inst.ClassCode, inst.Ticker)
		if inst.ApiTradeAvailableFlag {
			fmt.Printf("   ✅ Available for API trading\n")
		} else {
			fmt.Printf("   ❌ Not available for API trading\n")
		}
		fmt.Printf("   For IIS: %v\n", inst.ForIisFlag)
		fmt.Printf("   Lot Size: %d\n", inst.Lot)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Usage in your config:")
	fmt.Println("Use the 'Ticker' value in your configuration file.")
	fmt.Printf("For example, if you want to use the first result, set: \"%s\"\n", instruments[0].Ticker)
}
