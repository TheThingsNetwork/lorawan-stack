// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestAppendRecentUplink(t *testing.T) {
	ups := [...]*ttnpb.UplinkMessage{
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
		Recent   []*ttnpb.UplinkMessage
		Up       *ttnpb.UplinkMessage
		Window   int
		Expected []*ttnpb.UplinkMessage
	}{
		{
			Up:       ups[0],
			Window:   1,
			Expected: ups[:1],
		},
		{
			Recent:   ups[:1],
			Up:       ups[1],
			Window:   1,
			Expected: ups[1:2],
		},
		{
			Recent:   ups[:2],
			Up:       ups[2],
			Window:   1,
			Expected: ups[2:3],
		},
		{
			Recent:   ups[:1],
			Up:       ups[1],
			Window:   2,
			Expected: ups[:2],
		},
		{
			Recent:   ups[:2],
			Up:       ups[2],
			Window:   2,
			Expected: ups[1:3],
		},
	} {
		t.Run(fmt.Sprintf("recent_length:%d,window:%v", len(tc.Recent), tc.Window), func(t *testing.T) {
			a := assertions.New(t)
			recent := CopyUplinkMessages(tc.Recent...)
			up := CopyUplinkMessage(tc.Up)
			ret := appendRecentUplink(recent, up, tc.Window)
			a.So(recent, should.Resemble, tc.Recent)
			a.So(up, should.Resemble, tc.Up)
			a.So(ret, should.Resemble, tc.Expected)
		})
	}
}

