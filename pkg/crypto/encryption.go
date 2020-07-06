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

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var errMalformedCipherText = errors.DefineInvalidArgument("malformed_cipher_text", "malformed cipher text")

// Encrypt encrypts a plain text message.
// Uses AES128 keys in GCM (Galois/Counter Mode).
// Since GCM uses a nonce, the encrypted message will be different each time the operation is run for the same set of inputs.
// The returned cipher is in the format |nonce(12)|tag(16)|encrypted(plaintextLen)|.
func Encrypt(key types.AES128Key, plaintext []byte) ([]byte, error) {
	cipherBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts an encrypted message.
// Uses AES128 keys in GCM (Galois/Counter Mode).
func Decrypt(key types.AES128Key, encrypted []byte) ([]byte, error) {
	cipherBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, err
	}
	if len(encrypted) < gcm.NonceSize() {
		return nil, errMalformedCipherText
	}
	return gcm.Open(nil,
		encrypted[:gcm.NonceSize()],
		encrypted[gcm.NonceSize():],
		nil,
	)
}
