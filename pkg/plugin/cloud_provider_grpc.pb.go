// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: cloud_provider.proto

package plugin

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
	CloudProvider_GetInstanceInfo_FullMethodName    = "/plugin.CloudProvider/GetInstanceInfo"
	CloudProvider_StopInstance_FullMethodName       = "/plugin.CloudProvider/StopInstance"
	CloudProvider_StartInstance_FullMethodName      = "/plugin.CloudProvider/StartInstance"
	CloudProvider_GetProviderName_FullMethodName    = "/plugin.CloudProvider/GetProviderName"
	CloudProvider_GetProviderVersion_FullMethodName = "/plugin.CloudProvider/GetProviderVersion"
)

// CloudProviderClient is the client API for CloudProvider service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CloudProviderClient interface {
	GetInstanceInfo(ctx context.Context, in *GetInstanceInfoRequest, opts ...grpc.CallOption) (*GetInstanceInfoResponse, error)
	StopInstance(ctx context.Context, in *StopInstanceRequest, opts ...grpc.CallOption) (*StopInstanceResponse, error)
	StartInstance(ctx context.Context, in *StartInstanceRequest, opts ...grpc.CallOption) (*StartInstanceResponse, error)
	GetProviderName(ctx context.Context, in *GetProviderNameRequest, opts ...grpc.CallOption) (*GetProviderNameResponse, error)
	GetProviderVersion(ctx context.Context, in *GetProviderVersionRequest, opts ...grpc.CallOption) (*GetProviderVersionResponse, error)
}

type cloudProviderClient struct {
	cc grpc.ClientConnInterface
}

func NewCloudProviderClient(cc grpc.ClientConnInterface) CloudProviderClient {
	return &cloudProviderClient{cc}
}

func (c *cloudProviderClient) GetInstanceInfo(ctx context.Context, in *GetInstanceInfoRequest, opts ...grpc.CallOption) (*GetInstanceInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetInstanceInfoResponse)
	err := c.cc.Invoke(ctx, CloudProvider_GetInstanceInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudProviderClient) StopInstance(ctx context.Context, in *StopInstanceRequest, opts ...grpc.CallOption) (*StopInstanceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StopInstanceResponse)
	err := c.cc.Invoke(ctx, CloudProvider_StopInstance_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudProviderClient) StartInstance(ctx context.Context, in *StartInstanceRequest, opts ...grpc.CallOption) (*StartInstanceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StartInstanceResponse)
	err := c.cc.Invoke(ctx, CloudProvider_StartInstance_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudProviderClient) GetProviderName(ctx context.Context, in *GetProviderNameRequest, opts ...grpc.CallOption) (*GetProviderNameResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetProviderNameResponse)
	err := c.cc.Invoke(ctx, CloudProvider_GetProviderName_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudProviderClient) GetProviderVersion(ctx context.Context, in *GetProviderVersionRequest, opts ...grpc.CallOption) (*GetProviderVersionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetProviderVersionResponse)
	err := c.cc.Invoke(ctx, CloudProvider_GetProviderVersion_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CloudProviderServer is the server API for CloudProvider service.
// All implementations must embed UnimplementedCloudProviderServer
// for forward compatibility.
type CloudProviderServer interface {
	GetInstanceInfo(context.Context, *GetInstanceInfoRequest) (*GetInstanceInfoResponse, error)
	StopInstance(context.Context, *StopInstanceRequest) (*StopInstanceResponse, error)
	StartInstance(context.Context, *StartInstanceRequest) (*StartInstanceResponse, error)
	GetProviderName(context.Context, *GetProviderNameRequest) (*GetProviderNameResponse, error)
	GetProviderVersion(context.Context, *GetProviderVersionRequest) (*GetProviderVersionResponse, error)
	mustEmbedUnimplementedCloudProviderServer()
}

// UnimplementedCloudProviderServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCloudProviderServer struct{}

func (UnimplementedCloudProviderServer) GetInstanceInfo(context.Context, *GetInstanceInfoRequest) (*GetInstanceInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstanceInfo not implemented")
}
func (UnimplementedCloudProviderServer) StopInstance(context.Context, *StopInstanceRequest) (*StopInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopInstance not implemented")
}
func (UnimplementedCloudProviderServer) StartInstance(context.Context, *StartInstanceRequest) (*StartInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartInstance not implemented")
}
func (UnimplementedCloudProviderServer) GetProviderName(context.Context, *GetProviderNameRequest) (*GetProviderNameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProviderName not implemented")
}
func (UnimplementedCloudProviderServer) GetProviderVersion(context.Context, *GetProviderVersionRequest) (*GetProviderVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProviderVersion not implemented")
}
func (UnimplementedCloudProviderServer) mustEmbedUnimplementedCloudProviderServer() {}
func (UnimplementedCloudProviderServer) testEmbeddedByValue()                       {}

// UnsafeCloudProviderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CloudProviderServer will
// result in compilation errors.
type UnsafeCloudProviderServer interface {
	mustEmbedUnimplementedCloudProviderServer()
}

func RegisterCloudProviderServer(s grpc.ServiceRegistrar, srv CloudProviderServer) {
	// If the following call pancis, it indicates UnimplementedCloudProviderServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CloudProvider_ServiceDesc, srv)
}

func _CloudProvider_GetInstanceInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInstanceInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudProviderServer).GetInstanceInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CloudProvider_GetInstanceInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudProviderServer).GetInstanceInfo(ctx, req.(*GetInstanceInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudProvider_StopInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudProviderServer).StopInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CloudProvider_StopInstance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudProviderServer).StopInstance(ctx, req.(*StopInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudProvider_StartInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudProviderServer).StartInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CloudProvider_StartInstance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudProviderServer).StartInstance(ctx, req.(*StartInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudProvider_GetProviderName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProviderNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudProviderServer).GetProviderName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CloudProvider_GetProviderName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudProviderServer).GetProviderName(ctx, req.(*GetProviderNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudProvider_GetProviderVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProviderVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudProviderServer).GetProviderVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CloudProvider_GetProviderVersion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudProviderServer).GetProviderVersion(ctx, req.(*GetProviderVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CloudProvider_ServiceDesc is the grpc.ServiceDesc for CloudProvider service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CloudProvider_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "plugin.CloudProvider",
	HandlerType: (*CloudProviderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetInstanceInfo",
			Handler:    _CloudProvider_GetInstanceInfo_Handler,
		},
		{
			MethodName: "StopInstance",
			Handler:    _CloudProvider_StopInstance_Handler,
		},
		{
			MethodName: "StartInstance",
			Handler:    _CloudProvider_StartInstance_Handler,
		},
		{
			MethodName: "GetProviderName",
			Handler:    _CloudProvider_GetProviderName_Handler,
		},
		{
			MethodName: "GetProviderVersion",
			Handler:    _CloudProvider_GetProviderVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cloud_provider.proto",
}
