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

	conn, err := grpc.NewClient(c.config.ServerURL, opts...)
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

// FindInstrument searches for instruments by query string using real API
func (c *RealClient) FindInstrument(ctx context.Context, query string, instrumentType *investapi.InstrumentType, apiTradeAvailableOnly bool) ([]*investapi.InstrumentShort, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.FindInstrumentRequest{
		Query: query,
	}

	if instrumentType != nil {
		req.InstrumentKind = instrumentType
	}

	if apiTradeAvailableOnly {
		req.ApiTradeAvailableFlag = &apiTradeAvailableOnly
	}

	resp, err := c.instrumentsClient.FindInstrument(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find instruments for query '%s': %w", query, err)
	}

	return resp.Instruments, nil
}

// GetBonds returns all bonds from Tinkoff Investment API
func (c *RealClient) GetBonds(ctx context.Context) (*investapi.BondsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.InstrumentsRequest{}

	resp, err := c.instrumentsClient.Bonds(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bonds: %w", err)
	}

	return resp, nil
}

// GetBondCoupons returns coupon calendar for a bond
func (c *RealClient) GetBondCoupons(ctx context.Context, instrumentID string, from, to *time.Time) (*investapi.GetBondCouponsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetBondCouponsRequest{
		InstrumentId: instrumentID,
	}

	if from != nil {
		req.From = timestamppb.New(*from)
	}
	if to != nil {
		req.To = timestamppb.New(*to)
	}

	resp, err := c.instrumentsClient.GetBondCoupons(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bond coupons: %w", err)
	}

	return resp, nil
}

// GetBondEvents returns events for a bond
func (c *RealClient) GetBondEvents(ctx context.Context, instrumentID string, from, to *time.Time, eventType investapi.GetBondEventsRequest_EventType) (*investapi.GetBondEventsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetBondEventsRequest{
		InstrumentId: instrumentID,
		Type:         eventType,
	}

	if from != nil {
		req.From = timestamppb.New(*from)
	}
	if to != nil {
		req.To = timestamppb.New(*to)
	}

	resp, err := c.instrumentsClient.GetBondEvents(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bond events: %w", err)
	}

	return resp, nil
}

// GetAssetBy returns asset information by AssetUID using real API
// This method can be used to get emitent (brand) information from bond data
func (c *RealClient) GetAssetBy(ctx context.Context, assetUID string) (*investapi.AssetResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.AssetRequest{
		Id: assetUID,
	}

	resp, err := c.instrumentsClient.GetAssetBy(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset by UID %s: %w", assetUID, err)
	}

	return resp, nil
}

