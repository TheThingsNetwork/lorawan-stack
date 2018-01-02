// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"context"
	"net/http"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/labstack/echo"
	"github.com/soheilhy/cmux"
)

func (c *Component) initGRPC() {
	c.grpc = rpcserver.New(
		c.ctx,
		rpcserver.WithContextFiller(func(ctx context.Context) context.Context {
			// TODO: Fill globals in call context (data stores, config, ...)
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
		return errors.NewWithCause("Could not start loopback connection", err)
	}
	c.logger.Debug("Setting up gRPC gateway")
	for _, sub := range c.grpcSubsystems {
		sub.RegisterHandlers(c.grpc.ServeMux, c.loopback)
	}
	c.web.Any(rpcserver.APIPrefix, echo.WrapHandler(http.StripPrefix(rpcserver.APIPrefix, c.grpc)))
	return nil
}

func (c *Component) listenGRPC() (err error) {
	if c.config.GRPC.Listen != "" {
		l, err := c.Listen(c.config.GRPC.Listen)
		if err != nil {
			return errors.NewWithCause("Could not listen on gRPC port", err)
		}
		mux, err := l.TCP()
		if err != nil {
			return errors.NewWithCause("Could not create TCP mux on top of gRPC listener", err)
		}
		lis := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		go func() {
			if err := c.grpc.Serve(lis); err != nil {
				c.logger.WithError(err).Errorf("Error serving gRPC on %s", lis.Addr())
			}
		}()
	}
	if c.config.GRPC.ListenTLS != "" {
		l, err := c.Listen(c.config.GRPC.ListenTLS)
		if err != nil {
			return errors.NewWithCause("Could not listen on gRPC/tls port", err)
		}
		mux, err := l.TLS()
		if err != nil {
			return errors.NewWithCause("Could not create TLS mux on top of gRPC/tls listener", err)
		}
		lis := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
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
