// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package rpcserver initializes The Things Network's base gRPC server
package rpcserver

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/fillcontext"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/rpclog"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/sentry"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver/internal/jsonpb"
	"github.com/getsentry/raven-go"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// APIPrefix for the HTTP handler
const APIPrefix = "/api/v3"

func init() {
	grpc.EnableTracing = false
}

type options struct {
	contextFiller  fillcontext.Filler
	fieldExtractor grpc_ctxtags.RequestFieldExtractorFunc
	serverOptions  []grpc.ServerOption
	sentry         *raven.Client
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

// WithSentry sets a sentry server
func WithSentry(sentry *raven.Client) Option {
	return func(o *options) {
		o.sentry = sentry
	}
}

// New returns a new RPC server with a set of middlewares.
// The given context is used in some of the middlewares, the given server options are passed to gRPC
//
// Currently the following middlewares are included: tag extraction, Prometheus metrics,
// logging, sending errors to Sentry, validation, errors, panic recovery
func New(ctx context.Context, opts ...Option) *Server {
	options := new(options)
	for _, opt := range opts {
		opt(options)
	}
	server := &Server{ctx: ctx}
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
			sentry.StreamServerInterceptor(options.sentry),
			grpc_validator.StreamServerInterceptor(),

			// Recovery handler must be on bottom
			grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(ctxtagsOpts...),
			fillcontext.UnaryServerInterceptor(options.contextFiller),
			grpc_prometheus.UnaryServerInterceptor,
			rpclog.UnaryServerInterceptor(ctx),
			sentry.UnaryServerInterceptor(options.sentry),
			grpc_validator.UnaryServerInterceptor(),

			// Recovery handler must be on bottom
			grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		)),
	}
	server.Server = grpc.NewServer(append(baseOptions, options.serverOptions...)...)
	server.ServeMux = runtime.NewServeMux(runtime.WithMarshalerOption("*", &jsonpb.GoGoJSONPb{
		OrigName: true,
	}))
	return server
}

// Registerer allows components to register their services to the gRPC server and the HTTP gateway
type Registerer interface {
	RegisterServices(s *grpc.Server)
	RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn)
}

// Server wraps the gRPC server
type Server struct {
	ctx context.Context
	*grpc.Server
	*runtime.ServeMux
}

// ServeHTTP forwards requests to the gRPC gateway
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ServeMux.ServeHTTP(w, r)
}
