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
	"runtime/trace"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var allowedFieldMaskPaths = make(map[string]map[string]struct{})

// RegisterAllowedFieldMaskPaths registers the allowed field mask paths for an
// RPC. Note that all allowed paths and sub-paths must be registered.
// This function is not safe for concurrent use.
func RegisterAllowedFieldMaskPaths(rpcFullMethod string, allowedPaths ...string) {
	allowedFieldMaskPaths[rpcFullMethod] = make(map[string]struct{})
	for _, allowedPath := range allowedPaths {
		allowedFieldMaskPaths[rpcFullMethod][allowedPath] = struct{}{}
	}
}

var errForbiddenFieldMaskPaths = errors.DefineInvalidArgument("field_mask_paths", "forbidden path(s) in field mask", "forbidden_paths")

func forbiddenPaths(requestedPaths []string, allowedPaths map[string]struct{}) (invalidPaths []string) {
nextRequestedPath:
	for _, requestedPath := range requestedPaths {
		if _, ok := allowedPaths[requestedPath]; ok {
			continue nextRequestedPath
		}
		invalidPaths = append(invalidPaths, requestedPath)
	}
	return
}

func convertError(err error) error {
	if ttnErr, ok := errors.From(err); ok {
		return ttnErr
	}
	return grpc.Errorf(codes.InvalidArgument, err.Error())
}

var errNoValidator = errors.DefineUnimplemented("no_validator", "validator not defined")

func validateMessage(ctx context.Context, fullMethod string, msg interface{}) error {
	if v, ok := msg.(interface {
		GetFieldMask() types.FieldMask
	}); ok {
		region := trace.StartRegion(ctx, "validate field mask")
		if forbiddenPaths := forbiddenPaths(v.GetFieldMask().Paths, allowedFieldMaskPaths[fullMethod]); len(forbiddenPaths) > 0 {
			region.End()
			return errForbiddenFieldMaskPaths.WithAttributes("forbidden_paths", forbiddenPaths)
		}
		region.End()
	}

	switch v := msg.(type) {
	case interface {
		ValidateContext(context.Context) error
	}:
		defer trace.StartRegion(ctx, "validate with context").End()
		if err := v.ValidateContext(ctx); err != nil {
			return convertError(err)
		}
		return nil

	case interface {
		Validate() error
	}:
		defer trace.StartRegion(ctx, "validate without context").End()
		if err := v.Validate(); err != nil {
			return convertError(err)
		}
		return nil

	case interface {
		ValidateFields(...string) error
	}:
		defer trace.StartRegion(ctx, "validate fields").End()
		if err := v.ValidateFields(); err != nil {
			return convertError(err)
		}
		return nil

	case *types.Empty:
		return nil

	default:
		return errNoValidator
	}
}

// UnaryServerInterceptor returns a new unary server interceptor that validates
// incoming messages if those incoming messages implement:
//   (A) ValidateContext(ctx context.Context) error
//   (B) Validate() error
//   (C) ValidateFields(...string) error
// If a message implements both, then (A) should call (B).
//
// Invalid messages will be rejected with the error returned from the validator,
// if that error is a TTN error, or with an `InvalidArgument` if it isn't.
//
// If the RPC's FullPath has a registered list of allowed field mask paths (see
// RegisterAllowedFieldMaskPaths) and the message implements GetFieldMask() types.FieldMask
// then the field mask paths are validated according to the registered list.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validateMessage(ctx, info.FullMethod, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

type recvWrapper struct {
	grpc.ServerStream
	fullMethod string
}

func (s *recvWrapper) RecvMsg(msg interface{}) error {
	if err := s.ServerStream.RecvMsg(msg); err != nil {
		return err
	}
	if err := validateMessage(s.Context(), s.fullMethod, msg); err != nil {
		return err
	}
	return nil
}

// StreamServerInterceptor returns a new streaming server interceptor that validates
// incoming messages if those incoming messages implement:
//   (A) ValidateContext(ctx context.Context) error
//   (B) Validate() error
//   (C) ValidateFields(...string) error
// If a message implements both, then (A) should call (B).
//
// Invalid messages will be rejected with the error returned from the validator,
// if that error is a TTN error, or with an `InvalidArgument` if it isn't.
//
// The stage at which invalid messages will be rejected with `InvalidArgument` varies
// based on the type of the RPC. For `ServerStream` (1:m) requests, it will happen
// before reaching any userspace handlers. For `ClientStream` (n:1) or `BidiStream` (n:m)
// RPCs, the messages will be rejected on calls to `stream.Recv()`.
//
// If the RPC's FullPath has a registered list of allowed field mask paths (see
// RegisterAllowedFieldMaskPaths) and the message implements GetFieldMask() types.FieldMask
// then the field mask paths are validated according to the registered list.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &recvWrapper{
			ServerStream: stream,
			fullMethod:   info.FullMethod,
		})
	}
}
