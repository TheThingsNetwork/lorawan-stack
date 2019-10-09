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
	"crypto/x509"
	"net"
	"net/http"

	"go.thethings.network/lorawan-stack/pkg/interop"
)

// RegisterInterop registers an interop subsystem to the component.
func (c *Component) RegisterInterop(s interop.Registerer) {
	c.interopSubsystems = append(c.interopSubsystems, s)
}

func (c *Component) serveInterop(lis net.Listener) error {
	return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.interop.ServeHTTP(w, r)
	}))
}

func (c *Component) interopEndpoints() []Endpoint {
	certPool := x509.NewCertPool()
	for _, certs := range c.interop.SenderClientCAs {
		for _, cert := range certs {
			certPool.AddCert(cert)
		}
	}
	return []Endpoint{
		// TODO: Enable TCP endpoint (https://github.com/TheThingsNetwork/lorawan-stack/issues/717)
		NewTLSEndpoint(c.config.Interop.ListenTLS, "Interop",
			WithTLSClientAuth(tls.RequireAndVerifyClientCert, certPool, nil),
			WithNextProtos("h2", "http/1.1"),
		),
	}
}

func (c *Component) listenInterop() error {
	return c.serveOnEndpoints(c.interopEndpoints(), (*Component).serveInterop, "interop")
}
