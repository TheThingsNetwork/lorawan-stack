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

package crypto

import (
	"context"
	"crypto/tls"
)

// KeyVault provides wrapping and unwrapping keys using KEK labels.
type KeyVault interface {
	ComponentKEKLabeler

	// Wrap implements the RFC 3394 AES Key Wrap algorithm. Only keys of 16, 24 or 32 bytes are accepted.
	// Keys are referenced using the KEK labels.
	Wrap(ctx context.Context, plaintext []byte, kekLabel string) ([]byte, error)
	// UnwrapKey implements the RFC 3394 AES Key Unwrap algorithm. Only keys of 16, 24 or 32 bytes are accepted.
	// Keys are referenced using the KEK labels.
	Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error)

	// Encrypt encrypts messages of variable length using AES 128 GCM.
	// The encryption key is referenced using the ID.
	Encrypt(ctx context.Context, plaintext []byte, id string) ([]byte, error)
	// Decrypt decrypts messages of variable length using AES 128 GCM.
	// The encryption key is referenced using the ID.
	Decrypt(ctx context.Context, ciphertext []byte, id string) ([]byte, error)

	// ExportCertificate exports the X.509 certificate and private key of the given identifier.
	ExportCertificate(ctx context.Context, id string) (*tls.Certificate, error)

	// HMACHash calculates the Keyed-Hash Message Authentication Code (HMAC, RFC 2104) hash of the data.
	// The AES key used for hashing is referenced using the ID.
	HMACHash(ctx context.Context, payload []byte, id string) ([]byte, error)
}
