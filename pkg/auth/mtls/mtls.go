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

// Package mtls provides functions to authenticate client TLS certificates.
package mtls

import (
	"context"
	"crypto/x509"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"gopkg.in/yaml.v2"
)

// CAStore is a store of CA Certs.
type CAStore struct {
	commonPool *x509.CertPool
}

var (
	errParseIndexFile   = errors.DefineCorruption("parse_index_file", "parse index file")
	errFetchFile        = errors.Define("fetch_file", "fetch file `{path}`")
	errReadCertFromFile = errors.DefineInvalidArgument(
		"read_certificate_from_file",
		"read certificate from file `{path}`",
	)
	errNoCAPool               = errors.DefineInvalidArgument("no_ca_pool", "no CA pool configured")
	errCertificateNotVerified = errors.DefineInvalidArgument(
		"certificate_not_verified",
		"certificate not verified",
	)
	errCommonNameMismatch = errors.DefineInvalidArgument(
		"common_name_mismatch",
		"common name mismatch. Expected `{exp}`, got `{got}`",
	)
)

// NewCAStore creates a new CAStore.
// If the fetcher is given, the index file must be present. If the index file contains a common pool, it will be loaded.
func NewCAStore(_ context.Context, fetcher fetch.Interface) (*CAStore, error) {
	const commonCertPoolKey = "common"
	s := &CAStore{
		commonPool: x509.NewCertPool(),
	}
	if fetcher != nil {
		raw, err := fetcher.File("index.yml")
		if err != nil {
			return nil, errFetchFile.WithAttributes("path", "index.yml").WithCause(err)
		}
		var index struct {
			Common []string `yaml:"common"`
		}
		err = yaml.Unmarshal(raw, &index)
		if err != nil {
			return nil, errParseIndexFile.WithCause(err)
		}
		for _, fileName := range index.Common {
			pathElements := []string{commonCertPoolKey, fileName}
			raw, err := fetcher.File(pathElements...)
			if err != nil {
				return nil, errFetchFile.WithAttributes("path", strings.Join(pathElements, "/")).WithCause(err)
			}
			if ok := s.commonPool.AppendCertsFromPEM(raw); !ok {
				return nil, errReadCertFromFile.WithAttributes("path", strings.Join(pathElements, "/"))
			}
		}
	}
	return s, nil
}

func (c CAStore) certPools(_ context.Context) []*x509.CertPool {
	var res []*x509.CertPool
	if c.commonPool != nil {
		res = append(res, c.commonPool)
	}
	return res
}

// Verify verifies the certificate against the list of configured certificate pools.
// The function also checks that common name in the certificate matches the provided value.
func (c *CAStore) Verify(ctx context.Context, cn string, cert *x509.Certificate) error {
	if cert.Subject.CommonName != cn {
		return errCommonNameMismatch.WithAttributes(
			"exp", cn,
			"got", cert.Subject.CommonName,
		)
	}

	certPools := c.certPools(ctx)
	if len(certPools) == 0 {
		return errNoCAPool.New()
	}
	opts := x509.VerifyOptions{
		Roots: certPool,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}
	if _, err := cert.Verify(opts); err == nil {
		return nil
	}
	return errCertificateNotVerified.New()
}
