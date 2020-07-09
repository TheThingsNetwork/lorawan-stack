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
	"io/ioutil"

	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/universal"
)

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
	if err = ioutil.WriteFile("ca.pem", caCert, 0644); err != nil {
		return err
	}
	if err = ioutil.WriteFile("ca-key.pem", caKey, 0644); err != nil {
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
	if err = ioutil.WriteFile("cert.pem", cert, 0644); err != nil {
		return err
	}
	if err = ioutil.WriteFile("key.pem", key, 0644); err != nil {
		return err
	}
	return nil
}
