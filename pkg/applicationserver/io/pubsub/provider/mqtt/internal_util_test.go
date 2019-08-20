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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/server"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

var (
	timeout = (1 << 5) * test.Delay
)

func startMQTTServer(ctx context.Context, tlsConfig *tls.Config) (mqttnet.Listener, mqttnet.Listener, error) {
	logger := log.FromContext(ctx)
	s := server.New(ctx)

	lis, err := mqttnet.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, err
	}
	logger.Infof("Listening on %v", lis.Addr())
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				logger.WithError(err).Error("Could not accept connection")
				return
			}
			go s.Handle(conn)
		}
	}()

	if tlsConfig != nil {
		tlsTCPLis, err := tls.Listen("tcp", ":0", tlsConfig)
		if err != nil {
			lis.Close()
			return nil, nil, err
		}
		tlsLis := mqttnet.NewListener(tlsTCPLis, "tls")
		logger.Infof("Listening on TLS %v", tlsLis.Addr())
		go func() {
			for {
				conn, err := tlsLis.Accept()
				if err != nil {
					logger.WithError(err).Error("Could not accept connection")
					return
				}
				go s.Handle(conn)
			}
		}()
		return lis, tlsLis, nil
	}

	return lis, nil, nil
}

func createPKI() (ca []byte, clientCert []byte, clientKey []byte, err error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
			CommonName:   "Test-CA",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("::1"), net.ParseIP("::")},
		IsCA:        true,
	}

	privCA, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	caDERBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privCA.PublicKey, privCA)
	if err != nil {
		return
	}
	caCert, err := x509.ParseCertificate(caDERBytes)
	if err != nil {
		return
	}
	var caBuf bytes.Buffer
	err = pem.Encode(&caBuf, &pem.Block{Type: "CERTIFICATE", Bytes: caDERBytes})
	if err != nil {
		return
	}
	ca = caBuf.Bytes()

	template.Subject.CommonName = "Test-Client"
	template.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}
	privClient, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	clientDERBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privClient.PublicKey, privCA)
	if err != nil {
		return
	}
	var clientCertBuf bytes.Buffer
	err = pem.Encode(&clientCertBuf, &pem.Block{Type: "CERTIFICATE", Bytes: clientDERBytes})
	if err != nil {
		return
	}
	clientCert = clientCertBuf.Bytes()
	var clientKeyBuf bytes.Buffer
	err = pem.Encode(&clientKeyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privClient)})
	if err != nil {
		return
	}
	clientKey = clientKeyBuf.Bytes()
	return
}
