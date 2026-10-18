package client

import (
	"context"
	"fmt"

	investapi "github.com/buurzx/tinkoff-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MarketDataStream opens an authenticated stream governed by the caller context.
// Canceling it leaves other callers and the shared connection usable.
func (c *RealClient) MarketDataStream(ctx context.Context, opts ...grpc.CallOption) (investapi.MarketDataStreamService_MarketDataStreamClient, error) {
	if c == nil {
		return nil, fmt.Errorf("client not connected")
	}
	if err := checkCallerContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to start market data stream: %w", err)
	}
	c.mu.RLock()
	if !c.connected || c.marketDataStreamClient == nil {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client, md := c.marketDataStreamClient, c.metadata.Copy()
	c.mu.RUnlock()

	stream, err := client.MarketDataStream(authenticatedCallerContext(ctx, md), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start market data stream: %w", err)
	}
	return stream, nil
}

// GetUserTariff returns the provider's raw tariff without interpreting its limits.
func (c *RealClient) GetUserTariff(ctx context.Context, request *investapi.GetUserTariffRequest) (*investapi.GetUserTariffResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("client not connected")
	}
	if err := checkCallerContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to get user tariff: %w", err)
	}
	if request == nil {
		return nil, fmt.Errorf("user tariff request is nil")
	}
	c.mu.RLock()
	if !c.connected || c.usersClient == nil {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client, md := c.usersClient, c.metadata.Copy()
	c.mu.RUnlock()

	response, err := client.GetUserTariff(authenticatedCallerContext(ctx, md), request)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tariff: %w", err)
	}
	return response, nil
}

func checkCallerContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("caller context is nil")
	}
	if err := ctx.Err(); err != nil {
		return status.FromContextError(err).Err()
	}
	return nil
}

func authenticatedCallerContext(ctx context.Context, owned metadata.MD) context.Context {
	caller, _ := metadata.FromOutgoingContext(ctx)
	md := caller.Copy()
	if md == nil {
		md = metadata.MD{}
	}
	for key, values := range owned {
		md[key] = values
	}
	return metadata.NewOutgoingContext(ctx, md)
}
