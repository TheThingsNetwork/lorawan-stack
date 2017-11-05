// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package sentry implements gRPC middleware that forwards errors in RPCs to Sentry
package sentry

import (
	"fmt"

	"context"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	grpcerrors "github.com/TheThingsNetwork/ttn/pkg/errors/grpcerrors" // Note: actual package name there is "errors"
	raven "github.com/getsentry/raven-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type forwarder struct {
	client *raven.Client
}

func (f *forwarder) forward(ctx context.Context, method string, err error) {
	code := codes.Unknown
	if status, ok := status.FromError(err); ok {
		code = status.Code()
	}

	ttnErr := grpcerrors.FromGRPC(err)
	ttnErrType := ttnErr.Type()

	switch ttnErrType {
	case errors.InvalidArgument,
		errors.OutOfRange,
		errors.NotFound,
		errors.Conflict,
		errors.AlreadyExists,
		errors.Unauthorized,
		errors.PermissionDenied,
		errors.Timeout,
		errors.Canceled:
		return // ignore
	}

	var details = make(map[string]string)
	details["grpc.code"] = code.String()
	details["grpc.method"] = method
	if tags := grpc_ctxtags.Extract(ctx); tags != nil {
		for k, v := range tags.Values() {
			details[k] = fmt.Sprint(v)
		}
	}
	details["ttn.error.code"] = ttnErr.Code().String()
	details["ttn.error.type"] = ttnErrType.String()
	details["ttn.error.namespace"] = ttnErr.Namespace()
	for k, v := range ttnErr.Attributes() {
		details["ttn.error."+k] = fmt.Sprint(v)
	}

	f.client.CaptureError(err, details, nil)
}

// UnaryServerInterceptor forwards errors in Unary RPCs to Sentry
func UnaryServerInterceptor(client *raven.Client) grpc.UnaryServerInterceptor {
	if client == nil {
		client = raven.DefaultClient
	}
	f := &forwarder{client}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			f.forward(ctx, info.FullMethod, err)
		}
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
		if err != nil {
			f.forward(ss.Context(), info.FullMethod, err)
		}
		return
	}
}
