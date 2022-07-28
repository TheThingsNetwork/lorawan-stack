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

package networkserver

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestAppendRecentDownlink(t *testing.T) {
	downs := [...]*ttnpb.MACState_DownlinkMessage{
		{
			Payload: &ttnpb.MACState_DownlinkMessage_Message{
				MHdr: &ttnpb.MACState_DownlinkMessage_Message_MHDR{
					MType: 0x01,
				},
			},
		},
		{
			Payload: &ttnpb.MACState_DownlinkMessage_Message{
				MHdr: &ttnpb.MACState_DownlinkMessage_Message_MHDR{
					MType: 0x02,
				},
			},
		},
		{
			Payload: &ttnpb.MACState_DownlinkMessage_Message{
				MHdr: &ttnpb.MACState_DownlinkMessage_Message_MHDR{
					MType: 0x03,
				},
			},
		},
	}
	for _, tc := range []struct {
		Recent   []*ttnpb.MACState_DownlinkMessage
		Down     *ttnpb.DownlinkMessage
		Window   int
		Expected []*ttnpb.MACState_DownlinkMessage
	}{
		{
			Down: &ttnpb.DownlinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: 0x01,
					},
				},
			},
			Window:   1,
			Expected: downs[:1],
		},
		{
			Recent: downs[:1],
			Down: &ttnpb.DownlinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: 0x02,
					},
				},
			},
			Window:   1,
			Expected: downs[1:2],
		},
		{
			Recent: downs[:2],
			Down: &ttnpb.DownlinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: 0x03,
					},
				},
			},
			Window:   1,
			Expected: downs[2:3],
		},
		{
			Recent: downs[:1],
			Down: &ttnpb.DownlinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: 0x02,
					},
				},
			},
			Window:   2,
			Expected: downs[:2],
		},
		{
			Recent: downs[:2],
			Down: &ttnpb.DownlinkMessage{
				Payload: &ttnpb.Message{
					MHdr: &ttnpb.MHDR{
						MType: 0x03,
					},
				},
			},
			Window:   2,
			Expected: downs[1:3],
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     fmt.Sprintf("recent_length:%d,window:%v", len(tc.Recent), tc.Window),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				recent := ttnpb.CloneSlice(tc.Recent)
				down := ttnpb.Clone(tc.Down)
				ret := appendRecentDownlink(recent, down, tc.Window)
				a.So(recent, should.Resemble, tc.Recent)
				a.So(down, should.Resemble, tc.Down)
				a.So(ret, should.Resemble, tc.Expected)
			},
		})
	}
}

