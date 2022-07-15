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

package interop_test

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbroker"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	RootCAPath = filepath.Join("testdata", "rootCA.pem")
	RootCA     = test.Must(os.ReadFile(RootCAPath)).([]byte)

	ClientCertPath = filepath.Join("testdata", "clientcert.pem")
	ClientCert     = test.Must(os.ReadFile(ClientCertPath)).([]byte)

	ClientKeyPath = filepath.Join("testdata", "clientkey.pem")
	ClientKey     = test.Must(os.ReadFile(ClientKeyPath)).([]byte)

	ServerCertPath = filepath.Join("testdata", "servercert.pem")
	ServerCert     = test.Must(os.ReadFile(ServerCertPath)).([]byte)

	ServerKeyPath = filepath.Join("testdata", "serverkey.pem")
	ServerKey     = test.Must(os.ReadFile(ServerKeyPath)).([]byte)
)

func makeCertPool() *x509.CertPool {
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(RootCA)
	return certpool
}

func makeClientCertificates() []tls.Certificate {
	return []tls.Certificate{
		test.Must(tls.X509KeyPair(
			ClientCert,
			ClientKey,
		)).(tls.Certificate),
	}
}

func makeServerCertificates() []tls.Certificate {
	return []tls.Certificate{
		test.Must(tls.X509KeyPair(
			ServerCert,
			ServerKey,
		)).(tls.Certificate),
	}
}

func makeClientTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: makeClientCertificates(),
		RootCAs:      makeCertPool(),
	}
}

func makeServerTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: makeServerCertificates(),
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    makeCertPool(),
	}
}

func newTLSServer(port int, hdl http.Handler) *httptest.Server {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	srv := httptest.NewUnstartedServer(hdl)
	srv.Listener = lis
	srv.TLS = makeServerTLSConfig()
	srv.StartTLS()
	return srv
}

func makePacketBrokerTokenIssuer(ctx context.Context, subject string) (iss string, tok func(aud string) string) {
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	var (
		publicJWK = jose.JSONWebKey{
			Algorithm: string(jose.EdDSA),
			Key:       public,
			KeyID:     "test",
		}
		privateJWK = jose.JSONWebKey{
			Algorithm: string(jose.EdDSA),
			Key:       private,
			KeyID:     "test",
		}
		sig = test.Must(jose.NewSigner(jose.SigningKey{
			Algorithm: jose.SignatureAlgorithm(privateJWK.Algorithm),
			Key:       privateJWK,
		}, new(jose.SignerOptions).WithType("JWT"))).(jose.Signer)
	)

	router := mux.NewRouter()
	router.Handle("/.well-known/jwks.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&jose.JSONWebKeySet{ //nolint:errcheck
			Keys: []jose.JSONWebKey{publicJWK},
		})
	}))
	srv := httptest.NewServer(router)

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return srv.URL, func(aud string) string {
		claims := packetbroker.TokenClaims{
			Claims: jwt.Claims{
				Issuer:   srv.URL,
				Subject:  subject,
				Audience: jwt.Audience{aud},
			},
			PacketBroker: packetbroker.IAMTokenClaims{
				Cluster: true,
			},
		}
		return test.Must(jwt.Signed(sig).Claims(claims).CompactSerialize()).(string)
	}
}
