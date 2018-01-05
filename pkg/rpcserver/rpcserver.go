// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package rpcserver initializes The Things Network's base gRPC server
package rpcserver

import (
	"context"
	"math"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/errors/grpcerrors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/claims"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/fillcontext"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/rpclog"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/sentry"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver/internal/jsonpb"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
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
	contextFillers     []fillcontext.Filler
	fieldExtractor     grpc_ctxtags.RequestFieldExtractorFunc
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
	serverOptions      []grpc.ServerOption
	sentry             *raven.Client
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
func WithContextFiller(contextFillers ...fillcontext.Filler) Option {
	return func(o *options) {
		o.contextFillers = append(o.contextFillers, contextFillers...)
	}
}

// WithFieldExtractor sets a field extractor
func WithFieldExtractor(fieldExtractor grpc_ctxtags.RequestFieldExtractorFunc) Option {
	return func(o *options) {
		o.fieldExtractor = fieldExtractor
	}
}

// WithStreamInterceptors adds gRPC stream interceptors
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamInterceptors = append(o.streamInterceptors, interceptors...)
	}
}

// WithUnaryInterceptors adds gRPC unary interceptors
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInterceptors = append(o.unaryInterceptors, interceptors...)
	}
}

// WithTokenKeyInfoProvider adds both an unary and stream claims interceptor
// with the given claims.TokenKeyInfoProvider.
func WithTokenKeyInfoProvider(provider claims.TokenKeyInfoProvider) Option {
	return func(o *options) {
		o.unaryInterceptors = append(o.unaryInterceptors, claims.UnaryServerInterceptor(provider))
		o.streamInterceptors = append(o.streamInterceptors, claims.StreamServerInterceptor(provider))
	}
}

// WithSentry sets a sentry server
func WithSentry(sentry *raven.Client) Option {
	return func(o *options) {
		o.sentry = sentry
	}
}

func init() {
	ErrRPCRecovered.Register()
}

// ErrRPCRecovered is returned when we recovered from a panic
var ErrRPCRecovered = &errors.ErrDescriptor{
	MessageFormat:  "Internal Server Error",
	Code:           500,
	Type:           errors.Internal,
	SafeAttributes: nil, // We don't want to give any information to the clients
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
			return ErrRPCRecovered.New(errors.Attributes{
				"panic": p,
				"stack": string(debug.Stack()),
			})
		}),
	}
	grpc_prometheus.EnableHandlingTimeHistogram()

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpcerrors.StreamServerInterceptor(),
		grpc_ctxtags.StreamServerInterceptor(ctxtagsOpts...),
		fillcontext.StreamServerInterceptor(options.contextFillers...),
		grpc_prometheus.StreamServerInterceptor,
		rpclog.StreamServerInterceptor(ctx), // Gets logger from global context
		sentry.StreamServerInterceptor(options.sentry),
		grpc_validator.StreamServerInterceptor(),
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpcerrors.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(ctxtagsOpts...),
		fillcontext.UnaryServerInterceptor(options.contextFillers...),
		grpc_prometheus.UnaryServerInterceptor,
		rpclog.UnaryServerInterceptor(ctx), // Gets logger from global context
		sentry.UnaryServerInterceptor(options.sentry),
		grpc_validator.UnaryServerInterceptor(),
	}

	baseOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint16),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 6 * time.Hour,
			MaxConnectionAge:  24 * time.Hour,
		}),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			append(
				append(streamInterceptors, options.streamInterceptors...),
				grpc_recovery.StreamServerInterceptor(recoveryOpts...),
			)...,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			append(
				append(unaryInterceptors, options.unaryInterceptors...),
				grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
			)...,
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
	Roles() []ttnpb.PeerInfo_Role
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
