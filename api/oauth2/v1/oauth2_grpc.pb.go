// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

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

// Oauth2Client is the client API for Oauth2 service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type Oauth2Client interface {
	IssueOauth2Token(ctx context.Context, in *IssueOauth2TokenRequest, opts ...grpc.CallOption) (*IssueOauth2TokenResponse, error)
	AddWhiteList(ctx context.Context, in *AddWhiteListRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type oauth2Client struct {
	cc grpc.ClientConnInterface
}

func NewOauth2Client(cc grpc.ClientConnInterface) Oauth2Client {
	return &oauth2Client{cc}
}

func (c *oauth2Client) IssueOauth2Token(ctx context.Context, in *IssueOauth2TokenRequest, opts ...grpc.CallOption) (*IssueOauth2TokenResponse, error) {
	out := new(IssueOauth2TokenResponse)
	err := c.cc.Invoke(ctx, "/api.oauth2.v1.Oauth2/IssueOauth2Token", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oauth2Client) AddWhiteList(ctx context.Context, in *AddWhiteListRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api.oauth2.v1.Oauth2/AddWhiteList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Oauth2Server is the server API for Oauth2 service.
// All implementations must embed UnimplementedOauth2Server
// for forward compatibility
type Oauth2Server interface {
	IssueOauth2Token(context.Context, *IssueOauth2TokenRequest) (*IssueOauth2TokenResponse, error)
	AddWhiteList(context.Context, *AddWhiteListRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedOauth2Server()
}

// UnimplementedOauth2Server must be embedded to have forward compatible implementations.
type UnimplementedOauth2Server struct {
}

func (UnimplementedOauth2Server) IssueOauth2Token(context.Context, *IssueOauth2TokenRequest) (*IssueOauth2TokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IssueOauth2Token not implemented")
}
func (UnimplementedOauth2Server) AddWhiteList(context.Context, *AddWhiteListRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddWhiteList not implemented")
}
func (UnimplementedOauth2Server) mustEmbedUnimplementedOauth2Server() {}

// UnsafeOauth2Server may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to Oauth2Server will
// result in compilation errors.
type UnsafeOauth2Server interface {
	mustEmbedUnimplementedOauth2Server()
}

func RegisterOauth2Server(s grpc.ServiceRegistrar, srv Oauth2Server) {
	s.RegisterService(&Oauth2_ServiceDesc, srv)
}

func _Oauth2_IssueOauth2Token_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssueOauth2TokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Oauth2Server).IssueOauth2Token(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.oauth2.v1.Oauth2/IssueOauth2Token",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Oauth2Server).IssueOauth2Token(ctx, req.(*IssueOauth2TokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Oauth2_AddWhiteList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddWhiteListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Oauth2Server).AddWhiteList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.oauth2.v1.Oauth2/AddWhiteList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Oauth2Server).AddWhiteList(ctx, req.(*AddWhiteListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Oauth2_ServiceDesc is the grpc.ServiceDesc for Oauth2 service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Oauth2_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.oauth2.v1.Oauth2",
	HandlerType: (*Oauth2Server)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IssueOauth2Token",
			Handler:    _Oauth2_IssueOauth2Token_Handler,
		},
		{
			MethodName: "AddWhiteList",
			Handler:    _Oauth2_AddWhiteList_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/oauth2/v1/oauth2.proto",
}
