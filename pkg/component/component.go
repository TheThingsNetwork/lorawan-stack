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

// Package component contains the methods and structures common to all components.
package component

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/getsentry/raven-go"
	"github.com/heptiolabs/healthcheck"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/fillcontext"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log/middleware/sentry"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/version"
	"go.thethings.network/lorawan-stack/pkg/web"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
)

// Config is the type of configuration for Components
type Config struct {
	config.ServiceBase `name:",squash" yaml:",inline"`
}

// Component is a base component for The Things Network cluster
type Component struct {
	ctx                context.Context
	cancelCtx          context.CancelFunc
	terminationSignals chan os.Signal

	config        *Config
	getBaseConfig func(ctx context.Context) config.ServiceBase

	acme *autocert.Manager

	logger log.Stack
	sentry *raven.Client

	cluster    cluster.Cluster
	clusterNew func(ctx context.Context, config *config.Cluster, options ...cluster.Option) (cluster.Cluster, error)

	grpc           *rpcserver.Server
	grpcSubsystems []rpcserver.Registerer

	web           *web.Server
	webSubsystems []web.Registerer

	interop           *interop.Server
	interopSubsystems []interop.Registerer

	healthHandler healthcheck.Handler

	loopback *grpc.ClientConn

	tcpListeners map[string]*listener

	fillers []fillcontext.Filler

	FrequencyPlans *frequencyplans.Store
	KeyVault       crypto.KeyVault

	rightsFetcher rights.Fetcher

	tasks []task
}

// Option allows extending the component when it is instantiated with New.
type Option func(*Component)

// WithClusterNew returns an option that overrides the component's function for
// setting up the cluster.
// This allows extending the cluster configuration with custom logic based on
// information in the context.
func WithClusterNew(f func(ctx context.Context, config *config.Cluster, options ...cluster.Option) (cluster.Cluster, error)) Option {
	return func(c *Component) {
		c.clusterNew = f
	}
}

// WithBaseConfigGetter returns an option that overrides the component's function
// for getting the base config.
// This allows overriding the configuration with information in the context.
func WithBaseConfigGetter(f func(ctx context.Context) config.ServiceBase) Option {
	return func(c *Component) {
		c.getBaseConfig = f
	}
}

// New returns a new component.
func New(logger log.Stack, config *Config, opts ...Option) (c *Component, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	ctx = log.NewContext(ctx, logger)

	fps, err := config.FrequencyPlans.Store()
	if err != nil {
		return nil, err
	}

	c = &Component{
		ctx:                ctx,
		cancelCtx:          cancel,
		terminationSignals: make(chan os.Signal),

		config: config,
		logger: logger,

		healthHandler: healthcheck.NewHandler(),

		tcpListeners: make(map[string]*listener),

		FrequencyPlans: fps,
		KeyVault:       config.KeyVault.KeyVault(),
	}

	if config.Sentry.DSN != "" {
		c.sentry, _ = raven.New(config.Sentry.DSN)
		c.sentry.SetIncludePaths([]string{"go.thethings.network/lorawan-stack"})
		c.sentry.SetRelease(version.String())
		c.logger.Use(sentry.New(c.sentry))
	}

	for _, opt := range opts {
		opt(c)
	}
	if c.clusterNew == nil {
		c.clusterNew = cluster.New
	}

	if err = c.initWeb(); err != nil {
		return nil, err
	}

	if err := c.initACME(); err != nil {
		return nil, err
	}

	c.interop, err = interop.NewServer(c.ctx, config.Interop)
	if err != nil {
		return nil, err
	}

	c.initRights()

	c.initGRPC()

	return c, nil
}

// MustNew calls New and returns a new component or panics on an error.
// In most cases, you should just use New.
func MustNew(logger log.Stack, config *Config, opts ...Option) *Component {
	c, err := New(logger, config, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// Logger returns the logger of the component.
func (c *Component) Logger() log.Stack {
	return c.logger
}

// LogDebug returns whether the component should log debug messages.
func (c *Component) LogDebug() bool {
	return c.config.Log.Level == log.DebugLevel
}

// Context returns the context of the component.
func (c *Component) Context() context.Context {
	return c.ctx
}

// GetBaseConfig gets the base config of the component.
func (c *Component) GetBaseConfig(ctx context.Context) config.ServiceBase {
	if c.getBaseConfig != nil {
		return c.getBaseConfig(ctx)
	}
	return c.config.ServiceBase
}

// FillContext fills the context.
// This method should only be used for request contexts.
func (c *Component) FillContext(ctx context.Context) context.Context {
	for _, filler := range c.fillers {
		ctx = filler(ctx)
	}
	return ctx
}

// AddContextFiller adds the specified filler.
func (c *Component) AddContextFiller(f fillcontext.Filler) {
	c.fillers = append(c.fillers, f)
}

// Start starts the component.
func (c *Component) Start() (err error) {
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

	c.logger.Debug("Initializing cluster...")
	if err := c.initCluster(); err != nil {
		return err
	}

	c.logger.Debug("Initializing web server...")
	for _, sub := range c.webSubsystems {
		sub.RegisterRoutes(c.web)
	}

	c.logger.Debug("Initializing interop server...")
	for _, sub := range c.interopSubsystems {
		sub.RegisterInterop(c.interop)
	}

	if c.grpc != nil {
		c.logger.Debug("Starting gRPC server...")
		if err = c.listenGRPC(); err != nil {
			c.logger.WithError(err).Error("Could not start gRPC server")
			return err
		}
	}
	c.logger.Debug("Started gRPC server")

	c.logger.Debug("Starting web server...")
	if err = c.listenWeb(); err != nil {
		c.logger.WithError(err).Error("Could not start web server")
		return err
	}
	c.logger.Debug("Started web server")

	c.logger.Debug("Starting interop server")
	if err = c.listenInterop(); err != nil {
		c.logger.WithError(err).Error("Could not start interop server")
	}
	c.logger.Debug("Started interop server")

	c.logger.Debug("Joining cluster...")
	if err := c.cluster.Join(); err != nil {
		c.logger.WithError(err).Error("Could not join cluster")
		return err
	}
	c.logger.Debug("Joined cluster")

	c.logger.Debug("Starting tasks")
	c.startTasks()
	c.logger.Debug("Started tasks")

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

	signal.Notify(c.terminationSignals, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case sig := <-c.terminationSignals:
			fmt.Println()
			c.logger.WithField("signal", sig).Info("Received signal, exiting...")
			return nil
		}
	}
}

// Close closes the server.
func (c *Component) Close() {
	c.cancelCtx()

	for _, l := range c.tcpListeners {
		err := l.lis.Close()
		if err != nil && c.ctx.Err() == nil {
			c.logger.WithError(err).Errorf("Error while stopping to listen on %s", l.lis.Addr())
			continue
		}
		c.logger.Debugf("Stopped listening on %s", l.lis.Addr())
	}

	if c.grpc != nil {
		c.logger.Debug("Stopping gRPC server...")
		c.grpc.Stop()
		c.logger.Debug("Stopped gRPC server")
	}
}

// AllowInsecureForCredentials returns `true` if the component was configured to allow transmission of credentials
// over insecure transports.
func (c *Component) AllowInsecureForCredentials() bool {
	return c.config.GRPC.AllowInsecureForCredentials
}

// ServeHTTP serves an HTTP request.
// If the Content-Type is application/grpc, the request is routed to gRPC.
// Otherwise, the request is routed to the default web server.
func (c *Component) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
		c.grpc.Server.ServeHTTP(w, r)
	} else {
		c.web.ServeHTTP(w, r)
	}
}
