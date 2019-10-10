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
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	errKEKNotFound         = errors.DefineNotFound("kek_not_found", "KEK with label `{label}` not found")
	errCertificateNotFound = errors.DefineNotFound("certificate_not_found", "certificate with ID `{id}` not found")
)

// WrapAES128Key performs the RFC 3394 Wrap algorithm on the given key using the given key vault and KEK label.
// If the KEK label is empty, the key will be returned in the clear.
func WrapAES128Key(ctx context.Context, key types.AES128Key, kekLabel string, v crypto.KeyVault) (ttnpb.KeyEnvelope, error) {
	if kekLabel == "" {
		return ttnpb.KeyEnvelope{
			EncryptedKey: key[:],
		}, nil
	}
	wrapped, err := v.Wrap(ctx, key[:], kekLabel)
	if err != nil {
		return ttnpb.KeyEnvelope{}, err
	}
	return ttnpb.KeyEnvelope{
		EncryptedKey: wrapped,
		KEKLabel:     kekLabel,
	}, nil
}

var errInvalidLength = errors.DefineInvalidArgument("invalid_length", "invalid slice length")

// UnwrapAES128Key performs the RFC 3394 Unwrap algorithm on the given key envelope using the given key vault.
// If the KEK label is empty, the key is assumed to be stored in the clear.
func UnwrapAES128Key(ctx context.Context, wrapped ttnpb.KeyEnvelope, v crypto.KeyVault) (key types.AES128Key, err error) {
	if wrapped.Key != nil {
		return *wrapped.Key, nil
	}
	if wrapped.KEKLabel == "" {
		if len(wrapped.EncryptedKey) != 16 {
			return key, errInvalidLength
		}
		copy(key[:], wrapped.EncryptedKey)
	} else {
		keyBytes, err := v.Unwrap(ctx, wrapped.EncryptedKey, wrapped.KEKLabel)
		if err != nil {
			return key, err
		}
		if len(keyBytes) != 16 {
			return key, errInvalidLength
		}
		copy(key[:], keyBytes)
	}
	return key, nil
}
