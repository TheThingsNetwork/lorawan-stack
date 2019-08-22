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

package mqtt

import (
	"crypto/tls"
	"crypto/x509"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

var errInvalidCA = errors.DefineInvalidArgument("ca_pem_data", "CA PEM data is invalid")

func createTLSConfig(caPEM []byte, certPEM []byte, keyPEM []byte) (*tls.Config, error) {
	// Change the CA certificate pool only if a CA has been provided.
	// This allows the system-wide CA pool to be used.
	var certPool *x509.CertPool
	if len(caPEM) != 0 {
		certPool = x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caPEM) {
			return nil, errInvalidCA
		}
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
