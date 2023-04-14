// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.22.2
// source: lorawan-stack/api/oauth_services.proto

package ttnpb

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
	OAuthAuthorizationRegistry_List_FullMethodName        = "/ttn.lorawan.v3.OAuthAuthorizationRegistry/List"
	OAuthAuthorizationRegistry_ListTokens_FullMethodName  = "/ttn.lorawan.v3.OAuthAuthorizationRegistry/ListTokens"
	OAuthAuthorizationRegistry_Delete_FullMethodName      = "/ttn.lorawan.v3.OAuthAuthorizationRegistry/Delete"
	OAuthAuthorizationRegistry_DeleteToken_FullMethodName = "/ttn.lorawan.v3.OAuthAuthorizationRegistry/DeleteToken"
)

// OAuthAuthorizationRegistryClient is the client API for OAuthAuthorizationRegistry service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OAuthAuthorizationRegistryClient interface {
	// List OAuth clients that are authorized by the user.
	List(ctx context.Context, in *ListOAuthClientAuthorizationsRequest, opts ...grpc.CallOption) (*OAuthClientAuthorizations, error)
	// List OAuth access tokens issued to the OAuth client on behalf of the user.
	ListTokens(ctx context.Context, in *ListOAuthAccessTokensRequest, opts ...grpc.CallOption) (*OAuthAccessTokens, error)
	// Delete (de-authorize) an OAuth client for the user.
	Delete(ctx context.Context, in *OAuthClientAuthorizationIdentifiers, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Delete (invalidate) an OAuth access token.
	DeleteToken(ctx context.Context, in *OAuthAccessTokenIdentifiers, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type oAuthAuthorizationRegistryClient struct {
	cc grpc.ClientConnInterface
}

func NewOAuthAuthorizationRegistryClient(cc grpc.ClientConnInterface) OAuthAuthorizationRegistryClient {
	return &oAuthAuthorizationRegistryClient{cc}
}

func (c *oAuthAuthorizationRegistryClient) List(ctx context.Context, in *ListOAuthClientAuthorizationsRequest, opts ...grpc.CallOption) (*OAuthClientAuthorizations, error) {
	out := new(OAuthClientAuthorizations)
	err := c.cc.Invoke(ctx, OAuthAuthorizationRegistry_List_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oAuthAuthorizationRegistryClient) ListTokens(ctx context.Context, in *ListOAuthAccessTokensRequest, opts ...grpc.CallOption) (*OAuthAccessTokens, error) {
	out := new(OAuthAccessTokens)
	err := c.cc.Invoke(ctx, OAuthAuthorizationRegistry_ListTokens_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oAuthAuthorizationRegistryClient) Delete(ctx context.Context, in *OAuthClientAuthorizationIdentifiers, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, OAuthAuthorizationRegistry_Delete_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oAuthAuthorizationRegistryClient) DeleteToken(ctx context.Context, in *OAuthAccessTokenIdentifiers, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, OAuthAuthorizationRegistry_DeleteToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OAuthAuthorizationRegistryServer is the server API for OAuthAuthorizationRegistry service.
// All implementations must embed UnimplementedOAuthAuthorizationRegistryServer
// for forward compatibility
type OAuthAuthorizationRegistryServer interface {
	// List OAuth clients that are authorized by the user.
	List(context.Context, *ListOAuthClientAuthorizationsRequest) (*OAuthClientAuthorizations, error)
	// List OAuth access tokens issued to the OAuth client on behalf of the user.
	ListTokens(context.Context, *ListOAuthAccessTokensRequest) (*OAuthAccessTokens, error)
	// Delete (de-authorize) an OAuth client for the user.
	Delete(context.Context, *OAuthClientAuthorizationIdentifiers) (*emptypb.Empty, error)
	// Delete (invalidate) an OAuth access token.
	DeleteToken(context.Context, *OAuthAccessTokenIdentifiers) (*emptypb.Empty, error)
	mustEmbedUnimplementedOAuthAuthorizationRegistryServer()
}

// UnimplementedOAuthAuthorizationRegistryServer must be embedded to have forward compatible implementations.
type UnimplementedOAuthAuthorizationRegistryServer struct {
}

func (UnimplementedOAuthAuthorizationRegistryServer) List(context.Context, *ListOAuthClientAuthorizationsRequest) (*OAuthClientAuthorizations, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedOAuthAuthorizationRegistryServer) ListTokens(context.Context, *ListOAuthAccessTokensRequest) (*OAuthAccessTokens, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTokens not implemented")
}
func (UnimplementedOAuthAuthorizationRegistryServer) Delete(context.Context, *OAuthClientAuthorizationIdentifiers) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedOAuthAuthorizationRegistryServer) DeleteToken(context.Context, *OAuthAccessTokenIdentifiers) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteToken not implemented")
}
func (UnimplementedOAuthAuthorizationRegistryServer) mustEmbedUnimplementedOAuthAuthorizationRegistryServer() {
}

// UnsafeOAuthAuthorizationRegistryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OAuthAuthorizationRegistryServer will
// result in compilation errors.
type UnsafeOAuthAuthorizationRegistryServer interface {
	mustEmbedUnimplementedOAuthAuthorizationRegistryServer()
}

func RegisterOAuthAuthorizationRegistryServer(s grpc.ServiceRegistrar, srv OAuthAuthorizationRegistryServer) {
	s.RegisterService(&OAuthAuthorizationRegistry_ServiceDesc, srv)
}

func _OAuthAuthorizationRegistry_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOAuthClientAuthorizationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OAuthAuthorizationRegistryServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OAuthAuthorizationRegistry_List_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OAuthAuthorizationRegistryServer).List(ctx, req.(*ListOAuthClientAuthorizationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OAuthAuthorizationRegistry_ListTokens_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOAuthAccessTokensRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OAuthAuthorizationRegistryServer).ListTokens(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OAuthAuthorizationRegistry_ListTokens_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OAuthAuthorizationRegistryServer).ListTokens(ctx, req.(*ListOAuthAccessTokensRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OAuthAuthorizationRegistry_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OAuthClientAuthorizationIdentifiers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OAuthAuthorizationRegistryServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OAuthAuthorizationRegistry_Delete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OAuthAuthorizationRegistryServer).Delete(ctx, req.(*OAuthClientAuthorizationIdentifiers))
	}
	return interceptor(ctx, in, info, handler)
}

func _OAuthAuthorizationRegistry_DeleteToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OAuthAccessTokenIdentifiers)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OAuthAuthorizationRegistryServer).DeleteToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OAuthAuthorizationRegistry_DeleteToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OAuthAuthorizationRegistryServer).DeleteToken(ctx, req.(*OAuthAccessTokenIdentifiers))
	}
	return interceptor(ctx, in, info, handler)
}

// OAuthAuthorizationRegistry_ServiceDesc is the grpc.ServiceDesc for OAuthAuthorizationRegistry service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OAuthAuthorizationRegistry_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ttn.lorawan.v3.OAuthAuthorizationRegistry",
	HandlerType: (*OAuthAuthorizationRegistryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _OAuthAuthorizationRegistry_List_Handler,
		},
		{
			MethodName: "ListTokens",
			Handler:    _OAuthAuthorizationRegistry_ListTokens_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _OAuthAuthorizationRegistry_Delete_Handler,
		},
		{
			MethodName: "DeleteToken",
			Handler:    _OAuthAuthorizationRegistry_DeleteToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lorawan-stack/api/oauth_services.proto",
}
