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

package ttnpb

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func (m *KeyEnvelope) GetKey() *types.AES128Key {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KeyEnvelope) IsZero() bool {
	return m == nil || m.Key.IsZero() && m.KEKLabel == "" && len(m.EncryptedKey) == 0
}

// FieldIsZero returns whether path p is zero.
func (v *KeyEnvelope) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "encrypted_key":
		return v.EncryptedKey == nil
	case "kek_label":
		return v.KEKLabel == ""
	case "key":
		return v.Key == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *SessionKeys) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "app_s_key":
		return v.AppSKey == nil
	case "app_s_key.encrypted_key":
		return v.AppSKey.FieldIsZero("encrypted_key")
	case "app_s_key.kek_label":
		return v.AppSKey.FieldIsZero("kek_label")
	case "app_s_key.key":
		return v.AppSKey.FieldIsZero("key")
	case "f_nwk_s_int_key":
		return v.FNwkSIntKey == nil
	case "f_nwk_s_int_key.encrypted_key":
		return v.FNwkSIntKey.FieldIsZero("encrypted_key")
	case "f_nwk_s_int_key.kek_label":
		return v.FNwkSIntKey.FieldIsZero("kek_label")
	case "f_nwk_s_int_key.key":
		return v.FNwkSIntKey.FieldIsZero("key")
	case "nwk_s_enc_key":
		return v.NwkSEncKey == nil
	case "nwk_s_enc_key.encrypted_key":
		return v.NwkSEncKey.FieldIsZero("encrypted_key")
	case "nwk_s_enc_key.kek_label":
		return v.NwkSEncKey.FieldIsZero("kek_label")
	case "nwk_s_enc_key.key":
		return v.NwkSEncKey.FieldIsZero("key")
	case "s_nwk_s_int_key":
		return v.SNwkSIntKey == nil
	case "s_nwk_s_int_key.encrypted_key":
		return v.SNwkSIntKey.FieldIsZero("encrypted_key")
	case "s_nwk_s_int_key.kek_label":
		return v.SNwkSIntKey.FieldIsZero("kek_label")
	case "s_nwk_s_int_key.key":
		return v.SNwkSIntKey.FieldIsZero("key")
	case "session_key_id":
		return v.SessionKeyID == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *RootKeys) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "app_key":
		return v.AppKey == nil
	case "app_key.encrypted_key":
		return v.AppKey.FieldIsZero("encrypted_key")
	case "app_key.kek_label":
		return v.AppKey.FieldIsZero("kek_label")
	case "app_key.key":
		return v.AppKey.FieldIsZero("key")
	case "nwk_key":
		return v.NwkKey == nil
	case "nwk_key.encrypted_key":
		return v.NwkKey.FieldIsZero("encrypted_key")
	case "nwk_key.kek_label":
		return v.NwkKey.FieldIsZero("kek_label")
	case "nwk_key.key":
		return v.NwkKey.FieldIsZero("key")
	case "root_key_id":
		return v.RootKeyID == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}