func TestGenerateDataDownlink(t *testing.T) {
	const appIDString = "generate-data-downlink-test-app-id"
	appID := &ttnpb.ApplicationIdentifiers{ApplicationId: appIDString}
	const devID = "generate-data-downlink-test-dev-id"

	devAddr := types.DevAddr{0x42, 0xff, 0xff, 0xff}

	fNwkSIntKey := types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nwkSEncKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	sNwkSIntKey := types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	encodeMessage := func(msg *ttnpb.Message, ver ttnpb.MACVersion, confFCnt uint32) []byte {
		msg = ttnpb.Clone(msg)
		pld := msg.GetMacPayload()

		b, err := lorawan.MarshalMessage(msg)
		if err != nil {
			t.Fatal("Failed to marshal downlink")
		}

		var key types.AES128Key
		switch ver {
		case ttnpb.MACVersion_MAC_V1_0, ttnpb.MACVersion_MAC_V1_0_1, ttnpb.MACVersion_MAC_V1_0_2:
			key = fNwkSIntKey
		case ttnpb.MACVersion_MAC_V1_1:
			key = sNwkSIntKey
		default:
			panic(fmt.Errorf("unknown version %s", ver))
		}

		mic, err := crypto.ComputeDownlinkMIC(
			key,
			types.MustDevAddr(pld.FHdr.DevAddr).OrZero(),
			confFCnt,
			pld.FHdr.FCnt,
			b,
		)
		if err != nil {
			t.Fatal("Failed to compute MIC")
		}
		return append(b, mic[:]...)
	}

	for _, tc := range []struct {
		Name                         string
		Device                       *ttnpb.EndDevice
		Payload                      *ttnpb.Message
		ConfFCnt                     uint32
		ApplicationDownlinkAssertion func(t *testing.T, down *ttnpb.ApplicationDownlink) bool
		DeviceAssertion              func(*testing.T, *ttnpb.EndDevice) bool
		Error                        error
	}{
		{
			Name: "1.1/no app downlink/no MAC/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
				},
				Session:           generateSession(),
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Error: errNoDownlink,
		},
		{
			Name: "1.1/no app downlink/status after 1 downlink/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
				},
				MacState: &ttnpb.MACState{
					CurrentParameters:   &ttnpb.MACParameters{},
					DesiredParameters:   &ttnpb.MACParameters{},
					LorawanVersion:      ttnpb.MACVersion_MAC_V1_1,
					LastDevStatusFCntUp: 2,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
				},
				Session: &ttnpb.Session{
					LastFCntUp: 4,
				},
				LorawanPhyVersion:       ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:         band.EU_863_870,
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(time.Unix(42, 0)),
			},
			Error: errNoDownlink,
		},
		{
			Name: "1.1/no app downlink/status after an hour/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: ttnpb.ProtoDurationPtr(24 * time.Hour),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
				},
				LorawanPhyVersion:       ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:         band.EU_863_870,
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(time.Now()),
				Session:                 generateSession(),
			},
			Error: errNoDownlink,
		},
		{
			Name: "1.1/no app downlink/no MAC/ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{
								MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCnt: 24,
									},
									FullFCnt: 24,
								},
							},
						},
					}},
					RecentDownlinks: ToMACStateDownlinkMessages(
						MakeDataDownlink(&DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MACVersion_MAC_V1_1,
						}),
					),
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr:       devAddr.Bytes(),
					LastNFCntDown: 41,
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: true,
								Adr: true,
							},
							FCnt: 42,
						},
						FullFCnt: 42,
					},
				},
			},
			ConfFCnt: 24,
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacState: &ttnpb.MACState{
						CurrentParameters: &ttnpb.MACParameters{},
						DesiredParameters: &ttnpb.MACParameters{},
						LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_CONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{
									MacPayload: &ttnpb.MACPayload{
										FHdr: &ttnpb.FHDR{
											FCnt: 24,
										},
										FullFCnt: 24,
									},
								},
							},
						}},
						RecentDownlinks: ToMACStateDownlinkMessages(
							MakeDataDownlink(&DataDownlinkConfig{
								DecodePayload: true,
								MACVersion:    ttnpb.MACVersion_MAC_V1_1,
							}),
						),
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr.Bytes(),
						LastNFCntDown: 41,
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/unconfirmed app downlink/no MAC/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr.Bytes(),
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  false,
							FCnt:       42,
							FPort:      1,
							FrmPayload: []byte("test"),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: false,
								Adr: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FrmPayload: []byte("test"),
					},
				},
			},
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  false,
					FCnt:       42,
					FPort:      1,
					FrmPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacState: &ttnpb.MACState{
						CurrentParameters: &ttnpb.MACParameters{},
						DesiredParameters: &ttnpb.MACParameters{},
						LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCtrl: &ttnpb.FCtrl{},
									},
								}},
							},
						}},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr: devAddr.Bytes(),
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/unconfirmed app downlink/no MAC/ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{
								MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCnt: 24,
									},
									FullFCnt: 24,
								},
							},
						},
					}},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr.Bytes(),
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  false,
							FCnt:       42,
							FPort:      1,
							FrmPayload: []byte("test"),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: true,
								Adr: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FrmPayload: []byte("test"),
					},
				},
			},
			ConfFCnt: 24,
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  false,
					FCnt:       42,
					FPort:      1,
					FrmPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacState: &ttnpb.MACState{
						CurrentParameters: &ttnpb.MACParameters{},
						DesiredParameters: &ttnpb.MACParameters{},
						LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_CONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{
									MacPayload: &ttnpb.MACPayload{
										FHdr: &ttnpb.FHDR{
											FCnt: 24,
										},
										FullFCnt: 24,
									},
								},
							},
						}},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr: devAddr.Bytes(),
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/confirmed app downlink/no MAC/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr.Bytes(),
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  true,
							FCnt:       42,
							FPort:      1,
							FrmPayload: []byte("test"),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: false,
								Adr: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FrmPayload: []byte("test"),
					},
				},
			},
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  true,
					FCnt:       42,
					FPort:      1,
					FrmPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MacState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacState: &ttnpb.MACState{
						CurrentParameters: &ttnpb.MACParameters{},
						DesiredParameters: &ttnpb.MACParameters{},
						LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCtrl: &ttnpb.FCtrl{},
									},
								}},
							},
						}},
					},
					Session: &ttnpb.Session{
						DevAddr: devAddr.Bytes(),
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/confirmed app downlink/no MAC/ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{
								MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCnt: 24,
									},
									FullFCnt: 24,
								},
							},
						},
					}},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr.Bytes(),
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  true,
							FCnt:       42,
							FPort:      1,
							FrmPayload: []byte("test"),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: true,
								Adr: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FrmPayload: []byte("test"),
					},
				},
			},
			ConfFCnt: 24,
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  true,
					FCnt:       42,
					FPort:      1,
					FrmPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MacState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacState: &ttnpb.MACState{
						CurrentParameters: &ttnpb.MACParameters{},
						DesiredParameters: &ttnpb.MACParameters{},
						LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_CONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{
									MacPayload: &ttnpb.MACPayload{
										FHdr: &ttnpb.FHDR{
											FCnt: 24,
										},
										FullFCnt: 24,
									},
								},
							},
						}},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr: devAddr.Bytes(),
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/no app downlink/status(count)/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
				},
				MacState: &ttnpb.MACState{
					CurrentParameters:   &ttnpb.MACParameters{},
					DesiredParameters:   &ttnpb.MACParameters{},
					LorawanVersion:      ttnpb.MACVersion_MAC_V1_1,
					LastDevStatusFCntUp: 4,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
					RecentDownlinks: ToMACStateDownlinkMessages(
						MakeDataDownlink(&DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MACVersion_MAC_V1_1,
						}),
					),
				},
				Session: &ttnpb.Session{
					DevAddr:       devAddr.Bytes(),
					LastFCntUp:    99,
					LastNFCntDown: 41,
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: false,
								Adr: true,
							},
							FCnt: 42,
							FOpts: MustEncryptDownlink(nwkSEncKey, devAddr, 42,
								macspec.EncryptionOptions(ttnpb.MACVersion_MAC_V1_1, macspec.DownlinkFrame, 0, true),
								MakeDownlinkMACBuffer(
									LoRaWANBands[band.EU_863_870][ttnpb.PHYVersion_RP001_V1_1_REV_B],
									ttnpb.MACCommandIdentifier_CID_DEV_STATUS,
								)...),
						},
						FullFCnt: 42,
					},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MacState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacSettings: &ttnpb.MACSettings{
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
					},
					MacState: &ttnpb.MACState{
						CurrentParameters:   &ttnpb.MACParameters{},
						DesiredParameters:   &ttnpb.MACParameters{},
						LorawanVersion:      ttnpb.MACVersion_MAC_V1_1,
						LastDevStatusFCntUp: 4,
						PendingRequests: []*ttnpb.MACCommand{
							ttnpb.MACCommandIdentifier_CID_DEV_STATUS.MACCommand(),
						},
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCtrl: &ttnpb.FCtrl{},
									},
								}},
							},
						}},
						RecentDownlinks: ToMACStateDownlinkMessages(
							MakeDataDownlink(&DataDownlinkConfig{
								DecodePayload: true,
								MACVersion:    ttnpb.MACVersion_MAC_V1_1,
							}),
						),
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr.Bytes(),
						LastFCntUp:    99,
						LastNFCntDown: 41,
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/no app downlink/status(time/zero time)/no ack",
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       devID,
					DevAddr:        devAddr.Bytes(),
				},
				MacSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: ttnpb.ProtoDurationPtr(time.Nanosecond),
				},
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
						Payload: &ttnpb.Message{
							MHdr: &ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
								FHdr: &ttnpb.FHDR{
									FCtrl: &ttnpb.FCtrl{},
								},
							}},
						},
					}},
					RecentDownlinks: ToMACStateDownlinkMessages(
						MakeDataDownlink(&DataDownlinkConfig{
							DecodePayload: true,
							MACVersion:    ttnpb.MACVersion_MAC_V1_1,
						}),
					),
				},
				Session: &ttnpb.Session{
					DevAddr:       devAddr.Bytes(),
					LastNFCntDown: 41,
					Keys: &ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: nwkSEncKey.Bytes(),
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: sNwkSIntKey.Bytes(),
						},
					},
				},
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				FrequencyPlanId:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FHdr: &ttnpb.FHDR{
							DevAddr: devAddr.Bytes(),
							FCtrl: &ttnpb.FCtrl{
								Ack: false,
								Adr: true,
							},
							FCnt: 42,
							FOpts: MustEncryptDownlink(nwkSEncKey, devAddr, 42,
								macspec.EncryptionOptions(ttnpb.MACVersion_MAC_V1_1, macspec.DownlinkFrame, 0, true),
								MakeDownlinkMACBuffer(
									LoRaWANBands[band.EU_863_870][ttnpb.PHYVersion_RP001_V1_1_REV_B],
									ttnpb.MACCommandIdentifier_CID_DEV_STATUS,
								)...),
						},
						FullFCnt: 42,
					},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MacState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appID,
						DeviceId:       devID,
						DevAddr:        devAddr.Bytes(),
					},
					MacSettings: &ttnpb.MACSettings{
						StatusTimePeriodicity: ttnpb.ProtoDurationPtr(time.Nanosecond),
					},
					MacState: &ttnpb.MACState{
						CurrentParameters: &ttnpb.MACParameters{},
						DesiredParameters: &ttnpb.MACParameters{},
						LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
						PendingRequests: []*ttnpb.MACCommand{
							ttnpb.MACCommandIdentifier_CID_DEV_STATUS.MACCommand(),
						},
						RecentUplinks: []*ttnpb.MACState_UplinkMessage{{
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
									FHdr: &ttnpb.FHDR{
										FCtrl: &ttnpb.FCtrl{},
									},
								}},
							},
						}},
						RecentDownlinks: ToMACStateDownlinkMessages(
							MakeDataDownlink(&DataDownlinkConfig{
								DecodePayload: true,
								MACVersion:    ttnpb.MACVersion_MAC_V1_1,
							}),
						),
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr.Bytes(),
						LastNFCntDown: 41,
						Keys: &ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: nwkSEncKey.Bytes(),
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: sNwkSIntKey.Bytes(),
							},
						},
					},
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					FrequencyPlanId:   band.EU_863_870,
				})
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				c := component.MustNew(
					log.Noop,
					&component.Config{
						ServiceBase: config.ServiceBase{
							FrequencyPlans: config.FrequencyPlansConfig{
								ConfigSource: "static",
								Static:       test.StaticFrequencyPlans,
							},
						},
					},
					component.WithClusterNew(func(context.Context, *cluster.Config, ...cluster.Option) (cluster.Cluster, error) {
						return &test.MockCluster{
							JoinFunc: test.ClusterJoinNilFunc,
						}, nil
					}),
				)

				componenttest.StartComponent(t, c)

				ns := &NetworkServer{
					Component: c,
					ctx:       ctx,
					defaultMACSettings: &ttnpb.MACSettings{
						StatusTimePeriodicity:  ttnpb.ProtoDurationPtr(0),
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
					},
				}

				dev := ttnpb.Clone(tc.Device)
				fps, err := ns.FrequencyPlansStore(ctx)
				if !a.So(err, should.BeNil) {
					t.Fail()
					return
				}

				phy, err := DeviceBand(dev, fps)
				if !a.So(err, should.BeNil) {
					t.Fail()
					return
				}

				genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, dev.MacState.DeviceClass, time.Now(), math.MaxUint16, math.MaxUint16)
				if tc.Error != nil {
					a.So(err, should.EqualErrorOrDefinition, tc.Error)
					a.So(genDown, should.BeNil)
					return
				}
				// TODO: Assert AS uplinks generated(https://github.com/TheThingsNetwork/lorawan-stack/issues/631).

				if !a.So(err, should.BeNil) || !a.So(genDown, should.NotBeNil) {
					t.Fail()
					return
				}

				b := encodeMessage(tc.Payload, dev.MacState.LorawanVersion, tc.ConfFCnt)
				a.So(genDown.RawPayload, should.Resemble, b)
				pld := ttnpb.Clone(tc.Payload)
				pld.Mic = b[len(b)-4:]
				a.So(genDown.Payload, should.Resemble, pld)
				if tc.ApplicationDownlinkAssertion != nil {
					a.So(tc.ApplicationDownlinkAssertion(t, genState.ApplicationDownlink), should.BeTrue)
				} else {
					a.So(genState.ApplicationDownlink, should.BeNil)
				}

				if tc.DeviceAssertion != nil {
					a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
				} else {
					a.So(dev, should.Resemble, tc.Device)
				}
			},
		})
	}
}

