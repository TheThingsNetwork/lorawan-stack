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

package joinserver_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestMemCryptoService(t *testing.T) {
	ctx := test.Context()
	svc := &joinserver.MemCryptoService{
		KeyVault: cryptoutil.NewMemKeyVault(map[string][]byte{}),
		RootKeys: ttnpb.RootKeys{
			RootKeyID: "test",
			NwkKey: &ttnpb.KeyEnvelope{
				KEKLabel: "",
				Key:      []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1},
			},
			AppKey: &ttnpb.KeyEnvelope{
				KEKLabel: "",
				Key:      []byte{0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2, 0x2},
			},
		},
	}
	makeIDs := func(version ttnpb.MACVersion) ttnpb.CryptoServiceEndDeviceIdentifiers {
		return ttnpb.CryptoServiceEndDeviceIdentifiers{
			LoRaWANVersion: version,
			JoinEUI:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		}
	}

	t.Run("JoinRequestMIC", func(t *testing.T) {
		for _, tc := range []struct {
			Version ttnpb.MACVersion
			Payload []byte
			Result  [4]byte
		}{
			{
				Version: ttnpb.MAC_V1_1,
				Payload: bytes.Repeat([]byte{0x1}, 19),
				Result:  [4]byte{0x21, 0x4d, 0x19, 0x7d},
			},
			{
				Version: ttnpb.MAC_V1_0_2,
				Payload: bytes.Repeat([]byte{0x1}, 19),
				Result:  [4]byte{0x87, 0x14, 0x9f, 0xd},
			},
			{
				Version: ttnpb.MAC_V1_0_1,
				Payload: bytes.Repeat([]byte{0x1}, 19),
				Result:  [4]byte{0x87, 0x14, 0x9f, 0xd},
			},
			{
				Version: ttnpb.MAC_V1_0,
				Payload: bytes.Repeat([]byte{0x1}, 19),
				Result:  [4]byte{0x87, 0x14, 0x9f, 0xd},
			},
		} {
			t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
				a := assertions.New(t)
				res, err := svc.JoinRequestMIC(ctx, makeIDs(tc.Version), tc.Payload)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Result)
			})
		}
	})

	t.Run("JoinAcceptMIC", func(t *testing.T) {
		for _, tc := range []struct {
			Version     ttnpb.MACVersion
			JoinReqType byte
			DevNonce    types.DevNonce
			Payload     []byte
			Result      [4]byte
		}{
			{
				Version:     ttnpb.MAC_V1_1,
				JoinReqType: 0xff,
				DevNonce:    types.DevNonce{0x1, 0x2},
				Payload:     bytes.Repeat([]byte{0x1}, 13),
				Result:      [4]byte{0x1, 0xdf, 0x1e, 0xff},
			},
			{
				Version:     ttnpb.MAC_V1_1,
				JoinReqType: 0x0,
				DevNonce:    types.DevNonce{0x1, 0x2},
				Payload:     bytes.Repeat([]byte{0x1}, 13),
				Result:      [4]byte{0xa, 0x9c, 0x88, 0x33},
			},
			{
				Version:     ttnpb.MAC_V1_1,
				JoinReqType: 0x1,
				DevNonce:    types.DevNonce{0x1, 0x2},
				Payload:     bytes.Repeat([]byte{0x1}, 13),
				Result:      [4]byte{0xae, 0x2d, 0xdc, 0xd1},
			},
			{
				Version:     ttnpb.MAC_V1_1,
				JoinReqType: 0x2,
				DevNonce:    types.DevNonce{0x1, 0x2},
				Payload:     bytes.Repeat([]byte{0x1}, 13),
				Result:      [4]byte{0x18, 0x32, 0x16, 0xb1},
			},
			{
				Version: ttnpb.MAC_V1_0_2,
				Payload: bytes.Repeat([]byte{0x1}, 13),
				Result:  [4]byte{0x3, 0x1b, 0x42, 0x0},
			},
			{
				Version: ttnpb.MAC_V1_0_1,
				Payload: bytes.Repeat([]byte{0x1}, 13),
				Result:  [4]byte{0x3, 0x1b, 0x42, 0x0},
			},
			{
				Version: ttnpb.MAC_V1_0,
				Payload: bytes.Repeat([]byte{0x1}, 13),
				Result:  [4]byte{0x3, 0x1b, 0x42, 0x0},
			},
		} {
			t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
				a := assertions.New(t)
				res, err := svc.JoinAcceptMIC(ctx, makeIDs(tc.Version), tc.JoinReqType, tc.DevNonce, tc.Payload)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Result)
			})
		}
	})

	t.Run("EncryptJoinAccept", func(t *testing.T) {
		for _, tc := range []struct {
			Version ttnpb.MACVersion
			Payload []byte
			Result  []byte
		}{
			{
				Version: ttnpb.MAC_V1_1,
				Payload: bytes.Repeat([]byte{0x1}, 16),
				Result:  []byte{0xbc, 0x6e, 0x2b, 0xaf, 0x23, 0xca, 0x1e, 0x66, 0xaa, 0xd7, 0xb3, 0x95, 0xc1, 0xd6, 0xa6, 0xa},
			},
			{
				Version: ttnpb.MAC_V1_0_2,
				Payload: bytes.Repeat([]byte{0x1}, 16),
				Result:  []byte{0xe3, 0xcd, 0xe2, 0x37, 0xc8, 0xf2, 0xd9, 0x7b, 0x8d, 0x79, 0xf9, 0x17, 0x1d, 0x4b, 0xda, 0xc1},
			},
			{
				Version: ttnpb.MAC_V1_0_1,
				Payload: bytes.Repeat([]byte{0x1}, 16),
				Result:  []byte{0xe3, 0xcd, 0xe2, 0x37, 0xc8, 0xf2, 0xd9, 0x7b, 0x8d, 0x79, 0xf9, 0x17, 0x1d, 0x4b, 0xda, 0xc1},
			},
			{
				Version: ttnpb.MAC_V1_0,
				Payload: bytes.Repeat([]byte{0x1}, 16),
				Result:  []byte{0xe3, 0xcd, 0xe2, 0x37, 0xc8, 0xf2, 0xd9, 0x7b, 0x8d, 0x79, 0xf9, 0x17, 0x1d, 0x4b, 0xda, 0xc1},
			},
		} {
			t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
				a := assertions.New(t)
				res, err := svc.EncryptJoinAccept(ctx, makeIDs(tc.Version), tc.Payload)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Result)
			})
		}
	})

	t.Run("EncryptRejoinAccept", func(t *testing.T) {
		for _, tc := range []struct {
			Version ttnpb.MACVersion
			Payload []byte
			Result  []byte
		}{
			{
				Version: ttnpb.MAC_V1_1,
				Payload: bytes.Repeat([]byte{0x1}, 16),
				Result:  []byte{0x61, 0x58, 0x25, 0x46, 0x6a, 0x90, 0xec, 0xce, 0xf5, 0xd1, 0xf1, 0xc5, 0xba, 0x56, 0x6b, 0xe7},
			},
		} {
			t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
				a := assertions.New(t)
				res, err := svc.EncryptRejoinAccept(ctx, makeIDs(tc.Version), tc.Payload)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Result)
			})
		}

		for _, version := range []ttnpb.MACVersion{
			ttnpb.MAC_V1_0_2,
			ttnpb.MAC_V1_0_1,
			ttnpb.MAC_V1_0,
		} {
			t.Run(fmt.Sprintf("%v", version), func(t *testing.T) {
				a := assertions.New(t)
				a.So(func() {
					svc.EncryptRejoinAccept(ctx, makeIDs(version), bytes.Repeat([]byte{0x1}, 16))
				}, should.Panic)
			})
		}
	})

	t.Run("DeriveNwkSKeys", func(t *testing.T) {
		for _, tc := range []struct {
			Version     ttnpb.MACVersion
			JoinNonce   types.JoinNonce
			DevNonce    types.DevNonce
			NetID       types.NetID
			FNwkSIntKey types.AES128Key
			SNwkSIntKey types.AES128Key
			NwkSEncKey  types.AES128Key
		}{
			{
				Version:     ttnpb.MAC_V1_1,
				JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:    types.DevNonce{0x1, 0x2},
				FNwkSIntKey: types.AES128Key{0xf8, 0xd8, 0xb8, 0xb9, 0xb1, 0xec, 0x36, 0xe8, 0xb8, 0x10, 0x84, 0x29, 0xd3, 0xf7, 0x3d, 0xd2},
				SNwkSIntKey: types.AES128Key{0x72, 0xde, 0xab, 0x55, 0x40, 0x3, 0xd2, 0x29, 0xc, 0xec, 0x8, 0x6, 0x81, 0x71, 0x92, 0x5d},
				NwkSEncKey:  types.AES128Key{0x31, 0x87, 0x9c, 0xf0, 0x93, 0xc2, 0x41, 0x11, 0xe3, 0x99, 0x5, 0xc7, 0x72, 0x76, 0xbf, 0xd8},
			},
			{
				Version:     ttnpb.MAC_V1_0_2,
				JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:    types.DevNonce{0x1, 0x2},
				NetID:       types.NetID{0x1, 0x2, 0x3},
				FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
				SNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
				NwkSEncKey:  types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
			},
			{
				Version:     ttnpb.MAC_V1_0_1,
				JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:    types.DevNonce{0x1, 0x2},
				NetID:       types.NetID{0x1, 0x2, 0x3},
				FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
				SNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
				NwkSEncKey:  types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
			},
			{
				Version:     ttnpb.MAC_V1_0,
				JoinNonce:   types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:    types.DevNonce{0x1, 0x2},
				NetID:       types.NetID{0x1, 0x2, 0x3},
				FNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
				SNwkSIntKey: types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
				NwkSEncKey:  types.AES128Key{0x77, 0x51, 0x9b, 0x3, 0x2d, 0x33, 0x6, 0x44, 0xe7, 0x6c, 0xe4, 0xd9, 0x4e, 0x93, 0x3c, 0xc5},
			},
		} {
			t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
				a := assertions.New(t)
				fNwkSIntKey, sNwkSIntKey, nwkSEncKey, err := svc.DeriveNwkSKeys(ctx, makeIDs(tc.Version), tc.JoinNonce, tc.DevNonce, tc.NetID)
				a.So(err, should.BeNil)
				a.So(fNwkSIntKey, should.Resemble, tc.FNwkSIntKey)
				a.So(sNwkSIntKey, should.Resemble, tc.SNwkSIntKey)
				a.So(nwkSEncKey, should.Resemble, tc.NwkSEncKey)
			})
		}
	})

	t.Run("DeriveAppSKey", func(t *testing.T) {
		for _, tc := range []struct {
			Version   ttnpb.MACVersion
			JoinNonce types.JoinNonce
			DevNonce  types.DevNonce
			NetID     types.NetID
			AppSKey   types.AES128Key
		}{
			{
				Version:   ttnpb.MAC_V1_1,
				JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:  types.DevNonce{0x1, 0x2},
				AppSKey:   types.AES128Key{0x4, 0x30, 0x89, 0x5c, 0x7b, 0xa7, 0xb1, 0x51, 0xcf, 0x97, 0x36, 0x84, 0xf6, 0x22, 0xff, 0xc1},
			},
			{
				Version:   ttnpb.MAC_V1_0_2,
				JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:  types.DevNonce{0x1, 0x2},
				NetID:     types.NetID{0x1, 0x2, 0x3},
				AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
			},
			{
				Version:   ttnpb.MAC_V1_0_1,
				JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:  types.DevNonce{0x1, 0x2},
				NetID:     types.NetID{0x1, 0x2, 0x3},
				AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
			},
			{
				Version:   ttnpb.MAC_V1_0,
				JoinNonce: types.JoinNonce{0x1, 0x2, 0x3},
				DevNonce:  types.DevNonce{0x1, 0x2},
				NetID:     types.NetID{0x1, 0x2, 0x3},
				AppSKey:   types.AES128Key{0xeb, 0x55, 0x14, 0xa2, 0x16, 0x6, 0xd8, 0x3d, 0x49, 0xec, 0x12, 0x73, 0x1, 0xf0, 0x7a, 0x91},
			},
		} {
			t.Run(fmt.Sprintf("%v", tc.Version), func(t *testing.T) {
				a := assertions.New(t)
				appSKey, err := svc.DeriveAppSKey(ctx, makeIDs(tc.Version), tc.JoinNonce, tc.DevNonce, tc.NetID)
				a.So(err, should.BeNil)
				a.So(appSKey, should.Resemble, tc.AppSKey)
			})
		}
	})
}
