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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gpstime"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNewMACState(t *testing.T) {
	for _, tc := range []struct {
		Name               string
		Device             *ttnpb.EndDevice
		MACState           *ttnpb.MACState
		FrequencyPlanStore *frequencyplans.Store
		ErrorAssertion     func(*testing.T, error) bool
	}{
		{
			Name: "1.0.2/EU868",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_0_2,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2, ttnpb.PHY_V1_0_2_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.0.2/EU868/multicast/class A",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_0_2,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
				Multicast: true,
			},
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errClassAMulticast)
			},
		},
		{
			Name: "1.0.2/EU868/multicast/class B",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_0_2,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
				Multicast:      true,
				SupportsClassB: true,
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2, ttnpb.PHY_V1_0_2_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				macState.DeviceClass = ttnpb.CLASS_B
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.0.2/EU868/multicast/class C",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_0_2,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
				Multicast:      true,
				SupportsClassC: true,
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2, ttnpb.PHY_V1_0_2_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				macState.DeviceClass = ttnpb.CLASS_C
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.1/EU868",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.1/EU868/multicast/class A",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
				Multicast: true,
			},
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errClassAMulticast)
			},
		},
		{
			Name: "1.1/EU868/multicast/class B",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
				Multicast:      true,
				SupportsClassB: true,
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				macState.DeviceClass = ttnpb.CLASS_B
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.1/EU868/multicast/class C",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
				Multicast:      true,
				SupportsClassC: true,
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				macState.DeviceClass = ttnpb.CLASS_C
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.0.2/US915_FSB2",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_0_2,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2, ttnpb.PHY_V1_0_2_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.1/US915_FSB2",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
			},
			MACState: func() *ttnpb.MACState {
				macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B)
				macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_13
				return macState
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			pb := CopyEndDevice(tc.Device)

			macState, err := newMACState(pb, tc.FrequencyPlanStore, ttnpb.MACSettings{})
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(macState, should.Resemble, tc.MACState)
			a.So(pb, should.Resemble, tc.Device)
		})
	}
}

func TestBeaconTimeBefore(t *testing.T) {
	for _, tc := range []struct {
		Time     time.Time
		Expected time.Duration
	}{
		{
			Time:     gpstime.Parse(0),
			Expected: 0,
		},
		{
			Time:     gpstime.Parse(time.Nanosecond),
			Expected: 0,
		},
		{
			Time:     gpstime.Parse(beaconPeriod - time.Second),
			Expected: 0,
		},
		{
			Time:     gpstime.Parse(beaconPeriod),
			Expected: beaconPeriod,
		},
		{
			Time:     gpstime.Parse(beaconPeriod + time.Second),
			Expected: beaconPeriod,
		},
		{
			Time:     gpstime.Parse(2*beaconPeriod - time.Second),
			Expected: beaconPeriod,
		},
		{
			Time:     gpstime.Parse(10 * beaconPeriod),
			Expected: 10 * beaconPeriod,
		},
		{
			Time:     gpstime.Parse(10*beaconPeriod + time.Second),
			Expected: 10 * beaconPeriod,
		},
	} {
		t.Run(tc.Time.String(), func(t *testing.T) {
			a := assertions.New(t)
			ret := beaconTimeBefore(tc.Time)
			a.So(ret, should.Equal, tc.Expected)
		})
	}
}

func computePingOffset(beaconTime uint32, devAddr types.DevAddr, pingPeriod uint16) uint16 {
	return test.Must(crypto.ComputePingOffset(beaconTime, devAddr, pingPeriod)).(uint16)
}

func TestNextPingSlotAt(t *testing.T) {
	const beaconTime = 10000 * beaconPeriod
	beaconAt := gpstime.Parse(beaconTime)
	devAddr := types.DevAddr{0x01, 0x34, 0x07, 0x29}

	pingSlotTime := func(pingPeriod uint16, n uint16) time.Time {
		return beaconAt.Add(tBeaconDelay + beaconReserved + time.Duration(computePingOffset(uint32(beaconTime/time.Second), devAddr, pingPeriod)+n*pingPeriod)*pingSlotLen)
	}

	for _, tc := range []struct {
		Name         string
		Device       *ttnpb.EndDevice
		EarliestAt   time.Time
		ExpectedTime time.Time
		ExpectedOk   bool
	}{
		{
			Name:   "no MAC state/no session/no devAddr",
			Device: &ttnpb.EndDevice{},
		},
		{
			Name: "no session/no devAddr",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
		},
		{
			Name: "no devAddr",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
				Session:  &ttnpb.Session{},
			},
		},
		{
			Name:       "earliestAt:beaconAt;periodicity:0",
			EarliestAt: beaconAt,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_1S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<5, 0),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt;periodicity:1",
			EarliestAt: beaconAt,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_2S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<6, 0),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt;periodicity:2",
			EarliestAt: beaconAt,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_4S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<7, 0),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt;periodicity:3",
			EarliestAt: beaconAt,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_8S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<8, 0),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt;periodicity:4",
			EarliestAt: beaconAt,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_16S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<9, 0),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt+12s120ms15ns;periodicity:3",
			EarliestAt: beaconAt.Add(12*time.Second + 120*time.Millisecond + 1*time.Microsecond + 500*time.Nanosecond),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_8S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<8, 1),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt+20s;periodicity:4",
			EarliestAt: beaconAt.Add(20 * time.Second),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_16S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<9, 1),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt+38s;periodicity:4",
			EarliestAt: beaconAt.Add(38 * time.Second),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_16S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<9, 2),
			ExpectedOk:   true,
		},
		{
			Name:       "earliestAt:beaconAt+50s;periodicity:4",
			EarliestAt: beaconAt.Add(50 * time.Second),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: ttnpb.PING_EVERY_16S},
				},
				Session: &ttnpb.Session{
					DevAddr: devAddr,
				},
			},
			ExpectedTime: pingSlotTime(1<<9, 3),
			ExpectedOk:   true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			ctx = log.NewContext(ctx, test.GetLogger(t))

			ret, ok := nextPingSlotAt(ctx, tc.Device, tc.EarliestAt)
			if a.So(ok, should.Equal, tc.ExpectedOk) {
				a.So(ret, should.Resemble, tc.ExpectedTime)
			}
			if ok {
				earliestAt := ret
				ret, ok = nextPingSlotAt(ctx, tc.Device, earliestAt)
				a.So(ok, should.BeTrue)
				a.So(ret, should.Resemble, earliestAt)
			}
		})
	}
}

