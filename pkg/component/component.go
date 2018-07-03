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

// Package component contains the methods and structures common to all components.
package component

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	raven "github.com/getsentry/raven-go"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log/middleware/sentry"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/web"
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

	tcpListeners map[string]*listener

	FrequencyPlans *frequencyplans.Store

	rightsHook *rights.Hook
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
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.NewContext(ctx, logger)

	c := &Component{
		ctx:       ctx,
		cancelCtx: cancel,

		config: config,
		logger: logger,

		tcpListeners: make(map[string]*listener),

		FrequencyPlans: config.FrequencyPlans.Store(),
	}

	if config.Sentry.DSN != "" {
		c.sentry, _ = raven.New(config.Sentry.DSN)
		c.sentry.SetIncludePaths([]string{"go.thethings.network/lorawan-stack"})
		c.logger.Use(sentry.New(c.sentry))
	}

	c.web, err = web.New(c.ctx, config.HTTP)
	if err != nil {
		return nil, err
	}

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

	for _, l := range c.tcpListeners {
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

// RightsHook returns the hook that preload rights in the context based an authorization value.
func (c *Component) RightsHook() (*rights.Hook, error) {
	if c.rightsHook == nil {
		hook, err := rights.New(c.ctx, rightsFetchingConnector{Component: c}, c.config.Rights)
		if err != nil {
			return nil, errors.NewWithCause(err, "Could not initialize rights hook")
		}
		c.rightsHook = hook
	}

	return c.rightsHook, nil
}

// AllowInsecureRPCs returns `true` if the component was configured to allow connection over insecure protocols.
func (c *Component) AllowInsecureRPCs() bool {
	return c.config.Rights.AllowInsecure
}
