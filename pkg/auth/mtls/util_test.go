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

package mtls_test

import (
	"os"
	"path/filepath"

	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/universal"
)

// MockCA is a certificate authority for tests.
type MockCA struct {
	RootDir string
	CAs     []string
}

var caName = []csr.Name{
	{C: "NL", ST: "Noord-Holland", L: "Amsterdam", O: "The Things Gateway Demo"},
}

// New generates certificate authorities and writes certificates and keys to files.
func (m *MockCA) New() error {
	caReq := csr.CertificateRequest{
		Names:      caName,
		KeyRequest: csr.NewKeyRequest(),
	}
	for _, ca := range m.CAs {
		caCert, _, caKey, err := initca.New(&caReq)
		if err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Join(m.RootDir, ca), 0o755); err != nil {
			return err
		}
		if err = os.WriteFile( // nolint:gosec
			filepath.Join(m.RootDir, ca, "ca.pem"),
			caCert,
			0o644,
		); err != nil {
			return err
		}
		if err = os.WriteFile( // nolint:gosec
			filepath.Join(m.RootDir, ca, "ca-key.pem"),
			caKey,
			0o644,
		); err != nil {
			return err
		}
	}
	return nil
}

// GenerateCertificate generates a client certificate for the EUI with the given CA.
func (m *MockCA) GenerateCertificate(euiString string, ca string) ([]byte, error) {
	certReq := csr.CertificateRequest{
		CN:         euiString,
		Names:      caName,
		KeyRequest: csr.NewKeyRequest(),
	}
	g := &csr.Generator{Validator: func(req *csr.CertificateRequest) error { return nil }}
	req, _, err := g.ProcessRequest(&certReq)
	if err != nil {
		return nil, err
	}
	signReq := signer.SignRequest{
		Request: string(req),
		Hosts:   certReq.Hosts,
	}

	s, err := universal.NewSigner(universal.Root{
		Config: map[string]string{
			"cert-file": filepath.Join(m.RootDir, ca, "ca.pem"),
			"key-file":  filepath.Join(m.RootDir, ca, "ca-key.pem"),
		},
	}, &config.Signing{
		Profiles: map[string]*config.SigningProfile{},
		Default:  config.DefaultConfig(),
	})
	if err != nil {
		return nil, err
	}
	return s.Sign(signReq)
}

// Clean removes the generated certificates.
func (m *MockCA) Clean() error {
	for _, ca := range m.CAs {
		if err := os.RemoveAll(filepath.Join(m.RootDir, ca)); err != nil {
			return err
		}
	}
	return nil
}
