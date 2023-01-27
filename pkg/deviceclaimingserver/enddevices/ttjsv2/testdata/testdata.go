// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// NOTE: Set CAROOT=. when running go generate, i.e.:
// $ CAROOT=. go generate .

//go:generate mkcert -cert-file servercert.pem -key-file serverkey.pem localhost 127.0.0.1 ::1
//go:generate mkcert -cert-file clientcert-1.pem -key-file clientkey-1.pem -client client1.local
//go:generate mkcert -cert-file clientcert-2.pem -key-file clientkey-2.pem -client client2.local

// Package testdata provides test data.
package testdata

import (
	"crypto/x509"
	_ "embed"
	"encoding/pem"
)

var (
	//go:embed clientcert-1.pem
	client1CertData []byte
	//go:embed clientcert-2.pem
	client2CertData []byte
)

func x509Certificate(pemData []byte) *x509.Certificate {
	b, _ := pem.Decode(pemData)
	if b == nil || b.Type != "CERTIFICATE" {
		panic("invalid PEM data")
	}
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}

// Client certificates.
var (
	Client1Cert = x509Certificate(client1CertData)
	Client2Cert = x509Certificate(client2CertData)
)
