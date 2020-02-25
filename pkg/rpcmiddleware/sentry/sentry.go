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

// Package sentry implements gRPC middleware that forwards errors in RPCs to Sentry
package sentry

import (
	"context"
	"fmt"
	"strings"

	"github.com/getsentry/sentry-go"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.thethings.network/lorawan-stack/pkg/errors"

	sentryerrors "go.thethings.network/lorawan-stack/pkg/errors/sentry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func reportError(ctx context.Context, method string, err error) {
	if err == nil {
		return
	}

	code := codes.Code(errors.Code(err))
	switch code {
	case codes.Unknown,
		codes.Internal,
		codes.Unimplemented,
		codes.DataLoss:
	default:
		return // ignore
	}

	errEvent := sentryerrors.NewEvent(err)

	// Request Tags.
	errEvent.Transaction = method
	errEvent.Request.URL = method
	errEvent.Request.Headers = make(map[string]string)
	errEvent.Tags["grpc.method"] = method
	errEvent.Tags["grpc.code"] = code.String()
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if requestID := md["request-id"]; len(requestID) > 0 {
			errEvent.Tags["grpc.request_id"] = requestID[0]
		}
		for k, v := range md {
			if k == "grpc-trace-bin" {
				continue
			}
			errEvent.Request.Headers[k] = strings.Join(v, " ")
		}
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID := md["request-id"]; len(requestID) > 0 {
			errEvent.Tags["grpc.request_id"] = requestID[0]
		}
		for k, v := range md {
			if k == "grpc-trace-bin" {
				continue
			}
			errEvent.Request.Headers[k] = strings.Join(v, " ")
		}
	}
	for k, v := range grpc_ctxtags.Extract(ctx).Values() {
		if val := fmt.Sprint(v); len(val) < 64 {
			errEvent.Tags["tag."+k] = val
		}
	}

	// Capture the event.
	sentry.CaptureEvent(errEvent)
}

// UnaryServerInterceptor forwards errors in Unary RPCs to Sentry
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		reportError(ctx, info.FullMethod, err)
		return
	}
}

// StreamServerInterceptor forwards errors in Stream RPCs to Sentry
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		err = handler(srv, ss)
		reportError(ss.Context(), info.FullMethod, err)
		return
	}
}
