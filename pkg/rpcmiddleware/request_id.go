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

package rpcmiddleware

import (
	"context"
	"crypto/rand"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	ulid "github.com/oklog/ulid/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const requestIDKey = "request-id"

// RequestIDUnaryServerInterceptor returns a new unary server interceptor that inserts Request IDs if not present.
func RequestIDUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		if len(md.Get(requestIDKey)) == 0 {
			md.Set(requestIDKey, ulid.MustNew(ulid.Now(), rand.Reader).String())
			ctx = metadata.NewIncomingContext(ctx, md)
		}
		if err := grpc.SetHeader(ctx, metadata.Pairs(requestIDKey, md.Get(requestIDKey)[0])); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// RequestIDStreamServerInterceptor returns a new streaming server interceptor that that inserts Request IDs if not present.
func RequestIDStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		wrapped := grpc_middleware.WrapServerStream(stream)
		md, _ := metadata.FromIncomingContext(ctx)
		if len(md.Get(requestIDKey)) == 0 {
			md.Set(requestIDKey, ulid.MustNew(ulid.Now(), rand.Reader).String())
			wrapped.WrappedContext = metadata.NewIncomingContext(ctx, md)
		}
		if err := grpc.SetHeader(wrapped.WrappedContext, metadata.Pairs(requestIDKey, md.Get(requestIDKey)[0])); err != nil {
			return err
		}
		return handler(srv, wrapped)
	}
}
