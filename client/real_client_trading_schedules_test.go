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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRealClient_GetTradingSchedules_PreservesRequestAndRawPayload(t *testing.T) {
	empty, exchange := "", "FIXED_EXCHANGE"
	from := time.Date(2026, time.October, 4, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name              string
		request, expected *investapi.TradingSchedulesRequest
		response          *investapi.TradingSchedulesResponse
	}{
		{name: "success_absent_optionals_empty_response", request: &investapi.TradingSchedulesRequest{}, expected: &investapi.TradingSchedulesRequest{}, response: &investapi.TradingSchedulesResponse{}},
		{name: "success_present_empty_exchange", request: &investapi.TradingSchedulesRequest{Exchange: &empty}, expected: &investapi.TradingSchedulesRequest{Exchange: proto.String("")}, response: &investapi.TradingSchedulesResponse{}},
		{name: "success_fixed_bounds_rich_raw_response", request: &investapi.TradingSchedulesRequest{Exchange: &exchange, From: timestamppb.New(from), To: timestamppb.New(to)}, expected: &investapi.TradingSchedulesRequest{Exchange: proto.String("FIXED_EXCHANGE"), From: timestamppb.New(time.Date(2026, time.October, 4, 0, 0, 0, 0, time.UTC)), To: timestamppb.New(time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC))}, response: scheduleTestRawResponse()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			c := newScheduleTestClient(t, tc.expected, func(ctx context.Context) (*investapi.TradingSchedulesResponse, error) {
				calls.Add(1)
				if _, ok := ctx.Deadline(); !ok {
					t.Error("caller deadline did not reach server")
				}
				return tc.response, nil
			})
			ctx, cancel := scheduleTestCallerContext(scheduleTestTimeout)
			defer cancel()
			got, err := c.GetTradingSchedules(ctx, tc.request)
			if err != nil {
				t.Errorf("GetTradingSchedules: %v", err)
				return
			}
			if got == nil || !proto.Equal(got, tc.response) {
				t.Errorf("raw response changed: got %v, want %v", got, tc.response)
			}
			if calls.Load() != 1 {
				t.Errorf("RPC calls = %d, want 1", calls.Load())
			}
		})
	}
}

func TestRealClient_GetTradingSchedules_WrapsRPCFailure(t *testing.T) {
	var calls atomic.Int32
	c := newScheduleTestClient(t, &investapi.TradingSchedulesRequest{}, func(context.Context) (*investapi.TradingSchedulesResponse, error) {
		calls.Add(1)
		return nil, status.Error(codes.NotFound, "schedule missing")
	})
	ctx, cancel := scheduleTestCallerContext(scheduleTestTimeout)
	defer cancel()
	got, err := c.GetTradingSchedules(ctx, &investapi.TradingSchedulesRequest{})
	if got != nil || status.Code(err) != codes.NotFound {
		t.Errorf("got (%v, %v), want nil and NotFound", got, err)
	}
	cause := errors.Unwrap(err)
	if cause == nil || status.Convert(cause).Message() != "schedule missing" || !errors.Is(err, cause) {
		t.Errorf("wrapped RPC cause = %v", cause)
	}
	if err == nil || !strings.Contains(err.Error(), "trading schedules") {
		t.Errorf("missing endpoint error context: %v", err)
	}
	if calls.Load() != 1 {
		t.Errorf("RPC calls = %d, want 1", calls.Load())
	}
}

func TestRealClient_GetTradingSchedules_HonorsCallerCancellationAndDeadline(t *testing.T) {
	for _, tc := range []struct {
		name     string
		duration time.Duration
		cancel   bool
		code     codes.Code
	}{
		{name: "failed_caller_canceled", duration: scheduleTestTimeout, cancel: true, code: codes.Canceled},
		{name: "failed_caller_deadline", duration: time.Second, code: codes.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started, stopped := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			c := newScheduleTestClient(t, &investapi.TradingSchedulesRequest{}, func(ctx context.Context) (*investapi.TradingSchedulesResponse, error) {
				calls.Add(1)
				close(started)
				defer close(stopped)
				<-ctx.Done()
				return nil, status.FromContextError(ctx.Err()).Err()
			})
			ctx, cancel := scheduleTestCallerContext(tc.duration)
			defer cancel()
			result := make(chan error, 1)
			done := make(chan struct{})
			go func() {
				defer close(done)
				response, err := c.GetTradingSchedules(ctx, &investapi.TradingSchedulesRequest{})
				if response != nil {
					t.Errorf("canceled call returned response: %v", response)
				}
				result <- err
			}()
			waitScheduleTestSignal(t, started)
			if c.mu.TryLock() {
				c.mu.Unlock()
				t.Error("getter did not retain read lock during RPC")
			}
			if tc.cancel {
				cancel()
			}
			waitCtx, waitCancel := context.WithTimeout(context.Background(), scheduleTestTimeout)
			defer waitCancel()
			select {
			case err := <-result:
				if status.Code(err) != tc.code || errors.Unwrap(err) == nil {
					t.Errorf("RPC error = %v, want wrapped %v", err, tc.code)
				}
			case <-waitCtx.Done():
				t.Error("getter did not finish after caller context ended")
			}
			waitScheduleTestSignal(t, stopped)
			waitScheduleTestSignal(t, done)
			if c.mu.TryLock() {
				c.mu.Unlock()
			} else {
				t.Error("getter did not release read lock after RPC")
			}
			if calls.Load() != 1 {
				t.Errorf("RPC calls = %d, want 1", calls.Load())
			}
		})
	}
}

