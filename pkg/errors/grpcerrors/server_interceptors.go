// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpcerrors

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor converts errors to gRPC errors.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		res, err := handler(ctx, req)
		if err != nil {
			err = ToGRPC(err)
		}
		return res, err
	}
}

// StreamServerInterceptor converts errors to gRPC errors.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil {
			err = ToGRPC(err)
		}
		return err
	}
}
