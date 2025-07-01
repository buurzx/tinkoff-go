package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/buurzx/tinkoff-go/config"
	"github.com/buurzx/tinkoff-go/types"
)

// Client represents the main Tinkoff API client
type Client struct {
	config   *config.Config
	conn     *grpc.ClientConn
	metadata metadata.MD

	// Service clients will be added here when we generate proto files
	// For now, we'll create placeholders

	// Event handlers
	onCandle    func(*types.Candle)
	onTrade     func(*types.Trade)
	onOrderBook func(*types.OrderBook)

	// Channels for real-time data
	candleCh    chan *types.Candle
	tradeCh     chan *types.Trade
	orderBookCh chan *types.OrderBook

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread safety
	mu sync.RWMutex

	// Connection state
	connected bool

	// Accounts cache
	accounts []*types.Account
}

// New creates a new Tinkoff client
func New(token string) (*Client, error) {
	return NewWithDemo(token, false)
}

// NewDemo creates a new Tinkoff client for demo trading
func NewDemo(token string) (*Client, error) {
	return NewWithDemo(token, true)
}

// NewWithDemo creates a new Tinkoff client with demo flag
func NewWithDemo(token string, isDemo bool) (*Client, error) {
	cfg, err := config.New(token, isDemo)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return NewWithConfig(cfg)
}

// NewWithConfig creates a new Tinkoff client with provided config
func NewWithConfig(cfg *config.Config) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config:      cfg,
		metadata:    metadata.Pairs("authorization", "Bearer "+cfg.Token),
		ctx:         ctx,
		cancel:      cancel,
		candleCh:    make(chan *types.Candle, 100),
		tradeCh:     make(chan *types.Trade, 100),
		orderBookCh: make(chan *types.OrderBook, 100),
	}

	// Set default handlers
	client.onCandle = client.defaultCandleHandler
	client.onTrade = client.defaultTradeHandler
	client.onOrderBook = client.defaultOrderBookHandler

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// connect establishes gRPC connection
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ServerName: "invest-public-api.tinkoff.ru",
	})

	// Dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(64*1024*1024), // 64MB
			grpc.MaxCallSendMsgSize(64*1024*1024), // 64MB
		),
	}

	conn, err := grpc.Dial(c.config.ServerURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	c.conn = conn
	c.connected = true

	log.Printf("Connected to Tinkoff API: %s (demo: %v)", c.config.ServerURL, c.config.IsDemo)

	return nil
}

// Close closes the client connection and stops all goroutines
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Cancel context to stop all goroutines
	c.cancel()

	// Close channels
	close(c.candleCh)
	close(c.tradeCh)
	close(c.orderBookCh)

	// Close gRPC connection
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	c.connected = false
	log.Println("Tinkoff client closed")

	return nil
}

// IsConnected returns true if client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetAccounts returns list of accounts (placeholder implementation)
func (c *Client) GetAccounts(ctx context.Context) ([]*types.Account, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// TODO: Implement actual gRPC call when proto files are generated
	// For now, return mock data
	if c.accounts == nil {
		c.accounts = []*types.Account{
			{
				ID:         "mock-account-1",
				Type:       "ACCOUNT_TYPE_TINKOFF",
				Name:       "Mock Account",
				Status:     "ACCOUNT_STATUS_OPEN",
				OpenedDate: time.Now().Add(-365 * 24 * time.Hour),
			},
		}
	}

	return c.accounts, nil
}

// Event handler setters
func (c *Client) OnCandle(handler func(*types.Candle)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onCandle = handler
}

func (c *Client) OnTrade(handler func(*types.Trade)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onTrade = handler
}

func (c *Client) OnOrderBook(handler func(*types.OrderBook)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onOrderBook = handler
}

// Default event handlers
func (c *Client) defaultCandleHandler(candle *types.Candle) {
	log.Printf("Received candle: %s %s O:%.4f H:%.4f L:%.4f C:%.4f V:%d",
		candle.FIGI, candle.Time.Format("15:04:05"),
		candle.Open.ToFloat(), candle.High.ToFloat(),
		candle.Low.ToFloat(), candle.Close.ToFloat(),
		candle.Volume)
}

func (c *Client) defaultTradeHandler(trade *types.Trade) {
	direction := "BUY"
	if trade.Direction == types.OrderDirectionSell {
		direction = "SELL"
	}
	log.Printf("Received trade: %s %s %s %.4f x%d",
		trade.FIGI, trade.Time.Format("15:04:05"),
		direction, trade.Price.ToFloat(), trade.Quantity)
}

func (c *Client) defaultOrderBookHandler(orderBook *types.OrderBook) {
	log.Printf("Received order book: %s depth=%d bids=%d asks=%d",
		orderBook.FIGI, orderBook.Depth,
		len(orderBook.Bids), len(orderBook.Asks))
}

// GetInstrumentByFIGI returns instrument information by FIGI (placeholder)
func (c *Client) GetInstrumentByFIGI(ctx context.Context, figi string) (*types.Instrument, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// TODO: Implement actual gRPC call
	// For now, return mock data for SBER
	if figi == "BBG004730N88" {
		return &types.Instrument{
			FIGI:                  "BBG004730N88",
			Ticker:                "SBER",
			ClassCode:             "TQBR",
			ISIN:                  "RU0009029540",
			Lot:                   10,
			Currency:              "rub",
			Name:                  "Sberbank",
			Exchange:              "MOEX",
			CountryOfRisk:         "RU",
			CountryOfRiskName:     "Российская Федерация",
			InstrumentType:        "share",
			TradingStatus:         "SECURITY_TRADING_STATUS_NORMAL_TRADING",
			OTCFlag:               false,
			BuyAvailableFlag:      true,
			SellAvailableFlag:     true,
			MinPriceIncrement:     types.NewMoneyValue(0.01, "rub"),
			APITradeAvailableFlag: true,
		}, nil
	}

	return nil, fmt.Errorf("instrument not found: %s", figi)
}

// GetInstrumentByTicker returns instrument information by ticker and class code (placeholder)
func (c *Client) GetInstrumentByTicker(ctx context.Context, ticker, classCode string) (*types.Instrument, error) {
	// TODO: Implement actual search by ticker and class code
	if ticker == "SBER" && classCode == "TQBR" {
		return c.GetInstrumentByFIGI(ctx, "BBG004730N88")
	}

	return nil, fmt.Errorf("instrument not found: %s.%s", classCode, ticker)
}

// Context returns the client's context
func (c *Client) Context() context.Context {
	return c.ctx
}
