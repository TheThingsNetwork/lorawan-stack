// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package tlsconfig

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// Option provides customization for TLS configuration.
type Option interface {
	apply(*tls.Config)
}

// ConfigOptionFunc is a Option.
type ConfigOptionFunc func(*tls.Config)

func (fn ConfigOptionFunc) apply(c *tls.Config) {
	fn(c)
}

// WithTLSClientAuth sets TLS client authentication options.
func WithTLSClientAuth(
	auth tls.ClientAuthType,
	cas *x509.CertPool,
	verify func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error,
) Option {
	return ConfigOptionFunc(func(c *tls.Config) {
		c.ClientAuth = auth
		c.ClientCAs = cas
		c.VerifyPeerCertificate = verify
	})
}

// WithTLSCertificates sets TLS certificates.
func WithTLSCertificates(certificates ...tls.Certificate) Option {
	return ConfigOptionFunc(func(c *tls.Config) {
		c.Certificates = certificates
	})
}

// WithNextProtos appends the given protocols to NextProtos.
func WithNextProtos(protos ...string) Option {
	return ConfigOptionFunc(func(c *tls.Config) {
		c.NextProtos = append(c.NextProtos, protos...)
	})
}

// ConfigurationProvider generates a Config from the provided context.
type ConfigurationProvider func(context.Context) Config

// GetTLSServerConfig gets the component's server TLS config and applies the given options.
func (p ConfigurationProvider) GetTLSServerConfig(ctx context.Context, opts ...Option) (*tls.Config, error) {
	conf := p(ctx)
	cipherSuites, err := conf.GetCipherSuites()
	if err != nil {
		return nil, err
	}
	if cipherSuites != nil {
		log.FromContext(ctx).Warn(
			"TLS is configured to use a custom set of cipher suites.",
			"Make sure that your list is up to date, or disable the custom cipher suites",
		)
	}
	res := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		CipherSuites: cipherSuites,
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
func (p ConfigurationProvider) GetTLSClientConfig(ctx context.Context, opts ...Option) (*tls.Config, error) {
	conf := p(ctx)
	res := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: conf.InsecureSkipVerify, //nolint:gosec
	}
	if err := conf.Client.ApplyTo(res); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt.apply(res)
	}
	return res, nil
}
