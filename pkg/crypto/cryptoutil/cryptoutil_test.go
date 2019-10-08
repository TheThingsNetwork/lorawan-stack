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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestWrapAES128Key(t *testing.T) {
	var appSKey types.AES128Key
	appSKey.UnmarshalText([]byte("00112233445566778899AABBCCDDEEFF"))
	kekSKey, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F")
	cipherSKey, _ := hex.DecodeString("1FA68B0A8112B447AEF34BD8FB5A7B829D3E862371D2CFE5")

	kekOther, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0F1011121314151617")
	cipherOther, _ := hex.DecodeString("031D33264E15D33268F24EC260743EDCE1C6C7DDEE725A936BA814915C6762D2")

	v := cryptoutil.NewMemKeyVault(map[string][]byte{
		"skey":  kekSKey,
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
			Key:  appSKey,
			Expected: ttnpb.KeyEnvelope{
				EncryptedKey: appSKey[:],
			},
		},
		{
			Name:     "WrapWithKEK",
			Key:      appSKey,
			KEKLabel: "skey",
			Expected: ttnpb.KeyEnvelope{
				EncryptedKey: cipherSKey,
				KEKLabel:     "skey",
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			env, err := cryptoutil.WrapAES128Key(test.Context(), tc.Key, tc.KEKLabel, v)
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
				EncryptedKey: appSKey[:],
			},
			ExpectedKey: appSKey,
		},
		{
			Name: "UnwrapWithKEK",
			Envelope: ttnpb.KeyEnvelope{
				KEKLabel:     "skey",
				EncryptedKey: cipherSKey,
			},
			ExpectedKey: appSKey,
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
			unwrapped, err := cryptoutil.UnwrapAES128Key(test.Context(), tc.Envelope, v)
			if tc.ExpectedError != nil {
				a.So(tc.ExpectedError(err), should.BeTrue)
				return
			}
			a.So(err, should.BeNil)
			a.So(unwrapped, should.Resemble, tc.ExpectedKey)
		})
	}
}
