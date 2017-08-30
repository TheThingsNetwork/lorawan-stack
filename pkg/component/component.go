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
	log    log.Interface

	handler http.Handler
	grpc    *rpcserver.Server

	httpL  net.Listener
	httpsL net.Listener
	grpcL  net.Listener
	grpcsL net.Listener
}

// New returns a new component
func New(logger log.Interface, config *Config) *Component {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.WithLogger(ctx, logger)

	c := &Component{
		ctx:       ctx,
		cancelCtx: cancel,

		config: config,
		log:    logger,
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
func (c *Component) Logger() log.Interface {
	return c.log
}

// Context returns the context of the component
func (c *Component) Context() context.Context {
	return c.ctx
}

// Start starts the component
func (c *Component) Start() error {
	c.log.Debug("Starting component")

	errors := make(chan error, 10)
	signals := make(chan os.Signal)

	defer c.Close()

	if c.handler == nil {
		http.Handle(rpcserver.APIPrefix, http.StripPrefix(rpcserver.APIPrefix, c.grpc))
	}

	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)

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
			c.log.WithField("Signal", sig).Info("Received signal, exiting...")
			return nil
		}
	}
}

// Close closes the server
func (c *Component) Close() {
	c.cancelCtx()

	if c.httpL != nil {
		_ = c.httpL.Close()
		c.log.Debug("Stopped listening on HTTP")
	}

	if c.httpsL != nil {
		_ = c.httpsL.Close()
		c.log.Debug("Stopped listening on HTTPS")
	}

	if c.grpc != nil {
		c.grpc.Stop()
		c.log.Debug("Stopped gRPC server")
	}

	if c.grpcL != nil {
		c.grpcL.Close()
		c.log.Debug("Stopped listening on gRPC")
	}

	if c.grpcsL != nil {
		c.grpcsL.Close()
		c.log.Debug("Stopped listening on gRPCs")
	}
}

func (c *Component) listenHTTP() error {
	addr := c.config.HTTP.HTTP
	if addr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.httpL = listener

	c.log.WithField("Address", addr).Debug("HTTP server listening")
	return http.Serve(listener, c.handler)
}

func (c *Component) listenHTTPS() error {
	addr := c.config.HTTP.HTTPS
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

	c.log.WithField("Address", addr).Debug("HTTPS server listening")
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
	addr := c.config.GRPC.TCP
	if addr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.grpcL = listener

	c.log.WithField("Address", addr).Debug("gRPC server listening")
	return c.grpc.Serve(listener)
}

func (c *Component) listenGRPCS() error {
	addr := c.config.GRPC.TLS
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

	c.grpcsL = listener

	c.log.WithField("Address", addr).Debug("gRPCs server listening")
	return c.grpc.Serve(listener)
}
