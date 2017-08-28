// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package sentry implements gRPC middleware that forwards errors in RPCs to Sentry
package sentry

import (
	"fmt"

	raven "github.com/getsentry/raven-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type forwarder struct {
	client *raven.Client
}

func (f *forwarder) forward(ctx context.Context, method string, err error) {
	code := grpc.Code(err)
	if code == codes.OK {
		return
	}
	var details = make(map[string]string)
	details["grpc.code"] = code.String()
	details["grpc.method"] = method
	if tags := grpc_ctxtags.Extract(ctx); tags != nil {
		for k, v := range tags.Values() {
			details[k] = fmt.Sprint(v)
		}
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
