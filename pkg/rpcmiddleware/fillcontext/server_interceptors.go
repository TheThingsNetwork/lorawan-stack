// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package fillcontext implements a gRPC middleware that fills global context into a call context
package fillcontext

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// Filler extends the context
type Filler func(context.Context) context.Context

// UnaryServerInterceptor returns a new unary server interceptor that modifies the context.
func UnaryServerInterceptor(fillers ...Filler) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if fillers != nil {
			for _, fill := range fillers {
				ctx = fill(ctx)
			}
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that that modifies the context.
func StreamServerInterceptor(fillers ...Filler) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if fillers != nil {
			ctx := stream.Context()
			for _, fill := range fillers {
				ctx = fill(ctx)
			}
			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = ctx
			return handler(srv, wrapped)
		}
		return handler(srv, stream)
	}
}
