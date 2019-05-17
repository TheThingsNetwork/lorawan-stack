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

package component

import (
	"crypto/tls"
	"net"

	"github.com/soheilhy/cmux"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
)

var (
	errListenEndpoint = errors.Define("listen_endpoint", "could not listen on `{endpoint}` address")
	errListener       = errors.Define("listener", "could not create `{protocol}` listener")
)

// Listener that accepts multiple protocols on the same port
type Listener interface {
	TLS(opts ...TLSConfigOption) (net.Listener, error)
	TCP() (net.Listener, error)
	Close() error
}

type listener struct {
	c       *Component
	lis     net.Listener
	mux     cmux.CMux
	tcp     net.Listener
	tcpUsed bool
	tls     net.Listener
	tlsUsed bool
}

func (l *listener) TLS(opts ...TLSConfigOption) (net.Listener, error) {
	if l.tlsUsed {
		return nil, errors.New("TLS listener already in use")
	}
	config, err := l.c.GetTLSConfig(l.c.Context(), opts...)
	if err != nil {
		return nil, err
	}
	l.tlsUsed = true
	return tls.NewListener(l.tls, config), nil
}

func (l *listener) TCP() (net.Listener, error) {
	if l.tcpUsed {
		return nil, errors.New("TCP listener already in use")
	}
	l.tcpUsed = true
	return l.tcp, nil
}

func (l *listener) Close() error {
	return l.lis.Close()
}

// ListenTCP listens on a TCP address and allows for TCP and TLS on the same port.
func (c *Component) ListenTCP(address string) (Listener, error) {
	l, ok := c.tcpListeners[address]
	if !ok {
		c.logger.WithField("address", address).Debug("Creating listener")
		lis, err := net.Listen("tcp", address)
		if err != nil {
			return nil, err
		}
		mux := cmux.New(lis)
		l = &listener{
			c:   c,
			lis: lis,
			mux: mux,
			tls: mux.Match(cmux.TLS()),
			tcp: mux.Match(cmux.Any()),
		}
		c.tcpListeners[address] = l
		go func() {
			c.logger.WithField("address", l.lis.Addr().String()).Debug("Start serving")
			if err := l.mux.Serve(); err != nil && c.ctx.Err() == nil {
				c.logger.WithError(err).Errorf("Error in Listener %s", l.lis.Addr())
			}
		}()
	}
	return l, nil
}

// ListenUDP starts a listener on a UDP address.
func (c *Component) ListenUDP(address string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP("udp", udpAddr)
}

// Endpoint represents an endpoint that can be listened on.
type Endpoint interface {
	Address() string
	Protocol() string
	Listen(Listener) (net.Listener, error)
}

type tcpEndpoint struct {
	address  string
	protocol string
}

// NewTCPEndpoint returns a new TCP endpoint.
func NewTCPEndpoint(address, protocol string) Endpoint {
	return &tcpEndpoint{address, protocol}
}

func (e tcpEndpoint) Address() string                       { return e.address }
func (e tcpEndpoint) Protocol() string                      { return e.protocol }
func (tcpEndpoint) Listen(l Listener) (net.Listener, error) { return l.TCP() }

type tlsEndpoint struct {
	address    string
	protocol   string
	configOpts []TLSConfigOption
}

// NewTLSEndpoint returns a new TLS endpoint.
func NewTLSEndpoint(address, protocol string, configOpts ...TLSConfigOption) Endpoint {
	return &tlsEndpoint{address, protocol, configOpts}
}

func (e tlsEndpoint) Address() string                         { return e.address }
func (e tlsEndpoint) Protocol() string                        { return e.protocol + "/tls" }
func (e tlsEndpoint) Listen(l Listener) (net.Listener, error) { return l.TLS(e.configOpts...) }

func (c *Component) serveOnEndpoints(endpoints []Endpoint, serve func(*Component, net.Listener) error, namespace string) error {
	for _, endpoint := range endpoints {
		if endpoint.Address() == "" {
			continue
		}
		err := c.serveOnEndpoint(endpoint, serve, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Component) serveOnEndpoint(endpoint Endpoint, serve func(*Component, net.Listener) error, namespace string) error {
	l, err := c.ListenTCP(endpoint.Address())
	if err != nil {
		return errListenEndpoint.WithAttributes("endpoint", endpoint.Address()).WithCause(err)
	}
	lis, err := endpoint.Listen(l)
	if err != nil {
		return errListener.WithAttributes("protocol", endpoint.Protocol()).WithCause(err)
	}
	logger := log.FromContext(c.ctx).WithFields(log.Fields(
		"namespace", namespace,
		"address", endpoint.Address(),
		"protocol", endpoint.Protocol(),
	))
	logger.Info("Listening for connections")
	go func() {
		err := serve(c, lis)
		if err != nil && c.ctx.Err() == nil {
			logger.WithError(err).Error("Failed to serve")
		}
	}()
	return nil
}
