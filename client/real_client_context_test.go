package client

import (
	"context"
	"errors"
	"net"
	"reflect"
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

func TestRealClient_MarketDataStream_AuthenticatedFlowOptionsAndIsolation(t *testing.T) {
	request := &investapi.MarketDataRequest{Payload: &investapi.MarketDataRequest_SubscribeCandlesRequest{SubscribeCandlesRequest: &investapi.SubscribeCandlesRequest{SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE, Instruments: []*investapi.CandleInstrument{{InstrumentId: "fixed-instrument", Interval: investapi.SubscriptionInterval_SUBSCRIPTION_INTERVAL_ONE_MINUTE}}}}}
	response := &investapi.MarketDataResponse{Payload: &investapi.MarketDataResponse_Candle{Candle: &investapi.Candle{Figi: "fixed-instrument"}}}
	stopped := make(chan struct{}, 3)
	server := &contextTestServer{stream: func(stream investapi.MarketDataStreamService_MarketDataStreamServer) error {
		defer func() { stopped <- struct{}{} }()
		checkContextTestMetadata(t, stream.Context())
		got, err := stream.Recv()
		if err != nil {
			return err
		}
		if !proto.Equal(got, request) {
			t.Errorf("request changed: %v", got)
		}
		if err := stream.Send(response); err != nil {
			return err
		}
		<-stream.Context().Done()
		return status.FromContextError(stream.Context().Err()).Err()
	}}
	c := newContextTestClient(t, server, nil)
	callerMD := metadata.Pairs("authorization", "Bearer caller-conflict", "sdk-header", "caller-conflict", "caller-header", "preserved")
	original := callerMD.Copy()
	ctx, cancel := context.WithTimeout(metadata.NewOutgoingContext(context.Background(), callerMD), scheduleTestTimeout)
	defer cancel()
	firstCtx, firstCancel := context.WithCancel(ctx)
	defer firstCancel()
	first, err := c.MarketDataStream(firstCtx)
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.MarketDataStream(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, stream := range []investapi.MarketDataStreamService_MarketDataStreamClient{first, second} {
		if err := stream.Send(request); err != nil {
			t.Fatal(err)
		}
		got, err := stream.Recv()
		if err != nil || !proto.Equal(got, response) {
			t.Fatalf("response=(%v,%v)", got, err)
		}
	}
	result := make(chan error, 1)
	go func() { _, err := first.Recv(); result <- err }()
	firstCancel()
	if err := waitContextTestError(t, result); status.Code(err) != codes.Canceled {
		t.Errorf("canceled Recv: %v", err)
	}
	waitScheduleTestSignal(t, stopped)
	if !c.IsConnected() {
		t.Error("caller cancellation closed SDK")
	}
	if err := second.Send(request); err != nil {
		t.Errorf("other caller Send: %v", err)
	}
	limited, err := c.MarketDataStream(ctx, grpc.MaxCallRecvMsgSize(1))
	if err != nil {
		t.Fatal(err)
	}
	if err := limited.Send(request); err != nil {
		t.Fatal(err)
	}
	if _, err := limited.Recv(); status.Code(err) != codes.ResourceExhausted {
		t.Errorf("call option not applied: %v", err)
	}
	if !reflect.DeepEqual(callerMD, original) {
		t.Errorf("caller metadata mutated: %v", callerMD)
	}
	if !reflect.DeepEqual(c.metadata, metadata.Pairs("authorization", scheduleTestAuthorization, "sdk-header", "owned")) {
		t.Error("SDK metadata mutated")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Recv(); err == nil {
		t.Error("Close left stream usable")
	}
	waitScheduleTestSignal(t, stopped)
	waitScheduleTestSignal(t, stopped)
}

func TestRealClient_GetUserTariff_PreservesRawRequestResponseAndStatus(t *testing.T) {
	request := &investapi.GetUserTariffRequest{}
	request.ProtoReflect().SetUnknown([]byte{0x98, 0x06, 0x07})
	response := &investapi.GetUserTariffResponse{UnaryLimits: []*investapi.UnaryLimit{{LimitPerMinute: 17, Methods: []string{"future-method"}}}, StreamLimits: []*investapi.StreamLimit{{Limit: 3, Open: 2, Streams: []string{"future-stream"}}}}
	response.ProtoReflect().SetUnknown([]byte{0x98, 0x06, 0x09})
	for _, tc := range []struct {
		name     string
		response *investapi.GetUserTariffResponse
		err      error
	}{
		{name: "success_raw_payload", response: response},
		{name: "failed_provider_status", err: status.Error(codes.PermissionDenied, "provider denied tariff")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newContextTestClient(t, &contextTestServer{tariff: func(ctx context.Context, got *investapi.GetUserTariffRequest) (*investapi.GetUserTariffResponse, error) {
				checkContextTestMetadata(t, ctx)
				if _, ok := ctx.Deadline(); !ok {
					t.Error("caller deadline lost")
				}
				if !proto.Equal(got, request) {
					t.Errorf("request changed: %v", got)
				}
				return tc.response, tc.err
			}}, nil)
			ctx, cancel := contextTestCallerContext(scheduleTestTimeout)
			defer cancel()
			got, err := c.GetUserTariff(ctx, request)
			if !proto.Equal(got, tc.response) || status.Code(err) != status.Code(tc.err) {
				t.Errorf("result=(%v,%v)", got, err)
			}
			if tc.err != nil && (errors.Unwrap(err) == nil || status.Convert(errors.Unwrap(err)).Message() != "provider denied tariff") {
				t.Errorf("status cause not preserved: %v", err)
			}
		})
	}
}

func TestRealClient_ContextOperations_CancellationDeadlineAndConcurrentClose(t *testing.T) {
	for _, operation := range []string{"stream", "tariff", "legacy"} {
		for _, ending := range []struct {
			name    string
			cancel  bool
			close   bool
			timeout time.Duration
			code    codes.Code
		}{
			{name: "failed_caller_cancel", cancel: true, timeout: scheduleTestTimeout, code: codes.Canceled},
			{name: "failed_caller_deadline", timeout: 100 * time.Millisecond, code: codes.DeadlineExceeded},
			{name: "failed_concurrent_close", close: true, timeout: scheduleTestTimeout, code: codes.Unavailable},
		} {
			if operation == "legacy" && !ending.close {
				continue
			}
			t.Run(operation+"/"+ending.name, func(t *testing.T) {
				started, stopped := make(chan struct{}), make(chan struct{})
				block := func(ctx context.Context) error {
					close(started)
					defer close(stopped)
					<-ctx.Done()
					return status.FromContextError(ctx.Err()).Err()
				}
				c := newContextTestClient(t, &contextTestServer{
					stream: func(stream investapi.MarketDataStreamService_MarketDataStreamServer) error {
						return block(stream.Context())
					},
					tariff: func(ctx context.Context, _ *investapi.GetUserTariffRequest) (*investapi.GetUserTariffResponse, error) {
						return nil, block(ctx)
					},
				}, nil)
				ctx, cancel := context.WithTimeout(context.Background(), ending.timeout)
				defer cancel()
				result := make(chan error, 1)
				go func() {
					if operation == "tariff" {
						_, err := c.GetUserTariff(ctx, &investapi.GetUserTariffRequest{})
						result <- err
						return
					}
					var stream investapi.MarketDataStreamService_MarketDataStreamClient
					var err error
					if operation == "legacy" {
						stream, err = c.StartMarketDataStream()
					} else {
						stream, err = c.MarketDataStream(ctx)
					}
					if err == nil {
						_, err = stream.Recv()
					}
					result <- err
				}()
				waitScheduleTestSignal(t, started)
				if !c.mu.TryLock() {
					t.Fatal("RPC retained SDK mutex")
				}
				c.mu.Unlock()
				if ending.cancel {
					cancel()
				}
				if ending.close {
					closed := make(chan error, 1)
					go func() { closed <- c.Close() }()
					if err := waitContextTestError(t, closed); err != nil {
						t.Error(err)
					}
				}
				err := waitContextTestError(t, result)
				if ending.close {
					if err == nil {
						t.Error("Close left RPC active")
					}
				} else if status.Code(err) != ending.code {
					t.Errorf("result %v, want %v", err, ending.code)
				}
				if operation == "tariff" && errors.Unwrap(err) == nil {
					t.Error("tariff error not wrapped")
				}
				waitScheduleTestSignal(t, stopped)
			})
		}
	}
}

func TestRealClient_MarketDataStream_BlockedOpenCancellationAndClose(t *testing.T) {
	for _, tc := range []struct {
		name             string
		closeSDK, cancel bool
		timeout          time.Duration
		code             codes.Code
	}{
		{name: "failed_caller_cancel", cancel: true, timeout: scheduleTestTimeout, code: codes.Canceled},
		{name: "failed_caller_deadline", timeout: 100 * time.Millisecond, code: codes.DeadlineExceeded},
		{name: "failed_concurrent_close", closeSDK: true, timeout: scheduleTestTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dialing := make(chan struct{})
			dialDone := make(chan struct{})
			c := newContextTestClient(t, &contextTestServer{}, func(ctx context.Context, _ string) (net.Conn, error) {
				close(dialing)
				defer close(dialDone)
				<-ctx.Done()
				return nil, ctx.Err()
			})
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			result := make(chan error, 1)
			go func() { _, err := c.MarketDataStream(ctx, grpc.WaitForReady(true)); result <- err }()
			waitScheduleTestSignal(t, dialing)
			if !c.mu.TryLock() {
				t.Fatal("blocked opener retained mutex")
			}
			c.mu.Unlock()
			if tc.closeSDK {
				closed := make(chan error, 1)
				go func() { closed <- c.Close() }()
				if err := waitContextTestError(t, closed); err != nil {
					t.Error(err)
				}
			} else if tc.cancel {
				cancel()
			}
			err := waitContextTestError(t, result)
			if err == nil || errors.Unwrap(err) == nil {
				t.Errorf("open error: %v", err)
			}
			if !tc.closeSDK && status.Code(err) != tc.code {
				t.Errorf("cancel error: %v", err)
			}
			if err := c.Close(); err != nil {
				t.Error(err)
			}
			waitScheduleTestSignal(t, dialDone)
		})
	}
}

func TestRealClient_ContextOperations_InvalidDisconnectedAndClosedIssueNoRPC(t *testing.T) {
	var calls atomic.Int32
	c := newContextTestClient(t, &contextTestServer{stream: func(investapi.MarketDataStreamService_MarketDataStreamServer) error { calls.Add(1); return nil }, tariff: func(context.Context, *investapi.GetUserTariffRequest) (*investapi.GetUserTariffResponse, error) {
		calls.Add(1)
		return &investapi.GetUserTariffResponse{}, nil
	}}, nil)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name   string
		client *RealClient
		ctx    context.Context
	}{
		{name: "failed_nil_receiver", ctx: context.Background()},
		{name: "failed_nil_context", client: c},
		{name: "failed_canceled_context", client: c, ctx: canceled},
		{name: "failed_uninitialized", client: &RealClient{}, ctx: context.Background()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.client.MarketDataStream(tc.ctx); err == nil {
				t.Error("stream accepted invalid input")
			}
			if _, err := tc.client.GetUserTariff(tc.ctx, &investapi.GetUserTariffRequest{}); err == nil {
				t.Error("tariff accepted invalid input")
			}
		})
	}
	if _, err := c.GetUserTariff(context.Background(), nil); err == nil {
		t.Error("accepted nil request")
	}
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	if _, err := c.MarketDataStream(context.Background()); err == nil {
		t.Error("disconnected stream accepted")
	}
	if _, err := c.GetUserTariff(context.Background(), &investapi.GetUserTariffRequest{}); err == nil {
		t.Error("disconnected tariff accepted")
	}
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.MarketDataStream(context.Background()); err == nil {
		t.Error("closed stream accepted")
	}
	if _, err := c.StartMarketDataStream(); err == nil {
		t.Error("closed legacy stream accepted")
	}
	if _, err := c.GetUserTariff(context.Background(), &investapi.GetUserTariffRequest{}); err == nil {
		t.Error("closed tariff accepted")
	}
	if calls.Load() != 0 {
		t.Errorf("invalid calls reached transport: %d", calls.Load())
	}
}

