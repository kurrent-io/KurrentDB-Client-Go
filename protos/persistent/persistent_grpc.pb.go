// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.0
// source: persistent.proto

package persistent

import (
	context "context"
	shared "github.com/EventStore/EventStore-Client-Go/v3/protos/shared"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PersistentSubscriptionsClient is the client API for PersistentSubscriptions service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PersistentSubscriptionsClient interface {
	Create(ctx context.Context, in *CreateReq, opts ...grpc.CallOption) (*CreateResp, error)
	Update(ctx context.Context, in *UpdateReq, opts ...grpc.CallOption) (*UpdateResp, error)
	Delete(ctx context.Context, in *DeleteReq, opts ...grpc.CallOption) (*DeleteResp, error)
	Read(ctx context.Context, opts ...grpc.CallOption) (PersistentSubscriptions_ReadClient, error)
	GetInfo(ctx context.Context, in *GetInfoReq, opts ...grpc.CallOption) (*GetInfoResp, error)
	ReplayParked(ctx context.Context, in *ReplayParkedReq, opts ...grpc.CallOption) (*ReplayParkedResp, error)
	List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*ListResp, error)
	RestartSubsystem(ctx context.Context, in *shared.Empty, opts ...grpc.CallOption) (*shared.Empty, error)
}

type persistentSubscriptionsClient struct {
	cc grpc.ClientConnInterface
}

func NewPersistentSubscriptionsClient(cc grpc.ClientConnInterface) PersistentSubscriptionsClient {
	return &persistentSubscriptionsClient{cc}
}

func (c *persistentSubscriptionsClient) Create(ctx context.Context, in *CreateReq, opts ...grpc.CallOption) (*CreateResp, error) {
	out := new(CreateResp)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *persistentSubscriptionsClient) Update(ctx context.Context, in *UpdateReq, opts ...grpc.CallOption) (*UpdateResp, error) {
	out := new(UpdateResp)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *persistentSubscriptionsClient) Delete(ctx context.Context, in *DeleteReq, opts ...grpc.CallOption) (*DeleteResp, error) {
	out := new(DeleteResp)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *persistentSubscriptionsClient) Read(ctx context.Context, opts ...grpc.CallOption) (PersistentSubscriptions_ReadClient, error) {
	stream, err := c.cc.NewStream(ctx, &PersistentSubscriptions_ServiceDesc.Streams[0], "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Read", opts...)
	if err != nil {
		return nil, err
	}
	x := &persistentSubscriptionsReadClient{stream}
	return x, nil
}

type PersistentSubscriptions_ReadClient interface {
	Send(*ReadReq) error
	Recv() (*ReadResp, error)
	grpc.ClientStream
}

type persistentSubscriptionsReadClient struct {
	grpc.ClientStream
}

func (x *persistentSubscriptionsReadClient) Send(m *ReadReq) error {
	return x.ClientStream.SendMsg(m)
}

