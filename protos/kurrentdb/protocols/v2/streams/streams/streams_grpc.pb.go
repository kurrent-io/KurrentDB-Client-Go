// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.26.0
// source: kurrentdb/protocols/v2/streams/streams.proto

package streams

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
	StreamsService_MultiStreamAppend_FullMethodName        = "/kurrentdb.protocol.v2.StreamsService/MultiStreamAppend"
	StreamsService_MultiStreamAppendSession_FullMethodName = "/kurrentdb.protocol.v2.StreamsService/MultiStreamAppendSession"
	StreamsService_ReadSession_FullMethodName              = "/kurrentdb.protocol.v2.StreamsService/ReadSession"
)

// StreamsServiceClient is the client API for StreamsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StreamsServiceClient interface {
	// Executes an atomic operation to append records to multiple streams.
	// This transactional method ensures that all appends either succeed
	// completely, or are entirely rolled back, thereby maintaining strict data
	// consistency across all involved streams.
	MultiStreamAppend(ctx context.Context, in *MultiStreamAppendRequest, opts ...grpc.CallOption) (*MultiStreamAppendResponse, error)
	// Streaming version of MultiStreamAppend that allows clients to send multiple
	// append requests over a single connection. When the stream completes, all
	// records are appended transactionally (all succeed or fail together).
	// Provides improved efficiency for high-throughput scenarios while
	// maintaining the same transactional guarantees.
	MultiStreamAppendSession(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[AppendStreamRequest, MultiStreamAppendResponse], error)
	// Retrieve batches of records continuously.
	ReadSession(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ReadResponse], error)
}

type streamsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewStreamsServiceClient(cc grpc.ClientConnInterface) StreamsServiceClient {
	return &streamsServiceClient{cc}
}

func (c *streamsServiceClient) MultiStreamAppend(ctx context.Context, in *MultiStreamAppendRequest, opts ...grpc.CallOption) (*MultiStreamAppendResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MultiStreamAppendResponse)
	err := c.cc.Invoke(ctx, StreamsService_MultiStreamAppend_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamsServiceClient) MultiStreamAppendSession(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[AppendStreamRequest, MultiStreamAppendResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamsService_ServiceDesc.Streams[0], StreamsService_MultiStreamAppendSession_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[AppendStreamRequest, MultiStreamAppendResponse]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamsService_MultiStreamAppendSessionClient = grpc.ClientStreamingClient[AppendStreamRequest, MultiStreamAppendResponse]

func (c *streamsServiceClient) ReadSession(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ReadResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamsService_ServiceDesc.Streams[1], StreamsService_ReadSession_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[ReadRequest, ReadResponse]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamsService_ReadSessionClient = grpc.ServerStreamingClient[ReadResponse]

// StreamsServiceServer is the server API for StreamsService service.
// All implementations must embed UnimplementedStreamsServiceServer
// for forward compatibility.
type StreamsServiceServer interface {
	// Executes an atomic operation to append records to multiple streams.
	// This transactional method ensures that all appends either succeed
	// completely, or are entirely rolled back, thereby maintaining strict data
	// consistency across all involved streams.
	MultiStreamAppend(context.Context, *MultiStreamAppendRequest) (*MultiStreamAppendResponse, error)
	// Streaming version of MultiStreamAppend that allows clients to send multiple
	// append requests over a single connection. When the stream completes, all
	// records are appended transactionally (all succeed or fail together).
	// Provides improved efficiency for high-throughput scenarios while
	// maintaining the same transactional guarantees.
	MultiStreamAppendSession(grpc.ClientStreamingServer[AppendStreamRequest, MultiStreamAppendResponse]) error
	// Retrieve batches of records continuously.
	ReadSession(*ReadRequest, grpc.ServerStreamingServer[ReadResponse]) error
	mustEmbedUnimplementedStreamsServiceServer()
}

// UnimplementedStreamsServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStreamsServiceServer struct{}

func (UnimplementedStreamsServiceServer) MultiStreamAppend(context.Context, *MultiStreamAppendRequest) (*MultiStreamAppendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MultiStreamAppend not implemented")
}
func (UnimplementedStreamsServiceServer) MultiStreamAppendSession(grpc.ClientStreamingServer[AppendStreamRequest, MultiStreamAppendResponse]) error {
	return status.Errorf(codes.Unimplemented, "method MultiStreamAppendSession not implemented")
}
func (UnimplementedStreamsServiceServer) ReadSession(*ReadRequest, grpc.ServerStreamingServer[ReadResponse]) error {
	return status.Errorf(codes.Unimplemented, "method ReadSession not implemented")
}
func (UnimplementedStreamsServiceServer) mustEmbedUnimplementedStreamsServiceServer() {}
func (UnimplementedStreamsServiceServer) testEmbeddedByValue()                        {}

// UnsafeStreamsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StreamsServiceServer will
// result in compilation errors.
type UnsafeStreamsServiceServer interface {
	mustEmbedUnimplementedStreamsServiceServer()
}

func RegisterStreamsServiceServer(s grpc.ServiceRegistrar, srv StreamsServiceServer) {
	// If the following call pancis, it indicates UnimplementedStreamsServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&StreamsService_ServiceDesc, srv)
}

func _StreamsService_MultiStreamAppend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MultiStreamAppendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamsServiceServer).MultiStreamAppend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: StreamsService_MultiStreamAppend_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamsServiceServer).MultiStreamAppend(ctx, req.(*MultiStreamAppendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamsService_MultiStreamAppendSession_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(StreamsServiceServer).MultiStreamAppendSession(&grpc.GenericServerStream[AppendStreamRequest, MultiStreamAppendResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamsService_MultiStreamAppendSessionServer = grpc.ClientStreamingServer[AppendStreamRequest, MultiStreamAppendResponse]

func _StreamsService_ReadSession_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ReadRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamsServiceServer).ReadSession(m, &grpc.GenericServerStream[ReadRequest, ReadResponse]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamsService_ReadSessionServer = grpc.ServerStreamingServer[ReadResponse]

// StreamsService_ServiceDesc is the grpc.ServiceDesc for StreamsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StreamsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kurrentdb.protocol.v2.StreamsService",
	HandlerType: (*StreamsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MultiStreamAppend",
			Handler:    _StreamsService_MultiStreamAppend_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "MultiStreamAppendSession",
			Handler:       _StreamsService_MultiStreamAppendSession_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "ReadSession",
			Handler:       _StreamsService_ReadSession_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "kurrentdb/protocols/v2/streams/streams.proto",
}
