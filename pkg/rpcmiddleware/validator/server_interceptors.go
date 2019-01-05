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

// Based on github.com/grpc-ecosystem/go-grpc-middleware/validator, which is
// Copyright 2016 Michal Witkowski and licensed under the Apache 2.0 License.

// Package validator implements a gRPC middleware that defines custom validators that are invoked when the service is called.
package validator

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type validatorWithContext interface {
	ValidateContext(ctx context.Context) error
}

type validator interface {
	Validate() error
}

func convertError(err error) error {
	if ttnErr, ok := errors.From(err); ok {
		return ttnErr
	}
	return grpc.Errorf(codes.InvalidArgument, err.Error())
}

// UnaryServerInterceptor returns a new unary server interceptor that validates
// incoming messages if those incoming messages implement:
//   (A) ValidateContext(ctx context.Context) error
//   (B) Validate() error
// If a message implements both, then (A) should call (B).
//
// Invalid messages will be rejected with the error returned from the validator,
// if that error is a TTN error, or with an `InvalidArgument` if it isn't.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validatorWithContext); ok {
			if err := v.ValidateContext(ctx); err != nil {
				return nil, convertError(err)
			}
		} else if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				return nil, convertError(err)
			}
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that validates
// incoming messages if those incoming messages implement:
//   (A) ValidateContext(ctx context.Context) error
//   (B) Validate() error
// If a message implements both, then (A) should call (B).
//
// Invalid messages will be rejected with the error returned from the validator,
// if that error is a TTN error, or with an `InvalidArgument` if it isn't.
//
// The stage at which invalid messages will be rejected with `InvalidArgument` varies
// based on the type of the RPC. For `ServerStream` (1:m) requests, it will happen
// before reaching any userspace handlers. For `ClientStream` (n:1) or `BidiStream` (n:m)
// RPCs, the messages will be rejected on calls to `stream.Recv()`.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &recvWrapper{stream}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
}

func (s *recvWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if v, ok := m.(validatorWithContext); ok {
		if err := v.ValidateContext(s.Context()); err != nil {
			return convertError(err)
		}
	} else if v, ok := m.(validator); ok {
		if err := v.Validate(); err != nil {
			return convertError(err)
		}
	}
	return nil
}
