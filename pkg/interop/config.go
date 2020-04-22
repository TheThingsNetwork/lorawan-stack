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

package interop

import (
	"crypto/tls"
	"crypto/x509"

	"go.thethings.network/lorawan-stack/pkg/fetch"
)

type tlsConfig struct {
	RootCA      string `yaml:"root-ca"`
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

func (conf tlsConfig) IsZero() bool {
	return conf == (tlsConfig{})
}

func (conf tlsConfig) TLSConfig(fetcher fetch.Interface) (*tls.Config, error) {
	var rootCAs *x509.CertPool
	if conf.RootCA != "" {
		caPEM, err := fetcher.File(conf.RootCA)
		if err != nil {
			return nil, err
		}
		rootCAs = x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(caPEM)
	}

	var getCert func(*tls.CertificateRequestInfo) (*tls.Certificate, error)
	if conf.Certificate != "" || conf.Key != "" {
		getCert = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			certPEM, err := fetcher.File(conf.Certificate)
			if err != nil {
				return nil, err
			}
			keyPEM, err := fetcher.File(conf.Key)
			if err != nil {
				return nil, err
			}
			cert, err := tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		}
	}
	return &tls.Config{
		RootCAs:              rootCAs,
		GetClientCertificate: getCert,
	}, nil
}

// InteropClientConfigurationName represents the filename of interop client configuration.
const InteropClientConfigurationName = "config.yml"
