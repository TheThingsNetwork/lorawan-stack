// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"google.golang.org/grpc"
)

// Subsystem that can be registered to the component
type Subsystem interface {
	rpcserver.Registerer
}

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

	handler http.Handler
	grpc    *rpcserver.Server

	subsystems []Subsystem

	loopback *grpc.ClientConn

	httpL  net.Listener
	httpsL net.Listener
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
	}

	c.grpc = rpcserver.New(
		c.ctx,
		rpcserver.WithContextFiller(func(ctx context.Context) context.Context {
			// TODO: Fill globals in call context (data stores, config, ...)
			return ctx
		}),
	)
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

// Register a subsystem to the component
func (c *Component) Register(s Subsystem) {
	c.subsystems = append(c.subsystems, s)
}

// Start starts the component
func (c *Component) Start() (err error) {
	defer c.Close()

	c.logger.Debug("Setting up gRPC server")

	for _, sub := range c.subsystems {
		sub.RegisterServices(c.grpc.Server)
	}

	c.logger.Debug("Starting loopback connection")

	c.loopback, err = rpcserver.StartLoopback(c.grpc.Server)
	if err != nil {
		return
	}

	c.logger.Debug("Setting up gRPC gateway")

	for _, sub := range c.subsystems {
		sub.RegisterHandlers(c.grpc.ServeMux, c.loopback)
	}
	if c.handler == nil {
		http.Handle(rpcserver.APIPrefix, http.StripPrefix(rpcserver.APIPrefix, c.grpc))
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)

	c.logger.Debug("Starting servers")

	errors := make(chan error, 10)
	go func() { errors <- c.listenHTTP() }()
	go func() { errors <- c.listenHTTPS() }()
	go func() { errors <- c.listenGRPC() }()
	go func() { errors <- c.listenGRPCS() }()

	for {
		select {
		case err := <-errors:
			if err != nil {
				return err
			}
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

	if c.httpL != nil {
		_ = c.httpL.Close()
		c.logger.Debug("Stopped listening on HTTP")
	}

	if c.httpsL != nil {
		_ = c.httpsL.Close()
		c.logger.Debug("Stopped listening on HTTPS")
	}

	if c.grpc != nil {
		c.grpc.Stop() // This also closes all gRPC listeners
		c.logger.Debug("Stopped gRPC server")
	}
}

func (c *Component) listenHTTP() error {
	addr := c.config.HTTP.Listen
	if addr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.httpL = listener

	c.logger.WithField("address", addr).Debug("HTTP server listening")
	return http.Serve(listener, c.handler)
}

func (c *Component) listenHTTPS() error {
	addr := c.config.HTTP.ListenTLS
	cert := c.config.TLS.Certificate
	key := c.config.TLS.Key

	if addr == "" {
		return nil
	}

	if cert == "" || key == "" {
		return fmt.Errorf("Cannot set up HTTPS server without certificate and key")
	}

	listener, err := c.listenTLS(addr, cert, key)
	if err != nil {
		return err
	}

	c.httpsL = listener

	c.logger.WithField("address", addr).Debug("HTTPS server listening")
	return http.Serve(listener, c.handler)
}

func (c *Component) listenTLS(addr string, certFile string, keyFile string) (net.Listener, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return tls.Listen("tcp", addr, &tls.Config{
		Certificates:             []tls.Certificate{certificate},
		PreferServerCipherSuites: true,
	})
}

func (c *Component) listenGRPC() error {
	addr := c.config.GRPC.Listen
	if addr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.logger.WithField("address", addr).Debug("gRPC server listening")
	return c.grpc.Serve(listener)
}

func (c *Component) listenGRPCS() error {
	addr := c.config.GRPC.ListenTLS
	cert := c.config.TLS.Certificate
	key := c.config.TLS.Key

	if addr == "" {
		return nil
	}

	if cert == "" || key == "" {
		return fmt.Errorf("Cannot set up HTTPS server without certificate and key")
	}

	listener, err := c.listenTLS(addr, cert, key)
	if err != nil {
		return err
	}

	c.logger.WithField("address", addr).Debug("gRPCs server listening")
	return c.grpc.Serve(listener)
}
