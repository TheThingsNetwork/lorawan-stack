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

// WithTLSCertificates sets TLS certificates.
func WithTLSCertificates(certificates ...tls.Certificate) TLSConfigOption {
	return TLSConfigOptionFunc(func(c *tls.Config) {
		c.Certificates = certificates
	})
}

// WithNextProtos appends the given protocols to NextProtos.
func WithNextProtos(protos ...string) TLSConfigOption {
	return TLSConfigOptionFunc(func(c *tls.Config) {
		c.NextProtos = append(c.NextProtos, protos...)
	})
}

// GetTLSServerConfig gets the component's server TLS config and applies the given options.
func (c *Component) GetTLSServerConfig(ctx context.Context, opts ...TLSConfigOption) (*tls.Config, error) {
	conf := c.GetBaseConfig(ctx).TLS
	res := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	// TODO: Remove detection mechanism (https://github.com/TheThingsNetwork/lorawan-stack/issues/1450)
	if conf.Source == "" {
		switch {
		case conf.ACME.Enable:
			conf.Source = "acme"
		case conf.Certificate != "" && conf.Key != "":
			conf.Source = "file"
		case !conf.KeyVault.IsZero():
			conf.Source = "key-vault"
		}
	}
	if conf.Source == "key-vault" {
		conf.KeyVault.KeyVault = c.KeyVault
	}
	if err := conf.ServerAuth.ApplyTo(res); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt.apply(res)
	}
	return res, nil
}

// GetTLSClientConfig gets the component's client TLS config and applies the given options.
func (c *Component) GetTLSClientConfig(ctx context.Context, opts ...TLSConfigOption) (*tls.Config, error) {
	conf := c.GetBaseConfig(ctx).TLS
	res := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: conf.InsecureSkipVerify,
	}
	if err := conf.Client.ApplyTo(res); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt.apply(res)
	}
	return res, nil
}
