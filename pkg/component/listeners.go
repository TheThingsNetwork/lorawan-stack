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

package component

import (
	"crypto/tls"
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/soheilhy/cmux"
)

func (c *Component) tlsConfig() (*tls.Config, error) {
	if c.config.TLS.Certificate == "" || c.config.TLS.Key == "" {
		return nil, errors.New("No TLS certificate or key specified")
	}
	certificate, err := tls.LoadX509KeyPair(c.config.TLS.Certificate, c.config.TLS.Key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates:             []tls.Certificate{certificate},
		PreferServerCipherSuites: true,
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

// Listen on an address
func (c *Component) Listen(address string) (Listener, error) {
	l, ok := c.listeners[address]
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
		c.listeners[address] = l
		go func() {
			c.logger.WithField("address", l.lis.Addr().String()).Debug("Start serving")
			if err := l.mux.Serve(); err != nil {
				c.logger.WithError(err).Errorf("Error in Listener %s", l.lis.Addr())
			}
		}()
	}
	return l, nil
}
