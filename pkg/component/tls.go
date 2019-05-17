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

	"golang.org/x/crypto/acme"
)

// GetTLSConfig gets the component's TLS config.
func (c *Component) GetTLSConfig(ctx context.Context) (*tls.Config, error) {
	if c.acme != nil {
		return &tls.Config{
			GetCertificate:           c.acme.GetCertificate,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
			NextProtos: []string{
				"h2", "http/1.1",
				acme.ALPNProto,
			},
		}, nil
	}
	return c.config.TLS.Config(ctx)
}
