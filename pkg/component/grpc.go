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

package component

import (
	"context"
	"net/http"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"google.golang.org/grpc"
)

func (c *Component) initGRPC() {
	rpclog.ReplaceGrpcLogger(c.logger.WithField("namespace", "grpc"))

	c.grpc = rpcserver.New(
		c.ctx,
		rpcserver.WithContextFiller(func(ctx context.Context) context.Context {
			ctx = rights.NewContextWithFetcher(ctx, c.rightsFetcher)
			// TODO: Fill other globals in call context (data stores, config, ...)
			return ctx
		}),
		rpcserver.WithSentry(c.sentry),
	)
}

func (c *Component) setupGRPC() (err error) {
	for _, sub := range c.grpcSubsystems {
		sub.RegisterServices(c.grpc.Server)
	}
	c.logger.Debug("Starting loopback connection")
	c.loopback, err = rpcserver.StartLoopback(c.ctx, c.grpc.Server)
	if err != nil {
		return errors.New("could not start loopback connection").WithCause(err)
	}
	c.logger.Debug("Setting up gRPC gateway")
	for _, sub := range c.grpcSubsystems {
		sub.RegisterHandlers(c.grpc.ServeMux, c.loopback)
	}
	c.web.RootGroup(rpcserver.APIPrefix).Any("/*", echo.WrapHandler(http.StripPrefix(rpcserver.APIPrefix, c.grpc)))
	return nil
}

func (c *Component) listenGRPC() (err error) {
	if c.config.GRPC.Listen != "" {
		l, err := c.ListenTCP(c.config.GRPC.Listen)
		if err != nil {
			return errors.New("could not listen on gRPC port").WithCause(err)
		}
		lis, err := l.TCP()
		if err != nil {
			return errors.New("could not create TCP gRPC listener").WithCause(err)
		}
		c.logger.WithFields(log.Fields("namespace", "grpc", "address", c.config.GRPC.Listen)).Info("Listening for TCP gRPC connections")
		go func() {
			if err := c.grpc.Serve(lis); err != nil {
				c.logger.WithError(err).Errorf("Error serving gRPC on %s", lis.Addr())
			}
		}()
	}
	if c.config.GRPC.ListenTLS != "" {
		l, err := c.ListenTCP(c.config.GRPC.ListenTLS)
		if err != nil {
			return errors.New("could not listen on gRPC/tls port").WithCause(err)
		}
		lis, err := l.TLS()
		if err != nil {
			return errors.New("could not create TLS gRPC listener").WithCause(err)
		}
		c.logger.WithFields(log.Fields("namespace", "grpc", "address", c.config.GRPC.ListenTLS)).Info("Listening for TLS gRPC connections")
		go func() {
			if err := c.grpc.Serve(lis); err != nil {
				c.logger.WithError(err).Errorf("Error serving gRPC/tls on %s", lis.Addr())
			}
		}()
	}

	return nil
}

// RegisterGRPC registers a gRPC subsystem to the component
func (c *Component) RegisterGRPC(s rpcserver.Registerer) {
	if c.grpc == nil {
		c.initGRPC()
	}
	c.grpcSubsystems = append(c.grpcSubsystems, s)
}

// ClusterAuth that can be used to identify a component within a cluster.
func (c *Component) ClusterAuth() grpc.CallOption {
	return c.cluster.Auth()
}

// UnaryHook ensuring the caller of an RPC is part of the cluster.
// If a call can't be identified as coming from the cluster, it will be discarded.
func (c *Component) UnaryHook() hooks.UnaryHandlerMiddleware {
	return func(next grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if err := c.cluster.VerifySource(ctx); err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}

// StreamHook ensuring the caller of an RPC is part of the cluster.
// If a call can't be identified as coming from the cluster, it will be discarded.
func (c *Component) StreamHook() hooks.StreamHandlerMiddleware {
	return func(hdl grpc.StreamHandler) grpc.StreamHandler {
		return func(srv interface{}, stream grpc.ServerStream) error {
			if err := c.cluster.VerifySource(stream.Context()); err != nil {
				return err
			}
			return hdl(srv, stream)
		}
	}
}
