// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package joinserver_test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	. "github.com/TheThingsNetwork/ttn/pkg/joinserver"
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
	joinEUI = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nwkKey  = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	appKey  = types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
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
	js := New(&Config{
		Component: component.New(test.GetLogger(t), &component.Config{shared.DefaultServiceBase}),
		Registry:  reg,
		JoinEUI:   joinEUI,
	})

	resp, err := js.HandleJoin(context.Background(), nil)
	a.So(err, should.NotBeNil)
	a.So(resp, should.BeNil)

	req := ttnpb.NewPopulatedJoinRequest(test.Randy, false)
	req.Payload = *ttnpb.NewPopulatedMessageDownlink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), false)
	resp, err = js.HandleJoin(context.Background(), req)
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

		JoinRequest  *ttnpb.JoinRequest
		JoinResponse *ttnpb.JoinResponse

		Error error
	}{
		{
			"1.1",
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
						KekLabel: "",
					},
					NwkKey: &ttnpb.KeyEnvelope{
						Key:      &nwkKey,
						KekLabel: "",
					},
				},
				LoRaWANVersion: ttnpb.MAC_V1_1,
			},
			1,
			1,
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
						KekLabel: "",
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveSNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KekLabel: "",
					},
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveFNwkSIntKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KekLabel: "",
					},
					NwkSEncKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveNwkSEncKey(
							nwkKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KekLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0.2",
			&ttnpb.EndDevice{
				NextDevNonce:  42,
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KekLabel: "",
					},
				},
				LoRaWANVersion: ttnpb.MAC_V1_0_2,
			},
			42,
			1,
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
						KekLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KekLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0.1",
			&ttnpb.EndDevice{
				NextDevNonce:  42,
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KekLabel: "",
					},
				},
				LoRaWANVersion: ttnpb.MAC_V1_0_1,
			},
			42,
			1,
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
						KekLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KekLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
		{
			"1.0",
			&ttnpb.EndDevice{
				NextDevNonce:  42,
				NextJoinNonce: 0,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				RootKeys: &ttnpb.RootKeys{
					AppKey: &ttnpb.KeyEnvelope{
						Key:      &appKey,
						KekLabel: "",
					},
				},
				LoRaWANVersion: ttnpb.MAC_V1_0,
			},
			42,
			1,
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
						KekLabel: "",
					},
					AppSKey: &ttnpb.KeyEnvelope{
						Key: KeyPointer(crypto.DeriveLegacyAppSKey(
							appKey,
							types.JoinNonce{0x00, 0x00, 0x00},
							types.NetID{0x42, 0xff, 0xff},
							types.DevNonce{0x00, 0x00})),
						KekLabel: "",
					},
				},
				Lifetime: nil,
			},
			nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
			js := New(&Config{
				Component: component.New(test.GetLogger(t), &component.Config{shared.DefaultServiceBase}),
				Registry:  reg,
				JoinEUI:   joinEUI,
			})

			_, err := reg.Create(tc.Device)
			if !a.So(err, should.BeNil) {
				return
			}

			dev, err := reg.FindDeviceByIdentifiers(&tc.Device.EndDeviceIdentifiers)
			a.So(err, should.BeNil)
			if a.So(dev, should.NotBeNil) && a.So(dev, should.HaveLength, 1) {
				a.So(dev[0].EndDevice, should.Resemble, tc.Device)
			}

			resp, err := js.HandleJoin(context.Background(), tc.JoinRequest)
			if tc.Error != nil {
				a.So(err, should.Resemble, tc.Error)
				a.So(resp, should.BeNil)
			} else {
				a.So(err, should.BeNil)
				if a.So(resp, should.NotBeNil) {
					a.So(pretty.Diff(resp, tc.JoinResponse), should.BeEmpty)
				}
			}

			// ensure the stored device nonce(s) are updated
			time.Sleep(time.Millisecond)

			dev, err = reg.FindDeviceByIdentifiers(&tc.Device.EndDeviceIdentifiers)
			a.So(err, should.BeNil)
			if a.So(dev, should.NotBeNil) && a.So(dev, should.HaveLength, 1) {
				a.So(dev[0].GetNextDevNonce(), should.Equal, tc.NextNextDevNonce)
				a.So(dev[0].GetNextJoinNonce(), should.Equal, tc.NextNextJoinNonce)
			}
		})
	}
}
