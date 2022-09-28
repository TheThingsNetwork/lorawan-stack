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

package ttnmage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/universal"
)

var gatewayCertsDir = "gateway-certs"

// Certificates generates certificates for development.
func (Dev) Certificates() error {
	caReq := csr.CertificateRequest{
		Names: []csr.Name{
			{C: "NL", ST: "Noord-Holland", L: "Amsterdam", O: "The Things Demo"},
		},
		KeyRequest: csr.NewKeyRequest(),
	}
	caCert, _, caKey, err := initca.New(&caReq)
	if err != nil {
		return err
	}
	if err = os.WriteFile("ca.pem", caCert, 0o644); err != nil {
		return err
	}
	if err = os.WriteFile("ca-key.pem", caKey, 0o644); err != nil {
		return err
	}
	certReq := csr.CertificateRequest{
		Hosts: []string{
			"localhost",
			"*.localhost",
		},
		Names:      caReq.Names, // Use the same names.
		KeyRequest: csr.NewKeyRequest(),
	}
	g := &csr.Generator{Validator: func(req *csr.CertificateRequest) error { return nil }}
	csr, key, err := g.ProcessRequest(&certReq)
	if err != nil {
		return err
	}
	signReq := signer.SignRequest{
		Request: string(csr),
		Hosts:   certReq.Hosts,
	}
	s, err := universal.NewSigner(universal.Root{
		Config: map[string]string{
			"cert-file": "ca.pem",
			"key-file":  "ca-key.pem",
		},
	}, &config.Signing{
		Profiles: map[string]*config.SigningProfile{},
		Default:  config.DefaultConfig(),
	})
	if err != nil {
		return err
	}
	cert, err := s.Sign(signReq)
	if err != nil {
		return err
	}
	if err = os.WriteFile("cert.pem", cert, 0o644); err != nil {
		return err
	}
	if err = os.WriteFile("key.pem", key, 0o644); err != nil {
		return err
	}
	return nil
}

var names = []csr.Name{
	{C: "NL", ST: "Noord-Holland", L: "Amsterdam", O: "The Things Gateway Demo"},
}

// GenGatewayCA generates a certificate authority to sign gateway certificates.
func (Dev) GenGatewayCA() error {
	caReq := csr.CertificateRequest{
		Names:      names,
		KeyRequest: csr.NewKeyRequest(),
	}
	caCert, _, caKey, err := initca.New(&caReq)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Join(devDir, gatewayCertsDir), 0o755); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(devDir, gatewayCertsDir, "ca.pem"), caCert, 0o644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(devDir, gatewayCertsDir, "ca-key.pem"), caKey, 0o644); err != nil {
		return err
	}
	return nil
}

// GenGatewayCerts generates a client certificates for the EUI.
func (Dev) GenGatewayCerts() error {
	eui := os.Getenv("GATEWAY_EUI")
	certReq := csr.CertificateRequest{
		CN:         eui,
		Names:      names, // Use the same names.
		KeyRequest: csr.NewKeyRequest(),
	}
	g := &csr.Generator{Validator: func(req *csr.CertificateRequest) error { return nil }}
	csr, key, err := g.ProcessRequest(&certReq)
	if err != nil {
		return err
	}
	signReq := signer.SignRequest{
		Request: string(csr),
		Hosts:   certReq.Hosts,
	}
	s, err := universal.NewSigner(universal.Root{
		Config: map[string]string{
			"cert-file": filepath.Join(devDir, gatewayCertsDir, "ca.pem"),
			"key-file":  filepath.Join(devDir, gatewayCertsDir, "ca-key.pem"),
		},
	}, &config.Signing{
		Profiles: map[string]*config.SigningProfile{},
		Default:  config.DefaultConfig(),
	})
	if err != nil {
		return err
	}
	cert, err := s.Sign(signReq)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(devDir, gatewayCertsDir, fmt.Sprintf("%s.crt", eui)), cert, 0o644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(devDir, gatewayCertsDir, fmt.Sprintf("%s.key", eui)), key, 0o644); err != nil {
		return err
	}
	return nil
}
