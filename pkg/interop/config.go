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
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
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

func (r fetcherFileReader) ReadFile(name string) ([]byte, error) {
	b, err := r.fetcher.File(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch %q: %w", name, err)
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

// InteropClientConfigurationName represents the filename of interop client configuration.
const InteropClientConfigurationName = "config.yml"
