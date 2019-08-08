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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"go.thethings.network/lorawan-stack/pkg/util/test"
)

var (
	RootCAPath = filepath.Join("testdata", "rootCA.pem")
	RootCA     = test.Must(ioutil.ReadFile(RootCAPath)).([]byte)

	ClientCertPath = filepath.Join("testdata", "clientcert.pem")
	ClientCert     = test.Must(ioutil.ReadFile(ClientCertPath)).([]byte)

	ClientKeyPath = filepath.Join("testdata", "clientkey.pem")
	ClientKey     = test.Must(ioutil.ReadFile(ClientKeyPath)).([]byte)

	ServerCertPath = filepath.Join("testdata", "servercert.pem")
	ServerCert     = test.Must(ioutil.ReadFile(ServerCertPath)).([]byte)

	ServerKeyPath = filepath.Join("testdata", "serverkey.pem")
	ServerKey     = test.Must(ioutil.ReadFile(ServerKeyPath)).([]byte)
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
		Certificates: makeClientCertificates(),
		RootCAs:      makeCertPool(),
	}
}

func makeServerTLSConfig() *tls.Config {
	return &tls.Config{
		Certificates: makeServerCertificates(),
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    makeCertPool(),
	}
}

func newTLSServer(hdl http.Handler) *httptest.Server {
	srv := httptest.NewUnstartedServer(hdl)
	srv.TLS = makeServerTLSConfig()
	srv.StartTLS()
	return srv
}