// GetAssetFundamentals returns financial fundamentals for assets using real API
// This method returns financial data like EBITDA, Revenue, NetIncome, PE Ratio, ROE, ROA, etc.
func (c *RealClient) GetAssetFundamentals(ctx context.Context, assetUIDs []string) (*investapi.GetAssetFundamentalsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetAssetFundamentalsRequest{
		Assets: assetUIDs,
	}

	resp, err := c.instrumentsClient.GetAssetFundamentals(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset fundamentals for %d assets: %w", len(assetUIDs), err)
	}

	return resp, nil
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

// GetLastPrices returns last prices for given FIGIs using real API
func (c *RealClient) GetLastPrices(ctx context.Context, figis []string) (*investapi.GetLastPricesResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetLastPricesRequest{
		Figi: figis,
	}

	resp, err := c.marketDataClient.GetLastPrices(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get last prices: %w", err)
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

// GetLastTrades returns last trades for an instrument using real API
func (c *RealClient) GetLastTrades(ctx context.Context, req *investapi.GetLastTradesRequest) (*investapi.GetLastTradesResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	resp, err := c.marketDataClient.GetLastTrades(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get last trades: %w", err)
	}

	return resp, nil
}

// GetOrderBook returns order book for an instrument using real API
func (c *RealClient) GetOrderBook(ctx context.Context, req *investapi.GetOrderBookRequest) (*investapi.GetOrderBookResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	resp, err := c.marketDataClient.GetOrderBook(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
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

// STREAMING FUNCTIONALITY

// StartMarketDataStream starts real-time market data streaming
func (c *RealClient) StartMarketDataStream() (investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(c.ctx, c.metadata)

	// Start bidirectional stream
	stream, err := c.marketDataStreamClient.MarketDataStream(ctxWithAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to start market data stream: %w", err)
	}

	log.Println("🚀 Market data stream started")
	return stream, nil
}

// SubscribeCandles subscribes to candle updates for instruments
func (c *RealClient) SubscribeCandles(stream investapi.MarketDataStreamService_MarketDataStreamClient, instruments []string, interval investapi.SubscriptionInterval, waitingClose bool) error {
	candleInstruments := make([]*investapi.CandleInstrument, len(instruments))
	for i, instrumentID := range instruments {
		candleInstruments[i] = &investapi.CandleInstrument{
			InstrumentId: instrumentID,
			Interval:     interval,
		}
	}

	req := &investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{
			SubscribeCandlesRequest: &investapi.SubscribeCandlesRequest{
				SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments:        candleInstruments,
				WaitingClose:       waitingClose,
			},
		},
	}

	err := stream.Send(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to candles: %w", err)
	}

	log.Printf("📊 Subscribed to candles for %d instruments", len(instruments))
	return nil
}

// SubscribeOrderBook subscribes to order book updates for instruments
func (c *RealClient) SubscribeOrderBook(stream investapi.MarketDataStreamService_MarketDataStreamClient, instruments []string, depth int32) error {
	orderBookInstruments := make([]*investapi.OrderBookInstrument, len(instruments))
	for i, instrumentID := range instruments {
		orderBookInstruments[i] = &investapi.OrderBookInstrument{
			InstrumentId: instrumentID,
			Depth:        depth,
		}
	}

	req := &investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeOrderBookRequest{
			SubscribeOrderBookRequest: &investapi.SubscribeOrderBookRequest{
				SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments:        orderBookInstruments,
			},
		},
	}

	err := stream.Send(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to order book: %w", err)
	}

	log.Printf("📖 Subscribed to order book for %d instruments", len(instruments))
	return nil
}

// SubscribeTrades subscribes to trade updates for instruments
func (c *RealClient) SubscribeTrades(stream investapi.MarketDataStreamService_MarketDataStreamClient, instruments []string) error {
	tradeInstruments := make([]*investapi.TradeInstrument, len(instruments))
	for i, instrumentID := range instruments {
		tradeInstruments[i] = &investapi.TradeInstrument{
			InstrumentId: instrumentID,
		}
	}

	req := &investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeTradesRequest{
			SubscribeTradesRequest: &investapi.SubscribeTradesRequest{
				SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments:        tradeInstruments,
			},
		},
	}

	err := stream.Send(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to trades: %w", err)
	}

	log.Printf("💰 Subscribed to trades for %d instruments", len(instruments))
	return nil
}

// SubscribeLastPrices subscribes to last price updates for instruments
func (c *RealClient) SubscribeLastPrices(stream investapi.MarketDataStreamService_MarketDataStreamClient, instruments []string) error {
	lastPriceInstruments := make([]*investapi.LastPriceInstrument, len(instruments))
	for i, instrumentID := range instruments {
		lastPriceInstruments[i] = &investapi.LastPriceInstrument{
			InstrumentId: instrumentID,
		}
	}

	req := &investapi.MarketDataRequest{
		Payload: &investapi.MarketDataRequest_SubscribeLastPriceRequest{
			SubscribeLastPriceRequest: &investapi.SubscribeLastPriceRequest{
				SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments:        lastPriceInstruments,
			},
		},
	}

	err := stream.Send(req)
	if err != nil {
		return fmt.Errorf("failed to subscribe to last prices: %w", err)
	}

	log.Printf("💲 Subscribed to last prices for %d instruments", len(instruments))
	return nil
}

// StartOrderStream starts order state streaming
func (c *RealClient) StartOrderStream(accountIDs []string) (investapi.OrdersStreamService_OrderStateStreamClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(c.ctx, c.metadata)

	req := &investapi.OrderStateStreamRequest{
		Accounts: accountIDs,
	}

	stream, err := c.ordersStreamClient.OrderStateStream(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start order stream: %w", err)
	}

	log.Printf("🚀 Order stream started for %d accounts", len(accountIDs))
	return stream, nil
}

// ADVANCED ORDER FUNCTIONALITY

// PostStopOrder places a stop order using real API
func (c *RealClient) PostStopOrder(ctx context.Context, req *investapi.PostStopOrderRequest) (*investapi.PostStopOrderResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	resp, err := c.stopOrdersClient.PostStopOrder(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to post stop order: %w", err)
	}

	return resp, nil
}

// GetStopOrders returns stop orders for an account using real API
func (c *RealClient) GetStopOrders(ctx context.Context, accountID string, status investapi.StopOrderStatusOption) (*investapi.GetStopOrdersResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetStopOrdersRequest{
		AccountId: accountID,
		Status:    status,
	}

	resp, err := c.stopOrdersClient.GetStopOrders(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stop orders for account %s: %w", accountID, err)
	}

	return resp, nil
}

// CancelStopOrder cancels a stop order using real API
func (c *RealClient) CancelStopOrder(ctx context.Context, accountID, stopOrderID string) (*investapi.CancelStopOrderResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.CancelStopOrderRequest{
		AccountId:   accountID,
		StopOrderId: stopOrderID,
	}

	resp, err := c.stopOrdersClient.CancelStopOrder(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel stop order %s: %w", stopOrderID, err)
	}

	return resp, nil
}

// GetMaxLots returns maximum available lots for trading
func (c *RealClient) GetMaxLots(ctx context.Context, accountID, instrumentID string, price *float64) (*investapi.GetMaxLotsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetMaxLotsRequest{
		AccountId:    accountID,
		InstrumentId: instrumentID,
	}

	if price != nil {
		req.Price = floatToQuotation(*price)
	}

	resp, err := c.ordersClient.GetMaxLots(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get max lots: %w", err)
	}

	return resp, nil
}

// GetOrderPrice returns estimated order price
func (c *RealClient) GetOrderPrice(ctx context.Context, accountID, instrumentID string, price float64, direction investapi.OrderDirection, quantity int64) (*investapi.GetOrderPriceResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.GetOrderPriceRequest{
		AccountId:    accountID,
		InstrumentId: instrumentID,
		Price:        floatToQuotation(price),
		Direction:    direction,
		Quantity:     quantity,
	}

	resp, err := c.ordersClient.GetOrderPrice(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get order price: %w", err)
	}

	return resp, nil
}

// ReplaceOrder replaces an existing order
func (c *RealClient) ReplaceOrder(ctx context.Context, accountID, orderID, newIdempotencyKey string, quantity int64, price *float64) (*investapi.PostOrderResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Create context with authorization
	ctxWithAuth := metadata.NewOutgoingContext(ctx, c.metadata)

	req := &investapi.ReplaceOrderRequest{
		AccountId:      accountID,
		OrderId:        orderID,
		IdempotencyKey: newIdempotencyKey,
		Quantity:       quantity,
	}

	if price != nil {
		req.Price = floatToQuotation(*price)
	}

	resp, err := c.ordersClient.ReplaceOrder(ctxWithAuth, req)
	if err != nil {
		return nil, fmt.Errorf("failed to replace order %s: %w", orderID, err)
	}

	return resp, nil
}

// Helper function to convert float64 to Quotation
func floatToQuotation(value float64) *investapi.Quotation {
	units := int64(value)
	nano := int32((value - float64(units)) * 1e9)

	return &investapi.Quotation{
		Units: units,
		Nano:  nano,
	}
}

// Helper function to convert Quotation to float64
func quotationToFloat(q *investapi.Quotation) float64 {
	if q == nil {
		return 0.0
	}
	return float64(q.Units) + float64(q.Nano)/1e9
}
