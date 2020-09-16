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
	"fmt"
	"math"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestAppendRecentDownlink(t *testing.T) {
	downs := [...]*ttnpb.DownlinkMessage{
		{
			RawPayload: []byte("test1"),
		},
		{
			RawPayload: []byte("test2"),
		},
		{
			RawPayload: []byte("test3"),
		},
	}
	for _, tc := range []struct {
		Recent   []*ttnpb.DownlinkMessage
		Down     *ttnpb.DownlinkMessage
		Window   int
		Expected []*ttnpb.DownlinkMessage
	}{
		{
			Down:     downs[0],
			Window:   1,
			Expected: downs[:1],
		},
		{
			Recent:   downs[:1],
			Down:     downs[1],
			Window:   1,
			Expected: downs[1:2],
		},
		{
			Recent:   downs[:2],
			Down:     downs[2],
			Window:   1,
			Expected: downs[2:3],
		},
		{
			Recent:   downs[:1],
			Down:     downs[1],
			Window:   2,
			Expected: downs[:2],
		},
		{
			Recent:   downs[:2],
			Down:     downs[2],
			Window:   2,
			Expected: downs[1:3],
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     fmt.Sprintf("recent_length:%d,window:%v", len(tc.Recent), tc.Window),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				recent := CopyDownlinkMessages(tc.Recent...)
				down := CopyDownlinkMessage(tc.Down)
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
	appID := ttnpb.ApplicationIdentifiers{ApplicationID: appIDString}
	const devID = "generate-data-downlink-test-dev-id"

	devAddr := types.DevAddr{0x42, 0xff, 0xff, 0xff}

	fNwkSIntKey := types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nwkSEncKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	sNwkSIntKey := types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	encodeMessage := func(msg *ttnpb.Message, ver ttnpb.MACVersion, confFCnt uint32) []byte {
		msg = deepcopy.Copy(msg).(*ttnpb.Message)
		pld := msg.GetMACPayload()

		b, err := lorawan.MarshalMessage(*msg)
		if err != nil {
			t.Fatal("Failed to marshal downlink")
		}

		var key types.AES128Key
		switch ver {
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			key = fNwkSIntKey
		case ttnpb.MAC_V1_1:
			key = sNwkSIntKey
		default:
			panic(fmt.Errorf("unknown version %s", ver))
		}

		mic, err := crypto.ComputeDownlinkMIC(key, pld.DevAddr, confFCnt, pld.FCnt, b)
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
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				},
				Session:           ttnpb.NewPopulatedSession(test.Randy, false),
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Error: errNoDownlink,
		},
		{
			Name: "1.1/no app downlink/status after 1 downlink/no ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					LastDevStatusFCntUp: 2,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				},
				Session: &ttnpb.Session{
					LastFCntUp: 4,
				},
				LoRaWANPHYVersion:       ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:         band.EU_863_870,
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
			},
			Error: errNoDownlink,
		},
		{
			Name: "1.1/no app downlink/status after an hour/no ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: DurationPtr(24 * time.Hour),
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				},
				LoRaWANPHYVersion:       ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:         band.EU_863_870,
				LastDevStatusReceivedAt: TimePtr(time.Now()),
				Session:                 ttnpb.NewPopulatedSession(test.Randy, false),
			},
			Error: errNoDownlink,
		},
		{
			Name: "1.1/no app downlink/no MAC/ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
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
					DevAddr:       devAddr,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
								ADR: true,
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
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_CONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{
										FHDR: ttnpb.FHDR{
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
						DevAddr:       devAddr,
						LastNFCntDown: 41,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/unconfirmed app downlink/no MAC/no ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
					RxWindowsAvailable: true,
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  false,
							FCnt:       42,
							FPort:      1,
							FRMPayload: []byte("test"),
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
								ADR: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			},
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  false,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						}},
						RxWindowsAvailable: true,
					},
					Session: &ttnpb.Session{
						DevAddr: devAddr,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/unconfirmed app downlink/no MAC/ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
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
					DevAddr: devAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  false,
							FCnt:       42,
							FPort:      1,
							FRMPayload: []byte("test"),
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
								ADR: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			},
			ConfFCnt: 24,
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  false,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_CONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{
										FHDR: ttnpb.FHDR{
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
						DevAddr: devAddr,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/confirmed app downlink/no MAC/no ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  true,
							FCnt:       42,
							FPort:      1,
							FRMPayload: []byte("test"),
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
								ADR: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			},
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  true,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						}},
					},
					Session: &ttnpb.Session{
						DevAddr: devAddr,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/confirmed app downlink/no MAC/ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
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
					DevAddr: devAddr,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							Confirmed:  true,
							FCnt:       42,
							FPort:      1,
							FRMPayload: []byte("test"),
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
								ADR: true,
							},
							FCnt: 42,
						},
						FullFCnt:   42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			},
			ConfFCnt: 24,
			ApplicationDownlinkAssertion: func(t *testing.T, down *ttnpb.ApplicationDownlink) bool {
				return assertions.New(t).So(down, should.Resemble, &ttnpb.ApplicationDownlink{
					Confirmed:  true,
					FCnt:       42,
					FPort:      1,
					FRMPayload: []byte("test"),
				})
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_CONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{
										FHDR: ttnpb.FHDR{
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
						DevAddr: devAddr,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/no app downlink/status(count)/no ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					LastDevStatusFCntUp: 4,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				},
				Session: &ttnpb.Session{
					DevAddr:       devAddr,
					LastFCntUp:    99,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
								ADR: true,
							},
							FCnt: 42,
							FOpts: MustEncryptDownlink(nwkSEncKey, devAddr, 42, true, MakeDownlinkMACBuffer(
								LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B],
								ttnpb.CID_DEV_STATUS,
							)...),
						},
						FullFCnt: 42,
					},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 3},
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:      ttnpb.MAC_V1_1,
						LastDevStatusFCntUp: 4,
						PendingRequests: []*ttnpb.MACCommand{
							ttnpb.CID_DEV_STATUS.MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						}},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastFCntUp:    99,
						LastNFCntDown: 41,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
				})
			},
		},
		{
			Name: "1.1/no app downlink/status(time/zero time)/no ack",
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: appID,
					DeviceID:               devID,
					DevAddr:                &devAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: DurationPtr(time.Nanosecond),
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					RecentUplinks: []*ttnpb.UplinkMessage{{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
					}},
				},
				Session: &ttnpb.Session{
					DevAddr:       devAddr,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &nwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &sNwkSIntKey,
						},
					},
				},
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				FrequencyPlanID:   band.EU_863_870,
			},
			Payload: &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: devAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
								ADR: true,
							},
							FCnt: 42,
							FOpts: MustEncryptDownlink(nwkSEncKey, devAddr, 42, true, MakeDownlinkMACBuffer(
								LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B],
								ttnpb.CID_DEV_STATUS,
							)...),
						},
						FullFCnt: 42,
					},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				if !a.So(dev.MACState, should.NotBeNil) {
					t.FailNow()
				}
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: appID,
						DeviceID:               devID,
						DevAddr:                &devAddr,
					},
					MACSettings: &ttnpb.MACSettings{
						StatusTimePeriodicity: DurationPtr(time.Nanosecond),
					},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						PendingRequests: []*ttnpb.MACCommand{
							ttnpb.CID_DEV_STATUS.MACCommand(),
						},
						RecentUplinks: []*ttnpb.UplinkMessage{{
							Payload: &ttnpb.Message{
								MHDR: ttnpb.MHDR{
									MType: ttnpb.MType_UNCONFIRMED_UP,
								},
								Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
							},
						}},
					},
					Session: &ttnpb.Session{
						DevAddr:       devAddr,
						LastNFCntDown: 41,
						SessionKeys: ttnpb.SessionKeys{
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &nwkSEncKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &sNwkSIntKey,
							},
						},
					},
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					FrequencyPlanID:   band.EU_863_870,
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
					&component.Config{},
					component.WithClusterNew(func(context.Context, *cluster.Config, ...cluster.Option) (cluster.Cluster, error) {
						return &test.MockCluster{
							JoinFunc: test.ClusterJoinNilFunc,
						}, nil
					}),
				)
				c.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

				componenttest.StartComponent(t, c)

				ns := &NetworkServer{
					Component: c,
					ctx:       ctx,
					defaultMACSettings: ttnpb.MACSettings{
						StatusTimePeriodicity:  DurationPtr(0),
						StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
					},
				}

				dev := CopyEndDevice(tc.Device)
				phy, err := DeviceBand(dev, ns.FrequencyPlans)
				if !a.So(err, should.BeNil) {
					t.Fail()
					return
				}

				genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, dev.MACState.DeviceClass, time.Now(), math.MaxUint16, math.MaxUint16)
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

				b := encodeMessage(tc.Payload, dev.MACState.LoRaWANVersion, tc.ConfFCnt)
				a.So(genDown.RawPayload, should.Resemble, b)
				pld := CopyMessage(tc.Payload)
				pld.MIC = b[len(b)-4:]
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
