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

package joinserver_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	joinEUIPrefixes = []types.EUI64Prefix{
		{EUI64: types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 42},
		{EUI64: types.EUI64{0x10, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 12},
		{EUI64: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, Length: 56},
	}
	nwkKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	appKey = types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nsAddr = net.IPv4(0x42, 0x42, 0x42, 0x42).String()
	asAddr = net.IPv4(0x42, 0x42, 0x42, 0xff).String()
)

func mustEncryptJoinAccept(key types.AES128Key, pld []byte) []byte {
	b, err := crypto.EncryptJoinAccept(key, pld)
	if err != nil {
		panic(fmt.Sprintf("failed to encrypt join-accept: %s", err))
	}
	return b
}

func TestHandleJoin(t *testing.T) {
	a := assertions.New(t)

	authorizedCtx := clusterauth.NewContext(test.Context(), nil)

	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	js := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:        reg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)).(*JoinServer)

	req := ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload = *ttnpb.NewPopulatedMessageDownlink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), false)
	resp, err := js.HandleJoin(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.EndDeviceIdentifiers.DevAddr = nil
	resp, err = js.HandleJoin(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.Payload = nil
	resp, err = js.HandleJoin(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().JoinEUI = types.EUI64{0x11, 0x12, 0x13, 0x14, 0x42, 0x42, 0x42, 0x42}
	resp, err = js.HandleJoin(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().JoinEUI = types.EUI64{}
	resp, err = js.HandleJoin(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().DevEUI = types.EUI64{}
	resp, err = js.HandleJoin(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	resp, err = js.HandleJoin(authorizedCtx, ttnpb.NewPopulatedJoinRequest(test.Randy, false))
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	for _, tc := range []struct {
		Name string

		Device *ttnpb.EndDevice

		NextNextDevNonce  uint32
		NextNextJoinNonce uint32
		NextUsedDevNonces []uint32

		JoinRequest  *ttnpb.JoinRequest
		JoinResponse *ttnpb.JoinResponse

		ValidErr func(error) bool
	}{
		{
			"1.1 new device",
			&ttnpb.EndDevice{
				NextDevNonce:  0,
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key:      &nwkKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				NetworkServerAddress: nsAddr,
			},
			1,
			1,
			[]uint32{0},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			&ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x00, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0x16, 0x41, 0x9f, 0x29,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.1 existing device",
			&ttnpb.EndDevice{
				NextDevNonce:  0x2442,
				UsedDevNonces: []uint32{0, 42, 0x2441},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key:      &nwkKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				NetworkServerAddress: nsAddr,
			},
			0x2443,
			0x42ffff,
			[]uint32{0, 42, 0x2441, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x6e, 0x54, 0x1b, 0x37,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			&ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0xfe, 0xff, 0x42,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc8, 0xf7, 0x62, 0xf4,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.1 DevNonce too small",
			&ttnpb.EndDevice{
				NextDevNonce:  0x2443,
				UsedDevNonces: []uint32{0, 42, 0x2441, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key:      &nwkKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				NetworkServerAddress: nsAddr,
			},
			0x2442,
			0x42fffe,
			[]uint32{0, 42, 0x2441, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x6e, 0x54, 0x1b, 0x37,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"1.1 address mismatch",
			&ttnpb.EndDevice{
				NextDevNonce:  0,
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key:      &nwkKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				NetworkServerAddress: net.IPv4(0x45, 0x44, 0x43, 0x43).String(),
			},
			1,
			1,
			[]uint32{0},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x55, 0x17, 0x54, 0x8e,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsPermissionDenied,
		},
		{
			"1.0.2 new device",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52},
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0_2,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			1,
			[]uint32{23, 41, 42, 52, 0},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0_2,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0xcc, 0x15, 0x6f, 0xa,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			&ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x00, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xad, 0x48, 0xaf, 0x94,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0.1 new device",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52},
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0_1,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			1,
			[]uint32{23, 41, 42, 52, 0},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0xcc, 0x15, 0x6f, 0xa,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			&ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x00, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xad, 0x48, 0xaf, 0x94,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0 new device",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52},
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			1,
			[]uint32{23, 41, 42, 52, 0},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0xcc, 0x15, 0x6f, 0xa,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			&ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x00, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xad, 0x48, 0xaf, 0x94,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0 existing device",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42ffff,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0xed, 0x8b, 0xd2, 0x24,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			&ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0xfe, 0xff, 0x42,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xf8, 0x4a, 0x11, 0x8e,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0 repeated DevNonce",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0xed, 0x8b, 0xd2, 0x24,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"1.0 no payload",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"1.0 not a join request payload",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				Payload: ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
					},
					Payload: &ttnpb.Message_JoinAcceptPayload{},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"1.0 unsupported lorawan version",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				Payload: ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
						Major: ttnpb.Major(10),
					},
					Payload: &ttnpb.Message_JoinRequestPayload{},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"1.0 no joineui",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				Payload: ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
						Major: ttnpb.Major_LORAWAN_R1,
					},
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							DevEUI: types.EUI64{0x27, 0x00, 0x00, 0x00, 0x00, 0xab, 0xaa, 0x00},
						},
					},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"1.0 raw payload that can't be unmarshalled",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					0x23, 0x42, 0xff, 0xff, 0xaa, 0x42, 0x42, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff,
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"1.0 invalid mtype",
			&ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				NextJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					LoRaWANVersion: ttnpb.MAC_V1_0,
				},
				NetworkServerAddress: nsAddr,
			},
			0,
			0x42fffe,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				Payload: ttnpb.Message{
					MHDR: ttnpb.MHDR{
						MType: ttnpb.MType_JOIN_REQUEST,
					},
					Payload: &ttnpb.Message_JoinRequestPayload{
						JoinRequestPayload: &ttnpb.JoinRequestPayload{
							DevEUI:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							JoinEUI: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						},
					},
				},
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					DevAddr: &types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			nil,
			errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Registry:        reg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)).(*JoinServer)

			dev, err := reg.Create(deepcopy.Copy(tc.Device).(*ttnpb.EndDevice))
			if !a.So(err, should.BeNil) {
				return
			}

			ctx := (rpcmetadata.MD{
				NetAddress: nsAddr,
			}).ToIncomingContext(authorizedCtx)

			start := time.Now()
			resp, err := js.HandleJoin(ctx, tc.JoinRequest)
			if tc.ValidErr != nil {
				if !a.So(tc.ValidErr(err), should.BeTrue) {
					t.Fatalf("Received an unexpected error: %s", err)
				}
				a.So(resp, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			if !a.So(resp, should.Resemble, tc.JoinResponse) {
				pretty.Ldiff(t, resp, tc.JoinResponse)
				return
			}

			// ensure the stored device nonces are updated
			time.Sleep(test.Delay)

			dev, err = dev.Load()
			if !a.So(err, should.BeNil) {
				return
			}

			ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
			ed.CreatedAt = dev.EndDevice.GetCreatedAt()
			ed.UpdatedAt = dev.EndDevice.GetUpdatedAt()
			ed.NextDevNonce = tc.NextNextDevNonce
			ed.NextJoinNonce = tc.NextNextJoinNonce
			ed.UsedDevNonces = tc.NextUsedDevNonces
			if tc.ValidErr == nil {
				a.So([]time.Time{start, dev.GetSession().GetStartedAt(), time.Now()}, should.BeChronological)
				ed.Session = &ttnpb.Session{
					DevAddr:     *tc.JoinRequest.EndDeviceIdentifiers.DevAddr,
					SessionKeys: resp.SessionKeys,
					StartedAt:   dev.GetSession().GetStartedAt(),
				}
			}

			a.So(pretty.Diff(ed, dev.EndDevice), should.BeEmpty)

			resp, err = js.HandleJoin(authorizedCtx, tc.JoinRequest)
			a.So(err, should.BeError)
			a.So(resp, should.BeNil)
		})
	}
}

