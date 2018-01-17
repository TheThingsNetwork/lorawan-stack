// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheThingsNetwork/ttn/pkg/cluster"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/middleware/sentry"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	raven "github.com/getsentry/raven-go"
	"google.golang.org/grpc"
)

// Config is the type of configuration for Components
type Config struct {
	config.ServiceBase `name:",squash" yaml:",inline"`
}

// Component is a base component for The Things Network cluster
type Component struct {
	ctx       context.Context
	cancelCtx context.CancelFunc

	config *Config
	logger log.Stack
	sentry *raven.Client

	cluster cluster.Cluster

	grpc           *rpcserver.Server
	grpcSubsystems []rpcserver.Registerer

	web           *web.Server
	webSubsystems []web.Registerer

	loopback *grpc.ClientConn

	listeners map[string]*listener
}

// New returns a new component
func New(logger log.Stack, config *Config) *Component {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.WithLogger(ctx, logger)

	c := &Component{
		ctx:       ctx,
		cancelCtx: cancel,

		config: config,
		logger: logger,

		listeners: make(map[string]*listener),
	}

	if config.Sentry.DSN != "" {
		c.sentry, _ = raven.New(config.Sentry.DSN)
		c.logger.Use(sentry.New(c.sentry))
	}

	c.web = web.New(c.logger)

	return c
}

// Logger returns the logger of the component
func (c *Component) Logger() log.Stack {
	return c.logger
}

// Context returns the context of the component
func (c *Component) Context() context.Context {
	return c.ctx
}

// Start starts the component
func (c *Component) Start() (err error) {
	defer c.Close()

	c.initGRPC()

	if c.grpc != nil {
		c.logger.Debug("Setting up gRPC server")
		if err = c.setupGRPC(); err != nil {
			return err
		}
	}

	c.logger.Debug("Setting up web server")
	for _, sub := range c.webSubsystems {
		sub.RegisterRoutes(c.web)
	}

	if c.grpc != nil {
		c.logger.Debug("Setting up gRPC server...")
		if err = c.listenGRPC(); err != nil {
			c.logger.WithError(err).Error("Could not start gRPC server")
			return err
		}
	}
	c.logger.Debug("Started gRPC server")

	c.logger.Debug("Setting up HTTP server...")
	if err = c.listenWeb(); err != nil {
		c.logger.WithError(err).Error("Could not start HTTP server")
		return err
	}
	c.logger.Debug("Started HTTP server")

	c.logger.Debug("Initializing cluster...")
	if err := c.initCluster(); err != nil {
		return err
	}

	if err := c.cluster.Join(); err != nil {
		c.logger.WithError(err).Error("Could not join cluster")
		return err
	}
	defer func() {
		c.logger.Debug("Leaving cluster...")
		if err := c.cluster.Leave(); err != nil {
			c.logger.WithError(err).Error("Could not leave cluster")
		}
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case sig := <-signals:
			fmt.Println()
			c.logger.WithField("signal", sig).Info("Received signal, exiting...")
			return nil
		}
	}
}

// Close closes the server
func (c *Component) Close() {
	c.cancelCtx()

	for _, l := range c.listeners {
		err := l.lis.Close()
		if err == nil {
			c.logger.Debugf("Stopped listening on %s", l.lis.Addr())
		} else {
			c.logger.Errorf("Error while stopping to listen on %s", l.lis.Addr())
		}
	}

	if c.grpc != nil {
		c.logger.Debug("Stopping gRPC server...")
		c.grpc.Stop()
		c.logger.Debug("Stopped gRPC server")
	}
}
