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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	yaml "gopkg.in/yaml.v2"
)

type tlsConfig struct {
	RootCA      string `yaml:"root-ca"`
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

func (conf tlsConfig) IsZero() bool {
	return conf == (tlsConfig{})
}

type fetcherFileReader struct {
	fetcher fetch.Interface
}

var errFetchFile = errors.Define("fetch_file", "fetch file `{name}`")

func (r fetcherFileReader) ReadFile(name string) ([]byte, error) {
	b, err := r.fetcher.File(name)
	if err != nil {
		return nil, errFetchFile.WithCause(err).WithAttributes("name", name)
	}
	return b, nil
}

func (conf tlsConfig) TLSConfig(fetcher fetch.Interface) (*tls.Config, error) {
	res := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if err := (&tlsconfig.Client{
		FileReader: fetcherFileReader{fetcher: fetcher},
		RootCA:     conf.RootCA,
	}).ApplyTo(res); err != nil {
		return nil, err
	}
	if err := (&tlsconfig.ClientAuth{
		Source:      "file",
		FileReader:  fetcherFileReader{fetcher: fetcher},
		Certificate: conf.Certificate,
		Key:         conf.Key,
	}).ApplyTo(res); err != nil {
		return nil, err
	}
	return res, nil
}

const (
	// InteropClientConfigurationName represents the filename of interop client configuration.
	InteropClientConfigurationName = "config.yml"
	// SenderClientCAsConfigurationName represents the filename of sender client CAs configuration.
	SenderClientCAsConfigurationName = "config.yml"
)

// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6026)
func fetchSenderClientCAs( //nolint:gocyclo
	ctx context.Context, conf config.InteropServer, httpClientProvider httpclient.Provider,
) (map[string][]*x509.Certificate, error) {
	decodeCerts := func(b []byte) (res []*x509.Certificate, err error) {
		for len(b) > 0 {
			var block *pem.Block
			block, b = pem.Decode(b)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
				continue
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			res = append(res, cert)
		}
		return res, nil
	}

	var senderClientCAs map[string][]*x509.Certificate
	if len(conf.SenderClientCADeprecated) > 0 {
		senderClientCAs = make(map[string][]*x509.Certificate, len(conf.SenderClientCA.Static))
		for id, filename := range conf.SenderClientCADeprecated {
			b, err := os.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			certs, err := decodeCerts(b)
			if err != nil {
				return nil, err
			}
			if len(certs) > 0 {
				senderClientCAs[id] = certs
			}
		}
	} else if len(conf.SenderClientCA.Static) > 0 {
		senderClientCAs = make(map[string][]*x509.Certificate, len(conf.SenderClientCA.Static))
		for id, b := range conf.SenderClientCA.Static {
			certs, err := decodeCerts(b)
			if err != nil {
				return nil, err
			}
			if len(certs) > 0 {
				senderClientCAs[id] = certs
			}
		}
	} else {
		fetcher, err := conf.SenderClientCA.Fetcher(ctx, httpClientProvider)
		if err != nil {
			return nil, err
		}
		if fetcher != nil {
			confFileBytes, err := fetcher.File(SenderClientCAsConfigurationName)
			if err != nil {
				return nil, err
			}

			var yamlConf map[string]string
			if err := yaml.UnmarshalStrict(confFileBytes, &yamlConf); err != nil {
				return nil, err
			}

			senderClientCAs = make(map[string][]*x509.Certificate, len(yamlConf))
			for senderID, filename := range yamlConf {
				b, err := fetcher.File(filename)
				if err != nil {
					return nil, err
				}
				certs, err := decodeCerts(b)
				if err != nil {
					return nil, err
				}
				if len(certs) > 0 {
					senderClientCAs[senderID] = certs
				}
			}
		}
	}

	return senderClientCAs, nil
}
