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

package cryptoutil

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// MemKeyVault is a KeyVault that uses secrets from memory.
// This implementation does not provide any security as secrets are stored in the clear.
type MemKeyVault struct {
	ComponentPrefixKEKLabeler
	m map[string][]byte
}

// NewMemKeyVault returns a MemKeyVault.
// For certificates,
func NewMemKeyVault(m map[string][]byte) *MemKeyVault {
	return &MemKeyVault{
		m: m,
	}
}

var errKEKNotFound = errors.DefineNotFound("kek_not_found", "KEK with label `{label}` not found")

// Wrap implements KeyVault.
func (v MemKeyVault) Wrap(plaintext []byte, kekLabel string) ([]byte, error) {
	kek, ok := v.m[kekLabel]
	if !ok {
		return nil, errKEKNotFound.WithAttributes("label", kekLabel)
	}
	return crypto.WrapKey(plaintext, kek)
}

// Unwrap implements KeyVault.
func (v MemKeyVault) Unwrap(ciphertext []byte, kekLabel string) ([]byte, error) {
	kek, ok := v.m[kekLabel]
	if !ok {
		return nil, errKEKNotFound.WithAttributes("label", kekLabel)
	}
	return crypto.UnwrapKey(ciphertext, kek)
}

var errCertificateNotFound = errors.DefineNotFound("certificate_not_found", "certificate with ID `{id}` not found")

// LoadCertificate implements KeyVault.
func (v MemKeyVault) LoadCertificate(id string) (*tls.Certificate, error) {
	raw, ok := v.m[id]
	if !ok {
		return nil, errCertificateNotFound.WithAttributes("id", id)
	}
	certPEMBlock, keyPEMBlock := &bytes.Buffer{}, &bytes.Buffer{}
	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			if err := pem.Encode(certPEMBlock, block); err != nil {
				return nil, err
			}
		} else if block.Type == "PRIVATE KEY" || strings.HasSuffix(block.Type, " PRIVATE KEY") {
			if err := pem.Encode(keyPEMBlock, block); err != nil {
				return nil, err
			}
		}
		raw = rest
	}
	res, err := tls.X509KeyPair(certPEMBlock.Bytes(), keyPEMBlock.Bytes())
	if err != nil {
		return nil, err
	}
	return &res, nil
}
