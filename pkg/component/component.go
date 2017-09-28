// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/web"
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

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)

	c.logger.Debug("Starting servers")

	if c.grpc != nil {
		if err = c.listenGRPC(); err != nil {
			return err
		}
	}

	if err = c.listenWeb(); err != nil {
		return err
	}

	c.startListeners()

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
		c.grpc.Stop()
		c.logger.Debug("Stopped gRPC server")
	}
}
