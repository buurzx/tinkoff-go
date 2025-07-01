package client

import (
	"context"
	"testing"
	"time"

	"github.com/buurzx/tinkoff-go/config"
	"github.com/buurzx/tinkoff-go/types"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "test-token",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	cfg, err := config.New("test-token", false)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	client, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}
}

func TestClient_GetAccounts(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accounts, err := client.GetAccounts(ctx)
	if err != nil {
		t.Fatalf("GetAccounts() error = %v", err)
	}

	if len(accounts) == 0 {
		t.Error("Expected at least one mock account")
	}

	// Verify mock account properties
	account := accounts[0]
	if account.ID == "" {
		t.Error("Account ID should not be empty")
	}
	if account.Name == "" {
		t.Error("Account Name should not be empty")
	}
	if account.Type == "" {
		t.Error("Account Type should not be empty")
	}
}

func TestClient_GetInstrumentByFIGI(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with SBER FIGI (mock data)
	instrument, err := client.GetInstrumentByFIGI(ctx, "BBG004730N88")
	if err != nil {
		t.Fatalf("GetInstrumentByFIGI() error = %v", err)
	}

	if instrument.FIGI != "BBG004730N88" {
		t.Errorf("Expected FIGI BBG004730N88, got %s", instrument.FIGI)
	}
	if instrument.Ticker != "SBER" {
		t.Errorf("Expected ticker SBER, got %s", instrument.Ticker)
	}
	if instrument.Lot != 10 {
		t.Errorf("Expected lot size 10, got %d", instrument.Lot)
	}

	// Test with non-existent FIGI
	_, err = client.GetInstrumentByFIGI(ctx, "INVALID_FIGI")
	if err == nil {
		t.Error("Expected error for invalid FIGI")
	}
}

func TestClient_GetInstrumentByTicker(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with SBER ticker (mock data)
	instrument, err := client.GetInstrumentByTicker(ctx, "SBER", "TQBR")
	if err != nil {
		t.Fatalf("GetInstrumentByTicker() error = %v", err)
	}

	if instrument.Ticker != "SBER" {
		t.Errorf("Expected ticker SBER, got %s", instrument.Ticker)
	}
	if instrument.ClassCode != "TQBR" {
		t.Errorf("Expected class code TQBR, got %s", instrument.ClassCode)
	}

	// Test with non-existent ticker
	_, err = client.GetInstrumentByTicker(ctx, "INVALID", "TQBR")
	if err == nil {
		t.Error("Expected error for invalid ticker")
	}
}

func TestClient_EventHandlers(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test setting event handlers
	candleCalled := false
	client.OnCandle(func(candle *types.Candle) {
		candleCalled = true
	})

	tradeCalled := false
	client.OnTrade(func(trade *types.Trade) {
		tradeCalled = true
	})

	orderBookCalled := false
	client.OnOrderBook(func(orderBook *types.OrderBook) {
		orderBookCalled = true
	})

	// Event handlers are set but not called in this test
	// In a real implementation, we would trigger events via the stream

	// For now, just verify they don't panic when set
	if candleCalled || tradeCalled || orderBookCalled {
		t.Error("Handlers should not be called without events")
	}
}

func TestClient_Close(t *testing.T) {
	client, err := New("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client should be connected initially")
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if client.IsConnected() {
		t.Error("Client should not be connected after Close()")
	}

	// Calling Close() again should not error
	err = client.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}