func TestGetAppSKey(t *testing.T) {
	a := assertions.New(t)

	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	js := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:        reg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)).(*JoinServer)

	authorizedCtx := clusterauth.NewContext(test.Context(), nil)

	req := ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.DevEUI = types.EUI64{}
	resp, err := js.GetAppSKey(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.SessionKeyID = ""
	resp, err = js.GetAppSKey(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	for _, tc := range []struct {
		Name string

		Device *ttnpb.EndDevice

		CustomCreateDevice func(*ttnpb.EndDevice) error

		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.AppSKeyResponse

		ValidErr func(error) bool
	}{
		{
			"Valid session",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						AppSKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			&ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
					KEKLabel: "test",
				},
			},
			nil,
		},
		{
			"Valid fallback",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "zest",
					},
				},
				SessionFallback: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						AppSKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			&ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
					KEKLabel: "test",
				},
			},
			nil,
		},
		{
			"No device",
			nil,
			func(*ttnpb.EndDevice) error { return nil },
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsNotFound,
		},
		{
			"No session",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"Corrupt session",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"ID mismatch",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "zest",
					},
				},
				SessionFallback: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "fest",
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"ID mismatch no fallback",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "zest",
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"Address mismatch",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: "test",
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsPermissionDenied,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Registry:        reg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)).(*JoinServer)

			ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
			if tc.CustomCreateDevice != nil {
				err = tc.CustomCreateDevice(ed)
			} else {
				_, err = reg.Create(ed)
			}
			if !a.So(err, should.BeNil) {
				return
			}

			ctx := (rpcmetadata.MD{
				NetAddress: asAddr,
			}).ToIncomingContext(authorizedCtx)

			resp, err := js.GetAppSKey(ctx, tc.KeyRequest)
			if tc.ValidErr != nil {
				a.So(tc.ValidErr(err), should.BeTrue)
				a.So(resp, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			if !a.So(resp, should.Resemble, tc.KeyResponse) {
				pretty.Ldiff(t, resp, tc.KeyResponse)
			}
		})
	}
}

