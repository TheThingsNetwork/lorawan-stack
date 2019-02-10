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
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

func (c *Component) initGRPC() {
	rpclog.ReplaceGrpcLogger(c.logger.WithField("namespace", "grpc"))

	c.grpc = rpcserver.New(
		c.ctx,
		rpcserver.WithContextFiller(c.FillContext),
		rpcserver.WithSentry(c.sentry),
	)
}

func (c *Component) setupGRPC() (err error) {
	for _, sub := range c.grpcSubsystems {
		sub.RegisterServices(c.grpc.Server)
	}
	metrics.InitializeServerMetrics(c.grpc.Server)
	c.logger.Debug("Starting loopback connection")
	c.loopback, err = rpcserver.StartLoopback(c.ctx, c.grpc.Server)
	if err != nil {
		return errors.New("could not start loopback connection").WithCause(err)
	}
	c.logger.Debug("Setting up gRPC gateway")
	for _, sub := range c.grpcSubsystems {
		sub.RegisterHandlers(c.grpc.ServeMux, c.loopback)
	}
	c.web.RootGroup(ttnpb.HTTPAPIPrefix).Any("/*", echo.WrapHandler(http.StripPrefix(ttnpb.HTTPAPIPrefix, c.grpc)), middleware.CORS())
	return nil
}

// LoopbackConn returns the loopback gRPC connection to the component.
// This conn must *not* be closed.
func (c *Component) LoopbackConn() *grpc.ClientConn {
	return c.loopback
}

func (c *Component) serveGRPC(lis net.Listener) error {
	return c.grpc.Serve(lis)
}

func (c *Component) grpcEndpoints() []endpoint {
	return []endpoint{
		{listen: Listener.TCP, address: c.config.GRPC.Listen, protocol: "gRPC"},
		{listen: Listener.TLS, address: c.config.GRPC.ListenTLS, protocol: "gRPC/tls"},
	}
}

func (c *Component) listenGRPC() (err error) {
	return c.serveOnEndpoints(c.grpcEndpoints(), (*Component).serveGRPC, "grpc")
}

// RegisterGRPC registers a gRPC subsystem to the component.
func (c *Component) RegisterGRPC(s rpcserver.Registerer) {
	if c.grpc == nil {
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
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = c.cluster.WithVerifiedSource(ctx)
			return next(ctx, req)
		}
	}
}

// ClusterAuthStreamHook ensuring the caller of an RPC is part of the cluster.
// If a call can't be identified as coming from the cluster, it will be discarded.
func (c *Component) ClusterAuthStreamHook() hooks.StreamHandlerMiddleware {
	return func(hdl grpc.StreamHandler) grpc.StreamHandler {
		return func(srv interface{}, stream grpc.ServerStream) error {
			wrapped := grpc_middleware.WrapServerStream(stream)
			ctx := c.cluster.WithVerifiedSource(stream.Context())
			wrapped.WrappedContext = ctx
			return hdl(srv, wrapped)
		}
	}
}
