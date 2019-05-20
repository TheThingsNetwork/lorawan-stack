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

// GetTLSConfig gets the component's TLS config and applies the given options.
func (c *Component) GetTLSConfig(ctx context.Context, opts ...TLSConfigOption) (*tls.Config, error) {
	var conf *tls.Config
	if c.acme != nil {
		conf = &tls.Config{
			GetCertificate: c.acme.GetCertificate,
		}
		opts = append(opts, WithNextProtos(acme.ALPNProto))
	} else {
		var err error
		conf, err = c.config.TLS.Config(ctx)
		if err != nil {
			return nil, err
		}
	}
	conf.MinVersion = tls.VersionTLS12
	conf.NextProtos = []string{"h2", "http/1.1"}
	conf.PreferServerCipherSuites = true
	for _, opt := range opts {
		opt.apply(conf)
	}
	return conf, nil
}
