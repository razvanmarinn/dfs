// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.3
// source: proto/dfs.proto

package dfs

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
	BatchService_SendBatch_FullMethodName   = "/dfs.BatchService/SendBatch"
	BatchService_GetBatch_FullMethodName    = "/dfs.BatchService/GetBatch"
	BatchService_GetWorkerID_FullMethodName = "/dfs.BatchService/GetWorkerID"
)

// BatchServiceClient is the client API for BatchService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BatchServiceClient interface {
	SendBatch(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchResponse, error)
	GetBatch(ctx context.Context, in *GetBatchRequest, opts ...grpc.CallOption) (*GetBatchResponse, error)
	// New RPC to retrieve the worker's UUID
	GetWorkerID(ctx context.Context, in *WorkerIDRequest, opts ...grpc.CallOption) (*WorkerIDResponse, error)
}

type batchServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBatchServiceClient(cc grpc.ClientConnInterface) BatchServiceClient {
	return &batchServiceClient{cc}
}

func (c *batchServiceClient) SendBatch(ctx context.Context, in *BatchRequest, opts ...grpc.CallOption) (*BatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchResponse)
	err := c.cc.Invoke(ctx, BatchService_SendBatch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batchServiceClient) GetBatch(ctx context.Context, in *GetBatchRequest, opts ...grpc.CallOption) (*GetBatchResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetBatchResponse)
	err := c.cc.Invoke(ctx, BatchService_GetBatch_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *batchServiceClient) GetWorkerID(ctx context.Context, in *WorkerIDRequest, opts ...grpc.CallOption) (*WorkerIDResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WorkerIDResponse)
	err := c.cc.Invoke(ctx, BatchService_GetWorkerID_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BatchServiceServer is the server API for BatchService service.
// All implementations must embed UnimplementedBatchServiceServer
// for forward compatibility.
type BatchServiceServer interface {
	SendBatch(context.Context, *BatchRequest) (*BatchResponse, error)
	GetBatch(context.Context, *GetBatchRequest) (*GetBatchResponse, error)
	// New RPC to retrieve the worker's UUID
	GetWorkerID(context.Context, *WorkerIDRequest) (*WorkerIDResponse, error)
	mustEmbedUnimplementedBatchServiceServer()
}

// UnimplementedBatchServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBatchServiceServer struct{}

func (UnimplementedBatchServiceServer) SendBatch(context.Context, *BatchRequest) (*BatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendBatch not implemented")
}
func (UnimplementedBatchServiceServer) GetBatch(context.Context, *GetBatchRequest) (*GetBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBatch not implemented")
}
func (UnimplementedBatchServiceServer) GetWorkerID(context.Context, *WorkerIDRequest) (*WorkerIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkerID not implemented")
}
func (UnimplementedBatchServiceServer) mustEmbedUnimplementedBatchServiceServer() {}
func (UnimplementedBatchServiceServer) testEmbeddedByValue()                      {}

// UnsafeBatchServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BatchServiceServer will
// result in compilation errors.
type UnsafeBatchServiceServer interface {
	mustEmbedUnimplementedBatchServiceServer()
}

func RegisterBatchServiceServer(s grpc.ServiceRegistrar, srv BatchServiceServer) {
	// If the following call pancis, it indicates UnimplementedBatchServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&BatchService_ServiceDesc, srv)
}

func _BatchService_SendBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatchServiceServer).SendBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BatchService_SendBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatchServiceServer).SendBatch(ctx, req.(*BatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BatchService_GetBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatchServiceServer).GetBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BatchService_GetBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatchServiceServer).GetBatch(ctx, req.(*GetBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BatchService_GetWorkerID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkerIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BatchServiceServer).GetWorkerID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BatchService_GetWorkerID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BatchServiceServer).GetWorkerID(ctx, req.(*WorkerIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BatchService_ServiceDesc is the grpc.ServiceDesc for BatchService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BatchService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dfs.BatchService",
	HandlerType: (*BatchServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendBatch",
			Handler:    _BatchService_SendBatch_Handler,
		},
		{
			MethodName: "GetBatch",
			Handler:    _BatchService_GetBatch_Handler,
		},
		{
			MethodName: "GetWorkerID",
			Handler:    _BatchService_GetWorkerID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/dfs.proto",
}