func (x *persistentSubscriptionsReadClient) Recv() (*ReadResp, error) {
	m := new(ReadResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *persistentSubscriptionsClient) GetInfo(ctx context.Context, in *GetInfoReq, opts ...grpc.CallOption) (*GetInfoResp, error) {
	out := new(GetInfoResp)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/GetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *persistentSubscriptionsClient) ReplayParked(ctx context.Context, in *ReplayParkedReq, opts ...grpc.CallOption) (*ReplayParkedResp, error) {
	out := new(ReplayParkedResp)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/ReplayParked", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *persistentSubscriptionsClient) List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*ListResp, error) {
	out := new(ListResp)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *persistentSubscriptionsClient) RestartSubsystem(ctx context.Context, in *shared.Empty, opts ...grpc.CallOption) (*shared.Empty, error) {
	out := new(shared.Empty)
	err := c.cc.Invoke(ctx, "/event_store.client.persistent_subscriptions.PersistentSubscriptions/RestartSubsystem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PersistentSubscriptionsServer is the server API for PersistentSubscriptions service.
// All implementations must embed UnimplementedPersistentSubscriptionsServer
// for forward compatibility
type PersistentSubscriptionsServer interface {
	Create(context.Context, *CreateReq) (*CreateResp, error)
	Update(context.Context, *UpdateReq) (*UpdateResp, error)
	Delete(context.Context, *DeleteReq) (*DeleteResp, error)
	Read(PersistentSubscriptions_ReadServer) error
	GetInfo(context.Context, *GetInfoReq) (*GetInfoResp, error)
	ReplayParked(context.Context, *ReplayParkedReq) (*ReplayParkedResp, error)
	List(context.Context, *ListReq) (*ListResp, error)
	RestartSubsystem(context.Context, *shared.Empty) (*shared.Empty, error)
	mustEmbedUnimplementedPersistentSubscriptionsServer()
}

// UnimplementedPersistentSubscriptionsServer must be embedded to have forward compatible implementations.
type UnimplementedPersistentSubscriptionsServer struct {
}

func (UnimplementedPersistentSubscriptionsServer) Create(context.Context, *CreateReq) (*CreateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) Update(context.Context, *UpdateReq) (*UpdateResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) Delete(context.Context, *DeleteReq) (*DeleteResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) Read(PersistentSubscriptions_ReadServer) error {
	return status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) GetInfo(context.Context, *GetInfoReq) (*GetInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) ReplayParked(context.Context, *ReplayParkedReq) (*ReplayParkedResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplayParked not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) List(context.Context, *ListReq) (*ListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) RestartSubsystem(context.Context, *shared.Empty) (*shared.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RestartSubsystem not implemented")
}
func (UnimplementedPersistentSubscriptionsServer) mustEmbedUnimplementedPersistentSubscriptionsServer() {
}

// UnsafePersistentSubscriptionsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PersistentSubscriptionsServer will
// result in compilation errors.
type UnsafePersistentSubscriptionsServer interface {
	mustEmbedUnimplementedPersistentSubscriptionsServer()
}

func RegisterPersistentSubscriptionsServer(s grpc.ServiceRegistrar, srv PersistentSubscriptionsServer) {
	s.RegisterService(&PersistentSubscriptions_ServiceDesc, srv)
}

func _PersistentSubscriptions_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).Create(ctx, req.(*CreateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PersistentSubscriptions_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).Update(ctx, req.(*UpdateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PersistentSubscriptions_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).Delete(ctx, req.(*DeleteReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PersistentSubscriptions_Read_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PersistentSubscriptionsServer).Read(&persistentSubscriptionsReadServer{stream})
}

type PersistentSubscriptions_ReadServer interface {
	Send(*ReadResp) error
	Recv() (*ReadReq, error)
	grpc.ServerStream
}

type persistentSubscriptionsReadServer struct {
	grpc.ServerStream
}

func (x *persistentSubscriptionsReadServer) Send(m *ReadResp) error {
	return x.ServerStream.SendMsg(m)
}

func (x *persistentSubscriptionsReadServer) Recv() (*ReadReq, error) {
	m := new(ReadReq)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _PersistentSubscriptions_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).GetInfo(ctx, req.(*GetInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PersistentSubscriptions_ReplayParked_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplayParkedReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).ReplayParked(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/ReplayParked",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).ReplayParked(ctx, req.(*ReplayParkedReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PersistentSubscriptions_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).List(ctx, req.(*ListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PersistentSubscriptions_RestartSubsystem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(shared.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PersistentSubscriptionsServer).RestartSubsystem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event_store.client.persistent_subscriptions.PersistentSubscriptions/RestartSubsystem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PersistentSubscriptionsServer).RestartSubsystem(ctx, req.(*shared.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// PersistentSubscriptions_ServiceDesc is the grpc.ServiceDesc for PersistentSubscriptions service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PersistentSubscriptions_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "event_store.client.persistent_subscriptions.PersistentSubscriptions",
	HandlerType: (*PersistentSubscriptionsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _PersistentSubscriptions_Create_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _PersistentSubscriptions_Update_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _PersistentSubscriptions_Delete_Handler,
		},
		{
			MethodName: "GetInfo",
			Handler:    _PersistentSubscriptions_GetInfo_Handler,
		},
		{
			MethodName: "ReplayParked",
			Handler:    _PersistentSubscriptions_ReplayParked_Handler,
		},
		{
			MethodName: "List",
			Handler:    _PersistentSubscriptions_List_Handler,
		},
		{
			MethodName: "RestartSubsystem",
			Handler:    _PersistentSubscriptions_RestartSubsystem_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Read",
			Handler:       _PersistentSubscriptions_Read_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "persistent.proto",
}
