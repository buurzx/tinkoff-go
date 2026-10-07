package client

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	investapi "github.com/buurzx/tinkoff-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

func TestRealClient_GetFutureByFIGI_PreservesRawPayload(t *testing.T) {
	for _, tc := range []struct {
		name   string
		future *investapi.Future
	}{
		{
			name: "success_raw_future",
			future: &investapi.Future{
				Figi: futureTestFIGI, Uid: "fixed-future-uid", Lot: 3, Currency: "rub",
				MinPriceIncrement:       &investapi.Quotation{Units: 5, Nano: 125000000},
				MinPriceIncrementAmount: &investapi.Quotation{Units: 10, Nano: 250000000},
				InitialMarginOnBuy:      &investapi.MoneyValue{Currency: "rub", Units: 100, Nano: 500000000},
				InitialMarginOnSell:     &investapi.MoneyValue{Currency: "rub", Units: 120, Nano: 750000000},
			},
		},
		{name: "success_empty_instrument"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			c := newFutureTestClient(t, func(ctx context.Context) (*investapi.FutureResponse, error) {
				calls.Add(1)
				if _, ok := ctx.Deadline(); !ok {
					t.Error("caller deadline did not reach server")
				}
				return &investapi.FutureResponse{Instrument: tc.future}, nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), futureTestTimeout)
			defer cancel()
			got, err := c.GetFutureByFIGI(ctx, futureTestFIGI)
			if err != nil {
				t.Errorf("GetFutureByFIGI: %v", err)
				return
			}
			if !proto.Equal(got, tc.future) {
				t.Errorf("raw future changed: got %v, want %v", got, tc.future)
			}
			if tc.future == nil && got != nil {
				t.Errorf("empty Instrument became %v", got)
			}
			if calls.Load() != 1 {
				t.Errorf("RPC calls = %d, want 1", calls.Load())
			}
		})
	}
}

func TestRealClient_GetFutureByFIGI_WrapsRPCFailure(t *testing.T) {
	c := newFutureTestClient(t, func(context.Context) (*investapi.FutureResponse, error) {
		return nil, status.Error(codes.NotFound, "future missing")
	})
	ctx, cancel := context.WithTimeout(context.Background(), futureTestTimeout)
	defer cancel()
	got, err := c.GetFutureByFIGI(ctx, futureTestFIGI)
	if got != nil || status.Code(err) != codes.NotFound {
		t.Errorf("got (%v, %v), want nil and NotFound", got, err)
	}
	if cause := errors.Unwrap(err); cause == nil || status.Convert(cause).Message() != "future missing" {
		t.Errorf("wrapped RPC cause = %v", cause)
	}
	if err == nil || !strings.Contains(err.Error(), "future by FIGI "+futureTestFIGI) {
		t.Errorf("missing future/FIGI error context: %v", err)
	}
}

func TestRealClient_GetFutureByFIGI_HonorsCallerCancellationAndDeadline(t *testing.T) {
	for _, tc := range []struct {
		name     string
		duration time.Duration
		cancel   bool
		code     codes.Code
	}{
		{name: "failed_caller_canceled", duration: futureTestTimeout, cancel: true, code: codes.Canceled},
		{name: "failed_caller_deadline", duration: time.Second, code: codes.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started, stopped := make(chan struct{}), make(chan struct{})
			c := newFutureTestClient(t, func(ctx context.Context) (*investapi.FutureResponse, error) {
				close(started)
				defer close(stopped)
				<-ctx.Done()
				return nil, status.FromContextError(ctx.Err()).Err()
			})
			ctx, cancel := context.WithTimeout(context.Background(), tc.duration)
			defer cancel()
			result := make(chan error, 1)
			go func() {
				_, err := c.GetFutureByFIGI(ctx, futureTestFIGI)
				result <- err
			}()
			waitFutureTestSignal(t, started)
			if c.mu.TryLock() {
				c.mu.Unlock()
				t.Error("getter did not retain read lock during RPC")
			}
			if tc.cancel {
				cancel()
			}
			waitCtx, waitCancel := context.WithTimeout(context.Background(), futureTestTimeout)
			defer waitCancel()
			select {
			case err := <-result:
				if status.Code(err) != tc.code || errors.Unwrap(err) == nil {
					t.Errorf("RPC error = %v, want wrapped %v", err, tc.code)
				}
			case <-waitCtx.Done():
				t.Error("getter did not finish after caller context ended")
			}
			waitFutureTestSignal(t, stopped)
		})
	}
}

