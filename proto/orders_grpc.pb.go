// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: orders.proto

package investapi

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	OrdersStreamService_TradesStream_FullMethodName     = "/tinkoff.public.invest.api.contract.v1.OrdersStreamService/TradesStream"
	OrdersStreamService_OrderStateStream_FullMethodName = "/tinkoff.public.invest.api.contract.v1.OrdersStreamService/OrderStateStream"
)

// OrdersStreamServiceClient is the client API for OrdersStreamService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrdersStreamServiceClient interface {
	// TradesStream — стрим сделок пользователя
	TradesStream(ctx context.Context, in *TradesStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[TradesStreamResponse], error)
	// OrderStateStream — стрим поручений пользователя
	// Перед работой прочитайте [статью](/invest/services/orders/orders_state_stream).
	OrderStateStream(ctx context.Context, in *OrderStateStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[OrderStateStreamResponse], error)
}

type ordersStreamServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOrdersStreamServiceClient(cc grpc.ClientConnInterface) OrdersStreamServiceClient {
	return &ordersStreamServiceClient{cc}
}

func (c *ordersStreamServiceClient) TradesStream(ctx context.Context, in *TradesStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[TradesStreamResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &OrdersStreamService_ServiceDesc.Streams[0], OrdersStreamService_TradesStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[TradesStreamRequest, TradesStreamResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type OrdersStreamService_TradesStreamClient = grpc.ServerStreamingClient[TradesStreamResponse]

func (c *ordersStreamServiceClient) OrderStateStream(ctx context.Context, in *OrderStateStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[OrderStateStreamResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &OrdersStreamService_ServiceDesc.Streams[1], OrdersStreamService_OrderStateStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[OrderStateStreamRequest, OrderStateStreamResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type OrdersStreamService_OrderStateStreamClient = grpc.ServerStreamingClient[OrderStateStreamResponse]

// OrdersStreamServiceServer is the server API for OrdersStreamService service.
// All implementations must embed UnimplementedOrdersStreamServiceServer
// for forward compatibility.
type OrdersStreamServiceServer interface {
	// TradesStream — стрим сделок пользователя
	TradesStream(*TradesStreamRequest, grpc.ServerStreamingServer[TradesStreamResponse]) error
	// OrderStateStream — стрим поручений пользователя
	// Перед работой прочитайте [статью](/invest/services/orders/orders_state_stream).
	OrderStateStream(*OrderStateStreamRequest, grpc.ServerStreamingServer[OrderStateStreamResponse]) error
	mustEmbedUnimplementedOrdersStreamServiceServer()
}

// UnimplementedOrdersStreamServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOrdersStreamServiceServer struct{}

func (UnimplementedOrdersStreamServiceServer) TradesStream(*TradesStreamRequest, grpc.ServerStreamingServer[TradesStreamResponse]) error {
	return status.Errorf(codes.Unimplemented, "method TradesStream not implemented")
}
func (UnimplementedOrdersStreamServiceServer) OrderStateStream(*OrderStateStreamRequest, grpc.ServerStreamingServer[OrderStateStreamResponse]) error {
	return status.Errorf(codes.Unimplemented, "method OrderStateStream not implemented")
}
func (UnimplementedOrdersStreamServiceServer) mustEmbedUnimplementedOrdersStreamServiceServer() {}
func (UnimplementedOrdersStreamServiceServer) testEmbeddedByValue()                             {}

// UnsafeOrdersStreamServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrdersStreamServiceServer will
// result in compilation errors.
type UnsafeOrdersStreamServiceServer interface {
	mustEmbedUnimplementedOrdersStreamServiceServer()
}

func RegisterOrdersStreamServiceServer(s grpc.ServiceRegistrar, srv OrdersStreamServiceServer) {
	// If the following call pancis, it indicates UnimplementedOrdersStreamServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OrdersStreamService_ServiceDesc, srv)
}

func _OrdersStreamService_TradesStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(TradesStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OrdersStreamServiceServer).TradesStream(m, &grpc.GenericServerStream[TradesStreamRequest, TradesStreamResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type OrdersStreamService_TradesStreamServer = grpc.ServerStreamingServer[TradesStreamResponse]

func _OrdersStreamService_OrderStateStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(OrderStateStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OrdersStreamServiceServer).OrderStateStream(m, &grpc.GenericServerStream[OrderStateStreamRequest, OrderStateStreamResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type OrdersStreamService_OrderStateStreamServer = grpc.ServerStreamingServer[OrderStateStreamResponse]

// OrdersStreamService_ServiceDesc is the grpc.ServiceDesc for OrdersStreamService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OrdersStreamService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tinkoff.public.invest.api.contract.v1.OrdersStreamService",
	HandlerType: (*OrdersStreamServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TradesStream",
			Handler:       _OrdersStreamService_TradesStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "OrderStateStream",
			Handler:       _OrdersStreamService_OrderStateStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "orders.proto",
}

const (
	OrdersService_PostOrder_FullMethodName      = "/tinkoff.public.invest.api.contract.v1.OrdersService/PostOrder"
	OrdersService_PostOrderAsync_FullMethodName = "/tinkoff.public.invest.api.contract.v1.OrdersService/PostOrderAsync"
	OrdersService_CancelOrder_FullMethodName    = "/tinkoff.public.invest.api.contract.v1.OrdersService/CancelOrder"
	OrdersService_GetOrderState_FullMethodName  = "/tinkoff.public.invest.api.contract.v1.OrdersService/GetOrderState"
	OrdersService_GetOrders_FullMethodName      = "/tinkoff.public.invest.api.contract.v1.OrdersService/GetOrders"
	OrdersService_ReplaceOrder_FullMethodName   = "/tinkoff.public.invest.api.contract.v1.OrdersService/ReplaceOrder"
	OrdersService_GetMaxLots_FullMethodName     = "/tinkoff.public.invest.api.contract.v1.OrdersService/GetMaxLots"
	OrdersService_GetOrderPrice_FullMethodName  = "/tinkoff.public.invest.api.contract.v1.OrdersService/GetOrderPrice"
)

// OrdersServiceClient is the client API for OrdersService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrdersServiceClient interface {
	// PostOrder — выставить заявку
	PostOrder(ctx context.Context, in *PostOrderRequest, opts ...grpc.CallOption) (*PostOrderResponse, error)
	// PostOrderAsync — выставить заявку асинхронным методом
	// Особенности работы приведены в [статье](/invest/services/orders/async).
	PostOrderAsync(ctx context.Context, in *PostOrderAsyncRequest, opts ...grpc.CallOption) (*PostOrderAsyncResponse, error)
	// CancelOrder — отменить заявку
	CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderResponse, error)
	// GetOrderState — получить статус торгового поручения
	GetOrderState(ctx context.Context, in *GetOrderStateRequest, opts ...grpc.CallOption) (*OrderState, error)
	// GetOrders — получить список активных заявок по счету
	GetOrders(ctx context.Context, in *GetOrdersRequest, opts ...grpc.CallOption) (*GetOrdersResponse, error)
	// ReplaceOrder — изменить выставленную заявку
	ReplaceOrder(ctx context.Context, in *ReplaceOrderRequest, opts ...grpc.CallOption) (*PostOrderResponse, error)
	// GetMaxLots — расчет количества доступных для покупки/продажи лотов
	GetMaxLots(ctx context.Context, in *GetMaxLotsRequest, opts ...grpc.CallOption) (*GetMaxLotsResponse, error)
	// GetOrderPrice — получить предварительную стоимость для лимитной заявки
	GetOrderPrice(ctx context.Context, in *GetOrderPriceRequest, opts ...grpc.CallOption) (*GetOrderPriceResponse, error)
}

type ordersServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOrdersServiceClient(cc grpc.ClientConnInterface) OrdersServiceClient {
	return &ordersServiceClient{cc}
}

func (c *ordersServiceClient) PostOrder(ctx context.Context, in *PostOrderRequest, opts ...grpc.CallOption) (*PostOrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PostOrderResponse)
	err := c.cc.Invoke(ctx, OrdersService_PostOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) PostOrderAsync(ctx context.Context, in *PostOrderAsyncRequest, opts ...grpc.CallOption) (*PostOrderAsyncResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PostOrderAsyncResponse)
	err := c.cc.Invoke(ctx, OrdersService_PostOrderAsync_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CancelOrderResponse)
	err := c.cc.Invoke(ctx, OrdersService_CancelOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) GetOrderState(ctx context.Context, in *GetOrderStateRequest, opts ...grpc.CallOption) (*OrderState, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OrderState)
	err := c.cc.Invoke(ctx, OrdersService_GetOrderState_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) GetOrders(ctx context.Context, in *GetOrdersRequest, opts ...grpc.CallOption) (*GetOrdersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetOrdersResponse)
	err := c.cc.Invoke(ctx, OrdersService_GetOrders_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) ReplaceOrder(ctx context.Context, in *ReplaceOrderRequest, opts ...grpc.CallOption) (*PostOrderResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PostOrderResponse)
	err := c.cc.Invoke(ctx, OrdersService_ReplaceOrder_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) GetMaxLots(ctx context.Context, in *GetMaxLotsRequest, opts ...grpc.CallOption) (*GetMaxLotsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMaxLotsResponse)
	err := c.cc.Invoke(ctx, OrdersService_GetMaxLots_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersServiceClient) GetOrderPrice(ctx context.Context, in *GetOrderPriceRequest, opts ...grpc.CallOption) (*GetOrderPriceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetOrderPriceResponse)
	err := c.cc.Invoke(ctx, OrdersService_GetOrderPrice_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrdersServiceServer is the server API for OrdersService service.
// All implementations must embed UnimplementedOrdersServiceServer
// for forward compatibility.
type OrdersServiceServer interface {
	// PostOrder — выставить заявку
	PostOrder(context.Context, *PostOrderRequest) (*PostOrderResponse, error)
	// PostOrderAsync — выставить заявку асинхронным методом
	// Особенности работы приведены в [статье](/invest/services/orders/async).
	PostOrderAsync(context.Context, *PostOrderAsyncRequest) (*PostOrderAsyncResponse, error)
	// CancelOrder — отменить заявку
	CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderResponse, error)
	// GetOrderState — получить статус торгового поручения
	GetOrderState(context.Context, *GetOrderStateRequest) (*OrderState, error)
	// GetOrders — получить список активных заявок по счету
	GetOrders(context.Context, *GetOrdersRequest) (*GetOrdersResponse, error)
	// ReplaceOrder — изменить выставленную заявку
	ReplaceOrder(context.Context, *ReplaceOrderRequest) (*PostOrderResponse, error)
	// GetMaxLots — расчет количества доступных для покупки/продажи лотов
	GetMaxLots(context.Context, *GetMaxLotsRequest) (*GetMaxLotsResponse, error)
	// GetOrderPrice — получить предварительную стоимость для лимитной заявки
	GetOrderPrice(context.Context, *GetOrderPriceRequest) (*GetOrderPriceResponse, error)
	mustEmbedUnimplementedOrdersServiceServer()
}

// UnimplementedOrdersServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedOrdersServiceServer struct{}

func (UnimplementedOrdersServiceServer) PostOrder(context.Context, *PostOrderRequest) (*PostOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostOrder not implemented")
}
func (UnimplementedOrdersServiceServer) PostOrderAsync(context.Context, *PostOrderAsyncRequest) (*PostOrderAsyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostOrderAsync not implemented")
}
func (UnimplementedOrdersServiceServer) CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelOrder not implemented")
}
func (UnimplementedOrdersServiceServer) GetOrderState(context.Context, *GetOrderStateRequest) (*OrderState, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrderState not implemented")
}
func (UnimplementedOrdersServiceServer) GetOrders(context.Context, *GetOrdersRequest) (*GetOrdersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrders not implemented")
}
func (UnimplementedOrdersServiceServer) ReplaceOrder(context.Context, *ReplaceOrderRequest) (*PostOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplaceOrder not implemented")
}
func (UnimplementedOrdersServiceServer) GetMaxLots(context.Context, *GetMaxLotsRequest) (*GetMaxLotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMaxLots not implemented")
}
func (UnimplementedOrdersServiceServer) GetOrderPrice(context.Context, *GetOrderPriceRequest) (*GetOrderPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrderPrice not implemented")
}
func (UnimplementedOrdersServiceServer) mustEmbedUnimplementedOrdersServiceServer() {}
func (UnimplementedOrdersServiceServer) testEmbeddedByValue()                       {}

// UnsafeOrdersServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrdersServiceServer will
// result in compilation errors.
type UnsafeOrdersServiceServer interface {
	mustEmbedUnimplementedOrdersServiceServer()
}

func RegisterOrdersServiceServer(s grpc.ServiceRegistrar, srv OrdersServiceServer) {
	// If the following call pancis, it indicates UnimplementedOrdersServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&OrdersService_ServiceDesc, srv)
}

func _OrdersService_PostOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).PostOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_PostOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).PostOrder(ctx, req.(*PostOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_PostOrderAsync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostOrderAsyncRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).PostOrderAsync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_PostOrderAsync_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).PostOrderAsync(ctx, req.(*PostOrderAsyncRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_CancelOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).CancelOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_CancelOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).CancelOrder(ctx, req.(*CancelOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_GetOrderState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).GetOrderState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_GetOrderState_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).GetOrderState(ctx, req.(*GetOrderStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_GetOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).GetOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_GetOrders_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).GetOrders(ctx, req.(*GetOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_ReplaceOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplaceOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).ReplaceOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_ReplaceOrder_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).ReplaceOrder(ctx, req.(*ReplaceOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_GetMaxLots_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMaxLotsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).GetMaxLots(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_GetMaxLots_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).GetMaxLots(ctx, req.(*GetMaxLotsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrdersService_GetOrderPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServiceServer).GetOrderPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrdersService_GetOrderPrice_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServiceServer).GetOrderPrice(ctx, req.(*GetOrderPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OrdersService_ServiceDesc is the grpc.ServiceDesc for OrdersService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OrdersService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tinkoff.public.invest.api.contract.v1.OrdersService",
	HandlerType: (*OrdersServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostOrder",
			Handler:    _OrdersService_PostOrder_Handler,
		},
		{
			MethodName: "PostOrderAsync",
			Handler:    _OrdersService_PostOrderAsync_Handler,
		},
		{
			MethodName: "CancelOrder",
			Handler:    _OrdersService_CancelOrder_Handler,
		},
		{
			MethodName: "GetOrderState",
			Handler:    _OrdersService_GetOrderState_Handler,
		},
		{
			MethodName: "GetOrders",
			Handler:    _OrdersService_GetOrders_Handler,
		},
		{
			MethodName: "ReplaceOrder",
			Handler:    _OrdersService_ReplaceOrder_Handler,
		},
		{
			MethodName: "GetMaxLots",
			Handler:    _OrdersService_GetMaxLots_Handler,
		},
		{
			MethodName: "GetOrderPrice",
			Handler:    _OrdersService_GetOrderPrice_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "orders.proto",
}