func TestRealClient_GetTradingSchedules_DisconnectedAndAfterCloseIssueNoRPC(t *testing.T) {
	var calls atomic.Int32
	c := newScheduleTestClient(t, &investapi.TradingSchedulesRequest{}, func(context.Context) (*investapi.TradingSchedulesResponse, error) {
		calls.Add(1)
		return &investapi.TradingSchedulesResponse{}, nil
	})
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
	ctx, cancel := scheduleTestCallerContext(scheduleTestTimeout)
	defer cancel()
	got, err := c.GetTradingSchedules(ctx, &investapi.TradingSchedulesRequest{})
	if got != nil || err == nil || err.Error() != "client not connected" || calls.Load() != 0 {
		t.Errorf("disconnected result = (%v, %v), RPC calls = %d", got, err, calls.Load())
	}
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()
	if _, err := c.GetTradingSchedules(ctx, &investapi.TradingSchedulesRequest{}); err != nil {
		t.Errorf("connected call: %v", err)
		return
	}
	if calls.Load() != 1 {
		t.Errorf("connected RPC calls = %d, want 1", calls.Load())
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close: %v", err)
		return
	}
	got, err = c.GetTradingSchedules(ctx, &investapi.TradingSchedulesRequest{})
	if got != nil || err == nil || err.Error() != "client not connected" || calls.Load() != 1 {
		t.Errorf("after Close result = (%v, %v), RPC calls = %d", got, err, calls.Load())
	}
	if c.IsConnected() {
		t.Error("client remains connected after Close")
	}
	select {
	case <-c.ctx.Done():
	default:
		t.Error("Close did not cancel owned context")
	}
}

func scheduleTestCallerContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer conflicting-caller", "caller-marker", "must-be-replaced"))
	return context.WithTimeout(ctx, timeout)
}

func scheduleTestRawResponse() *investapi.TradingSchedulesResponse {
	at := func(day, hour, minute int) *timestamppb.Timestamp {
		return timestamppb.New(time.Date(2026, time.October, day, hour, minute, 0, 0, time.UTC))
	}
	return &investapi.TradingSchedulesResponse{Exchanges: []*investapi.TradingSchedule{
		{Exchange: "FIXED_EXCHANGE", Days: []*investapi.TradingDay{
			{Date: at(4, 0, 0), IsTradingDay: true, StartTime: at(4, 19, 0), EndTime: at(5, 2, 0),
				OpeningAuctionStartTime: at(4, 18, 50), OpeningAuctionEndTime: at(4, 19, 0), ClosingAuctionStartTime: at(5, 1, 50), ClosingAuctionEndTime: at(5, 2, 0),
				EveningOpeningAuctionStartTime: at(4, 19, 5), EveningStartTime: at(4, 19, 10), EveningEndTime: at(5, 1, 30),
				ClearingStartTime: at(4, 21, 0), ClearingEndTime: at(4, 21, 5), PremarketStartTime: at(4, 18, 0), PremarketEndTime: at(4, 18, 45),
				Intervals: []*investapi.TradingInterval{{Type: "future-unknown-native-kind", Interval: &investapi.TradingInterval_TimeInterval{StartTs: at(4, 19, 0), EndTs: at(5, 2, 0)}}, {Type: "fixed-auction-kind", Interval: &investapi.TradingInterval_TimeInterval{StartTs: at(4, 18, 50), EndTs: at(4, 19, 0)}}}},
			{Date: at(5, 0, 0), IsTradingDay: false},
		}},
		{Exchange: "SECOND_FIXED_EXCHANGE", Days: []*investapi.TradingDay{{Date: at(6, 0, 0), IsTradingDay: true, StartTime: at(6, 7, 0), EndTime: at(6, 15, 0)}}},
	}}
}

func newScheduleTestClient(t *testing.T, expected *investapi.TradingSchedulesRequest, respond func(context.Context) (*investapi.TradingSchedulesResponse, error)) *RealClient {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, request any, info *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
		if info.FullMethod != investapi.InstrumentsService_TradingSchedules_FullMethodName {
			t.Errorf("RPC method = %q", info.FullMethod)
			return nil, status.Error(codes.Unimplemented, "unexpected method")
		}
		req, ok := request.(*investapi.TradingSchedulesRequest)
		if !ok || !proto.Equal(req, expected) {
			t.Errorf("TradingSchedules request = %v", request)
			return nil, status.Error(codes.InvalidArgument, "unexpected request")
		}
		md, _ := metadata.FromIncomingContext(ctx)
		auth := md.Get("authorization")
		if len(auth) != 1 || auth[0] != scheduleTestAuthorization || len(md.Get("caller-marker")) != 0 {
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
		waitScheduleTestSignal(t, serveDone)
	})
	conn, err := grpc.NewClient("passthrough:///schedule-test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.DialContext(ctx)
	}))
	if err != nil {
		t.Errorf("in-process connection: %v", err)
		t.FailNow()
	}
	c = &RealClient{
		conn: conn, instrumentsClient: investapi.NewInstrumentsServiceClient(conn),
		metadata: metadata.Pairs("authorization", scheduleTestAuthorization),
		ctx:      ctx, cancel: cancel, connected: true,
	}
	return c
}

func waitScheduleTestSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), scheduleTestTimeout)
	defer cancel()
	select {
	case <-signal:
	case <-ctx.Done():
		t.Error("timed out waiting for in-process RPC synchronization")
		t.FailNow()
	}
}

const (
	scheduleTestAuthorization = "Bearer in-process-test-only"
	scheduleTestTimeout       = 5 * time.Second
)