type contextTestServer struct {
	investapi.UnimplementedMarketDataStreamServiceServer
	investapi.UnimplementedUsersServiceServer
	stream func(investapi.MarketDataStreamService_MarketDataStreamServer) error
	tariff func(context.Context, *investapi.GetUserTariffRequest) (*investapi.GetUserTariffResponse, error)
}

func (s *contextTestServer) MarketDataStream(stream investapi.MarketDataStreamService_MarketDataStreamServer) error {
	return s.stream(stream)
}
func (s *contextTestServer) GetUserTariff(ctx context.Context, request *investapi.GetUserTariffRequest) (*investapi.GetUserTariffResponse, error) {
	return s.tariff(ctx, request)
}

func newContextTestClient(t *testing.T, service *contextTestServer, dialer func(context.Context, string) (net.Conn, error)) *RealClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	investapi.RegisterMarketDataStreamServiceServer(server, service)
	investapi.RegisterUsersServiceServer(server, service)
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("Serve: %v", err)
		}
	}()
	if dialer == nil {
		dialer = func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }
	}
	conn, err := grpc.NewClient("passthrough:///context-test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &RealClient{conn: conn, usersClient: investapi.NewUsersServiceClient(conn), marketDataStreamClient: investapi.NewMarketDataStreamServiceClient(conn), metadata: metadata.Pairs("authorization", scheduleTestAuthorization, "sdk-header", "owned"), ctx: ctx, cancel: cancel, connected: true}
	t.Cleanup(func() {
		cancel()
		if err := c.Close(); err != nil {
			t.Error(err)
		}
		server.Stop()
		_ = listener.Close()
		waitScheduleTestSignal(t, serveDone)
	})
	return c
}
func checkContextTestMetadata(t *testing.T, ctx context.Context) {
	t.Helper()
	md, _ := metadata.FromIncomingContext(ctx)
	for key, want := range map[string][]string{"authorization": {scheduleTestAuthorization}, "sdk-header": {"owned"}, "caller-header": {"preserved"}} {
		if !reflect.DeepEqual(md.Get(key), want) {
			t.Errorf("metadata %s=%v, want %v", key, md.Get(key), want)
		}
	}
}
func contextTestCallerContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer caller-conflict", "sdk-header", "caller-conflict", "caller-header", "preserved")), timeout)
}
func waitContextTestError(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(scheduleTestTimeout):
		t.Fatal("timed out waiting for RPC completion")
		return nil
	}
}
