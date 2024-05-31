// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package mtls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// HeaderReader is an interface for reading headers, typically HTTP headers and gRPC metadata.
type HeaderReader interface {
	Get(key string) string
}

var errProxyHeaderValue = errors.DefineCorruption("proxy_header_value", "invalid proxy header value")

// FromProxyHeaders extracts a client certificate from proxy headers.
// This function supports Envoy Proxy and Traefik.
// If a proxy's header is set, it expects the value to contain a client certificate, otherwise an error is returned.
// If no proxy headers are set, it returns nil, nil.
func FromProxyHeaders(h HeaderReader) (*x509.Certificate, error) {
	for _, proxy := range []struct {
		key   string
		parse func(string) (*x509.Certificate, error)
	}{
		{
			// Envoy Proxy
			// See https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-client-cert
			key: "x-forwarded-client-cert",
			parse: func(value string) (*x509.Certificate, error) {
				parts := strings.Split(value, ";")
				for _, part := range parts {
					chainPEM, found := strings.CutPrefix(part, "Chain=")
					if !found {
						continue
					}
					chainPEM = strings.Trim(chainPEM, `\"`)
					chainPEM, err := url.PathUnescape(chainPEM)
					if err != nil {
						return nil, errProxyHeaderValue.WithCause(err)
					}
					block, _ := pem.Decode([]byte(chainPEM))
					if block == nil {
						return nil, errProxyHeaderValue.New()
					}
					cert, err := x509.ParseCertificate(block.Bytes)
					if err != nil {
						return nil, errProxyHeaderValue.WithCause(err)
					}
					return cert, nil
				}
				return nil, errProxyHeaderValue.New()
			},
		},
		{
			// Traefik
			// See https://doc.traefik.io/traefik/middlewares/http/passtlsclientcert/
			key: "x-forwarded-tls-client-cert",
			parse: func(value string) (*x509.Certificate, error) {
				leafPEM := strings.Split(value, ",")[0]
				leafPEM = fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", leafPEM)
				block, _ := pem.Decode([]byte(leafPEM))
				if block == nil {
					return nil, errProxyHeaderValue.New()
				}
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, errProxyHeaderValue.WithCause(err)
				}
				return cert, nil
			},
		},
	} {
		if value := h.Get(proxy.key); value != "" {
			return proxy.parse(value)
		}
	}
	return nil, nil
}