func TestGetNwkSKeys(t *testing.T) {
	a := assertions.New(t)

	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	js := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:        reg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)).(*JoinServer)

	authorizedCtx := clusterauth.NewContext(test.Context(), nil)

	req := ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.DevEUI = types.EUI64{}
	resp, err := js.GetNwkSKeys(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.SessionKeyID = ""
	resp, err = js.GetNwkSKeys(authorizedCtx, req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	for _, tc := range []struct {
		Name string

		Device *ttnpb.EndDevice

		CustomCreateDevice func(*ttnpb.EndDevice) error

		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.NwkSKeysResponse

		ValidErr func(error) bool
	}{
		{
			"Valid request",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
							KEKLabel: "test",
						},
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			&ttnpb.NwkSKeysResponse{
				FNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
					KEKLabel: "test",
				},
				SNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
					KEKLabel: "test",
				},
				NwkSEncKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
					KEKLabel: "test",
				},
			},
			nil,
		},
		{
			"Valid fallback",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "zest",
					},
				},
				SessionFallback: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
							KEKLabel: "test",
						},
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			&ttnpb.NwkSKeysResponse{
				FNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
					KEKLabel: "test",
				},
				SNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
					KEKLabel: "test",
				},
				NwkSEncKey: ttnpb.KeyEnvelope{
					Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
					KEKLabel: "test",
				},
			},
			nil,
		},
		{
			"No device",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
							KEKLabel: "test",
						},
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			func(*ttnpb.EndDevice) error { return nil },
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsNotFound,
		},
		{
			"No NwkSEncKey",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"No FNwkSIntKey",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
							KEKLabel: "test",
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"No SNwkSIntKey",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "test",
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff},
							KEKLabel: "test",
						},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key:      &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff},
							KEKLabel: "test",
						},
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"No session",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsDataLoss,
		},
		{
			"ID mismatch",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "zest",
					},
				},
				SessionFallback: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "fest",
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"ID mismatch no fallback",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: "zest",
					},
				},
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsInvalidArgument,
		},
		{
			"Address mismatch",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: "test",
			},
			nil,
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			errors.IsPermissionDenied,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Registry:        reg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)).(*JoinServer)

			ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
			if tc.CustomCreateDevice != nil {
				err = tc.CustomCreateDevice(ed)
			} else {
				_, err = reg.Create(ed)
			}
			if !a.So(err, should.BeNil) {
				return
			}

			ctx := (rpcmetadata.MD{
				NetAddress: nsAddr,
			}).ToIncomingContext(authorizedCtx)

			resp, err := js.GetNwkSKeys(ctx, tc.KeyRequest)
			if tc.ValidErr != nil {
				a.So(tc.ValidErr(err), should.BeTrue)
				a.So(resp, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			if !a.So(resp, should.Resemble, tc.KeyResponse) {
				pretty.Ldiff(t, resp, tc.KeyResponse)
			}
		})
	}
}
