package main

import (
	"fmt"
	"log"
	"os"

	"github.com/buurzx/tinkoff-go/client"
)

func main() {
	fmt.Println("TinkoffGo - High-performance Go client for Tinkoff Invest API")
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  Set TINKOFF_TOKEN environment variable with your API token")
	fmt.Println("  Get your token at: https://www.tinkoff.ru/invest/settings/")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  make run-connect    # Connection and basic functionality")
	fmt.Println("  make run-accounts   # Account and portfolio management")
	fmt.Println("  make run-streaming  # Real-time market data streaming")
	fmt.Println()

	// Check if token is available
	token := os.Getenv("TINKOFF_TOKEN")
	if token == "" {
		fmt.Println("⚠️  TINKOFF_TOKEN environment variable not set")
		fmt.Println("   Run: export TINKOFF_TOKEN=your-token-here")
		os.Exit(1)
	}

	// Test basic connection
	fmt.Println("🔌 Testing connection...")
	c, err := client.New(token)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}
	defer c.Close()

	fmt.Println("✅ Successfully connected to Tinkoff Invest API")
	fmt.Printf("   Demo mode: %v\n", false) // We'll get this from config later
	fmt.Println()
	fmt.Println("Ready to use! Check out the examples in ./examples/")
}
