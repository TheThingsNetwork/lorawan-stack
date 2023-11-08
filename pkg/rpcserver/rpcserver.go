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

// Package rpcserver initializes The Things Network's base gRPC server
package rpcserver

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/textproto"
	"os"
	"runtime/debug"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/fillcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware"
	rpcfillcontext "go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/fillcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	sentrymiddleware "go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/sentry"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/validator"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // Register gzip compression.
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

func init() {
	grpc.EnableTracing = false
	for rpc, paths := range ttnpb.RPCFieldMaskPaths {
		validator.RegisterAllowedFieldMaskPaths(rpc, paths.Set, paths.All, paths.Allowed...)
	}
}

type options struct {
	contextFillers              []fillcontext.Filler
	streamInterceptors          []grpc.StreamServerInterceptor
	unaryInterceptors           []grpc.UnaryServerInterceptor
	serverOptions               []grpc.ServerOption
	trustedProxies              []string
	logIgnoreMethods            []string
	correlationIDsIgnoreMethods []string
	limiter                     ratelimit.Interface
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

// WithTrustedProxies adds trusted proxies from which proxy headers are trusted.
func WithTrustedProxies(cidrs ...string) Option {
	return func(o *options) {
		o.trustedProxies = append(o.trustedProxies, cidrs...)
	}
}

// WithLogIgnoreMethods sets a list of methods for which no log messages are printed on success.
func WithLogIgnoreMethods(methods []string) Option {
	return func(o *options) {
		o.logIgnoreMethods = methods
	}
}

// WithCorrelationIDsIgnoreMethods sets a list of methods for which no correlation IDs are injected.
func WithCorrelationIDsIgnoreMethods(methods []string) Option {
	return func(o *options) {
		o.correlationIDsIgnoreMethods = methods
	}
}

// WithRateLimiter configures a rate limiter on the server.
func WithRateLimiter(limiter ratelimit.Interface) Option {
	return func(o *options) {
		o.limiter = limiter
	}
}

// ErrRPCRecovered is returned when a panic is caught from an RPC.
var ErrRPCRecovered = errors.DefineInternal("rpc_recovered", "Internal Server Error")

// New returns a new RPC server with a set of middlewares.
// The given context is used in some of the middlewares, the given server options are passed to gRPC
//
// Currently the following middlewares are included: tag extraction, metrics,
// logging, sending errors to Sentry, validation, errors, panic recovery
func New(ctx context.Context, opts ...Option) *Server {
	options := new(options)
	options.limiter = &ratelimit.NoopRateLimiter{}
	for _, opt := range opts {
		opt(options)
	}
	server := &Server{ctx: ctx, Hooks: &hooks.Hooks{}}
	ctxtagsOpts := []grpc_ctxtags.Option{
		grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor),
	}
	var proxyHeaders rpcmiddleware.ProxyHeaders
	proxyHeaders.ParseAndAddTrusted(options.trustedProxies...)
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p any) (err error) {
			fmt.Fprintln(os.Stderr, p)
			os.Stderr.Write(debug.Stack())
			if pErr, ok := p.(error); ok {
				err = ErrRPCRecovered.WithCause(pErr)
			} else {
				err = ErrRPCRecovered.WithAttributes("panic", p)
			}
			return err
		}),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		rpcfillcontext.StreamServerInterceptor(options.contextFillers...),
		grpc_ctxtags.StreamServerInterceptor(ctxtagsOpts...),
		rpcmiddleware.RequestIDStreamServerInterceptor(),
		proxyHeaders.StreamServerInterceptor(),
		otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(tracing.FromContext(ctx))),
		events.StreamServerInterceptor(options.correlationIDsIgnoreMethods),
		rpclog.StreamServerInterceptor(ctx, rpclog.WithIgnoreMethods(options.logIgnoreMethods)),
		metrics.StreamServerInterceptor,
		errors.StreamServerInterceptor(),
		// NOTE: All middleware that works with lorawan-stack/pkg/errors errors must be placed below.
		sentrymiddleware.StreamServerInterceptor(),
		grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		validator.StreamServerInterceptor(),
		ratelimit.StreamServerInterceptor(options.limiter),
		server.Hooks.StreamServerInterceptor(),
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		rpcfillcontext.UnaryServerInterceptor(options.contextFillers...),
		grpc_ctxtags.UnaryServerInterceptor(ctxtagsOpts...),
		rpcmiddleware.RequestIDUnaryServerInterceptor(),
		proxyHeaders.UnaryServerInterceptor(),
		otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tracing.FromContext(ctx))),
		events.UnaryServerInterceptor(options.correlationIDsIgnoreMethods),
		rpclog.UnaryServerInterceptor(ctx, rpclog.WithIgnoreMethods(options.logIgnoreMethods)),
		metrics.UnaryServerInterceptor,
		errors.UnaryServerInterceptor(),
		// NOTE: All middleware that works with lorawan-stack/pkg/errors errors must be placed below.
		sentrymiddleware.UnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		validator.UnaryServerInterceptor(),
		ratelimit.UnaryServerInterceptor(options.limiter),
		server.Hooks.UnaryServerInterceptor(),
	}

	baseOptions := []grpc.ServerOption{
		grpc.StatsHandler(rpcmiddleware.StatsHandlers{new(ocgrpc.ServerHandler), metrics.StatsHandler}),
		grpc.MaxConcurrentStreams(math.MaxUint16),
		grpc.MaxRecvMsgSize(1024 * 1024 * 16),
		grpc.MaxSendMsgSize(1024 * 1024 * 16),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             1 * time.Minute,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 6 * time.Hour,
			MaxConnectionAge:  24 * time.Hour,
			Time:              1 * time.Minute,
			Timeout:           20 * time.Second,
		}),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			append(streamInterceptors, options.streamInterceptors...)...,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			append(unaryInterceptors, options.unaryInterceptors...)...,
		)),
	}
	server.Server = grpc.NewServer(append(baseOptions, options.serverOptions...)...)
	server.ServeMux = runtime.NewServeMux(
		runtime.WithMarshalerOption("*", jsonpb.TTN()),
		runtime.WithMarshalerOption("text/event-stream", jsonpb.TTNEventStream()),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			md := rpcmetadata.MD{
				Host: req.Host,
				URI:  req.RequestURI,
			}
			return md.ToMetadata()
		}),
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			s = textproto.CanonicalMIMEHeaderKey(s)
			switch s {
			case "Forwarded",
				"X-Request-Id",
				"X-Forwarded-For",
				"X-Real-Ip",
				"X-Forwarded-Host",
				"X-Forwarded-Proto",
				"X-Forwarded-Client-Cert",
				"X-Forwarded-Tls-Client-Cert",
				"X-Forwarded-Tls-Client-Cert-Info":
				return s, true
			}
			return runtime.DefaultHeaderMatcher(s)
		}),
		runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
			// NOTE: When adding headers, also add them to CORSConfig in ../component/grpc.go.
			switch s {
			case "x-total-count":
				return "X-Total-Count", true
			case "x-rate-limit-limit":
				return "X-Rate-Limit-Limit", true
			case "x-rate-limit-available":
				return "X-Rate-Limit-Available", true
			case "x-rate-limit-reset":
				return "X-Rate-Limit-Reset", true
			case "x-rate-limit-retry":
				return "X-Rate-Limit-Retry", true
			case "warning":
				// NOTE: the "Warning" header in HTTP is specified differently than our "warning" gRPC metadata.
				return "X-Warning", true
			}
			return s, false
		}),
		runtime.WithOutgoingTrailerMatcher(func(s string) (string, bool) {
			return s, false
		}),
		runtime.WithDisablePathLengthFallback(),
	)
	return server
}

// ServiceRegisterer allows components to register their services to the gRPC server and the HTTP gateway.
type ServiceRegisterer interface {
	RegisterServices(s *grpc.Server)
	RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn)
}

// Registerer extends ServiceRegisterer with the cluster roles fulfilled by the component.
type Registerer interface {
	Roles() []ttnpb.ClusterRole
	ServiceRegisterer
}

// Server wraps the gRPC server
type Server struct {
	ctx context.Context
	*grpc.Server
	*hooks.Hooks
	*runtime.ServeMux
}

// ServeHTTP forwards requests to the gRPC gateway
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ServeMux.ServeHTTP(w, r)
}
