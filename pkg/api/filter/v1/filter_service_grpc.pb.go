// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: filter/v1/filter_service.proto

package filterv1

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
	FilterService_SetFilter_FullMethodName = "/filter.v1.FilterService/SetFilter"
)

// FilterServiceClient is the client API for FilterService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FilterServiceClient interface {
	SetFilter(ctx context.Context, in *SetFilterRequest, opts ...grpc.CallOption) (*SetFilterResponse, error)
}

type filterServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFilterServiceClient(cc grpc.ClientConnInterface) FilterServiceClient {
	return &filterServiceClient{cc}
}

func (c *filterServiceClient) SetFilter(ctx context.Context, in *SetFilterRequest, opts ...grpc.CallOption) (*SetFilterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetFilterResponse)
	err := c.cc.Invoke(ctx, FilterService_SetFilter_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FilterServiceServer is the server API for FilterService service.
// All implementations should embed UnimplementedFilterServiceServer
// for forward compatibility.
type FilterServiceServer interface {
	SetFilter(context.Context, *SetFilterRequest) (*SetFilterResponse, error)
}

// UnimplementedFilterServiceServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedFilterServiceServer struct{}

func (UnimplementedFilterServiceServer) SetFilter(context.Context, *SetFilterRequest) (*SetFilterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetFilter not implemented")
}
func (UnimplementedFilterServiceServer) testEmbeddedByValue() {}

// UnsafeFilterServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FilterServiceServer will
// result in compilation errors.
type UnsafeFilterServiceServer interface {
	mustEmbedUnimplementedFilterServiceServer()
}

func RegisterFilterServiceServer(s grpc.ServiceRegistrar, srv FilterServiceServer) {
	// If the following call pancis, it indicates UnimplementedFilterServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&FilterService_ServiceDesc, srv)
}

func _FilterService_SetFilter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetFilterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FilterServiceServer).SetFilter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FilterService_SetFilter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FilterServiceServer).SetFilter(ctx, req.(*SetFilterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FilterService_ServiceDesc is the grpc.ServiceDesc for FilterService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FilterService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "filter.v1.FilterService",
	HandlerType: (*FilterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetFilter",
			Handler:    _FilterService_SetFilter_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "filter/v1/filter_service.proto",
}
