// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: riverpb/river.proto

package riverpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	River_Get_FullMethodName  = "/River/Get"
	River_TTL_FullMethodName  = "/River/TTL"
	River_Put_FullMethodName  = "/River/Put"
	River_Exp_FullMethodName  = "/River/Exp"
	River_Del_FullMethodName  = "/River/Del"
	River_Stat_FullMethodName = "/River/Stat"
)

// RiverClient is the client API for River service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RiverClient interface {
	Get(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*DataResult, error)
	TTL(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*TTLResult, error)
	Put(ctx context.Context, in *Record, opts ...grpc.CallOption) (*InfoResult, error)
	Exp(ctx context.Context, in *ExpRecord, opts ...grpc.CallOption) (*InfoResult, error)
	Del(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*InfoResult, error)
	Stat(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Status, error)
}

type riverClient struct {
	cc grpc.ClientConnInterface
}

func NewRiverClient(cc grpc.ClientConnInterface) RiverClient {
	return &riverClient{cc}
}

func (c *riverClient) Get(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*DataResult, error) {
	out := new(DataResult)
	err := c.cc.Invoke(ctx, River_Get_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *riverClient) TTL(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*TTLResult, error) {
	out := new(TTLResult)
	err := c.cc.Invoke(ctx, River_TTL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *riverClient) Put(ctx context.Context, in *Record, opts ...grpc.CallOption) (*InfoResult, error) {
	out := new(InfoResult)
	err := c.cc.Invoke(ctx, River_Put_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *riverClient) Exp(ctx context.Context, in *ExpRecord, opts ...grpc.CallOption) (*InfoResult, error) {
	out := new(InfoResult)
	err := c.cc.Invoke(ctx, River_Exp_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *riverClient) Del(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*InfoResult, error) {
	out := new(InfoResult)
	err := c.cc.Invoke(ctx, River_Del_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *riverClient) Stat(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, River_Stat_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RiverServer is the server API for River service.
// All implementations must embed UnimplementedRiverServer
// for forward compatibility
type RiverServer interface {
	Get(context.Context, *RawData) (*DataResult, error)
	TTL(context.Context, *RawData) (*TTLResult, error)
	Put(context.Context, *Record) (*InfoResult, error)
	Exp(context.Context, *ExpRecord) (*InfoResult, error)
	Del(context.Context, *RawData) (*InfoResult, error)
	Stat(context.Context, *emptypb.Empty) (*Status, error)
	mustEmbedUnimplementedRiverServer()
}

// UnimplementedRiverServer must be embedded to have forward compatible implementations.
type UnimplementedRiverServer struct {
}

func (UnimplementedRiverServer) Get(context.Context, *RawData) (*DataResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedRiverServer) TTL(context.Context, *RawData) (*TTLResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TTL not implemented")
}
func (UnimplementedRiverServer) Put(context.Context, *Record) (*InfoResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Put not implemented")
}
func (UnimplementedRiverServer) Exp(context.Context, *ExpRecord) (*InfoResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exp not implemented")
}
func (UnimplementedRiverServer) Del(context.Context, *RawData) (*InfoResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Del not implemented")
}
func (UnimplementedRiverServer) Stat(context.Context, *emptypb.Empty) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stat not implemented")
}
func (UnimplementedRiverServer) mustEmbedUnimplementedRiverServer() {}

// UnsafeRiverServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RiverServer will
// result in compilation errors.
type UnsafeRiverServer interface {
	mustEmbedUnimplementedRiverServer()
}

func RegisterRiverServer(s grpc.ServiceRegistrar, srv RiverServer) {
	s.RegisterService(&River_ServiceDesc, srv)
}

func _River_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RawData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiverServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: River_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiverServer).Get(ctx, req.(*RawData))
	}
	return interceptor(ctx, in, info, handler)
}

func _River_TTL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RawData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiverServer).TTL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: River_TTL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiverServer).TTL(ctx, req.(*RawData))
	}
	return interceptor(ctx, in, info, handler)
}

func _River_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Record)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiverServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: River_Put_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiverServer).Put(ctx, req.(*Record))
	}
	return interceptor(ctx, in, info, handler)
}

func _River_Exp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExpRecord)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiverServer).Exp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: River_Exp_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiverServer).Exp(ctx, req.(*ExpRecord))
	}
	return interceptor(ctx, in, info, handler)
}

func _River_Del_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RawData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiverServer).Del(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: River_Del_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiverServer).Del(ctx, req.(*RawData))
	}
	return interceptor(ctx, in, info, handler)
}

func _River_Stat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RiverServer).Stat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: River_Stat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RiverServer).Stat(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// River_ServiceDesc is the grpc.ServiceDesc for River service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var River_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "River",
	HandlerType: (*RiverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _River_Get_Handler,
		},
		{
			MethodName: "TTL",
			Handler:    _River_TTL_Handler,
		},
		{
			MethodName: "Put",
			Handler:    _River_Put_Handler,
		},
		{
			MethodName: "Exp",
			Handler:    _River_Exp_Handler,
		},
		{
			MethodName: "Del",
			Handler:    _River_Del_Handler,
		},
		{
			MethodName: "Stat",
			Handler:    _River_Stat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "riverpb/river.proto",
}