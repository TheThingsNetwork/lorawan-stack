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

package cryptoutil_test

import (
	"encoding/hex"
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestWrapAES128Key(t *testing.T) {
	var key types.AES128Key
	key.UnmarshalText([]byte("00112233445566778899AABBCCDDEEFF"))
	kekKey, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F")
	cipherKey, _ := hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")

	kekOther, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F1011121314151617")
	cipherOther, _ := hex.DecodeString("031D33264E15D33268F24EC260743EDCE1C6C7DDEE725A936BA814915C6762D2")

	v := NewMemKeyVault(map[string][]byte{
		"key":   kekKey,
		"other": kekOther,
	})

	for _, tc := range []struct {
		Name     string
		Key      types.AES128Key
		KEKLabel string
		Expected ttnpb.KeyEnvelope
	}{
		{
			Name: "WrapWithoutKEK",
			Key:  key,
			Expected: ttnpb.KeyEnvelope{
				EncryptedKey: key[:],
			},
		},
		{
			Name:     "WrapWithKEK",
			Key:      key,
			KEKLabel: "key",
			Expected: ttnpb.KeyEnvelope{
				EncryptedKey: cipherKey,
				KEKLabel:     "key",
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			env, err := WrapAES128Key(test.Context(), tc.Key, tc.KEKLabel, v)
			a.So(err, should.BeNil)
			a.So(env, should.Resemble, tc.Expected)
		})
	}

	for _, tc := range []struct {
		Name          string
		Envelope      ttnpb.KeyEnvelope
		ExpectedError func(error) bool
		ExpectedKey   types.AES128Key
	}{
		{
			Name: "UnwrapWithoutKEK",
			Envelope: ttnpb.KeyEnvelope{
				EncryptedKey: key[:],
			},
			ExpectedKey: key,
		},
		{
			Name: "UnwrapWithKEK",
			Envelope: ttnpb.KeyEnvelope{
				KEKLabel:     "key",
				EncryptedKey: cipherKey,
			},
			ExpectedKey: key,
		},
		{
			Name: "UnwrapInvalid",
			Envelope: ttnpb.KeyEnvelope{
				KEKLabel:     "other",
				EncryptedKey: cipherOther,
			},
			ExpectedError: errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			unwrapped, err := UnwrapAES128Key(test.Context(), tc.Envelope, v)
			if tc.ExpectedError != nil {
				a.So(tc.ExpectedError(err), should.BeTrue)
				return
			}
			a.So(err, should.BeNil)
			a.So(unwrapped, should.Resemble, tc.ExpectedKey)
		})
	}
}

func TestUnwrapSelectedSessionKeys(t *testing.T) {
	var key types.AES128Key
	test.Must(nil, key.UnmarshalText([]byte("00112233445566778899AABBCCDDEEFF")))
	kekKey := test.Must(hex.DecodeString("000102030405060708090A0B0C0D0E0F")).([]byte)
	cipherKey := test.Must(hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")).([]byte)

	v := NewMemKeyVault(map[string][]byte{
		"key": kekKey,
	})
	for _, tc := range []struct {
		Name                string
		SessionKeys         ttnpb.SessionKeys
		Prefix              string
		Paths               []string
		ExpectedSessionKeys ttnpb.SessionKeys
		ErrorAssertion      func(*testing.T, error) bool
	}{
		{
			Name:           "no keys/no prefix/no paths",
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/no prefix/no paths",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/no prefix/paths(nwk_s_enc_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Paths: []string{
				"nwk_s_enc_key",
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/no prefix/paths(app_s_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Paths: []string{
				"app_s_key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/no prefix/paths(app_s_key.key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Paths: []string{
				"app_s_key.key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "encrypted AppSKey/no prefix/paths(app_s_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "key",
					EncryptedKey: cipherKey,
				},
			},
			Paths: []string{
				"app_s_key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "encrypted AppSKey/no prefix/paths(app_s_key.key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "key",
					EncryptedKey: cipherKey,
				},
			},
			Paths: []string{
				"app_s_key.key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "encrypted AppSKey, decrypted Nwk keys/no prefix/paths(app_s_key.key,f_nwk_s_int_key.key,nwk_s_enc_key,s_nwk_s_int_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "key",
					EncryptedKey: cipherKey,
				},
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02},
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03},
				},
			},
			Paths: []string{
				"app_s_key.key",
				"f_nwk_s_int_key.key",
				"nwk_s_enc_key",
				"s_nwk_s_int_key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02},
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/prefix(test)/no paths",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Prefix:         "test",
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/prefix(test)/paths(app_s_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Prefix: "test",
			Paths: []string{
				"app_s_key",
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/prefix(test)/paths(app_s_key.key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Prefix: "test",
			Paths: []string{
				"app_s_key.key",
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/prefix(test)/paths(test.nwk_s_enc_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Prefix: "test",
			Paths: []string{
				"test.nwk_s_enc_key",
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/prefix(test)/paths(test.app_s_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Prefix: "test",
			Paths: []string{
				"test.app_s_key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "decrypted AppSKey/prefix(test)/paths(test.app_s_key.key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			Prefix: "test",
			Paths: []string{
				"test.app_s_key.key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "encrypted AppSKey/prefix(test)/paths(test.app_s_key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "key",
					EncryptedKey: cipherKey,
				},
			},
			Prefix: "test",
			Paths: []string{
				"test.app_s_key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "encrypted AppSKey/prefix(test)/paths(test.app_s_key.key)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "key",
					EncryptedKey: cipherKey,
				},
			},
			Prefix: "test",
			Paths: []string{
				"test.app_s_key.key",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "encrypted AppSKey, decrypted Nwk keys/prefix(test)/paths(test)",
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "key",
					EncryptedKey: cipherKey,
				},
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02},
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03},
				},
			},
			Prefix: "test",
			Paths: []string{
				"test",
			},
			ExpectedSessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &key,
				},
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02},
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			sk := deepcopy.Copy(tc.SessionKeys).(ttnpb.SessionKeys)
			ret, err := UnwrapSelectedSessionKeys(test.Context(), v, sk, tc.Prefix, tc.Paths...)
			a.So(sk, should.Resemble, tc.SessionKeys)
			a.So(ret, should.Resemble, tc.ExpectedSessionKeys)
			a.So(tc.ErrorAssertion(t, err), should.BeTrue)
		})
	}
}
