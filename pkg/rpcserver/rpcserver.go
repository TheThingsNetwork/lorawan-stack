// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package rpcserver initializes The Things Network's base gRPC server
package rpcserver

import (
	"context"
	"math"
	"time"

	"github.com/TheThingsNetwork/go-utils/errors"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/rpclog"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/sentry"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func init() {
	grpc.EnableTracing = false
}

// TODO: Move this to errors package
var (
	code     = errors.Range(0, 1000)
	rpcPanic = errors.ErrDescriptor{
		MessageFormat: "Internal Server Error",
		Type:          errors.Internal,
		Code:          code(500),
	}
)

// TODO: Move this to errors package
func init() {
	rpcPanic.Register()
}

// New returns a new gRPC server with a set of middlewares.
// The given context is used in some of the middlewares, the given server options are passed to gRPC
//
// Currently the following middlewares are included: tag extraction, Prometheus metrics,
// logging, sending errors to Sentry, validation, errors, panic recovery
func New(ctx context.Context, options ...grpc.ServerOption) *grpc.Server {
	ctxtagsOpts := []grpc_ctxtags.Option{
		grpc_ctxtags.WithFieldExtractor(nil), // TODO: Extract useful fields from the context or payload
	}
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			return rpcPanic.New(errors.Attributes{"panic": p}) // TODO: Use actual error
		}),
	}
	grpc_prometheus.EnableHandlingTimeHistogram()
	baseOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint16),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 6 * time.Hour,
			MaxConnectionAge:  24 * time.Hour,
		}),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(ctxtagsOpts...),
			grpc_prometheus.StreamServerInterceptor,
			rpclog.StreamServerInterceptor(ctx),
			sentry.StreamServerInterceptor,
			grpc_validator.StreamServerInterceptor(),

			// Recovery handler must be on bottom
			grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(ctxtagsOpts...),
			grpc_prometheus.UnaryServerInterceptor,
			rpclog.UnaryServerInterceptor(ctx),
			sentry.UnaryServerInterceptor,
			grpc_validator.UnaryServerInterceptor(),

			// Recovery handler must be on bottom
			grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		)),
	}
	server := grpc.NewServer(append(baseOptions, options...)...)
	return server
}
