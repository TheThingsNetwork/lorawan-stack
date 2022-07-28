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
	"fmt"
	"runtime/trace"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	errKEKNotFound         = errors.DefineNotFound("kek_not_found", "KEK with label `{label}` not found")
	errKeyNotFound         = errors.DefineNotFound("key_not_found", "key with ID `{id}` not found")
	errCertificateNotFound = errors.DefineNotFound("certificate_not_found", "certificate with ID `{id}` not found")
)

// WrapAES128Key performs the RFC 3394 Wrap algorithm on the given key using the given key vault and KEK label.
// If the KEK label is empty, the key will be returned in the clear.
func WrapAES128Key(ctx context.Context, key types.AES128Key, kekLabel string, v crypto.KeyVault) (*ttnpb.KeyEnvelope, error) {
	defer trace.StartRegion(ctx, "wrap AES-128 key").End()
	if kekLabel == "" {
		return &ttnpb.KeyEnvelope{
			EncryptedKey: key[:],
		}, nil
	}
	wrapped, err := v.Wrap(ctx, key[:], kekLabel)
	if err != nil {
		return nil, err
	}
	return &ttnpb.KeyEnvelope{
		EncryptedKey: wrapped,
		KekLabel:     kekLabel,
	}, nil
}

// WrapAES128KeyWithKEK is like WrapAES128Key, but takes a KEK instead of key vault.
func WrapAES128KeyWithKEK(ctx context.Context, key types.AES128Key, kekLabel string, kek types.AES128Key) (*ttnpb.KeyEnvelope, error) {
	defer trace.StartRegion(ctx, "wrap AES-128 key").End()
	if kekLabel == "" {
		return &ttnpb.KeyEnvelope{
			EncryptedKey: key[:],
		}, nil
	}
	wrapped, err := crypto.WrapKey(key[:], kek[:])
	if err != nil {
		return nil, err
	}
	return &ttnpb.KeyEnvelope{
		EncryptedKey: wrapped,
		KekLabel:     kekLabel,
	}, nil
}

var errInvalidLength = errors.DefineInvalidArgument("invalid_length", "invalid slice length")

// UnwrapAES128Key performs the RFC 3394 Unwrap algorithm on the given key envelope using the given key vault.
// If the KEK label is empty, the key is assumed to be stored in the clear.
func UnwrapAES128Key(ctx context.Context, wrapped *ttnpb.KeyEnvelope, v crypto.KeyVault) (key types.AES128Key, err error) {
	defer trace.StartRegion(ctx, "unwrap AES-128 key").End()
	if wrapped.Key != nil {
		return *types.MustAES128Key(wrapped.Key), nil
	}
	if wrapped.KekLabel == "" {
		if len(wrapped.EncryptedKey) != 16 {
			return key, errInvalidLength.New()
		}
		copy(key[:], wrapped.EncryptedKey)
	} else {
		keyBytes, err := v.Unwrap(ctx, wrapped.EncryptedKey, wrapped.KekLabel)
		if err != nil {
			return key, err
		}
		if len(keyBytes) != 16 {
			return key, errInvalidLength.New()
		}
		copy(key[:], keyBytes)
	}
	return key, nil
}

// UnwrapKeyEnvelope calls UnwrapAES128Key on the given key envelope using the given key vault if necessary and
// returns the result as a key envelope.
// NOTE: UnwrapKeyEnvelope returns ke if unwrapping is not necessary.
func UnwrapKeyEnvelope(ctx context.Context, ke *ttnpb.KeyEnvelope, v crypto.KeyVault) (*ttnpb.KeyEnvelope, error) {
	if !types.MustAES128Key(ke.GetKey()).OrZero().IsZero() || len(ke.GetEncryptedKey()) == 0 {
		return ke, nil
	}
	k, err := UnwrapAES128Key(ctx, ke, v)
	if err != nil {
		return nil, err
	}
	return &ttnpb.KeyEnvelope{
		Key: k.Bytes(),
	}, nil
}

func pathWithPrefix(prefix, path string) string {
	if prefix == "" {
		return path
	}
	return fmt.Sprintf("%s.%s", prefix, path)
}

func UnwrapSelectedSessionKeys(ctx context.Context, keyVault crypto.KeyVault, sk *ttnpb.SessionKeys, prefix string, paths ...string) (*ttnpb.SessionKeys, error) {
	var (
		fNwkSIntKeyEnvelope *ttnpb.KeyEnvelope
		sNwkSIntKeyEnvelope *ttnpb.KeyEnvelope
		nwkSEncKeyEnvelope  *ttnpb.KeyEnvelope
		appSKeyEnvelope     *ttnpb.KeyEnvelope

		err error
	)
	if ttnpb.HasAnyField(paths, pathWithPrefix(prefix, "app_s_key.key")) && sk.GetAppSKey() != nil {
		appSKeyEnvelope, err = UnwrapKeyEnvelope(ctx, sk.AppSKey, keyVault)
		if err != nil {
			return nil, err
		}
	}
	if ttnpb.HasAnyField(paths, pathWithPrefix(prefix, "f_nwk_s_int_key.key")) && sk.GetFNwkSIntKey() != nil {
		fNwkSIntKeyEnvelope, err = UnwrapKeyEnvelope(ctx, sk.FNwkSIntKey, keyVault)
		if err != nil {
			return nil, err
		}
	}
	if ttnpb.HasAnyField(paths, pathWithPrefix(prefix, "nwk_s_enc_key.key")) && sk.GetNwkSEncKey() != nil {
		nwkSEncKeyEnvelope, err = UnwrapKeyEnvelope(ctx, sk.NwkSEncKey, keyVault)
		if err != nil {
			return nil, err
		}
	}
	if ttnpb.HasAnyField(paths, pathWithPrefix(prefix, "s_nwk_s_int_key.key")) && sk.GetSNwkSIntKey() != nil {
		sNwkSIntKeyEnvelope, err = UnwrapKeyEnvelope(ctx, sk.SNwkSIntKey, keyVault)
		if err != nil {
			return nil, err
		}
	}
	return &ttnpb.SessionKeys{
		SessionKeyId: sk.GetSessionKeyId(),
		FNwkSIntKey:  fNwkSIntKeyEnvelope,
		NwkSEncKey:   nwkSEncKeyEnvelope,
		SNwkSIntKey:  sNwkSIntKeyEnvelope,
		AppSKey:      appSKeyEnvelope,
	}, nil
}
