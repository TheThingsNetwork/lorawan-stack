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
	"net"
	"path"
	"strings"

	"github.com/getsentry/sentry-go"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	sentryerrors "go.thethings.network/lorawan-stack/v3/pkg/errors/sentry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func reportError(ctx context.Context, method string, err error) {
	if err == nil {
		return
	}

	code := codes.Code(errors.Code(err))
	switch code {
	case codes.Unknown,
		codes.Internal,
		codes.DataLoss:
	default:
		return // ignore
	}

	errEvent := sentryerrors.NewEvent(err)

	// Request Tags.
	errEvent.Transaction = method
	if errEvent.Request == nil {
		errEvent.Request = &sentry.Request{}
	}
	errEvent.Request.URL = method
	errEvent.Tags["grpc.service"] = path.Dir(method)[1:]
	errEvent.Tags["grpc.method"] = path.Base(method)
	errEvent.Tags["grpc.code"] = code.String()

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil && p.Addr.String() != "pipe" {
		if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
			errEvent.User.IPAddress = host
		}
	}

	errEvent.Request.Headers = make(map[string]string)

	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		for k, v := range md {
			if len(v) == 0 {
				continue
			}
			switch strings.ToLower(k) {
			case "cookie", "grpc-trace-bin":
				continue // ingored header
			case "authorization":
				parts := strings.SplitN(v[len(v)-1], " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					if tokenType, tokenID, _, err := auth.SplitToken(parts[1]); err == nil {
						errEvent.Tags["auth.token_type"] = tokenType.String()
						errEvent.Tags["auth.token_id"] = tokenID
					}
				}
				continue // ignored header
			case "x-request-id":
				errEvent.Tags["request_id"] = v[len(v)-1]
			}
			errEvent.Request.Headers[k] = strings.Join(v, " ")
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			if len(v) == 0 {
				continue
			}
			switch strings.ToLower(k) {
			case "grpcgateway-authorization",
				"cookie", "grpcgateway-cookie",
				"grpc-trace-bin":
				continue // ingored header
			case "authorization":
				parts := strings.SplitN(v[len(v)-1], " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					if tokenType, tokenID, _, err := auth.SplitToken(parts[1]); err == nil {
						errEvent.Tags["auth.token_type"] = tokenType.String()
						errEvent.Tags["auth.token_id"] = tokenID
					}
				}
				continue // ingored header
			case "x-request-id":
				errEvent.Tags["request_id"] = v[len(v)-1]
			}
			errEvent.Request.Headers[k] = strings.Join(v, " ")
		}
	}

	for k, v := range grpc_ctxtags.Extract(ctx).Values() {
		if strings.HasPrefix(k, "grpc.request.") && (strings.HasSuffix(k, "_id") || strings.HasSuffix(k, "_uid") || strings.HasSuffix(k, "_eui")) {
			k = strings.TrimPrefix(k, "grpc.request.")
		}
		if len(k) > 32 {
			continue
		}
		val := fmt.Sprint(v)
		if len(val) > 200 {
			continue
		}
		errEvent.Tags[k] = val
	}

	// Capture the event.
	sentry.CaptureEvent(errEvent)
}

// UnaryServerInterceptor forwards errors in Unary RPCs to Sentry
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		reportError(ctx, info.FullMethod, err)
		return resp, err
	}
}

// StreamServerInterceptor forwards errors in Stream RPCs to Sentry
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		reportError(ss.Context(), info.FullMethod, err)
		return err
	}
}
