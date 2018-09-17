// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import errors "go.thethings.network/lorawan-stack/pkg/errorsv3"

// KeyVault provides wrapping and unwrapping keys using KEK labels.
type KeyVault interface {
	Wrap(plaintext []byte, kekLabel string) ([]byte, error)
	Unwrap(ciphertext []byte, kekLabel string) ([]byte, error)
}

type memKeyVault map[string][]byte

// NewMemKeyVault returns a KeyVault that uses KEKs from memory. This implementation does not provide any security as
// KEKs are stored in the clear.
func NewMemKeyVault(m map[string][]byte) KeyVault {
	return memKeyVault(m)
}

var errKEKNotFound = errors.DefineNotFound("kek_not_found", "KEK with label `{label}` not found")

func (v memKeyVault) Wrap(plaintext []byte, kekLabel string) ([]byte, error) {
	kek, ok := v[kekLabel]
	if !ok {
		return nil, errKEKNotFound.WithAttributes("label", kekLabel)
	}
	return WrapKey(plaintext, kek)
}

func (v memKeyVault) Unwrap(ciphertext []byte, kekLabel string) ([]byte, error) {
	kek, ok := v[kekLabel]
	if !ok {
		return nil, errKEKNotFound.WithAttributes("label", kekLabel)
	}
	return UnwrapKey(ciphertext, kek)
}
