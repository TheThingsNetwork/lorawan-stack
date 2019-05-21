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
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	joinEUIPrefixes = []*types.EUI64Prefix{
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
	ctx := test.Context()

	redisClient, flush := test.NewRedis(t, "joinserver_test")
	defer flush()
	defer redisClient.Close()
	devReg := &redis.DeviceRegistry{Redis: redisClient}
	keyReg := &redis.KeyRegistry{Redis: redisClient}

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	js := test.Must(New(
		c,
		&Config{
			Devices:         devReg,
			Keys:            keyReg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)).(*JoinServer)
	test.Must(nil, c.Start())

	// Invalid payload.
	req := ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload = ttnpb.NewPopulatedMessageDownlink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), false)
	res, err := js.HandleJoin(ctx, req)
	a.So(err, should.NotBeNil)
	a.So(res, should.BeNil)

	// No payload.
	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.Payload = nil
	res, err = js.HandleJoin(ctx, req)
	a.So(err, should.NotBeNil)
	a.So(res, should.BeNil)

	// JoinEUI out of range.
	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().JoinEUI = types.EUI64{0x11, 0x12, 0x13, 0x14, 0x42, 0x42, 0x42, 0x42}
	res, err = js.HandleJoin(ctx, req)
	a.So(err, should.NotBeNil)
	a.So(res, should.BeNil)

	// Empty JoinEUI.
	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().JoinEUI = types.EUI64{}
	res, err = js.HandleJoin(ctx, req)
	a.So(err, should.NotBeNil)
	a.So(res, should.BeNil)

	// Empty DevEUI.
	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().DevEUI = types.EUI64{}
	res, err = js.HandleJoin(ctx, req)
	a.So(err, should.NotBeNil)
	a.So(res, should.BeNil)

	// Random payload is invalid.
	res, err = js.HandleJoin(ctx, ttnpb.NewPopulatedJoinRequest(test.Randy, false))
	a.So(err, should.NotBeNil)
	a.So(res, should.BeNil)

	for _, tc := range []struct {
		Name string

		Device *ttnpb.EndDevice

		NextLastDevNonce  uint32
		NextLastJoinNonce uint32
		NextUsedDevNonces []uint32

		JoinRequest  *ttnpb.JoinRequest
		JoinResponse *ttnpb.JoinResponse

		ValidError func(error) bool
	}{
		{
			Name: "1.1.0/new device",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastDevNonce:  0,
			NextLastJoinNonce: 1,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
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
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(nwkKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xeb, 0xcd, 0x74, 0x59,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			},
		},
		{
			Name: "1.1.0/existing device",
			Device: &ttnpb.EndDevice{
				LastDevNonce:  0x2441,
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastDevNonce:  0x2442,
			NextLastJoinNonce: 0x42fffe,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
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
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: &ttnpb.JoinResponse{
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
						Key: KeyPtr(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			},
			ValidError: nil,
		},
		{
			Name: "1.1.0/DevNonce too small",
			Device: &ttnpb.EndDevice{
				LastDevNonce:  0x2442,
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastDevNonce:  0x2442,
			NextLastJoinNonce: 0x42fffd,
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_1,
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
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.2/new device",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0_2,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			NextUsedDevNonces: []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0_2,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			},
			ValidError: nil,
		},
		{
			Name: "1.0.1/new device",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0_1,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			NextUsedDevNonces: []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			},
			ValidError: nil,
		},
		{
			Name: "1.0.0/new device",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 1,
			NextUsedDevNonces: []uint32{1},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42,
					/** DevEUI **/
					0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42,
					/** DevNonce **/
					0x01, 0x00,

					/* MIC */
					0xc4, 0x8, 0x50, 0xcf,
				},
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: &ttnpb.JoinResponse{
				RawPayload: append([]byte{
					/* MHDR */
					0x20},
					mustEncryptJoinAccept(appKey, []byte{
						/* JoinNonce */
						0x01, 0x00, 0x00,
						/* NetID */
						0xff, 0xff, 0x42,
						/* DevAddr */
						0xff, 0xff, 0xff, 0x42,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xc9, 0x7a, 0x61, 0x04,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x01},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x01})),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			},
			ValidError: nil,
		},
		{
			Name: "1.0.0/existing device",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2444},
				LastJoinNonce: 0x42fffd,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442, 0x2444},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
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
				DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				NetID:   types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: &ttnpb.JoinResponse{
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
						Key: KeyPtr(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPtr(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x42, 0xff, 0xfe},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x24, 0x42})),
						KEKLabel: "",
					},
				},
				Lifetime: 0,
			},
			ValidError: nil,
		},
		{
			Name: "1.0.0/repeated DevNonce",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
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
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/no payload",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				NetID:              types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/not a join request payload",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
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
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/unsupported LoRaWAN version",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
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
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/no JoinEUI",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
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
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/raw payload that can't be unmarshalled",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					0x23, 0x42, 0xff, 0xff, 0xaa, 0x42, 0x42, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff,
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
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
		{
			Name: "1.0.0/invalid MType",
			Device: &ttnpb.EndDevice{
				UsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
				LastJoinNonce: 0x42fffe,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
					DeviceID:               "test-dev",
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KEKLabel: "",
					},
				},
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			NextLastJoinNonce: 0x42fffe,
			NextUsedDevNonces: []uint32{23, 41, 42, 52, 0x2442},
			JoinRequest: &ttnpb.JoinRequest{
				SelectedMACVersion: ttnpb.MAC_V1_0,
				Payload: &ttnpb.Message{
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
				NetID: types.NetID{0x42, 0xff, 0xff},
				DownlinkSettings: ttnpb.DLSettings{
					OptNeg:      true,
					Rx1DROffset: 0x7,
					Rx2DR:       0xf,
				},
				RxDelay: 0x42,
				CFList:  nil,
			},
			JoinResponse: nil,
			ValidError:   errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			redisClient, flush := test.NewRedis(t, "joinserver_test")
			defer flush()
			defer redisClient.Close()
			devReg := &redis.DeviceRegistry{Redis: redisClient}
			keyReg := &redis.KeyRegistry{Redis: redisClient}

			c := component.MustNew(test.GetLogger(t), &component.Config{})
			js := test.Must(New(
				c,
				&Config{
					Devices:         devReg,
					Keys:            keyReg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)).(*JoinServer)
			test.Must(nil, c.Start())

			pb := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			start := time.Now()

			ret, err := devReg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceID,
				[]string{
					"created_at",
					"ids.dev_eui",
					"ids.join_eui",
					"last_dev_nonce",
					"last_join_nonce",
					"lorawan_version",
					"network_server_address",
					"provisioner_id",
					"provisioning_data",
					"root_keys",
					"updated_at",
					"used_dev_nonces",
				},
				func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					if !a.So(stored, should.BeNil) {
						t.Fatal("Registry is not empty")
					}
					return CopyEndDevice(pb), []string{
						"ids.dev_eui",
						"ids.join_eui",
						"last_dev_nonce",
						"last_join_nonce",
						"lorawan_version",
						"network_server_address",
						"provisioner_id",
						"provisioning_data",
						"root_keys",
						"used_dev_nonces",
					}, nil
				},
			)
			if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
				t.Fatalf("Failed to create device: %s", err)
			}
			a.So(ret.CreatedAt, should.HappenAfter, start)
			a.So(ret.UpdatedAt, should.HappenAfter, start)
			a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
			pb.CreatedAt = ret.CreatedAt
			pb.UpdatedAt = ret.UpdatedAt
			a.So(ret, should.HaveEmptyDiff, pb)

			res, err := js.HandleJoin(ctx, deepcopy.Copy(tc.JoinRequest).(*ttnpb.JoinRequest))
			if tc.ValidError != nil {
				if !a.So(err, should.BeError) || !a.So(tc.ValidError(err), should.BeTrue) {
					t.Fatalf("Received an unexpected error: %s", err)
				}
				a.So(res, should.BeNil)
				return
			}

			if !a.So(err, should.BeNil) || !a.So(res, should.NotBeNil) {
				t.FailNow()
			}
			expectedResp := deepcopy.Copy(tc.JoinResponse).(*ttnpb.JoinResponse)
			a.So(res.SessionKeyID, should.NotBeEmpty)
			expectedResp.SessionKeyID = res.SessionKeyID
			a.So(res, should.Resemble, expectedResp)

			ret, err = devReg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
			if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
				t.FailNow()
			}
			a.So(ret.CreatedAt, should.Equal, pb.CreatedAt)
			a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
			pb.UpdatedAt = ret.UpdatedAt
			pb.LastJoinNonce = tc.NextLastJoinNonce
			if tc.JoinRequest.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				pb.UsedDevNonces = tc.NextUsedDevNonces
			} else {
				pb.LastDevNonce = tc.NextLastDevNonce
			}
			if !a.So(ret.Session, should.NotBeNil) {
				t.FailNow()
			}
			a.So([]time.Time{start, ret.GetSession().GetStartedAt(), time.Now()}, should.BeChronological)
			pb.Session = &ttnpb.Session{
				DevAddr:     tc.JoinRequest.DevAddr,
				SessionKeys: res.SessionKeys,
				StartedAt:   ret.GetSession().GetStartedAt(),
			}
			a.So(ret, should.HaveEmptyDiff, pb)

			res, err = js.HandleJoin(ctx, deepcopy.Copy(tc.JoinRequest).(*ttnpb.JoinRequest))
			a.So(err, should.BeError)
			a.So(res, should.BeNil)
		})
	}
}

