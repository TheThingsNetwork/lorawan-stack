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

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type clientCertificateContextKeyType struct{}

var clientCertificateContextKey clientCertificateContextKeyType

// NewContextWithClientCertificate returns a context derived from parent that contains
// the client TLS certificate.
func NewContextWithClientCertificate(parent context.Context, cert *x509.Certificate) context.Context {
	return context.WithValue(parent, clientCertificateContextKey, cert)
}

// ClientCertificateFromContext returns the certificate from the context if present.
// If the certificate is not present in the context, it tries to extract it from the peer.
func ClientCertificateFromContext(ctx context.Context) *x509.Certificate {
	if cert, ok := ctx.Value(clientCertificateContextKey).(*x509.Certificate); ok {
		return cert
	}
	if p, ok := peer.FromContext(ctx); ok {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			if len(tlsInfo.State.PeerCertificates) > 0 {
				return tlsInfo.State.PeerCertificates[0]
			}
		}
	}
	return nil
}

type rootCAsKeyType struct{}

var rootCAsKey rootCAsKeyType

// RootCAsFromContext returns the root CAs from the context if present.
func RootCAsFromContext(ctx context.Context) *x509.CertPool {
	if pool, ok := ctx.Value(rootCAsKey).(*x509.CertPool); ok {
		return pool
	}
	return nil
}

// AppendRootCAsToContext appends the given PEM encoded Root Certificates to the root CAs in the context.
func AppendRootCAsToContext(parent context.Context, pem []byte) context.Context {
	certPool := RootCAsFromContext(parent)
	if certPool != nil {
		certPool = certPool.Clone()
	} else {
		certPool = x509.NewCertPool()
	}
	certPool.AppendCertsFromPEM(pem)
	return context.WithValue(parent, rootCAsKey, certPool)
}
