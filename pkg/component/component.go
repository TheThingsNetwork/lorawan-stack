// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Config is the type of configuration for Components
type Config struct {
	config.ServiceBase `name:",squash"`
}

// Component is a base component for The Things Network cluster
type Component struct {
	config *Config
	log    log.Interface
	http   *http.Server
	https  *http.Server
	grpc   *grpc.Server
	grpcs  *grpc.Server
}

// New returns a new component
func New(log log.Interface, config *Config) (*Component, error) {

	cert := config.TLS.Certificate
	key := config.TLS.Key

	var grpcs *grpc.Server
	if cert != "" && key != "" {
		creds, err := credentials.NewServerTLSFromFile(cert, key)

		if err != nil {
			return nil, err
		}

		grpcs = grpc.NewServer(grpc.Creds(creds))
	}

	return &Component{
		config: config,
		log:    log,
		http:   &http.Server{Addr: config.ServiceBase.HTTP.HTTP},
		https:  &http.Server{Addr: config.ServiceBase.HTTP.HTTPS},
		grpc:   grpc.NewServer(),
		grpcs:  grpcs,
	}, nil
}

// Start starts the component
func (c *Component) Start() error {
	c.log.Debug("Starting component")

	errors := make(chan error, 10)
	signals := make(chan os.Signal)

	defer c.Close()

	signal.Notify(signals, os.Interrupt)

	go func() { errors <- c.startHTTP() }()
	go func() { errors <- c.startHTTPS() }()
	go func() { errors <- c.startGRPC() }()
	go func() { errors <- c.startGRPCS() }()

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
	if c.http != nil {
		_ = c.http.Close()
		c.log.Debug("Stopped HTTP server")
	}

	if c.https != nil {
		_ = c.https.Close()
		c.log.Debug("Stopped HTTPS server")
	}

	if c.grpc != nil {
		c.grpc.Stop()
		c.log.Debug("Stopped gRPC server")
	}

	if c.grpcs != nil {
		c.grpcs.Stop()
		c.log.Debug("Stopped gRPCs server")
	}
}

func (c *Component) startHTTP() error {
	addr := c.config.HTTP.HTTP
	if addr == "" {
		return nil
	}

	c.log.WithField("Address", addr).Debug("HTTP server listening")
	return c.http.ListenAndServe()
}

func (c *Component) startHTTPS() error {
	addr := c.config.HTTP.HTTPS
	cert := c.config.TLS.Certificate
	key := c.config.TLS.Key

	if addr == "" || cert == "" || key == "" {
		return nil
	}

	c.log.WithField("Address", c.https.Addr).Debug("HTTPS server listening")
	return c.https.ListenAndServeTLS(cert, key)
}

func (c *Component) startGRPC() error {
	if c.grpc == nil {
		return nil
	}

	addr := c.config.GRPC.TCP
	if addr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.log.WithField("Address", addr).Debug("gRPC server listening")
	return c.grpc.Serve(listener)
}

func (c *Component) startGRPCS() error {
	if c.grpcs == nil {
		return nil
	}

	addr := c.config.GRPC.TLS
	if addr == "" {
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.log.WithField("Address", addr).Debug("gRPCs server listening")
	return c.grpcs.Serve(listener)
}
