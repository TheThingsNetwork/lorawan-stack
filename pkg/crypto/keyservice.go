// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package crypto

import (
	"context"
	"crypto/tls"

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// KeyService provides common cryptographic operations.
type KeyService interface {
	// Wrap implements the RFC 3394 AES Key Wrap algorithm. Only keys of 16, 24 or 32 bytes are accepted.
	// Keys are referenced using the KEK labels.
	Wrap(ctx context.Context, plaintext []byte, kekLabel string) ([]byte, error)
	// Unwrap implements the RFC 3394 AES Key Unwrap algorithm. Only keys of 16, 24 or 32 bytes are accepted.
	// Keys are referenced using the KEK labels.
	Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error)

	// Encrypt encrypts messages of variable length using AES 128 GCM.
	// The encryption key is referenced using the label.
	Encrypt(ctx context.Context, plaintext []byte, label string) ([]byte, error)
	// Decrypt decrypts messages of variable length using AES 128 GCM.
	// The encryption key is referenced using the label.
	Decrypt(ctx context.Context, ciphertext []byte, label string) ([]byte, error)

	// ServerCertificate returns the X.509 certificate and private key of the given label.
	ServerCertificate(ctx context.Context, label string) (tls.Certificate, error)

	// ClientCertificate returns the X.509 client certificate and private key.
	ClientCertificate(ctx context.Context) (tls.Certificate, error)

	// HMACHash calculates the Keyed-Hash Message Authentication Code (HMAC, RFC 2104) hash of the data.
	// The AES key used for hashing is referenced using the label.
	HMACHash(ctx context.Context, payload []byte, label string) ([]byte, error)
}

type keyService struct {
	vault KeyVault
}

// NewKeyService returns a new KeyService.
func NewKeyService(vault KeyVault) KeyService {
	return &keyService{
		vault: vault,
	}
}

func (ks *keyService) aes128Key(ctx context.Context, label string) (types.AES128Key, error) {
	key, err := ks.vault.Key(ctx, label)
	if err != nil {
		return types.AES128Key{}, err
	}
	var res types.AES128Key
	if err := res.Unmarshal(key); err != nil {
		return types.AES128Key{}, err
	}
	return res, nil
}

func (ks *keyService) Decrypt(ctx context.Context, ciphertext []byte, label string) ([]byte, error) {
	key, err := ks.aes128Key(ctx, label)
	if err != nil {
		return nil, err
	}
	return Decrypt(key, ciphertext)
}

func (ks *keyService) Encrypt(ctx context.Context, plaintext []byte, label string) ([]byte, error) {
	key, err := ks.aes128Key(ctx, label)
	if err != nil {
		return nil, err
	}
	return Encrypt(key, plaintext)
}

func (ks *keyService) ServerCertificate(ctx context.Context, label string) (tls.Certificate, error) {
	return ks.vault.ServerCertificate(ctx, label)
}

func (ks *keyService) ClientCertificate(ctx context.Context) (tls.Certificate, error) {
	return ks.vault.ClientCertificate(ctx)
}

func (ks *keyService) HMACHash(ctx context.Context, payload []byte, label string) ([]byte, error) {
	key, err := ks.aes128Key(ctx, label)
	if err != nil {
		return nil, err
	}
	return HMACHash(key, payload)
}

func (ks *keyService) Unwrap(ctx context.Context, ciphertext []byte, label string) ([]byte, error) {
	key, err := ks.vault.Key(ctx, label)
	if err != nil {
		return nil, err
	}
	return UnwrapKey(ciphertext, key)
}

func (ks *keyService) Wrap(ctx context.Context, plaintext []byte, label string) ([]byte, error) {
	key, err := ks.vault.Key(ctx, label)
	if err != nil {
		return nil, err
	}
	return WrapKey(plaintext, key)
}
