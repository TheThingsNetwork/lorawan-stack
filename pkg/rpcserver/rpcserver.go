// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package rpcserver initializes The Things Network's base gRPC server
package rpcserver

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/fillcontext"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/sentry"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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

// WithSentry sets a sentry server
func WithSentry(sentry *raven.Client) Option {
	return func(o *options) {
		o.sentry = sentry
	}
}

// ErrRPCRecovered is returned when a panic is caught from an RPC.
var ErrRPCRecovered = errors.DefineInternal("rpc_recovered", "Internal Server Error")

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
			return ErrRPCRecovered.WithAttributes("panic", p)
		}),
	}
	grpc_prometheus.EnableHandlingTimeHistogram()

	streamInterceptors := []grpc.StreamServerInterceptor{
		fillcontext.StreamServerInterceptor(options.contextFillers...),
		grpc_ctxtags.StreamServerInterceptor(ctxtagsOpts...),
		rpcmiddleware.RequestIDStreamServerInterceptor(),
		events.StreamServerInterceptor,
		rpclog.StreamServerInterceptor(ctx),
		grpc_prometheus.StreamServerInterceptor,
		sentry.StreamServerInterceptor(options.sentry),
		errors.StreamServerInterceptor(),
		grpc_validator.StreamServerInterceptor(),
		hooks.StreamServerInterceptor(),
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		fillcontext.UnaryServerInterceptor(options.contextFillers...),
		grpc_ctxtags.UnaryServerInterceptor(ctxtagsOpts...),
		rpcmiddleware.RequestIDUnaryServerInterceptor(),
		events.UnaryServerInterceptor,
		rpclog.UnaryServerInterceptor(ctx),
		grpc_prometheus.UnaryServerInterceptor,
		sentry.UnaryServerInterceptor(options.sentry),
		errors.UnaryServerInterceptor(),
		grpc_validator.UnaryServerInterceptor(),
		hooks.UnaryServerInterceptor(),
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
	server.ServeMux = runtime.NewServeMux(
		runtime.WithMarshalerOption("*", jsonpb.TTN()),
		runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),
	)
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