func TestNextDataDownlinkAt(t *testing.T) {
	nextPingSlotAt := func(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) time.Time {
		pingSlotAt, ok := nextPingSlotAt(ctx, dev, earliestAt)
		if !ok {
			panic(fmt.Sprintf("failed to compute next ping slot starting from %v", earliestAt))
		}
		return pingSlotAt
	}

	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	type TestCase struct {
		Name          string
		Device        *ttnpb.EndDevice
		EarliestAt    time.Time
		ExpectedTime  time.Time
		ExpectedClass ttnpb.Class
		ExpectedOk    bool
	}
	for _, tc := range []TestCase{
		{
			Name:   "no MAC state",
			Device: &ttnpb.EndDevice{},
		},
		{
			Name:       "unicast/class A/Rx1,Rx2 available",
			EarliestAt: time.Unix(42, 0),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: ttnpb.RX_DELAY_4,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							Payload: &ttnpb.Message{
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{},
								},
							},
							ReceivedAt: time.Unix(41, 0),
						},
					},
				},
				Session: &ttnpb.Session{
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
			},
			ExpectedTime:  time.Unix(41+4, 0).Add(-infrastructureDelay / 2),
			ExpectedClass: ttnpb.CLASS_A,
			ExpectedOk:    true,
		},
		{
			Name:       "unicast/class A/Rx windows closed",
			EarliestAt: time.Unix(42, 0),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: ttnpb.RX_DELAY_4,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_A,
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							Payload: &ttnpb.Message{
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{},
								},
							},
							ReceivedAt: time.Unix(41, 0),
						},
					},
				},
				Session: &ttnpb.Session{
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
			},
		},
		{
			Name:       "unicast/class B/Rx1,Rx2 available",
			EarliestAt: gpstime.Parse(beaconPeriod + time.Second),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: ttnpb.RX_DELAY_4,
					},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PING_EVERY_2S,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_B,
					RxWindowsAvailable: true,
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							Payload: &ttnpb.Message{
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{},
								},
							},
							ReceivedAt: gpstime.Parse(beaconPeriod),
						},
					},
				},
				Session: &ttnpb.Session{
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
			},
			ExpectedTime:  gpstime.Parse(beaconPeriod + 4*time.Second).Add(-infrastructureDelay / 2),
			ExpectedClass: ttnpb.CLASS_A,
			ExpectedOk:    true,
		},
		func() TestCase {
			dev := &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: ttnpb.RX_DELAY_4,
					},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PING_EVERY_2S,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_B,
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							Payload: &ttnpb.Message{
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{},
								},
							},
							ReceivedAt: gpstime.Parse(beaconPeriod),
						},
					},
				},
				Session: &ttnpb.Session{
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
			}
			earliestAt := gpstime.Parse(beaconPeriod + time.Second)
			return TestCase{
				Name:          "unicast/class B/Rx windows closed",
				Device:        dev,
				EarliestAt:    earliestAt,
				ExpectedTime:  nextPingSlotAt(ctx, dev, earliestAt),
				ExpectedClass: ttnpb.CLASS_B,
				ExpectedOk:    true,
			}
		}(),
		{
			Name:       "unicast/class C/Rx1,Rx2 available",
			EarliestAt: time.Unix(42, 0),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: ttnpb.RX_DELAY_4,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_C,
					RxWindowsAvailable: true,
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							Payload: &ttnpb.Message{
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{},
								},
							},
							ReceivedAt: time.Unix(41, 0),
						},
					},
				},
				Session: &ttnpb.Session{
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
			},
			ExpectedTime:  time.Unix(41+4, 0).Add(-infrastructureDelay / 2),
			ExpectedClass: ttnpb.CLASS_A,
			ExpectedOk:    true,
		},
		{
			Name:       "unicast/class C/Rx windows closed",
			EarliestAt: time.Unix(42, 0),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: ttnpb.RX_DELAY_4,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_C,
					RecentUplinks: []*ttnpb.UplinkMessage{
						{
							Payload: &ttnpb.Message{
								Payload: &ttnpb.Message_MACPayload{
									MACPayload: &ttnpb.MACPayload{},
								},
							},
							ReceivedAt: time.Unix(41, 0),
						},
					},
				},
				Session: &ttnpb.Session{
					DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
				},
			},
			ExpectedTime:  time.Unix(42, 0),
			ExpectedClass: ttnpb.CLASS_C,
			ExpectedOk:    true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := log.NewContext(ctx, test.GetLogger(t))
			ret, class, ok := nextDataDownlinkAt(ctx, tc.Device, test.Must(band.GetByID(band.EU_863_870)).(band.Band), ttnpb.MACSettings{}, tc.EarliestAt)
			if a.So(ok, should.Equal, tc.ExpectedOk) {
				a.So(class, should.Resemble, tc.ExpectedClass)
				a.So(ret, should.Resemble, tc.ExpectedTime)
			}
		})
	}
}
