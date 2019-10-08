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
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/fs"
	"go.thethings.network/lorawan-stack/pkg/log"
	"golang.org/x/crypto/acme"
)

// TLSConfigOption provides customization for TLS configuration.
type TLSConfigOption interface {
	apply(*tls.Config)
}

// TLSConfigOptionFunc is a TLSConfigOption.
type TLSConfigOptionFunc func(*tls.Config)

func (fn TLSConfigOptionFunc) apply(c *tls.Config) {
	fn(c)
}

// WithTLSClientAuth sets TLS client authentication options.
func WithTLSClientAuth(auth tls.ClientAuthType, cas *x509.CertPool, verify func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error) TLSConfigOption {
	return TLSConfigOptionFunc(func(c *tls.Config) {
		c.ClientAuth = auth
		c.ClientCAs = cas
		c.VerifyPeerCertificate = verify
	})
}

// WithNextProtos appends the given protocols to NextProtos.
func WithNextProtos(protos ...string) TLSConfigOption {
	return TLSConfigOptionFunc(func(c *tls.Config) {
		c.NextProtos = append(c.NextProtos, protos...)
	})
}

var (
	errAmbiguousTLSConfig = errors.DefineFailedPrecondition("tls_config_ambiguous", "ambiguous TLS configuration")
	errEmptyTLSConfig     = errors.DefineFailedPrecondition("tls_config_empty", "empty TLS configuration")
	errTLSKeyVaultID      = errors.DefineFailedPrecondition("tls_key_vault_id", "invalid TLS key vault ID")
)

// GetTLSServerConfig gets the component's server TLS config and applies the given options.
func (c *Component) GetTLSServerConfig(ctx context.Context, opts ...TLSConfigOption) (*tls.Config, error) {
	var (
		logger = log.FromContext(ctx)
		conf   = c.GetBaseConfig(ctx).TLS
		res    *tls.Config
	)
	for _, src := range []struct {
		Enable            bool
		CertificateGetter func() (func(*tls.ClientHelloInfo) (*tls.Certificate, error), error)
	}{
		{
			Enable: conf.Certificate != "" && conf.Key != "",
			CertificateGetter: func() (func(*tls.ClientHelloInfo) (*tls.Certificate, error), error) {
				var cv atomic.Value
				loadCertificate := func() error {
					cert, err := tls.LoadX509KeyPair(conf.Certificate, conf.Key)
					if err != nil {
						return err
					}
					cv.Store(&cert)
					logger.Debug("Loaded TLS certificate")
					return nil
				}
				if err := loadCertificate(); err != nil {
					return nil, err
				}
				debounce := make(chan struct{}, 1)
				fs.Watch(conf.Certificate, events.HandlerFunc(func(evt events.Event) {
					if evt.Name() != "fs.write" {
						return
					}
					// We have to debounce this; OpenSSL typically causes a lot of write events.
					select {
					case debounce <- struct{}{}:
						time.AfterFunc(5*time.Second, func() {
							if err := loadCertificate(); err != nil {
								logger.WithError(err).Error("Could not reload TLS certificate")
								return
							}
							<-debounce
						})
					default:
					}
				}))
				return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
					return cv.Load().(*tls.Certificate), nil
				}, nil
			},
		},
		{
			Enable: conf.ACME.Enable,
			CertificateGetter: func() (func(*tls.ClientHelloInfo) (*tls.Certificate, error), error) {
				opts = append(opts, WithNextProtos(acme.ALPNProto))
				return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
					if hello.ServerName == "" {
						hello.ServerName = conf.ACME.DefaultHost
					}
					return c.acme.GetCertificate(hello)
				}, nil
			},
		},
		{
			Enable: conf.KeyVault.Enable,
			CertificateGetter: func() (func(*tls.ClientHelloInfo) (*tls.Certificate, error), error) {
				if conf.KeyVault.ID == "" {
					return nil, errTLSKeyVaultID
				}
				return func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
					return c.KeyVault.LoadCertificate(conf.KeyVault.ID)
				}, nil
			},
		},
	} {
		if !src.Enable {
			continue
		}
		if res != nil {
			return nil, errAmbiguousTLSConfig
		}
		fn, err := src.CertificateGetter()
		if err != nil {
			return nil, err
		}
		res = &tls.Config{
			GetCertificate: fn,
		}
	}
	if res == nil {
		return nil, errEmptyTLSConfig
	}
	res.MinVersion = tls.VersionTLS12
	res.PreferServerCipherSuites = true
	for _, opt := range opts {
		opt.apply(res)
	}
	return res, nil
}

// GetTLSClientConfig gets the component's client TLS config and applies the given options.
func (c *Component) GetTLSClientConfig(ctx context.Context, opts ...TLSConfigOption) (*tls.Config, error) {
	conf := c.GetBaseConfig(ctx).TLS
	res := &tls.Config{}
	if conf.RootCA != "" {
		pem, err := ioutil.ReadFile(conf.RootCA)
		if err != nil {
			return nil, err
		}
		res.RootCAs = x509.NewCertPool()
		res.RootCAs.AppendCertsFromPEM(pem)
	}
	res.InsecureSkipVerify = conf.InsecureSkipVerify
	res.MinVersion = tls.VersionTLS12
	for _, opt := range opts {
		opt.apply(res)
	}
	return res, nil
}
