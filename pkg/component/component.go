// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"sync"
	"syscall"

	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/mtls"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/fillcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/healthcheck"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
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

	cluster    cluster.Cluster
	clusterNew func(ctx context.Context, config *cluster.Config, options ...cluster.Option) (cluster.Cluster, error)

	GRPC           *rpcserver.Server
	grpcLogger     log.Interface
	grpcSubsystems []rpcserver.Registerer

	web           *web.Server
	webSubsystems []web.Registerer

	interop           *interop.Server
	interopSubsystems []interop.Registerer

	healthHandler healthcheck.HealthChecker

	loopback *grpc.ClientConn

	tcpListeners   map[string]*listener
	tcpListenersMu sync.Mutex

	fillers []fillcontext.Filler

	frequencyPlans *frequencyplans.Store

	componentKEKLabeler crypto.ComponentKEKLabeler
	keyService          crypto.KeyService

	rightsFetcher rights.Fetcher

	taskStarter task.Starter
	taskConfigs []*task.Config

	caStore *mtls.CAStore

	limiter ratelimit.Interface
}

// Option allows extending the component when it is instantiated with New.
type Option func(*Component)

// WithClusterNew returns an option that overrides the component's function for
// setting up the cluster.
// This allows extending the cluster configuration with custom logic based on
// information in the context.
func WithClusterNew(f func(ctx context.Context, config *cluster.Config, options ...cluster.Option) (cluster.Cluster, error)) Option {
	return func(c *Component) {
		c.clusterNew = f
	}
}

// WithGRPCLogger returns an option that overrides the component's gRPC logger.
func WithGRPCLogger(l log.Interface) Option {
	return func(c *Component) {
		c.grpcLogger = l
	}
}

// WithTaskStarter returns an option that overrides the component's TaskStarter for
// starting tasks.
func WithTaskStarter(s task.Starter) Option {
	return func(c *Component) {
		c.taskStarter = s
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

// WithTracerProvider returns an option that stores the given trace provider
// in the component's context.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *Component) {
		c.ctx = tracing.NewContextWithTracerProvider(c.ctx, tp)
		c.AddContextFiller(func(ctx context.Context) context.Context {
			return tracing.NewContextWithTracerProvider(ctx, tp)
		})
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

	c = &Component{
		ctx:                ctx,
		cancelCtx:          cancel,
		terminationSignals: make(chan os.Signal),

		config: config,
		logger: logger,

		tcpListeners: make(map[string]*listener),

		taskStarter: task.StartTaskFunc(task.DefaultStartTask),
	}

	c.healthHandler, err = healthcheck.NewDefaultHealthChecker()
	if err != nil {
		return nil, err
	}

	c.componentKEKLabeler, err = config.KeyVault.ComponentKEKLabeler()
	if err != nil {
		return nil, err
	}
	c.keyService, err = config.KeyVault.KeyService(ctx, c)
	if err != nil {
		return nil, err
	}

	c.limiter, err = ratelimit.New(ctx, config.RateLimiting, config.Blob, c)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(c)
	}

	fpsFetcher, err := config.FrequencyPlansFetcher(ctx, c)
	if err != nil {
		return nil, err
	}
	c.frequencyPlans = frequencyplans.NewStore(fpsFetcher)

	caStoreFetcher, err := config.MTLSAuthCAStoreFetcher(ctx, c)
	if err != nil {
		return nil, err
	}
	c.caStore, err = mtls.NewCAStore(ctx, caStoreFetcher)
	if err != nil {
		return nil, err
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

	config.Interop.SenderClientCA.BlobConfig = config.Blob
	c.interop, err = interop.NewServer(c, c.FillContext, config.Interop)
	if err != nil {
		return nil, err
	}

	config.TTGC.TLS.KeyVault.CertificateProvider = c.keyService

	c.initRights()

	c.initGRPC()

	if !config.ServiceBase.SkipVersionCheck {
		c.RegisterTask(versionCheckTask(ctx, c))
	}

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

// FromRequestContext returns a derived context from the component context with key values from the request context.
// This can be used to decouple the lifetime from the request context while keeping security information.
func (c *Component) FromRequestContext(ctx context.Context) context.Context {
	return &crossContext{
		valueCtx:  ctx,
		cancelCtx: c.ctx,
	}
}

// ComponentKEKLabeler returns the component's ComponentKEKLabeler
func (c *Component) ComponentKEKLabeler() crypto.ComponentKEKLabeler {
	return c.componentKEKLabeler
}

// KeyService returns the component's KeyService.
func (c *Component) KeyService() crypto.KeyService {
	return c.keyService
}

// FrequencyPlansStore returns the component's frequencyPlans Store
func (c *Component) FrequencyPlansStore(ctx context.Context) (*frequencyplans.Store, error) {
	return c.frequencyPlans, nil
}

// GRPCServer returns the component's gRPC server.
func (c *Component) GRPCServer() *rpcserver.Server {
	return c.GRPC
}

// Start starts the component.
func (c *Component) Start() (err error) {
	if c.GRPC != nil {
		c.logger.Debug("Initializing gRPC server...")
		if err = c.setupGRPC(); err != nil {
			return err
		}
		serviceInfo := c.GRPC.Server.GetServiceInfo()
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
	c.web.RootRouter().PathPrefix("/").Handler(c.web.Router())

	c.logger.Debug("Initializing interop server...")
	for _, sub := range c.interopSubsystems {
		sub.RegisterInterop(c.interop)
	}

	if c.GRPC != nil {
		c.logger.Debug("Starting gRPC server...")
		if err = c.listenGRPC(); err != nil {
			c.logger.WithError(err).Error("Could not start gRPC server")
			return err
		}
		c.web.Prefix(ttnpb.HTTPAPIPrefix + "/").Handler(http.StripPrefix(ttnpb.HTTPAPIPrefix, c.GRPC))
		c.logger.Debug("Started gRPC server")
	}

	c.logger.Debug("Starting web server...")
	if err = c.listenWeb(); err != nil {
		c.logger.WithError(err).Error("Could not start web server")
		return err
	}
	c.logger.Debug("Started web server")

	c.logger.Debug("Starting interop server")
	if err = c.listenInterop(); err != nil {
		c.logger.WithError(err).Error("Could not start interop server")
		return err
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

	signal.Notify(c.terminationSignals, os.Interrupt, syscall.SIGTERM)

	sig := <-c.terminationSignals
	fmt.Println()
	c.logger.WithField("signal", sig).Info("Received signal, exiting...")
	return nil
}

// Close closes the server.
func (c *Component) Close() {
	c.cancelCtx()

	c.tcpListenersMu.Lock()
	defer c.tcpListenersMu.Unlock()
	for _, l := range c.tcpListeners {
		err := l.Close()
		if err != nil && c.ctx.Err() == nil {
			c.logger.WithError(err).Errorf("Error while stopping to listen on %s", l.Addr())
			continue
		}
		c.logger.Debugf("Stopped listening on %s", l.Addr())
	}

	if c.loopback != nil {
		c.logger.Debug("Stopping gRPC client...")
		c.loopback.Close()
		c.logger.Debug("Stopped gRPC client")
	}

	if c.GRPC != nil {
		c.logger.Debug("Stopping gRPC server...")
		c.GRPC.Stop()
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
		c.GRPC.Server.ServeHTTP(w, r)
	} else {
		c.web.ServeHTTP(w, r)
	}
}

// CAStore returns the component's CA Store.
func (c *Component) CAStore() *mtls.CAStore {
	return c.caStore
}
