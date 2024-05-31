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
	"context"
	"crypto/x509"
)

type clientCertificateContextKeyType struct{}

var clientCertificateContextKey clientCertificateContextKeyType

// NewContextWithClientCertificate returns a context derived from parent that contains
// the client TLS certificate.
func NewContextWithClientCertificate(parent context.Context, cert *x509.Certificate) context.Context {
	return context.WithValue(parent, clientCertificateContextKey, cert)
}

// ClientCertificateFromContext returns the certificate from the context if present.
func ClientCertificateFromContext(ctx context.Context) *x509.Certificate {
	if cert, ok := ctx.Value(clientCertificateContextKey).(*x509.Certificate); ok {
		return cert
	}
	return nil
}
