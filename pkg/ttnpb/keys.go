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
	"encoding/hex"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	types "go.thethings.network/lorawan-stack/v3/pkg/types"
)

func (m *KeyEnvelope) IsZero() bool {
	return m == nil || types.MustAES128Key(m.Key).OrZero().IsZero() && m.KekLabel == "" && len(m.EncryptedKey) == 0
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
		return v.KekLabel == ""
	case "key":
		return v.Key == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// EncodeMsgpack implements msgpack.CustomEncoder interface.
func (v KeyEnvelope) EncodeMsgpack(enc *msgpack.Encoder) error {
	var n uint8
	if v.Key != nil {
		n++
	}
	if v.KekLabel != "" {
		n++
	}
	if len(v.EncryptedKey) > 0 {
		n++
	}
	if err := enc.EncodeMapLen(int(n)); err != nil {
		return err
	}

	if v.Key != nil {
		if err := enc.EncodeString("key"); err != nil {
			return err
		}
		if err := enc.EncodeString(hex.EncodeToString(v.Key)); err != nil {
			return err
		}
	}
	if v.KekLabel != "" {
		if err := enc.EncodeString("kek_label"); err != nil {
			return err
		}
		if err := enc.EncodeString(v.KekLabel); err != nil {
			return err
		}
	}
	if len(v.EncryptedKey) > 0 {
		if err := enc.EncodeString("encrypted_key"); err != nil {
			return err
		}
		if err := enc.EncodeString(hex.EncodeToString(v.EncryptedKey)); err != nil {
			return err
		}
	}
	return nil
}

// DecodeMsgpack implements msgpack.CustomDecoder interface.
func (v *KeyEnvelope) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}
	*v = KeyEnvelope{}
	for i := 0; i < n; i++ {
		s, err := dec.DecodeString()
		if err != nil {
			return err
		}
		switch s {
		case "key":
			s, err := dec.DecodeString()
			if err != nil {
				return err
			}
			fv, err := hex.DecodeString(s)
			if err != nil {
				return err
			}
			v.Key = fv

		case "kek_label":
			fv, err := dec.DecodeString()
			if err != nil {
				return err
			}
			v.KekLabel = fv

		case "encrypted_key":
			s, err := dec.DecodeString()
			if err != nil {
				return err
			}
			fv, err := hex.DecodeString(s)
			if err != nil {
				return err
			}
			v.EncryptedKey = fv

		default:
			return errInvalidField.WithAttributes("field", s)
		}
	}
	return nil
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
		return v.SessionKeyId == nil
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
		return v.RootKeyId == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}
