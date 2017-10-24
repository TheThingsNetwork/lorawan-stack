// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpclog

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(ctx context.Context, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	logger := log.FromContext(ctx)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx := newLoggerForCall(ctx, logger, method)
		startTime := time.Now()
		err := invoker(newCtx, method, req, reply, cc, opts...)
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
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(ctx context.Context, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	logger := log.FromContext(ctx)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		newCtx := newLoggerForCall(ctx, logger, method)
		startTime := time.Now()
		clientStream, err := streamer(newCtx, desc, cc, method, opts...)
		if err != nil {
			code := o.codeFunc(err)
			level := o.levelFunc(code)
			entry := log.FromContext(newCtx).WithError(err)
			commit(entry, level, "Failed streaming call")
		}
		if err == nil {
			go func() {
				<-clientStream.Context().Done()
				err := clientStream.Context().Err()
				code := o.codeFunc(err)
				level := o.levelFunc(code)
				entry := log.FromContext(ctx).WithFields(log.Fields(
					"grpc_code", code.String(),
					"duration", time.Since(startTime),
				))
				if err != nil {
					entry = entry.WithError(err)
				}
				commit(entry, level, "Finished streaming call")
			}()
		}
		return clientStream, err
	}
}
