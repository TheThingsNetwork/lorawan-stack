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

package component

import (
	"context"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"google.golang.org/grpc"
)

func (c *Component) initGRPC() {
	if c.grpcLogger == nil {
		c.grpcLogger = c.logger.WithField("namespace", "grpc")
	}
	rpclog.ReplaceGrpcLogger(c.grpcLogger)

	c.GRPC = rpcserver.New(
		c.ctx,
		rpcserver.WithContextFiller(c.FillContext),
		rpcserver.WithTrustedProxies(c.config.GRPC.TrustedProxies...),
		rpcserver.WithLogIgnoreMethods(c.config.GRPC.LogIgnoreMethods),
		rpcserver.WithCorrelationIDsIgnoreMethods(c.config.GRPC.CorrelationIDsIgnoreMethods),
		rpcserver.WithRateLimiter(c.RateLimiter()),
	)
}

func (c *Component) setupGRPC() (err error) {
	for _, sub := range c.grpcSubsystems {
		sub.RegisterServices(c.GRPC.Server)
	}
	metrics.InitializeServerMetrics(c.GRPC.Server)
	c.logger.Debug("Starting loopback connection")
	c.loopback, err = rpcserver.StartLoopback(
		c.ctx, c.GRPC.Server,
		rpcclient.DefaultDialOptions(
			// Suppress loopback client logs, because we already have server logs.
			log.NewContext(c.ctx, log.Noop),
		)...,
	)
	if err != nil {
		return errors.New("could not start loopback connection").WithCause(err)
	}
	c.logger.Debug("Setting up gRPC gateway")
	for _, sub := range c.grpcSubsystems {
		sub.RegisterHandlers(c.GRPC.ServeMux, c.loopback)
	}
	return nil
}

// LoopbackConn returns the loopback gRPC connection to the component.
// This conn must *not* be closed.
func (c *Component) LoopbackConn() *grpc.ClientConn {
	return c.loopback
}

func (c *Component) serveGRPC(lis net.Listener) error {
	return c.GRPC.Serve(lis)
}

func (c *Component) grpcEndpoints() []Endpoint {
	return []Endpoint{
		NewTCPEndpoint(c.config.GRPC.Listen, "gRPC"),
		NewTLSEndpoint(c.config.GRPC.ListenTLS, "gRPC", tlsconfig.WithNextProtos("h2", "http/1.1")),
	}
}

func (c *Component) listenGRPC() (err error) {
	return c.serveOnEndpoints(c.grpcEndpoints(), (*Component).serveGRPC, "grpc")
}

// RegisterGRPC registers a gRPC subsystem to the component.
func (c *Component) RegisterGRPC(s rpcserver.Registerer) {
	if c.GRPC == nil {
		c.initGRPC()
	}
	c.grpcSubsystems = append(c.grpcSubsystems, s)
}

// WithClusterAuth that can be used to identify a component within a cluster.
func (c *Component) WithClusterAuth() grpc.CallOption {
	return c.cluster.Auth()
}

// ClusterAuthUnaryHook ensuring the caller of an RPC is part of the cluster.
// If a call can't be identified as coming from the cluster, it will be discarded.
func (c *Component) ClusterAuthUnaryHook() hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req any) (any, error) {
			ctx = c.cluster.WithVerifiedSource(ctx)
			return next(ctx, req)
		}
	}
}

// ClusterAuthStreamHook ensuring the caller of an RPC is part of the cluster.
// If a call can't be identified as coming from the cluster, it will be discarded.
func (c *Component) ClusterAuthStreamHook() hooks.StreamHandlerMiddleware {
	return func(hdl grpc.StreamHandler) grpc.StreamHandler {
		return func(srv any, stream grpc.ServerStream) error {
			wrapped := grpc_middleware.WrapServerStream(stream)
			ctx := c.cluster.WithVerifiedSource(stream.Context())
			wrapped.WrappedContext = ctx
			return hdl(srv, wrapped)
		}
	}
}
