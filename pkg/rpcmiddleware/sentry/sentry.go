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

	raven "github.com/getsentry/raven-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/sentry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type forwarder struct {
	client *raven.Client
}

func (fw *forwarder) forward(ctx context.Context, method string, err error) {
	if err == nil {
		return
	}

	code := codes.Code(errors.Code(err))
	switch code {
	case codes.Canceled,
		codes.InvalidArgument,
		codes.DeadlineExceeded,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated:
		return // ignore
	}

	// Request Tags
	var tags = map[string]string{
		"grpc.method": method,
		"grpc.code":   code.String(),
	}
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if requestID := md["request-id"]; len(requestID) > 0 {
			tags["grpc.request_id"] = requestID[0]
		}
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID := md["request-id"]; len(requestID) > 0 {
			tags["grpc.request_id"] = requestID[0]
		}
	}
	for k, v := range grpc_ctxtags.Extract(ctx).Values() {
		if val := fmt.Sprint(v); len(val) < 64 {
			tags["tag."+k] = val
		}
	}

	// Error Tags
	var correlationID string
	if ttnErr, ok := errors.From(err); ok {
		if ttnErr == nil {
			return
		}
		tags["error.namespace"] = ttnErr.Namespace()
		tags["error.name"] = ttnErr.Name()
		correlationID = ttnErr.CorrelationID()
	}

	// Error Attributes
	for k, v := range errors.Attributes(err) {
		if val := fmt.Sprint(v); len(val) < 64 {
			tags["error.attributes."+k] = val
		}
	}

	// Capture the error
	pkt := raven.NewPacket(err.Error(), sentry.ErrorAsExceptions(err, "go.thethings.network/lorawan-stack"))
	pkt.Culprit = method
	pkt.EventID = correlationID
	fw.client.Capture(pkt, tags)
}

// UnaryServerInterceptor forwards errors in Unary RPCs to Sentry
func UnaryServerInterceptor(client *raven.Client) grpc.UnaryServerInterceptor {
	if client == nil {
		client = raven.DefaultClient
	}
	f := &forwarder{client}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		f.forward(ctx, info.FullMethod, err)
		return
	}
}

// StreamServerInterceptor forwards errors in Stream RPCs to Sentry
func StreamServerInterceptor(client *raven.Client) grpc.StreamServerInterceptor {
	if client == nil {
		client = raven.DefaultClient
	}
	f := &forwarder{client}
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		err = handler(srv, ss)
		f.forward(ss.Context(), info.FullMethod, err)
		return
	}
}
