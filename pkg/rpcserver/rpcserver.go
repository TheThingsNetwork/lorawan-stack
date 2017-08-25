// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package rpcserver initializes The Things Network's base gRPC server
package rpcserver

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/fillcontext"
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

type options struct {
	contextFiller  fillcontext.Filler
	fieldExtractor grpc_ctxtags.RequestFieldExtractorFunc
	serverOptions  []grpc.ServerOption
}

// Option for the gRPC server
type Option func(*options)

// WithServerOptions adds gRPC ServerOptions
func WithServerOptions(serverOptions ...grpc.ServerOption) Option {
	return func(o *options) {
		o.serverOptions = append(o.serverOptions, serverOptions...)
	}
}

// WithContextFiller sets a context filler
func WithContextFiller(contextFiller fillcontext.Filler) Option {
	return func(o *options) {
		o.contextFiller = contextFiller
	}
}

// WithFieldExtractor sets a field extractor
func WithFieldExtractor(fieldExtractor grpc_ctxtags.RequestFieldExtractorFunc) Option {
	return func(o *options) {
		o.fieldExtractor = fieldExtractor
	}
}

// New returns a new gRPC server with a set of middlewares.
// The given context is used in some of the middlewares, the given server options are passed to gRPC
//
// Currently the following middlewares are included: tag extraction, Prometheus metrics,
// logging, sending errors to Sentry, validation, errors, panic recovery
func New(ctx context.Context, opts ...Option) *grpc.Server {
	options := new(options)
	for _, opt := range opts {
		opt(options)
	}

	ctxtagsOpts := []grpc_ctxtags.Option{
		grpc_ctxtags.WithFieldExtractor(options.fieldExtractor),
	}
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			return errors.New(fmt.Sprint(p))
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
			fillcontext.StreamServerInterceptor(options.contextFiller),
			grpc_prometheus.StreamServerInterceptor,
			rpclog.StreamServerInterceptor(ctx),
			sentry.StreamServerInterceptor,
			grpc_validator.StreamServerInterceptor(),

			// Recovery handler must be on bottom
			grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(ctxtagsOpts...),
			fillcontext.UnaryServerInterceptor(options.contextFiller),
			grpc_prometheus.UnaryServerInterceptor,
			rpclog.UnaryServerInterceptor(ctx),
			sentry.UnaryServerInterceptor,
			grpc_validator.UnaryServerInterceptor(),

			// Recovery handler must be on bottom
			grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		)),
	}
	server := grpc.NewServer(append(baseOptions, options.serverOptions...)...)
	return server
}
