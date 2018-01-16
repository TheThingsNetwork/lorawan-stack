// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver_test

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	. "github.com/TheThingsNetwork/ttn/pkg/joinserver"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"golang.org/x/net/context"
)

const appID = "test"

var (
	joinEUIPrefixes = []types.EUI64Prefix{
		{types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 42},
		{types.EUI64{0x10, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 12},
		{types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, 56},
	}
	nwkKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	appKey = types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nsAddr = net.IPv4(0x42, 0x42, 0x42, 0x42).String()
	asAddr = net.IPv4(0x42, 0x42, 0x42, 0xff).String()
)

func mustEncryptJoinAccept(key types.AES128Key, pld []byte) []byte {
	b, err := crypto.EncryptJoinAccept(key, pld)
	if err != nil {
		panic(errors.NewWithCause("failed to encrypt join accept", err))
	}
	return b
}

func TestHandleJoin(t *testing.T) {
	a := assertions.New(t)

	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	js := New(
		component.New(test.GetLogger(t), &component.Config{ServiceBase: shared.DefaultServiceBase}),
		&Config{
			Registry:        reg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)

	req := ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload = *ttnpb.NewPopulatedMessageDownlink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), false)
	resp, err := js.HandleJoin(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.EndDeviceIdentifiers.DevAddr = nil
	resp, err = js.HandleJoin(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.Payload = nil
	resp, err = js.HandleJoin(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().JoinEUI = types.EUI64{0x11, 0x12, 0x13, 0x14, 0x42, 0x42, 0x42, 0x42}
	resp, err = js.HandleJoin(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().JoinEUI = types.EUI64{}
	resp, err = js.HandleJoin(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload.GetJoinRequestPayload().DevEUI = types.EUI64{}
	resp, err = js.HandleJoin(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	resp, err = js.HandleJoin(context.Background(), ttnpb.NewPopulatedJoinRequest(test.Randy, false))
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

		Error error
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
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
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x19, 0x86, 0x7c, 0x5f,
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
						0x42, 0xff, 0xff,
						/* DevAddr */
						0x42, 0xff, 0xff, 0xff,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0x3e, 0xd7, 0x4a, 0x70,
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
				NextJoinNonce: 0x424242,
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			0x2443,
			0x424243,
			[]uint32{0, 42, 0x2441, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x79, 0x8, 0xfd, 0x3d,
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
						0x42, 0x42, 0x42,
						/* NetID */
						0x42, 0xff, 0xff,
						/* DevAddr */
						0x42, 0xff, 0xff, 0xff,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0xa5, 0x2c, 0x95, 0x4c,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveAppSKey(
							appKey,
							types.JoinNonce{0x42, 0x42, 0x42},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x42, 0x24})),
						KEKLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0x42, 0x42},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x42, 0x24})),
						KEKLabel: "",
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x42, 0x42, 0x42},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x42, 0x24})),
						KEKLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x42, 0x42, 0x42},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x42, 0x24})),
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
				NextJoinNonce: 0x424242,
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
				NetworkServerAddress: nsAddr,
			},
			0x2442,
			0x424242,
			[]uint32{0, 42, 0x2441, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_1,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x79, 0x8, 0xfd, 0x3d,
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
			ErrDevNonceTooSmall.New(nil),
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
				LoRaWANVersion:       ttnpb.MAC_V1_1,
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
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x19, 0x86, 0x7c, 0x5f,
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
			ErrAddressMismatch.New(errors.Attributes{
				"component": "network server",
			}),
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
				LoRaWANVersion:       ttnpb.MAC_V1_0_2,
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
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x6b, 0x94, 0x91, 0x59,
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
						0x42, 0xff, 0xff,
						/* DevAddr */
						0x42, 0xff, 0xff, 0xff,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0x2a, 0xa5, 0xbf, 0x25,
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
				LoRaWANVersion:       ttnpb.MAC_V1_0_1,
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
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x6b, 0x94, 0x91, 0x59,
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
						0x42, 0xff, 0xff,
						/* DevAddr */
						0x42, 0xff, 0xff, 0xff,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0x2a, 0xa5, 0xbf, 0x25,
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
				LoRaWANVersion:       ttnpb.MAC_V1_0,
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
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x00, 0x00,

					/* MIC */
					0x6b, 0x94, 0x91, 0x59,
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
						0x42, 0xff, 0xff,
						/* DevAddr */
						0x42, 0xff, 0xff, 0xff,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0x2a, 0xa5, 0xbf, 0x25,
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
				NextJoinNonce: 0x424242,
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
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			0,
			0x424243,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x8, 0xc4, 0x4, 0x4a,
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
						0x42, 0x42, 0x42,
						/* NetID */
						0x42, 0xff, 0xff,
						/* DevAddr */
						0x42, 0xff, 0xff, 0xff,
						/* DLSettings */
						0xff,
						/* RxDelay */
						0x42,

						/* MIC */
						0x54, 0x90, 0x4, 0xb6,
					})...),
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyNwkSKey(
							appKey,
							types.JoinNonce{0x42, 0x42, 0x42},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x42, 0x24})),
						KEKLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x42, 0x42, 0x42},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x42, 0x24})),
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
				NextJoinNonce: 0x424242,
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
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				NetworkServerAddress: nsAddr,
			},
			0,
			0x424242,
			[]uint32{23, 41, 42, 52, 0x2442},
			&ttnpb.JoinRequest{
				SelectedMacVersion: ttnpb.MAC_V1_0,
				RawPayload: []byte{
					/* MHDR */
					0x00,

					/* MACPayload */
					/** JoinEUI **/
					0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevEUI **/
					0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					/** DevNonce **/
					0x42, 0x24,

					/* MIC */
					0x8, 0xc4, 0x4, 0x4a,
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
			ErrDevNonceReused.New(nil),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
			js := New(
				component.New(test.GetLogger(t), &component.Config{ServiceBase: shared.DefaultServiceBase}),
				&Config{
					Registry:        reg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)

			_, err := reg.Create(tc.Device)
			if !a.So(err, should.BeNil) {
				return
			}

			dev, err := deviceregistry.FindDeviceByIdentifiers(reg, &tc.Device.EndDeviceIdentifiers)
			a.So(err, should.BeNil)
			if a.So(dev, should.NotBeNil) && a.So(dev, should.HaveLength, 1) {
				a.So(pretty.Diff(dev[0].EndDevice, tc.Device), should.BeEmpty)
			}

			ctx := (rpcmetadata.MD{
				NetAddress: nsAddr,
			}).ToIncomingContext(context.Background())

			start := time.Now()
			resp, err := js.HandleJoin(ctx, tc.JoinRequest)
			if tc.Error != nil {
				a.So(errors.From(err).Attributes(), should.Resemble, errors.From(tc.Error).Attributes())
				a.So(errors.From(err).Code(), should.Resemble, errors.From(tc.Error).Code())
				a.So(resp, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			if !a.So(resp, should.Resemble, tc.JoinResponse) {
				pretty.Ldiff(t, resp, tc.JoinResponse)
				return
			}

			// ensure the stored device nonces are updated
			time.Sleep(time.Millisecond)

			dev, err = deviceregistry.FindDeviceByIdentifiers(reg, &tc.Device.EndDeviceIdentifiers)
			a.So(err, should.BeNil)
			if a.So(dev, should.NotBeNil) && a.So(dev, should.HaveLength, 1) {
				a.So(dev[0].GetNextDevNonce(), should.Equal, tc.NextNextDevNonce)
				a.So(dev[0].GetNextJoinNonce(), should.Equal, tc.NextNextJoinNonce)
				a.So(pretty.Diff(dev[0].GetUsedDevNonces(), tc.NextUsedDevNonces), should.BeEmpty)
				if s := dev[0].GetSession(); tc.Error == nil && a.So(s, should.NotBeNil) {
					a.So(s.DevAddr, should.Resemble, tc.JoinRequest.EndDeviceIdentifiers.DevAddr)
					a.So(s.SessionKeys, should.Resemble, resp.SessionKeys)
					a.So([]time.Time{start, s.StartedAt, time.Now()}, should.BeChronological)
				}
			}

			resp, err = js.HandleJoin(context.Background(), tc.JoinRequest)
			a.So(err, should.BeError)
			a.So(resp, should.BeNil)
		})
	}
}

func TestGetAppSKey(t *testing.T) {
	a := assertions.New(t)

	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	js := New(
		component.New(test.GetLogger(t), &component.Config{ServiceBase: shared.DefaultServiceBase}),
		&Config{
			Registry:        reg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)

	req := ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.DevEUI = types.EUI64{}
	resp, err := js.GetAppSKey(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.SessionKeyID = ""
	resp, err = js.GetAppSKey(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	for _, tc := range []struct {
		Name string

		Device *ttnpb.EndDevice

		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.AppSKeyResponse

		Error error
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
			"No session",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: asAddr,
			},
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrNoSession.New(nil),
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
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrSessionKeyIDMismatch.New(nil),
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
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrSessionKeyIDMismatch.New(nil),
		},
		{
			"Address mismatch",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				ApplicationServerAddress: "test",
			},
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrAddressMismatch.New(errors.Attributes{
				"component": "application server",
			}),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
			js := New(
				component.New(test.GetLogger(t), &component.Config{ServiceBase: shared.DefaultServiceBase}),
				&Config{
					Registry:        reg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)

			_, err := reg.Create(tc.Device)
			if !a.So(err, should.BeNil) {
				return
			}

			dev, err := deviceregistry.FindDeviceByIdentifiers(reg, &tc.Device.EndDeviceIdentifiers)
			a.So(err, should.BeNil)
			if a.So(dev, should.NotBeNil) && a.So(dev, should.HaveLength, 1) {
				a.So(pretty.Diff(dev[0].EndDevice, tc.Device), should.BeEmpty)
			}

			ctx := (rpcmetadata.MD{
				NetAddress: asAddr,
			}).ToIncomingContext(context.Background())

			resp, err := js.GetAppSKey(ctx, tc.KeyRequest)
			if tc.Error != nil {
				a.So(errors.From(err).Attributes(), should.Resemble, errors.From(tc.Error).Attributes())
				a.So(errors.From(err).Code(), should.Resemble, errors.From(tc.Error).Code())
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

	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	js := New(
		component.New(test.GetLogger(t), &component.Config{ServiceBase: shared.DefaultServiceBase}),
		&Config{
			Registry:        reg,
			JoinEUIPrefixes: joinEUIPrefixes,
		},
	)

	req := ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.DevEUI = types.EUI64{}
	resp, err := js.GetNwkSKeys(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req = ttnpb.NewPopulatedSessionKeyRequest(test.Randy, false)
	req.SessionKeyID = ""
	resp, err = js.GetNwkSKeys(context.Background(), req)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	for _, tc := range []struct {
		Name string

		Device *ttnpb.EndDevice

		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.NwkSKeysResponse

		Error error
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
			"No session",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: nsAddr,
			},
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrNoSession.New(nil),
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
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrSessionKeyIDMismatch.New(nil),
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
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrSessionKeyIDMismatch.New(nil),
		},
		{
			"Address mismatch",
			&ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI: &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				NetworkServerAddress: "test",
			},
			&ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: "test",
			},
			nil,
			ErrAddressMismatch.New(errors.Attributes{
				"component": "network server",
			}),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
			js := New(
				component.New(test.GetLogger(t), &component.Config{ServiceBase: shared.DefaultServiceBase}),
				&Config{
					Registry:        reg,
					JoinEUIPrefixes: joinEUIPrefixes,
				},
			)

			_, err := reg.Create(tc.Device)
			if !a.So(err, should.BeNil) {
				return
			}

			dev, err := deviceregistry.FindDeviceByIdentifiers(reg, &tc.Device.EndDeviceIdentifiers)
			a.So(err, should.BeNil)
			if a.So(dev, should.NotBeNil) && a.So(dev, should.HaveLength, 1) {
				a.So(pretty.Diff(dev[0].EndDevice, tc.Device), should.BeEmpty)
			}

			ctx := (rpcmetadata.MD{
				NetAddress: nsAddr,
			}).ToIncomingContext(context.Background())

			resp, err := js.GetNwkSKeys(ctx, tc.KeyRequest)
			if tc.Error != nil {
				a.So(errors.From(err).Attributes(), should.Resemble, errors.From(tc.Error).Attributes())
				a.So(errors.From(err).Code(), should.Resemble, errors.From(tc.Error).Code())
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
