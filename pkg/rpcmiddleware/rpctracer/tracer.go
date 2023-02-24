// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package rpctracer implements a gRPC tracing middleware.
package rpctracer

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracer"
	"google.golang.org/grpc"
)

// TracerHook is the name of the namespace hook.
const TracerHook = "tracer"

// UnaryTracerHook adds the tracer to the context of the unary call.
func UnaryTracerHook(name string, opts ...trace.TracerOption) hooks.UnaryHandlerMiddleware {
	return func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = tracer.NewContextWithTracer(ctx, name, opts...)
			return h(ctx, req)
		}
	}
}

// StreamTracerHook adds the tracer to the context of the stream.
func StreamTracerHook(name string, opts ...trace.TracerOption) hooks.StreamHandlerMiddleware {
	return func(h grpc.StreamHandler) grpc.StreamHandler {
		return func(srv interface{}, stream grpc.ServerStream) error {
			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = tracer.NewContextWithTracer(stream.Context(), name, opts...)
			return h(srv, wrapped)
		}
	}
}