func TestRealClient_GetFutureByFIGI_DisconnectedAndAfterCloseIssueNoRPC(t *testing.T) {
	var calls atomic.Int32
	c := newFutureTestClient(t, func(context.Context) (*investapi.FutureResponse, error) {
		calls.Add(1)
		return &investapi.FutureResponse{}, nil
	})
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), futureTestTimeout)
	defer cancel()
	got, err := c.GetFutureByFIGI(ctx, futureTestFIGI)
	if got != nil || err == nil || err.Error() != "client not connected" || calls.Load() != 0 {
		t.Errorf("disconnected result = (%v, %v), RPC calls = %d", got, err, calls.Load())
	}
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	if _, err := c.GetFutureByFIGI(ctx, futureTestFIGI); err != nil {
		t.Errorf("connected call: %v", err)
		return
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close: %v", err)
		return
	}
	got, err = c.GetFutureByFIGI(ctx, futureTestFIGI)
	if got != nil || err == nil || err.Error() != "client not connected" || calls.Load() != 1 {
		t.Errorf("after Close result = (%v, %v), RPC calls = %d", got, err, calls.Load())
	}
	if c.IsConnected() {
		t.Error("client remains connected after Close")
	}
}

func newFutureTestClient(t *testing.T, respond func(context.Context) (*investapi.FutureResponse, error)) *RealClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, request any, info *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
		if info.FullMethod != investapi.InstrumentsService_FutureBy_FullMethodName {
			t.Errorf("RPC method = %q", info.FullMethod)
			return nil, status.Error(codes.Unimplemented, "unexpected method")
		}
		req, ok := request.(*investapi.InstrumentRequest)
		if !ok || req.IdType != investapi.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI || req.Id != futureTestFIGI || req.ClassCode != nil {
			t.Errorf("FutureBy request = %v", request)
			return nil, status.Error(codes.InvalidArgument, "unexpected request")
		}
		md, _ := metadata.FromIncomingContext(ctx)
		auth := md.Get("authorization")
		if len(auth) != 1 || auth[0] != futureTestAuthorization {
			t.Errorf("authorization = %v", auth)
			return nil, status.Error(codes.Unauthenticated, "unexpected authorization")
		}
		return respond(ctx)
	}))
	investapi.RegisterInstrumentsServiceServer(server, &investapi.UnimplementedInstrumentsServiceServer{})
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("in-process server: %v", err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	var c *RealClient
	t.Cleanup(func() {
		cancel()
		server.Stop()
		if c != nil {
			if err := c.Close(); err != nil {
				t.Errorf("client cleanup: %v", err)
			}
		}
		if err := listener.Close(); err != nil {
			t.Errorf("listener cleanup: %v", err)
		}
		waitFutureTestSignal(t, serveDone)
	})
	conn, err := grpc.NewClient("passthrough:///future-test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}))
	if err != nil {
		t.Errorf("in-process connection: %v", err)
		t.FailNow()
	}
	c = &RealClient{
		conn: conn, instrumentsClient: investapi.NewInstrumentsServiceClient(conn),
		metadata: metadata.Pairs("authorization", futureTestAuthorization),
		ctx:      ctx, cancel: cancel, connected: true,
	}
	return c
}

func waitFutureTestSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), futureTestTimeout)
	defer cancel()
	select {
	case <-signal:
	case <-ctx.Done():
		t.Error("timed out waiting for in-process RPC synchronization")
		t.FailNow()
	}
}

const (
	futureTestFIGI          = "fixed-future-figi"
	futureTestAuthorization = "Bearer in-process-test-only"
	futureTestTimeout       = 5 * time.Second
)