func TestMatchAndHandleUplink(t *testing.T) {
	type TestCase struct {
		Name            string
		Uplink          *ttnpb.UplinkMessage
		Deduplicated    bool
		MakeDevices     func(context.Context) []contextualEndDevice
		DeviceAssertion func(*testing.T, *matchedDevice) bool
		Error           error
	}
	var tcs []TestCase

	fpID := test.EUFrequencyPlanID
	phyVersion := ttnpb.PHY_V1_1_REV_B
	fp := test.Must(frequencyplans.NewStore(test.FrequencyPlansFetcher).GetByID(fpID)).(*frequencyplans.FrequencyPlan)
	phy := test.Must(test.Must(band.GetByID(fp.BandID)).(band.Band).Version(phyVersion)).(band.Band)
	chIdx := uint8(len(phy.UplinkChannels) - 1)
	ch := phy.UplinkChannels[chIdx]
	drIdx := ch.MaxDataRate
	dr := phy.DataRates[drIdx].Rate

	for _, deduplicated := range [2]bool{
		true,
		false,
	} {
		deduplicated := deduplicated
		makeName := func(parts ...string) string {
			return MakeTestCaseName(append(parts, fmt.Sprintf("Deduplicated:%v", deduplicated))...)
		}
		macVersion := ttnpb.MAC_V1_0_4
		tcs = append(tcs,
			TestCase{
				Name: makeName("Payload too short"),
				Uplink: &ttnpb.UplinkMessage{
					Settings: MakeUplinkSettings(dr, ch.Frequency),
				},
				MakeDevices: func(ctx context.Context) []contextualEndDevice {
					return []contextualEndDevice{
						{
							Context: ctx,
							EndDevice: &ttnpb.EndDevice{
								EndDeviceIdentifiers: *MakeABPIdentifiers(true),
								FrequencyPlanID:      test.EUFrequencyPlanID,
								LoRaWANPHYVersion:    phyVersion,
								LoRaWANVersion:       macVersion,
								MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
								Session: &ttnpb.Session{
									DevAddr:     DevAddr,
									SessionKeys: *MakeSessionKeys(macVersion, false),
								},
							},
						},
					}
				},
				Deduplicated: deduplicated,
				DeviceAssertion: func(t *testing.T, dev *matchedDevice) bool {
					return assertions.New(t).So(dev, should.BeNil)
				},
				Error: errRawPayloadTooShort,
			},
		)
	}
	ForEachMACVersion(func(makeName func(...string) string, macVersion ttnpb.MACVersion) {
		makeSession := func(lastFCntUp, lastConfFCntDown uint32) *ttnpb.Session {
			return &ttnpb.Session{
				DevAddr:          DevAddr,
				LastConfFCntDown: lastConfFCntDown,
				LastFCntUp:       lastFCntUp,
				SessionKeys:      *MakeSessionKeys(macVersion, false),
				StartedAt:        time.Unix(0, 1),
			}
		}
		makeIdentifiers := func() ttnpb.EndDeviceIdentifiers { return *MakeABPIdentifiers(macVersion.RequireDevEUIForABP()) }
		makeNilEvents := func(bool) []events.DefinitionDataClosure { return nil }

		type devConfig struct {
			Name   string
			Device *ttnpb.EndDevice

			Error                    error
			FCnt                     uint32
			FCntReset                bool
			NbTrans                  uint32
			Pending                  bool
			ApplyDeviceDiff          func(deduplicated bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice
			MakeQueuedEvents         func(deduplicated bool) []events.DefinitionDataClosure
			QueuedApplicationUplinks []*ttnpb.ApplicationUp
			SetPaths                 []string
		}
		makeConditionalConfigFunc := func(cond func() bool) func(conf devConfig) devConfig {
			return func(conf devConfig) devConfig {
				if cond() {
					return conf
				}
				return devConfig{
					Name:   conf.Name,
					Device: conf.Device,
					Error:  errDeviceNotFound,
				}
			}
		}
		ifNoFCntGap := makeConditionalConfigFunc(func() bool { return !macVersion.HasMaxFCntGap() })
		ifPre1_1 := makeConditionalConfigFunc(func() bool { return macVersion.Compare(ttnpb.MAC_V1_1) < 0 })

		makeApplicationDownlink := func(confirmed bool, fCnt uint16) *ttnpb.ApplicationDownlink {
			return &ttnpb.ApplicationDownlink{
				SessionKeyID: []byte("test-id"),
				FPort:        0x42,
				FCnt:         uint32(fCnt),
				FRMPayload:   []byte("test-payload"),
				Confirmed:    confirmed,
			}
		}

		upRecvAt := time.Unix(0, 42)
		for upConf, devConfs := range map[*struct {
			Confirmed         bool
			Ack               bool
			ADR               bool
			ADRAckReq         bool
			ClassB            bool
			FCnt              uint32
			ConfFCntDown      uint32
			FPort             uint8
			FRMPayload, FOpts []byte
		}][]devConfig{
			{}: {
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0", "NbTrans=1"),
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: makeIdentifiers(),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    phyVersion,
						LoRaWANVersion:       macVersion,
						MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
						Session:              makeSession(0, 0),
					},
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				},
				ifPre1_1(devConfig{
					Name: MakeTestCaseName("Pending session", "Class A", "NbTrans=1"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							PendingMACState:      MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							PendingSession:       makeSession(0x00, 0x00),
						}
						dev.PendingMACState.PendingJoinRequest = MakeNsJsJoinRequest(macVersion, phyVersion, fp, &DevAddr, ttnpb.RX_DELAY_4, 1, ttnpb.DATA_RATE_3)
						return dev
					}(),
					Pending: true,
					NbTrans: 1,
					SetPaths: []string{
						"ids.dev_addr",
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState = dev.PendingMACState
						dev.MACState.RxWindowsAvailable = true
						dev.MACState.CurrentParameters.Rx1DataRateOffset = 1
						dev.MACState.CurrentParameters.Rx1Delay = ttnpb.RX_DELAY_4
						dev.MACState.CurrentParameters.Rx2DataRateIndex = ttnpb.DATA_RATE_3
						dev.MACState.CurrentParameters.Channels = MakeDefaultEU868DesiredChannels()
						dev.Session = dev.PendingSession
						dev.Session.StartedAt = upRecvAt
						dev.MACState.PendingJoinRequest = nil
						dev.PendingMACState = nil
						dev.PendingSession = nil
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				}),
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0x00", "NbTrans=2", "MaxNbTrans=1"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:              makeSession(0x00, 0x00),
						}
						dev.MACState.PendingRequests = []*ttnpb.MACCommand{ttnpb.CID_DEV_STATUS.MACCommand()}
						up := MakeDataUplink(macVersion, true, false, DevAddr, ttnpb.FCtrl{}, 0x00, 0x00, 0x00, nil, nil, dr, drIdx, phy.UplinkChannels[chIdx-1].Frequency, chIdx-1)
						up.ReceivedAt = upRecvAt.Add(-time.Nanosecond)
						dev.MACState.RecentUplinks = appendRecentUplink(dev.MACState.RecentUplinks, up, recentUplinkCount)
						dev.RecentUplinks = appendRecentUplink(dev.RecentUplinks, up, recentUplinkCount)
						return dev
					}(),
					Error: errDeviceNotFound,
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0x00", "NbTrans=2", "MaxNbTrans=2"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:              makeSession(0x00, 0x00),
						}
						dev.MACState.CurrentParameters.ADRNbTrans = 2
						dev.MACState.PendingRequests = []*ttnpb.MACCommand{ttnpb.CID_DEV_STATUS.MACCommand()}
						up := MakeDataUplink(macVersion, true, false, DevAddr, ttnpb.FCtrl{}, 0x00, 0x00, 0x00, nil, nil, dr, drIdx, phy.UplinkChannels[chIdx-1].Frequency, chIdx-1)
						up.ReceivedAt = upRecvAt.Add(-time.Nanosecond)
						dev.MACState.RecentUplinks = appendRecentUplink(dev.MACState.RecentUplinks, up, recentUplinkCount)
						dev.RecentUplinks = appendRecentUplink(dev.RecentUplinks, up, recentUplinkCount)
						return dev
					}(),
					NbTrans: 2,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState.PendingRequests = nil
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				},
				{
					Name: MakeTestCaseName("Current session", "Class B", "LastFCntUp=0xfffe", "NbTrans=1", "FCnt reset"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACSettings: &ttnpb.MACSettings{
								ResetsFCnt: &pbtypes.BoolValue{Value: true},
							},
							MACState:       MakeDefaultEU868MACState(ttnpb.CLASS_B, macVersion, phyVersion),
							Session:        makeSession(0xfffe, 0x02),
							SupportsClassB: true,
							SupportsClassC: true,
						}
						dev.MACState.DesiredParameters = MakeDefaultUS915FSB2DesiredMACParameters(phyVersion)
						dev.Session.LastNFCntDown = 0x42
						return dev
					}(),
					FCntReset: true,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						class := ttnpb.CLASS_A
						if macVersion.Compare(ttnpb.MAC_V1_1) < 0 {
							class = ttnpb.CLASS_C
						}
						dev.MACState = MakeDefaultEU868MACState(class, macVersion, phyVersion)
						dev.MACState.RxWindowsAvailable = true
						dev.Session = makeSession(0x00, 0x02)
						dev.Session.LastNFCntDown = 0x42
						dev.Session.StartedAt = upRecvAt
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xfef0", "NbTrans=1", "FCnt reset", "Pending application downlink"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACSettings: &ttnpb.MACSettings{
								ResetsFCnt: &pbtypes.BoolValue{Value: true},
							},
							MACState: MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:  makeSession(0xfef0, 0x02),
						}
						dev.MACState.DesiredParameters = MakeDefaultUS915FSB2DesiredMACParameters(phyVersion)
						dev.MACState.PendingApplicationDownlink = makeApplicationDownlink(true, 0x02)
						dev.Session.LastNFCntDown = 0x42
						return dev
					}(),
					FCntReset: true,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState = MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion)
						dev.MACState.RxWindowsAvailable = true
						dev.Session = makeSession(0x00, 0x02)
						dev.Session.LastNFCntDown = 0x42
						dev.Session.StartedAt = upRecvAt
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
					QueuedApplicationUplinks: []*ttnpb.ApplicationUp{
						{
							CorrelationIDs:       DataUplinkCorrelationIDs[:],
							EndDeviceIdentifiers: makeIdentifiers(),
							Up: &ttnpb.ApplicationUp_DownlinkNack{
								DownlinkNack: makeApplicationDownlink(true, 0x02),
							},
						},
					},
				},
			},
			{
				FCnt: 0x22,
				FRMPayload: MakeUplinkMACBuffer(phy,
					ttnpb.CID_LINK_CHECK,
					ttnpb.CID_BEACON_TIMING,
					&ttnpb.MACCommand_PingSlotInfoReq{
						Period: ttnpb.PING_EVERY_2S,
					},
					ttnpb.CID_DEVICE_TIME,
				),
			}: {
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0", "NbTrans=1"),
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: makeIdentifiers(),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    phyVersion,
						LoRaWANVersion:       macVersion,
						MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
						Session:              makeSession(0, 0),
					},
					FCnt:    0x22,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0x22
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.QueuedResponses = AppendMACCommanders(dev.MACState.QueuedResponses,
							ttnpb.CID_PING_SLOT_INFO,
						)
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: func(deduplicated bool) []events.DefinitionDataClosure {
						if deduplicated {
							return []events.DefinitionDataClosure{
								evtReceiveLinkCheckRequest.BindData(nil),
								evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
									Period: ttnpb.PING_EVERY_2S,
								}),
								evtEnqueuePingSlotInfoAnswer.BindData(nil),
								evtReceiveDeviceTimeRequest.BindData(nil),
							}
						}
						return []events.DefinitionDataClosure{
							evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
								Period: ttnpb.PING_EVERY_2S,
							}),
							evtEnqueuePingSlotInfoAnswer.BindData(nil),
						}
					},
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0x30002", "NbTrans=1", "FCnt reset"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACSettings: &ttnpb.MACSettings{
								ResetsFCnt: &pbtypes.BoolValue{Value: true},
							},
							MACState: MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:  makeSession(0x30014, 0x02),
						}
						dev.MACState.DesiredParameters = MakeDefaultUS915FSB2DesiredMACParameters(phyVersion)
						dev.Session.LastNFCntDown = 0x42
						return dev
					}(),
					FCntReset: true,
					FCnt:      0x22,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState = MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion)
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.QueuedResponses = AppendMACCommanders(dev.MACState.QueuedResponses,
							ttnpb.CID_PING_SLOT_INFO,
						)
						dev.MACState.RxWindowsAvailable = true
						dev.Session = makeSession(0x22, 0x02)
						dev.Session.LastNFCntDown = 0x42
						dev.Session.StartedAt = upRecvAt
						return dev
					},
					MakeQueuedEvents: func(deduplicated bool) []events.DefinitionDataClosure {
						if deduplicated {
							return []events.DefinitionDataClosure{
								evtReceiveLinkCheckRequest.BindData(nil),
								evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
									Period: ttnpb.PING_EVERY_2S,
								}),
								evtEnqueuePingSlotInfoAnswer.BindData(nil),
								evtReceiveDeviceTimeRequest.BindData(nil),
							}
						}
						return []events.DefinitionDataClosure{
							evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
								Period: ttnpb.PING_EVERY_2S,
							}),
							evtEnqueuePingSlotInfoAnswer.BindData(nil),
						}
					},
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xff02", "NbTrans=1", "FCnt reset"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACSettings: &ttnpb.MACSettings{
								ResetsFCnt:        &pbtypes.BoolValue{Value: true},
								Supports32BitFCnt: &pbtypes.BoolValue{Value: false},
							},
							MACState: MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:  makeSession(0xff02, 0x02),
						}
						dev.MACState.DesiredParameters = MakeDefaultUS915FSB2DesiredMACParameters(phyVersion)
						dev.Session.LastNFCntDown = 0x42
						return dev
					}(),
					FCntReset: true,
					FCnt:      0x22,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState = MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion)
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.QueuedResponses = AppendMACCommanders(dev.MACState.QueuedResponses,
							ttnpb.CID_PING_SLOT_INFO,
						)
						dev.MACState.RxWindowsAvailable = true
						dev.Session = makeSession(0x22, 0x02)
						dev.Session.LastNFCntDown = 0x42
						dev.Session.StartedAt = upRecvAt
						return dev
					},
					MakeQueuedEvents: func(deduplicated bool) []events.DefinitionDataClosure {
						if deduplicated {
							return []events.DefinitionDataClosure{
								evtReceiveLinkCheckRequest.BindData(nil),
								evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
									Period: ttnpb.PING_EVERY_2S,
								}),
								evtEnqueuePingSlotInfoAnswer.BindData(nil),
								evtReceiveDeviceTimeRequest.BindData(nil),
							}
						}
						return []events.DefinitionDataClosure{
							evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
								Period: ttnpb.PING_EVERY_2S,
							}),
							evtEnqueuePingSlotInfoAnswer.BindData(nil),
						}
					},
				},
			},
			{
				FCnt:         0xff00,
				ConfFCntDown: 0x02,
				Ack:          true,
				FPort:        0x01,
				FRMPayload:   []byte("test-payload"),
				FOpts: MakeUplinkMACBuffer(phy,
					&ttnpb.MACCommand_PingSlotInfoReq{
						Period: ttnpb.PING_EVERY_2S,
					},
				),
			}: {
				{
					Name: MakeTestCaseName("Pending session", "Class A", "NbTrans=1"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							PendingMACState:      MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							PendingSession:       makeSession(0x00, 0x00),
						}
						dev.PendingMACState.PendingJoinRequest = MakeNsJsJoinRequest(macVersion, phyVersion, fp, &DevAddr, ttnpb.RX_DELAY_4, 1, ttnpb.DATA_RATE_3)
						return dev
					}(),
					Error: errDeviceNotFound,
				},
				ifNoFCntGap(devConfig{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0x01", "NbTrans=1"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:              makeSession(0x01, 0x02),
						}
						dev.MACState.RecentDownlinks = []*ttnpb.DownlinkMessage{{}}
						return dev
					}(),
					FCnt:    0xff00,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0xff00
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.QueuedResponses = AppendMACCommanders(dev.MACState.QueuedResponses,
							ttnpb.CID_PING_SLOT_INFO,
						)
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: func(bool) []events.DefinitionDataClosure {
						return []events.DefinitionDataClosure{
							evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
								Period: ttnpb.PING_EVERY_2S,
							}),
							evtEnqueuePingSlotInfoAnswer.BindData(nil),
						}
					},
				}),
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xfef0", "NbTrans=1", "Pending application downlink"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:              makeSession(0xfef0, 0x02),
						}
						dev.MACState.RecentDownlinks = []*ttnpb.DownlinkMessage{{}}
						dev.MACState.PendingApplicationDownlink = makeApplicationDownlink(true, 0x02)
						return dev
					}(),
					FCnt:    0xff00,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0xff00
						dev.MACState.PendingApplicationDownlink = nil
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.QueuedResponses = AppendMACCommanders(dev.MACState.QueuedResponses,
							ttnpb.CID_PING_SLOT_INFO,
						)
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					QueuedApplicationUplinks: []*ttnpb.ApplicationUp{
						{
							CorrelationIDs:       DataUplinkCorrelationIDs[:],
							EndDeviceIdentifiers: makeIdentifiers(),
							Up: &ttnpb.ApplicationUp_DownlinkAck{
								DownlinkAck: makeApplicationDownlink(true, 0x02),
							},
						},
					},
					MakeQueuedEvents: func(bool) []events.DefinitionDataClosure {
						return []events.DefinitionDataClosure{
							evtReceivePingSlotInfoRequest.BindData(&ttnpb.MACCommand_PingSlotInfoReq{
								Period: ttnpb.PING_EVERY_2S,
							}),
							evtEnqueuePingSlotInfoAnswer.BindData(nil),
						}
					},
				},
			},
			{
				FCnt:         0xff00,
				ADR:          true,
				ADRAckReq:    true,
				ConfFCntDown: 0x02,
				ClassB:       true,
				FPort:        0x01,
				FRMPayload:   []byte("test-payload"),
			}: {
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xfef0", "NbTrans=1", "Pending application downlink"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:              makeSession(0xfef0, 0x02),
						}
						dev.MACState.PendingRequests = []*ttnpb.MACCommand{{}, {}}
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.PendingApplicationDownlink = makeApplicationDownlink(true, 0x02)
						return dev
					}(),
					FCnt:    0xff00,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0xff00
						dev.MACState.PendingApplicationDownlink = nil
						dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
					QueuedApplicationUplinks: []*ttnpb.ApplicationUp{
						{
							CorrelationIDs:       DataUplinkCorrelationIDs[:],
							EndDeviceIdentifiers: makeIdentifiers(),
							Up: &ttnpb.ApplicationUp_DownlinkNack{
								DownlinkNack: makeApplicationDownlink(true, 0x02),
							},
						},
					},
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xfef0", "NbTrans=1"),
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: makeIdentifiers(),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    phyVersion,
						LoRaWANVersion:       macVersion,
						MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
						Session:              makeSession(0xfef0, 0x02),
					},
					FCnt:    0xff00,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0xff00
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xfef0", "NbTrans=1", "Supports class B"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
							Session:              makeSession(0xfef0, 0x02),
							SupportsClassB:       true,
						}
						dev.MACState.PendingRequests = []*ttnpb.MACCommand{{}, {}}
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						return dev
					}(),
					FCnt:    0xff00,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0xff00
						dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]
						dev.MACState.DeviceClass = ttnpb.CLASS_B
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: func(bool) []events.DefinitionDataClosure {
						return []events.DefinitionDataClosure{
							evtClassBSwitch.BindData(ttnpb.CLASS_A),
						}
					},
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xff00", "NbTrans=2", "Supports class B", "MaxNbTrans=3"),
					Device: func() *ttnpb.EndDevice {
						dev := &ttnpb.EndDevice{
							EndDeviceIdentifiers: makeIdentifiers(),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    phyVersion,
							LoRaWANVersion:       macVersion,
							MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_B, macVersion, phyVersion),
							Session:              makeSession(0xff00, 0x02),
							SupportsClassB:       true,
						}
						dev.MACState.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
							Value: ttnpb.PING_EVERY_2S,
						}
						dev.MACState.CurrentParameters.ADRNbTrans = 3
						dev.MACState.PendingRequests = []*ttnpb.MACCommand{{}, {}}
						up := MakeDataUplink(macVersion, true, false, DevAddr, ttnpb.FCtrl{
							ADR:       true,
							ADRAckReq: true,
							ClassB:    true,
						}, 0xff00, 0x02, 0x01, []byte("test-payload"), []byte{}, dr, drIdx, phy.UplinkChannels[chIdx-1].Frequency, chIdx-1)
						up.ReceivedAt = upRecvAt.Add(-time.Nanosecond)
						dev.MACState.RecentUplinks = appendRecentUplink(dev.MACState.RecentUplinks, up, recentUplinkCount)
						dev.RecentUplinks = appendRecentUplink(dev.RecentUplinks, up, recentUplinkCount)
						return dev
					}(),
					FCnt:    0xff00,
					NbTrans: 2,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.MACState.PendingRequests = nil
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				},
			},
			{
				FCnt:       0x10000,
				Confirmed:  true,
				ADR:        true,
				ADRAckReq:  true,
				FPort:      0x02,
				FRMPayload: []byte("test-payload"),
			}: {
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xffff", "NbTrans=1"),
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: makeIdentifiers(),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    phyVersion,
						LoRaWANVersion:       macVersion,
						MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
						Session:              makeSession(0xffff, 0x02),
					},
					FCnt:    0x10000,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0x10000
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				},
				{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0xffff", "NbTrans=1", "Does not support 32-bit FCnt"),
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: makeIdentifiers(),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    phyVersion,
						LoRaWANVersion:       macVersion,
						MACSettings: &ttnpb.MACSettings{
							Supports32BitFCnt: &pbtypes.BoolValue{Value: false},
						},
						MACState: MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
						Session:  makeSession(0xffff, 0x02),
					},
					Error: errDeviceNotFound,
				},
				ifNoFCntGap(devConfig{
					Name: MakeTestCaseName("Current session", "Class A", "LastFCntUp=0x01", "NbTrans=1"),
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: makeIdentifiers(),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    phyVersion,
						LoRaWANVersion:       macVersion,
						MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, macVersion, phyVersion),
						Session:              makeSession(0x01, 0x02),
					},
					FCnt:    0x10000,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					ApplyDeviceDiff: func(_ bool, dev *ttnpb.EndDevice) *ttnpb.EndDevice {
						dev.Session.LastFCntUp = 0x10000
						dev.MACState.RxWindowsAvailable = true
						return dev
					},
					MakeQueuedEvents: makeNilEvents,
				}),
			},
		} {
			upName := makeName(fmt.Sprintf("confirmed:%v,ack:%v,adr:%v,adr_ack_req:%v,class_b:%v,f_cnt:0x%x,conf_f_cnt_down:0x%x,f_port:%v,frm_payload:%v,fOpts:%v",
				upConf.Confirmed, upConf.Ack, upConf.ADR, upConf.ADRAckReq, upConf.ClassB, upConf.FCnt, upConf.ConfFCntDown, upConf.FPort, hex.EncodeToString(upConf.FRMPayload), hex.EncodeToString(upConf.FOpts)))
			up := MakeDataUplink(macVersion, true, upConf.Confirmed, DevAddr, ttnpb.FCtrl{
				Ack:       upConf.Ack,
				ADR:       upConf.ADR,
				ADRAckReq: upConf.ADRAckReq,
				ClassB:    upConf.ClassB,
			}, upConf.FCnt, upConf.ConfFCntDown, upConf.FPort, upConf.FRMPayload, upConf.FOpts, dr, drIdx, ch.Frequency, chIdx)
			up.ReceivedAt = upRecvAt
			for _, deduplicated := range [2]bool{
				true,
				false,
			} {
				deduplicated := deduplicated
				makeName := func(parts ...string) string {
					return MakeTestCaseName(append(append([]string{upName}, parts...), fmt.Sprintf("Deduplicated:%v", deduplicated))...)
				}
				tcs = append(tcs,
					TestCase{
						Name:         makeName("No devices"),
						Uplink:       CopyUplinkMessage(up),
						MakeDevices:  func(context.Context) []contextualEndDevice { return nil },
						Deduplicated: deduplicated,
						DeviceAssertion: func(t *testing.T, matched *matchedDevice) bool {
							return assertions.New(t).So(matched, should.BeNil)
						},
						Error: errDeviceNotFound,
					},
					TestCase{
						Name:         makeName("No matching device"),
						Uplink:       CopyUplinkMessage(up),
						Deduplicated: deduplicated,
						MakeDevices: func(ctx context.Context) []contextualEndDevice {
							return []contextualEndDevice{
								{
									Context: ctx,
									EndDevice: &ttnpb.EndDevice{
										EndDeviceIdentifiers: makeIdentifiers(),
										FrequencyPlanID:      test.EUFrequencyPlanID,
										LoRaWANVersion:       macVersion,
										LoRaWANPHYVersion:    phyVersion,
									},
								},
								{
									Context: ctx,
									EndDevice: &ttnpb.EndDevice{
										EndDeviceIdentifiers: makeIdentifiers(),
										FrequencyPlanID:      test.EUFrequencyPlanID,
										LoRaWANVersion:       macVersion,
										LoRaWANPHYVersion:    phyVersion,
										MACState:             &ttnpb.MACState{},
									},
								},
								{
									Context: ctx,
									EndDevice: &ttnpb.EndDevice{
										EndDeviceIdentifiers: makeIdentifiers(),
										FrequencyPlanID:      test.EUFrequencyPlanID,
										LoRaWANVersion:       macVersion,
										LoRaWANPHYVersion:    phyVersion,
										Session:              &ttnpb.Session{},
									},
								},
								{
									Context: ctx,
									EndDevice: &ttnpb.EndDevice{
										EndDeviceIdentifiers: makeIdentifiers(),
										FrequencyPlanID:      test.EUFrequencyPlanID,
										LoRaWANVersion:       macVersion,
										LoRaWANPHYVersion:    phyVersion,
										Session:              &ttnpb.Session{},
										PendingMACState:      &ttnpb.MACState{},
									},
								},
								{
									Context: ctx,
									EndDevice: &ttnpb.EndDevice{
										EndDeviceIdentifiers: makeIdentifiers(),
										FrequencyPlanID:      test.EUFrequencyPlanID,
										LoRaWANVersion:       macVersion,
										LoRaWANPHYVersion:    phyVersion,
										Session:              &ttnpb.Session{},
										MACState:             &ttnpb.MACState{},
										PendingSession:       &ttnpb.Session{},
										PendingMACState:      &ttnpb.MACState{},
									},
								},
							}
						},
						DeviceAssertion: func(t *testing.T, matched *matchedDevice) bool {
							return assertions.New(t).So(matched, should.BeNil)
						},
						Error: errDeviceNotFound,
					},
				)
				for _, devConf := range devConfs {
					devConf := devConf

					noSessionDev := CopyEndDevice(devConf.Device)
					noSessionDev.MACState = nil
					noSessionDev.PendingMACState = nil
					noSessionDev.Session = nil
					noSessionDev.PendingSession = nil

					invalidKeyDev := CopyEndDevice(devConf.Device)
					if invalidKeyDev.Session != nil {
						if invalidKeyDev.Session.SessionKeys.GetFNwkSIntKey().GetKey() != nil {
							invalidKeyDev.Session.SessionKeys.FNwkSIntKey.Key[0]++
						}
						if invalidKeyDev.Session.SessionKeys.GetFNwkSIntKey().GetEncryptedKey() != nil {
							invalidKeyDev.Session.SessionKeys.FNwkSIntKey.EncryptedKey[0]++
						}
					}

					matchDev := CopyEndDevice(devConf.Device)
					ctxKey := &struct{}{}
					ctxValue := &struct{}{}
					makeDeviceContext := func(ctx context.Context) context.Context {
						return context.WithValue(ctx, ctxKey, ctxValue)
					}
					devAssertion := func(t *testing.T, matched *matchedDevice) bool {
						t.Helper()
						a := assertions.New(t)
						if devConf.Error != nil {
							return a.So(matched, should.BeNil)
						}

						if !a.So(AllTrue(
							a.So(matched, should.NotBeNil),
							a.So(matched.Context.Value(ctxKey), should.Equal, ctxValue),
							a.So(matched.phy, should.HaveEmptyDiff, phy),
						), should.BeTrue) {
							return false
						}
						matched.Context = nil     // Comparing context with should.Resemble results in infinite recursion.
						matched.phy = band.Band{} // band.Band cannot be compared with neither should.Resemble, nor should.Equal.
						if !a.So(AllTrue(
							a.So(matched.SetPaths, should.HaveSameElementsDeep, devConf.SetPaths),
							a.So(matched.QueuedEventClosures, should.ResembleEventDefinitionDataClosures, devConf.MakeQueuedEvents(deduplicated)),
							a.So(matched, should.Resemble, &matchedDevice{
								ChannelIndex:             chIdx,
								DataRateIndex:            drIdx,
								DeferredMACHandlers:      matched.DeferredMACHandlers,
								Device:                   devConf.ApplyDeviceDiff(deduplicated, CopyEndDevice(devConf.Device)),
								FCnt:                     devConf.FCnt,
								FCntReset:                devConf.FCntReset,
								NbTrans:                  devConf.NbTrans,
								Pending:                  devConf.Pending,
								QueuedApplicationUplinks: devConf.QueuedApplicationUplinks,
								QueuedEventClosures:      matched.QueuedEventClosures,
								SetPaths:                 matched.SetPaths,
							}),
						), should.BeTrue) {
							return false
						}

						if deduplicated || len(matched.DeferredMACHandlers) == 0 {
							return true
						}
						queuedEvents := matched.QueuedEventClosures
						for _, f := range matched.DeferredMACHandlers {
							evs, err := f(matched.Context, matched.Device, CopyUplinkMessage(up))
							if !a.So(err, should.BeNil) {
								return false
							}
							queuedEvents = append(queuedEvents, evs...)
						}
						return AllTrue(
							a.So(queuedEvents, should.HaveSameElementsFunc, test.EventDefinitionDataClosureEqual, devConf.MakeQueuedEvents(true)),
							a.So(matched.Device, should.Resemble, devConf.ApplyDeviceDiff(true, CopyEndDevice(devConf.Device))),
						)
					}
					tcs = append(tcs,
						TestCase{
							Name:         makeName("Only device", devConf.Name),
							Uplink:       CopyUplinkMessage(up),
							Deduplicated: deduplicated,
							MakeDevices: func(ctx context.Context) []contextualEndDevice {
								return []contextualEndDevice{
									{
										Context:   makeDeviceContext(ctx),
										EndDevice: CopyEndDevice(matchDev),
									},
								}
							},
							DeviceAssertion: devAssertion,
							Error:           devConf.Error,
						},
						TestCase{
							Name:         makeName("Multiple devices", devConf.Name),
							Uplink:       CopyUplinkMessage(up),
							Deduplicated: deduplicated,
							MakeDevices: func(ctx context.Context) []contextualEndDevice {
								return []contextualEndDevice{
									{
										Context:   ctx,
										EndDevice: CopyEndDevice(noSessionDev),
									},
									{
										Context:   makeDeviceContext(ctx),
										EndDevice: CopyEndDevice(matchDev),
									},
									{
										Context:   ctx,
										EndDevice: CopyEndDevice(invalidKeyDev),
									},
								}
							},
							DeviceAssertion: devAssertion,
							Error:           devConf.Error,
						},
					)
				}
			}
		}
	})
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns, ctx, env, stop := StartTest(t, component.Config{}, Config{
				NetID: NetID,
			}, (1<<5)*test.Delay)
			defer stop()

			<-env.DownlinkTasks.Pop

			dev, err := ns.matchAndHandleDataUplink(CopyUplinkMessage(tc.Uplink), tc.Deduplicated, tc.MakeDevices(ctx)...)
			if a.So(err, should.EqualErrorOrDefinition, tc.Error) {
				a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
			}
		})
	}
}
