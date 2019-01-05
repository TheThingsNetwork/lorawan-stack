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
	"sync/atomic"
	"time"

	"github.com/soheilhy/cmux"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/fs"
	"go.thethings.network/lorawan-stack/pkg/log"
)

var (
	errListenEndpoint = errors.Define("listen_endpoint", "could not listen on `{endpoint}` address")
	errListener       = errors.Define("listener", "could not create `{protocol}` listener")
)

func (c *Component) tlsConfig() (*tls.Config, error) {
	if c.config.TLS.Certificate == "" || c.config.TLS.Key == "" {
		return nil, errors.New("No TLS certificate or key specified")
	}
	var cv atomic.Value
	loadCertificate := func() error {
		certificate, err := tls.LoadX509KeyPair(c.config.TLS.Certificate, c.config.TLS.Key)
		if err != nil {
			return err
		}
		cv.Store([]tls.Certificate{certificate})
		c.Logger().Debug("Loaded TLS certificate")
		return nil
	}
	if err := loadCertificate(); err != nil {
		return nil, err
	}
	debounce := make(chan struct{}, 1)
	fs.Watch(c.config.TLS.Certificate, events.HandlerFunc(func(evt events.Event) {
		if evt.Name() != "fs.write" {
			return
		}
		// We have to debounce this; OpenSSL typically causes a lot of write events.
		select {
		case debounce <- struct{}{}:
			time.AfterFunc(5*time.Second, func() {
				if err := loadCertificate(); err != nil {
					c.Logger().WithError(err).Error("Could not reload TLS certificate")
					return
				}
				<-debounce
			})
		default:
		}
	}))

	return &tls.Config{
		GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			tlsConfig := &tls.Config{
				Certificates:             cv.Load().([]tls.Certificate),
				PreferServerCipherSuites: true,
				MinVersion:               tls.VersionTLS12,
			}
			for _, proto := range info.SupportedProtos {
				if proto == "h2" {
					tlsConfig.NextProtos = []string{"h2"}
					break
				}
			}
			return tlsConfig, nil
		},
	}, nil
}

// Listener that accepts multiple protocols on the same port
type Listener interface {
	TLS() (net.Listener, error)
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

func (l *listener) TLS() (net.Listener, error) {
	if l.tlsUsed {
		return nil, errors.New("TLS listener already in use")
	}
	config, err := l.c.tlsConfig()
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

type endpoint struct {
	toNativeListener func(Listener) (net.Listener, error)

	address, protocol string
}

func (c *Component) serveOnListeners(endpoints []endpoint, serve func(*Component, net.Listener) error, namespace string) error {
	for _, endpoint := range endpoints {
		if endpoint.address == "" {
			continue
		}
		err := c.serveOnListener(endpoint, serve, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Component) serveOnListener(endpoint endpoint, serve func(*Component, net.Listener) error, namespace string) error {
	l, err := c.ListenTCP(endpoint.address)
	if err != nil {
		return errListenEndpoint.WithAttributes("endpoint", endpoint.address).WithCause(err)
	}
	lis, err := endpoint.toNativeListener(l)
	if err != nil {
		return errListener.WithAttributes("protocol", endpoint.protocol).WithCause(err)
	}
	logger := log.FromContext(c.ctx).WithFields(log.Fields("namespace", namespace, "address", endpoint.address))
	logger.Infof("Listening for %s connections", endpoint.protocol)
	go func() {
		err := serve(c, lis)
		if err != nil && c.ctx.Err() == nil {
			logger.WithError(err).Errorf("Error serving %s on %s", endpoint.protocol, lis.Addr())
		}
	}()
	return nil
}
