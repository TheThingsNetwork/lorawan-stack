// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
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

// MustNew calls New and returns a new component or panics on an error.
// In most cases, you should just use New.
func MustNew(logger log.Stack, config *Config) *Component {
	c, err := New(logger, config)
	if err != nil {
		panic(err)
	}
	return c
}

// New returns a new component
func New(logger log.Stack, config *Config) (*Component, error) {
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

	hash, block, err := config.HTTP.Cookie.Keys()
	if err != nil {
		return nil, err
	}

	c.web = web.New(c.logger, web.WithCookieSecrets(hash, block))

	return c, nil
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
	c.initGRPC()

	if c.grpc != nil {
		c.logger.Debug("Initializing gRPC server...")
		if err = c.setupGRPC(); err != nil {
			return err
		}
		serviceInfo := c.grpc.Server.GetServiceInfo()
		services := make([]string, 0, len(serviceInfo))
		for service := range serviceInfo {
			services = append(services, service)
		}
		sort.Strings(services)
		c.logger.WithFields(log.Fields(
			"namespace", "grpc",
			"services", services,
		)).Debug("Exposed services")
	}

	c.logger.Debug("Initializing web server...")
	for _, sub := range c.webSubsystems {
		sub.RegisterRoutes(c.web)
	}

	if c.grpc != nil {
		c.logger.Debug("Starting gRPC server...")
		if err = c.listenGRPC(); err != nil {
			c.logger.WithError(err).Error("Could not start gRPC server")
			return err
		}
	}
	c.logger.Debug("Started gRPC server")

	c.logger.Debug("Starting HTTP server...")
	if err = c.listenWeb(); err != nil {
		c.logger.WithError(err).Error("Could not start HTTP server")
		return err
	}
	c.logger.Debug("Started HTTP server")

	c.logger.Debug("Initializing cluster...")
	if err := c.initCluster(); err != nil {
		return err
	}

	c.logger.Debug("Joining cluster...")
	if err := c.cluster.Join(); err != nil {
		c.logger.WithError(err).Error("Could not join cluster")
		return err
	}
	c.logger.Debug("Joined cluster")

	return nil
}

// Run starts the component, and returns when a stop signal has been received by the process.
func (c *Component) Run() error {
	defer c.Close()

	if err := c.Start(); err != nil {
		return err
	}

	defer func() {
		c.logger.Debug("Leaving cluster...")
		if err := c.cluster.Leave(); err != nil {
			c.logger.WithError(err).Error("Could not leave cluster")
		}
		c.logger.Debug("Left cluster")
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
