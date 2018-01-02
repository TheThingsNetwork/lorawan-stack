// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpcerrors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryClientInterceptor converts gRPC errors to regular errors.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			err = FromGRPC(err)
		}
		return err
	}
}

type wrappedStream struct {
	grpc.ClientStream
}

func (w wrappedStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)
	if err != nil {
		err = FromGRPC(err)
	}
	return err
}
func (w wrappedStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if err != nil {
		err = FromGRPC(err)
	}
	return err
}

// StreamClientInterceptor converts gRPC errors to regular errors.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, FromGRPC(err)
		}
		return wrappedStream{s}, nil
	}
}