func TestGetNwkSKeys(t *testing.T) {
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name string

		GetByID     func(context.Context, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.NwkSKeysResponse

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name: "Registry error",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return nil, errTest
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, ErrRegistryOperation.WithCause(errTest)) {
					t.FailNow()
				}
				return a.So(errors.IsInternal(err), should.BeTrue)
			},
		},
		{
			Name: "No SNwkSIntKey",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					FNwkSIntKey: ttnpb.NewPopulatedKeyEnvelope(test.Randy, false),
					NwkSEncKey:  ttnpb.NewPopulatedKeyEnvelope(test.Randy, false),
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoSNwkSIntKey)
			},
		},
		{
			Name: "No NwkSEncKey",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					FNwkSIntKey: ttnpb.NewPopulatedKeyEnvelope(test.Randy, false),
					SNwkSIntKey: ttnpb.NewPopulatedKeyEnvelope(test.Randy, false),
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoNwkSEncKey)
			},
		},
		{
			Name: "No FNwkSIntKey",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					SNwkSIntKey: ttnpb.NewPopulatedKeyEnvelope(test.Randy, false),
					NwkSEncKey:  ttnpb.NewPopulatedKeyEnvelope(test.Randy, false),
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoFNwkSIntKey)
			},
		},
		{
			Name: "Matching request",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"f_nwk_s_int_key",
					"nwk_s_enc_key",
					"s_nwk_s_int_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      KeyPtr(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key:      KeyPtr(types.AES128Key{0x43, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel: "NwkSEncKey-kek",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key:      KeyPtr(types.AES128Key{0x44, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel: "SNwkSIntKey-kek",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.NwkSKeysResponse{
				FNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      KeyPtr(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel: "",
				},
				NwkSEncKey: ttnpb.KeyEnvelope{
					Key:      KeyPtr(types.AES128Key{0x43, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel: "NwkSEncKey-kek",
				},
				SNwkSIntKey: ttnpb.KeyEnvelope{
					Key:      KeyPtr(types.AES128Key{0x44, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel: "SNwkSIntKey-kek",
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.ContextWithT(test.Context(), t)

			c := component.MustNew(test.GetLogger(t), &component.Config{})
			js := test.Must(New(
				c,
				&Config{
					Keys:    &MockKeyRegistry{GetByIDFunc: tc.GetByID},
					Devices: &MockDeviceRegistry{},
				},
			)).(*JoinServer)
			test.Must(nil, c.Start())
			res, err := js.GetNwkSKeys(ctx, tc.KeyRequest)

			if tc.ErrorAssertion != nil {
				if !tc.ErrorAssertion(t, err) {
					t.Errorf("Received unexpected error: %s", err)
				}
				a.So(res, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			a.So(res, should.Resemble, tc.KeyResponse)
		})
	}
}
