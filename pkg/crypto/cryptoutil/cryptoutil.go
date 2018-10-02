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

package cryptoutil

import (
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// WrapAES128Key performs the RFC 3394 Wrap algorithm on the given key using the given key vault and KEK label.
// If the KEK label is empty, the key will be returned in the clear.
func WrapAES128Key(key types.AES128Key, kekLabel string, v crypto.KeyVault) (ttnpb.KeyEnvelope, error) {
	if kekLabel == "" {
		return ttnpb.KeyEnvelope{Key: key[:]}, nil
	}
	wrapped, err := v.Wrap(key[:], kekLabel)
	if err != nil {
		return ttnpb.KeyEnvelope{}, err
	}
	return ttnpb.KeyEnvelope{
		Key:      wrapped,
		KEKLabel: kekLabel,
	}, nil
}

var errInvalidLength = errors.DefineInvalidArgument("invalid_length", "invalid slice length")

// UnwrapAES128Key performs the RFC 3394 Unwrap algorithm on the given key envelope using the given key vault.
// If the KEK label is empty, the key is assumed to be stored in the clear.
func UnwrapAES128Key(wrapped ttnpb.KeyEnvelope, v crypto.KeyVault) (types.AES128Key, error) {
	var key []byte
	if wrapped.KEKLabel == "" {
		key = wrapped.Key
	} else {
		var err error
		key, err = v.Unwrap(wrapped.Key, wrapped.KEKLabel)
		if err != nil {
			return types.AES128Key{}, err
		}
	}
	if len(key) != 16 {
		return types.AES128Key{}, errInvalidLength
	}
	unwrapped := types.AES128Key{}
	copy(unwrapped[:], key)
	return unwrapped, nil
}
