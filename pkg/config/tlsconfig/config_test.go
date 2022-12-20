// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package tlsconfig_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockFileReader map[string][]byte

func (m mockFileReader) ReadFile(name string) ([]byte, error) {
	if f, ok := m[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("not found")
}

func genCert() (certPEM []byte, keyPEM []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(err)
	}
	now := time.Now()
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"The Things Testing Co"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, key.Public(), key)
	if err != nil {
		panic(err)
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		panic(err)
	}
	return pem.EncodeToMemory(&pem.Block{
			Type: "CERTIFICATE", Bytes: certBytes,
		}), pem.EncodeToMemory(&pem.Block{
			Type: "PRIVATE KEY", Bytes: keyBytes,
		})
}

func TestApplyTLSClientConfig(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	caCert, _ := genCert()
	tlsConfig := &tls.Config{}
	err := (&Client{
		FileReader: mockFileReader{
			"ca.pem": caCert,
		},
		RootCA:             "ca.pem",
		InsecureSkipVerify: true,
	}).ApplyTo(tlsConfig)
	a.So(err, should.BeNil)
	a.So(tlsConfig.RootCAs, should.NotBeNil)
	a.So(tlsConfig.InsecureSkipVerify, should.BeTrue)

	t.Run("Empty", func(t *testing.T) {
		a := assertions.New(t)
		tlsConfig := &tls.Config{} //nolint:gosec
		err := (&Client{}).ApplyTo(tlsConfig)
		a.So(err, should.BeNil)
		a.So(tlsConfig.RootCAs, should.BeNil)
		a.So(tlsConfig.InsecureSkipVerify, should.BeFalse)
	})
}

func TestApplyTLSServerAuth(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	cert, key := genCert()
	tlsConfig := &tls.Config{} //nolint:gosec
	err := (&ServerAuth{
		Source: "file",
		FileReader: mockFileReader{
			"cert.pem": cert,
			"key.pem":  key,
		},
		Certificate: "cert.pem",
		Key:         "key.pem",
	}).ApplyTo(tlsConfig)
	a.So(err, should.BeNil)
	a.So(tlsConfig.GetCertificate, should.NotBeNil)
}

func TestApplyTLSClientAuth(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	cert, key := genCert()
	tlsConfig := &tls.Config{} //nolint:gosec
	err := (&ClientAuth{
		Source: "file",
		FileReader: mockFileReader{
			"cert.pem": cert,
			"key.pem":  key,
		},
		Certificate: "cert.pem",
		Key:         "key.pem",
	}).ApplyTo(tlsConfig)
	a.So(err, should.BeNil)
	a.So(tlsConfig.GetClientCertificate, should.NotBeNil)
}
