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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buurzx/tinkoff-go/config"
	investapi "github.com/buurzx/tinkoff-go/proto"
)

// RealClient represents the real Tinkoff API client using generated proto types
type RealClient struct {
	config   *config.Config
	conn     *grpc.ClientConn
	metadata metadata.MD

	// gRPC service clients
	usersClient       investapi.UsersServiceClient
	instrumentsClient investapi.InstrumentsServiceClient
	marketDataClient  investapi.MarketDataServiceClient
	ordersClient      investapi.OrdersServiceClient
	operationsClient  investapi.OperationsServiceClient
	stopOrdersClient  investapi.StopOrdersServiceClient

	// Streaming clients
	marketDataStreamClient investapi.MarketDataStreamServiceClient
	ordersStreamClient     investapi.OrdersStreamServiceClient
	operationsStreamClient investapi.OperationsStreamServiceClient

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread safety
	mu sync.RWMutex

	// Connection state
	connected bool

	// Accounts cache
	accounts []*investapi.Account
}

// NewReal creates a new real Tinkoff client using actual API
func NewReal(token string) (*RealClient, error) {
	return NewRealWithDemo(token, false)
}

// NewRealDemo creates a new real Tinkoff client for demo trading
func NewRealDemo(token string) (*RealClient, error) {
	return NewRealWithDemo(token, true)
}

// NewRealWithDemo creates a new real Tinkoff client with demo flag
func NewRealWithDemo(token string, isDemo bool) (*RealClient, error) {
	cfg, err := config.New(token, isDemo)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return NewRealWithConfig(cfg)
}

// NewRealWithConfig creates a new real Tinkoff client with provided config
func NewRealWithConfig(cfg *config.Config) (*RealClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &RealClient{
		config:   cfg,
		metadata: metadata.Pairs("authorization", "Bearer "+cfg.Token),
		ctx:      ctx,
		cancel:   cancel,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

// connect establishes gRPC connection and initializes service clients
func (c *RealClient) connect() error {
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

	// Initialize service clients
	c.usersClient = investapi.NewUsersServiceClient(conn)
	c.instrumentsClient = investapi.NewInstrumentsServiceClient(conn)
	c.marketDataClient = investapi.NewMarketDataServiceClient(conn)
	c.ordersClient = investapi.NewOrdersServiceClient(conn)
	c.operationsClient = investapi.NewOperationsServiceClient(conn)
	c.stopOrdersClient = investapi.NewStopOrdersServiceClient(conn)

	// Initialize streaming clients
	c.marketDataStreamClient = investapi.NewMarketDataStreamServiceClient(conn)
	c.ordersStreamClient = investapi.NewOrdersStreamServiceClient(conn)
	c.operationsStreamClient = investapi.NewOperationsStreamServiceClient(conn)

	c.connected = true

	log.Printf("Connected to Tinkoff API: %s (demo: %v)", c.config.ServerURL, c.config.IsDemo)

	return nil
}

// Close closes the client connection
func (c *RealClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Cancel context to stop all goroutines
	c.cancel()

	// Close gRPC connection
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	c.connected = false
	log.Println("Real Tinkoff client closed")

	return nil
}

// IsConnected returns true if client is connected
func (c *RealClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetAccounts returns list of accounts using real API
func (c *RealClient) GetAccounts(ctx context.Context) ([]*investapi.Account, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetAccountsRequest{}
	resp, err := c.usersClient.GetAccounts(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	// Cache accounts
	c.accounts = resp.Accounts

	return resp.Accounts, nil
}

// GetInstrumentByFIGI returns instrument information by FIGI using real API
func (c *RealClient) GetInstrumentByFIGI(ctx context.Context, figi string) (*investapi.Instrument, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.InstrumentRequest{
		IdType: investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     figi,
	}

	resp, err := c.instrumentsClient.GetInstrumentBy(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument by FIGI %s: %w", figi, err)
	}

	return resp.Instrument, nil
}

// GetInstrumentByTicker returns instrument information by ticker using real API
func (c *RealClient) GetInstrumentByTicker(ctx context.Context, ticker, classCode string) (*investapi.Instrument, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.InstrumentRequest{
		IdType:    investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_TICKER,
		Id:        ticker,
		ClassCode: &classCode,
	}

	resp, err := c.instrumentsClient.GetInstrumentBy(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument by ticker %s.%s: %w", classCode, ticker, err)
	}

	return resp.Instrument, nil
}

// GetPortfolio returns portfolio information for an account using real API
func (c *RealClient) GetPortfolio(ctx context.Context, accountID string) (*investapi.PortfolioResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	currency := investapi.PortfolioRequest_RUB
	req := &investapi.PortfolioRequest{
		AccountId: accountID,
		Currency:  &currency, // Default to RUB
	}

	resp, err := c.operationsClient.GetPortfolio(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio for account %s: %w", accountID, err)
	}

	return resp, nil
}

// GetPositions returns positions for an account using real API
func (c *RealClient) GetPositions(ctx context.Context, accountID string) (*investapi.PositionsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.PositionsRequest{
		AccountId: accountID,
	}

	resp, err := c.operationsClient.GetPositions(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions for account %s: %w", accountID, err)
	}

	return resp, nil
}

// GetOrders returns orders for an account using real API
func (c *RealClient) GetOrders(ctx context.Context, accountID string) (*investapi.GetOrdersResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetOrdersRequest{
		AccountId: accountID,
	}

	resp, err := c.ordersClient.GetOrders(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders for account %s: %w", accountID, err)
	}

	return resp, nil
}

// GetCandles returns historical candles using real API
func (c *RealClient) GetCandles(ctx context.Context, figi string, from, to time.Time, interval investapi.CandleInterval) (*investapi.GetCandlesResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetCandlesRequest{
		Figi:     &figi,
		From:     timestamppb.New(from),
		To:       timestamppb.New(to),
		Interval: interval,
	}

	resp, err := c.marketDataClient.GetCandles(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles for %s: %w", figi, err)
	}

	return resp, nil
}

// PostOrder places an order using real API
func (c *RealClient) PostOrder(ctx context.Context, req *investapi.PostOrderRequest) (*investapi.PostOrderResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	resp, err := c.ordersClient.PostOrder(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to post order: %w", err)
	}

	return resp, nil
}

// CancelOrder cancels an order using real API
func (c *RealClient) CancelOrder(ctx context.Context, accountID, orderID string) (*investapi.CancelOrderResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.CancelOrderRequest{
		AccountId: accountID,
		OrderId:   orderID,
	}

	resp, err := c.ordersClient.CancelOrder(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel order %s: %w", orderID, err)
	}

	return resp, nil
}

// GetUserInfo returns user information using real API
func (c *RealClient) GetUserInfo(ctx context.Context) (*investapi.GetInfoResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetInfoRequest{}
	resp, err := c.usersClient.GetInfo(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return resp, nil
}

// Context returns the client's context
func (c *RealClient) Context() context.Context {
	return c.ctx
}
