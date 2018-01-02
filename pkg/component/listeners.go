// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

type listener struct {
	c   *Component
	lis net.Listener
	tcp cmux.CMux
	tls cmux.CMux
}

// Listener that accepts multiple protocols on the same port
type Listener interface {
	TLS() (cmux.CMux, error)
	TCP() (cmux.CMux, error)
	Close() error
}

func (l *listener) TLS() (cmux.CMux, error) {
	if l.tls == nil {
		config, err := l.c.tlsConfig()
		if err != nil {
			return nil, err
		}
		tcp, err := l.TCP()
		if err != nil {
			return nil, err
		}
		l.tls = cmux.New(tls.NewListener(tcp.Match(cmux.TLS()), config))
	}
	return l.tls, nil
}

func (l *listener) TCP() (cmux.CMux, error) {
	if l.tcp == nil {
		l.tcp = cmux.New(l.lis)
	}
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
		l = &listener{c: c, lis: lis}
		c.listeners[address] = l
	}
	return l, nil
}

func (c *Component) startListeners() {
	for _, lis := range c.listeners {
		l := lis // shadow the listener
		if l.tcp != nil {
			go func() {
				c.logger.WithField("address", l.lis.Addr().String()).Debug("Start serving (TCP)")
				err := l.tcp.Serve()
				if err != nil {
					c.logger.WithError(err).Errorf("Error in TCP Listener %s", l.lis.Addr())
				}
			}()
		}
		if l.tls != nil {
			go func() {
				c.logger.WithField("address", l.lis.Addr().String()).Debug("Start serving (TLS)")
				err := l.tls.Serve()
				if err != nil {
					c.logger.WithError(err).Errorf("Error in TLS Listener %s", l.lis.Addr())
				}
			}()
		}
	}
}
