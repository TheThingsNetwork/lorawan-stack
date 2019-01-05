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

// Package warning communicates warnings over gRPC headers.
// The Add func is used by the server to add a warning header.
// The client interceptors log warnings to the logger in the context, or to the
// default logger.
//
// Note that headers are currently not supported by ServeHTTP of the gRPC server.
// This means that warnings may not be received by clients using the fallback server.
package warning

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const warning = "warning"

func logWarnings(ctx context.Context, md metadata.MD) {
	if warnings := md.Get(warning); len(warnings) > 0 {
		logger := log.FromContext(ctx)
		if logger == log.Noop {
			logger = log.Default
		}
		for _, warning := range warnings {
			logger.Warn(warning)
		}
	}
}

// Add a warning to the response headers.
func Add(ctx context.Context, message string) {
	grpc.SetHeader(ctx, metadata.Pairs(warning, message)) // nolint:gas
}

// UnaryClientInterceptor is a unary client interceptor that logs warnings sent by the server.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var md metadata.MD
	err := invoker(ctx, method, req, reply, cc, append(opts, grpc.Header(&md))...)
	logWarnings(ctx, md)
	return err
}

// StreamClientInterceptor is a streaming client interceptor that logs warnings sent by the server.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var md metadata.MD
	stream, err := streamer(ctx, desc, cc, method, append(opts, grpc.Header(&md))...)
	logWarnings(ctx, md)
	return stream, err
}
