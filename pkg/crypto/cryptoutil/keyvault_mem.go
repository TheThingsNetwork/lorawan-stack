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
	"context"
	"crypto/tls"
	"encoding/pem"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// MemKeyVault is a KeyVault that uses secrets from memory.
// This implementation does not provide any security as secrets are stored in the clear.
type MemKeyVault struct {
	ComponentPrefixKEKLabeler
	m map[string][]byte
}

// NewMemKeyVault returns a MemKeyVault.
// Certificates keys can be appended as PEM block.
func NewMemKeyVault(m map[string][]byte) *MemKeyVault {
	return &MemKeyVault{
		m: m,
	}
}

// Wrap implements KeyVault.
func (v MemKeyVault) Wrap(ctx context.Context, plaintext []byte, kekLabel string) ([]byte, error) {
	kek, ok := v.m[kekLabel]
	if !ok {
		return nil, errKEKNotFound.WithAttributes("label", kekLabel)
	}
	return crypto.WrapKey(plaintext, kek)
}

// Unwrap implements KeyVault.
func (v MemKeyVault) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	kek, ok := v.m[kekLabel]
	if !ok {
		return nil, errKEKNotFound.WithAttributes("label", kekLabel)
	}
	return crypto.UnwrapKey(ciphertext, kek)
}

// Encrypt implements KeyVault.
func (v MemKeyVault) Encrypt(ctx context.Context, plaintext []byte, id string) ([]byte, error) {
	rawKey, ok := v.m[id]
	if !ok {
		return nil, errKeyNotFound.WithAttributes("id", id)
	}
	var key types.AES128Key
	if err := key.UnmarshalBinary(rawKey); err != nil {
		return nil, err
	}
	return crypto.Encrypt(key, plaintext)
}

// Decrypt implements KeyVault.
func (v MemKeyVault) Decrypt(ctx context.Context, ciphertext []byte, id string) ([]byte, error) {
	rawKey, ok := v.m[id]
	if !ok {
		return nil, errKeyNotFound.WithAttributes("id", id)
	}
	var key types.AES128Key
	if err := key.UnmarshalBinary(rawKey); err != nil {
		return nil, err
	}
	return crypto.Decrypt(key, ciphertext)
}

// HMACHash implements KeyVault.
func (v MemKeyVault) HMACHash(_ context.Context, payload []byte, id string) ([]byte, error) {
	rawKey, ok := v.m[id]
	if !ok {
		return nil, errKeyNotFound.WithAttributes("id", id)
	}
	var key types.AES128Key
	if err := key.UnmarshalBinary(rawKey); err != nil {
		return nil, err
	}
	return crypto.HMACHash(key, payload)
}

// ExportCertificate implements KeyVault.
func (v MemKeyVault) ExportCertificate(ctx context.Context, id string) (*tls.Certificate, error) {
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
