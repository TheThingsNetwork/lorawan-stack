// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
