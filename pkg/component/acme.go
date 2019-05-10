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

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var (
	errMissingACMEDir      = errors.Define("missing_acme_dir", "missing ACME storage directory")
	errMissingACMEEndpoint = errors.Define("missing_acme_endpoint", "missing ACME endpoint")
)

func (c *Component) initACME() error {
	if !c.config.TLS.ACME.Enable {
		return nil
	}
	if c.config.TLS.ACME.Endpoint == "" {
		return errMissingACMEEndpoint
	}
	if c.config.TLS.ACME.Dir == "" {
		return errMissingACMEDir
	}
	c.acme = &autocert.Manager{
		Cache:      autocert.DirCache(c.config.TLS.ACME.Dir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(c.config.TLS.ACME.Hosts...),
		Client: &acme.Client{
			DirectoryURL: c.config.TLS.ACME.Endpoint,
		},
		Email: c.config.TLS.ACME.Email,
	}
	c.acmeTLS = &tls.Config{
		GetCertificate:           c.acme.GetCertificate,
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		NextProtos: []string{
			"h2", "http/1.1",
			acme.ALPNProto,
		},
	}
	c.web.Any(".well-known/acme-challenge/*", echo.WrapHandler(c.acme.HTTPHandler(nil)))
	return nil
}

// GetTLSConfig gets the component's TLS config.
func (c *Component) GetTLSConfig(ctx context.Context) (*tls.Config, error) {
	if c.acmeTLS != nil {
		return c.acmeTLS, nil
	}
	return c.config.TLS.Config(ctx)
}
