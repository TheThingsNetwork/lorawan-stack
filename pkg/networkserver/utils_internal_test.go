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
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNextDataDownlinkSlot(t *testing.T) {
	nextPingSlotAt := func(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) time.Time {
		pingSlotAt, ok := mac.NextPingSlotAt(ctx, dev, earliestAt)
		if !ok {
			panic(fmt.Sprintf("failed to compute next ping slot starting from %v", earliestAt))
		}
		return pingSlotAt
	}

	beaconTime := gpstime.Parse(mac.BeaconPeriod)

	up := &ttnpb.MACState_UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
			},
			Payload: &ttnpb.Message_MacPayload{
				MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						FCtrl: &ttnpb.FCtrl{},
					},
				},
			},
		},
		Settings:   ToMACStateTxSettings(DefaultTxSettings),
		RxMetadata: ToMACStateRxMetadata(DefaultRxMetadata[:]),
		ReceivedAt: ttnpb.ProtoTimePtr(beaconTime),
	}
	ups := []*ttnpb.MACState_UplinkMessage{up}

	rxDelay := ttnpb.RxDelay_RX_DELAY_4
	rx1 := ttnpb.StdTime(up.ReceivedAt).Add(rxDelay.Duration())
	rx2 := rx1.Add(time.Second)

	beforeRX1 := rx1.Add(-time.Millisecond)
	afterRX2 := rx2.Add(time.Microsecond)

	classA := &classADownlinkSlot{
		RxDelay: rxDelay.Duration(),
		Uplink:  up,
	}

	absTime := beaconTime.Add(time.Hour)

	type TestCase struct {
		Name         string
		Device       *ttnpb.EndDevice
		EarliestAt   time.Time
		ExpectedSlot downlinkSlot
		ExpectedOk   bool
	}
	for _, tc := range []TestCase{
		{
			Name:   "no MAC state",
			Device: &ttnpb.EndDevice{},
		},
		{
			Name:       "unicast/class A/MAC diff/RX1,RX2 available",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters:  &ttnpb.MACParameters{},
					LorawanVersion:     ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:        ttnpb.Class_CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class A/RX1,RX2 available/application downlink",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LorawanVersion:     ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:        ttnpb.Class_CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{},
					},
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class A/RX1,RX2 available/class BC application downlink",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LorawanVersion:     ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:        ttnpb.Class_CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				MacSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity:  ttnpb.ProtoDurationPtr(0),
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{},
						},
					},
				},
			},
		},
		{
			Name:       "unicast/class A/MAC diff/RX2 available",
			EarliestAt: rx2,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters:  &ttnpb.MACParameters{},
					LorawanVersion:     ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:        ttnpb.Class_CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class A/MAC diff/RX windows closed",
			EarliestAt: afterRX2,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:       ttnpb.Class_CLASS_A,
					RecentUplinks:     ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
		},
		{
			Name:       "unicast/class B/MAC diff/RX1,RX2 available",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PingSlotPeriod_PING_EVERY_2S,
					},
					LorawanVersion:     ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:        ttnpb.Class_CLASS_B,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class B/MAC diff/RX windows closed/no application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PingSlotPeriod_PING_EVERY_2S,
					},
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:    ttnpb.Class_CLASS_B,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
		},
		func() TestCase {
			dev := &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PingSlotPeriod_PING_EVERY_2S,
					},
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:    ttnpb.Class_CLASS_B,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{},
					},
				},
			}
			return TestCase{
				Name:       "unicast/class B/MAC diff/RX windows closed/application downlink",
				EarliestAt: beforeRX1,
				Device:     dev,
				ExpectedSlot: &networkInitiatedDownlinkSlot{
					Time:  nextPingSlotAt(log.NewContext(test.Context(), test.GetLogger(t)), dev, rx2),
					Class: ttnpb.Class_CLASS_B,
				},
				ExpectedOk: true,
			}
		}(),
		{
			Name:       "unicast/class C/RX1,RX2 available",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters:  &ttnpb.MACParameters{},
					LorawanVersion:     ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:        ttnpb.Class_CLASS_C,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class C/RX windows closed",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:       ttnpb.Class_CLASS_C,
					RecentUplinks:     ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
		},
		{
			Name:       "unicast/class C/RX windows closed/application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:       ttnpb.Class_CLASS_C,
					RecentUplinks:     ups,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Time:  rx2,
				Class: ttnpb.Class_CLASS_C,
			},
			ExpectedOk: true,
		},
		{
			Name:       "unicast/class C/no uplink/no application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:       ttnpb.Class_CLASS_C,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
				},
			},
		},
		{
			Name:       "unicast/class C/no uplink/application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:       ttnpb.Class_CLASS_C,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								Gateways: []*ttnpb.ClassBCGatewayIdentifiers{
									{
										GatewayIds: &ttnpb.GatewayIdentifiers{
											GatewayId: "test-gtw",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Class: ttnpb.Class_CLASS_C,
			},
			ExpectedOk: true,
		},
		{
			Name:       "unicast/class C/no uplink/absolute-time application downlink",
			EarliestAt: absTime,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					DeviceClass:       ttnpb.Class_CLASS_C,
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: ttnpb.ProtoTimePtr(absTime),
								Gateways: []*ttnpb.ClassBCGatewayIdentifiers{
									{
										GatewayIds: &ttnpb.GatewayIdentifiers{
											GatewayId: "test-gtw",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Time:              absTime,
				Class:             ttnpb.Class_CLASS_C,
				IsApplicationTime: true,
			},
			ExpectedOk: true,
		},
		{
			Name:       "unicast/class C/no uplink/expired absolute-time application downlink",
			EarliestAt: absTime.Add(time.Nanosecond),
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
					DeviceClass:       ttnpb.Class_CLASS_C,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: ttnpb.ProtoTimePtr(absTime),
								Gateways: []*ttnpb.ClassBCGatewayIdentifiers{
									{
										GatewayIds: &ttnpb.GatewayIdentifiers{
											GatewayId: "test-gtw",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:       "unicast/class C/no uplink/absolute-time application downlink/no paths",
			EarliestAt: absTime,
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					DeviceClass:       ttnpb.Class_CLASS_C,
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: ttnpb.ProtoTimePtr(absTime),
							},
						},
					},
				},
			},
		},
		{
			Name:       "multicast/class C/no uplink/absolute-time application downlink/no paths",
			EarliestAt: absTime,
			Device: &ttnpb.EndDevice{
				Multicast: true,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					DeviceClass:       ttnpb.Class_CLASS_C,
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: ttnpb.ProtoTimePtr(absTime),
							},
						},
					},
				},
			},
		},
		{
			Name: "multicast/class C/no uplink/application downlink with forced gateways",
			Device: &ttnpb.EndDevice{
				Multicast: true,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					DeviceClass:       ttnpb.Class_CLASS_C,
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: test.DefaultDevAddr.Bytes(),
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								Gateways: []*ttnpb.ClassBCGatewayIdentifiers{
									{
										GatewayIds: &ttnpb.GatewayIdentifiers{
											GatewayId: "test-gtw",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Class: ttnpb.Class_CLASS_C,
			},
			ExpectedOk: true,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				clock := test.NewMockClock(beaconTime.Add(time.Millisecond))
				defer SetMockClock(clock)()

				ret, ok := nextDataDownlinkSlot(ctx, tc.Device, LoRaWANBands[band.EU_863_870][ttnpb.PHYVersion_RP001_V1_1_REV_B], &ttnpb.MACSettings{}, tc.EarliestAt)
				if a.So(ok, should.Equal, tc.ExpectedOk) {
					a.So(ret, should.Resemble, tc.ExpectedSlot)
				}
			},
		})
	}
}
