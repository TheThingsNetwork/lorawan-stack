// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpclog

import (
	"time"

	"context"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds the logger from the global context to the call context.
func UnaryServerInterceptor(ctx context.Context, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	logger := log.FromContext(ctx)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, logger, info.FullMethod)
		startTime := time.Now()
		resp, err := handler(newCtx, req)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		entry := log.FromContext(newCtx).WithFields(log.Fields(
			"grpc_code", code.String(),
			"duration", time.Since(startTime),
		))
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
	logger := log.FromContext(ctx)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		startTime := time.Now()
		err := handler(srv, wrapped)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		entry := log.FromContext(newCtx).WithFields(log.Fields(
			"grpc_code", code.String(),
			"duration", time.Since(startTime),
		))
		if err != nil {
			entry = entry.WithError(err)
		}
		commit(entry, level, "Finished streaming call")
		return err
	}
}