func generateSession() *ttnpb.Session {
	randomVal := uint32(random.Int63n(100))
	var key types.AES128Key
	rand.Read(key[:])
	keys := &ttnpb.SessionKeys{
		SessionKeyId: []byte{0x01, 0x02, 0x03, 0x04},
		FNwkSIntKey: &ttnpb.KeyEnvelope{
			KekLabel: "FNwkSIntKey",
			Key:      key.Bytes(),
		},
		SNwkSIntKey: &ttnpb.KeyEnvelope{
			KekLabel: "SNwkSIntKey",
			Key:      key.Bytes(),
		},
		NwkSEncKey: &ttnpb.KeyEnvelope{
			KekLabel: "NwkSEncKey",
			Key:      key.Bytes(),
		},
		AppSKey: &ttnpb.KeyEnvelope{
			KekLabel: "AppSKey",
			Key:      key.Bytes(),
		},
	}
	queuedDownlinks := make([]*ttnpb.ApplicationDownlink, randomVal%5)
	for i := range queuedDownlinks {
		payload := make([]byte, randomVal%5)
		rand.Read(payload[:])
		queuedDownlinks[i] = &ttnpb.ApplicationDownlink{
			FPort:      uint32(i + 1),
			FCnt:       randomVal + uint32(i),
			FrmPayload: payload,
		}
	}
	return &ttnpb.Session{
		DevAddr:                    types.DevAddr{0x26, 0x01, 0xff, 0xff}.Bytes(),
		Keys:                       keys,
		LastFCntUp:                 randomVal,
		LastNFCntDown:              randomVal,
		LastAFCntDown:              randomVal,
		StartedAt:                  ttnpb.ProtoTimePtr(time.Now()),
		QueuedApplicationDownlinks: queuedDownlinks,
	}
}
