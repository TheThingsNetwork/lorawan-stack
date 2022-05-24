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

package events

import (
	"context"
	"fmt"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const requestIDKey = "x-request-id"

func extractOrGenerateCorrelationID(ctx context.Context, fullMethod string) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	var id string
	if xRequestID := md.Get(requestIDKey); len(xRequestID) > 0 {
		id = xRequestID[len(xRequestID)-1]
	} else {
		id = NewCorrelationID()
	}
	ctx = ContextWithCorrelationID(ctx, fmt.Sprintf("rpc:%s:%s", fullMethod, id))
	return ctx
}

// UnaryServerInterceptor returns a new unary server interceptor
// that modifies the context to include a correlation ID.
func UnaryServerInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	ctx = extractOrGenerateCorrelationID(ctx, info.FullMethod)
	return handler(ctx, req)
}

// StreamServerInterceptor returns a new streaming server interceptor
// that that modifies the context.
func StreamServerInterceptor(
	srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	wrapped := grpc_middleware.WrapServerStream(stream)
	wrapped.WrappedContext = extractOrGenerateCorrelationID(stream.Context(), info.FullMethod)
	return handler(srv, wrapped)
}
