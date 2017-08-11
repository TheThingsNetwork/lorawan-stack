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

// UnaryServerInterceptor forwards errors in Unary RPCs to Sentry
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		log(ctx, info.FullMethod, err)
	}
	return
}

// StreamServerInterceptor forwards errors in Stream RPCs to Sentry
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	err = handler(srv, ss)
	if err != nil {
		log(ss.Context(), info.FullMethod, err)
	}
	return
}

func log(ctx context.Context, method string, err error) {
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
	raven.CaptureError(err, details, nil)
}
