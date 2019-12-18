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

package rpclog

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.thethings.network/lorawan-stack/pkg/log"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds the logger from the global context to the call context.
func UnaryServerInterceptor(ctx context.Context, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	logger := log.FromContext(ctx).WithField("namespace", "grpc")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, logger, info.FullMethod)
		startTime := time.Now()
		resp, err := handler(newCtx, req)
		logFields := []interface{}{
			"duration", time.Since(startTime),
		}
		if err != nil {
			logFields = append(logFields, logFieldsForError(err)...)
		}
		level := o.levelFunc(grpc.Code(err))
		if err == context.Canceled {
			level = log.InfoLevel
		}
		entry := log.FromContext(newCtx).WithFields(log.Fields(logFields...))
		if err != nil {
			entry = entry.WithError(err)
		}
		commit(entry, level, "Finished unary call")
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds the logger from the global context to the call context.
func StreamServerInterceptor(ctx context.Context, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	logger := log.FromContext(ctx).WithField("namespace", "grpc")
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		startTime := time.Now()
		err := handler(srv, wrapped)
		logFields := []interface{}{
			"duration", time.Since(startTime),
		}
		if err != nil {
			logFields = append(logFields, logFieldsForError(err)...)
		}
		level := o.levelFunc(grpc.Code(err))
		if err == context.Canceled {
			level = log.InfoLevel
		}
		entry := log.FromContext(newCtx).WithFields(log.Fields(logFields...))
		if err != nil {
			entry = entry.WithError(err)
		}
		commit(entry, level, "Finished streaming call")
		return err
	}
}
