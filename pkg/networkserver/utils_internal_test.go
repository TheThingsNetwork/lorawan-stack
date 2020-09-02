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
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
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
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				pb := CopyEndDevice(tc.Device)

				macState, err := newMACState(pb, tc.FrequencyPlanStore, ttnpb.MACSettings{})
				if tc.ErrorAssertion != nil {
					a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				} else {
					a.So(err, should.BeNil)
				}
				a.So(macState, should.Resemble, tc.MACState)
				a.So(pb, should.Resemble, tc.Device)
			},
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
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Time.String(),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				a.So(beaconTimeBefore(tc.Time), should.Equal, tc.Expected)
			},
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
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ret, ok := nextPingSlotAt(ctx, tc.Device, tc.EarliestAt)
				if !a.So(ok, should.Equal, tc.ExpectedOk) {
					t.FailNow()
				}
				a.So(ret, should.Resemble, tc.ExpectedTime)
				if ok {
					earliestAt := ret
					ret, ok = nextPingSlotAt(ctx, tc.Device, earliestAt)
					a.So(ok, should.BeTrue)
					a.So(ret, should.Resemble, earliestAt)
				}
			},
		})
	}
}

func TestNextDataDownlinkSlot(t *testing.T) {
	nextPingSlotAt := func(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) time.Time {
		pingSlotAt, ok := nextPingSlotAt(ctx, dev, earliestAt)
		if !ok {
			panic(fmt.Sprintf("failed to compute next ping slot starting from %v", earliestAt))
		}
		return pingSlotAt
	}

	beaconTime := gpstime.Parse(beaconPeriod)

	clock := test.NewMockClock(beaconTime.Add(time.Millisecond))
	defer SetMockClock(clock)()

	up := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHDR: ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
			},
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{},
			},
		},
		RxMetadata: RxMetadata[:],
		ReceivedAt: beaconTime,
	}
	ups := []*ttnpb.UplinkMessage{up}

	rxDelay := ttnpb.RX_DELAY_4
	rx1 := up.ReceivedAt.Add(rxDelay.Duration())
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
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class A/RX1,RX2 available/application downlink",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
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
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DesiredParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity:  DurationPtr(0),
					StatusCountPeriodicity: &pbtypes.UInt32Value{Value: 0},
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
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
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_A,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class A/MAC diff/RX windows closed",
			EarliestAt: afterRX2,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_A,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
		},
		{
			Name:       "unicast/class B/MAC diff/RX1,RX2 available",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PING_EVERY_2S,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_B,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class B/MAC diff/RX windows closed/no application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PING_EVERY_2S,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_B,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
		},
		func() TestCase {
			dev := &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{
						Value: ttnpb.PING_EVERY_2S,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_B,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
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
					Class: ttnpb.CLASS_B,
				},
				ExpectedOk: true,
			}
		}(),
		{
			Name:       "unicast/class C/RX1,RX2 available",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion:     ttnpb.MAC_V1_0_3,
					DeviceClass:        ttnpb.CLASS_C,
					RxWindowsAvailable: true,
					RecentUplinks:      ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
			ExpectedSlot: classA,
			ExpectedOk:   true,
		},
		{
			Name:       "unicast/class C/RX windows closed",
			EarliestAt: beforeRX1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_C,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
		},
		{
			Name:       "unicast/class C/RX windows closed/application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_C,
					RecentUplinks:  ups,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Time:  rx2,
				Class: ttnpb.CLASS_C,
			},
			ExpectedOk: true,
		},
		{
			Name:       "unicast/class C/no uplink/no application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_C,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
				},
			},
		},
		{
			Name:       "unicast/class C/no uplink/application downlink",
			EarliestAt: rx1,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_C,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								Gateways: []ttnpb.GatewayAntennaIdentifiers{
									{
										GatewayIdentifiers: ttnpb.GatewayIdentifiers{
											GatewayID: "test-gtw",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Class: ttnpb.CLASS_C,
			},
			ExpectedOk: true,
		},
		{
			Name:       "unicast/class C/no uplink/absolute-time application downlink",
			EarliestAt: absTime,
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DeviceClass:    ttnpb.CLASS_C,
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: &absTime,
								Gateways: []ttnpb.GatewayAntennaIdentifiers{
									{
										GatewayIdentifiers: ttnpb.GatewayIdentifiers{
											GatewayID: "test-gtw",
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
				Class:             ttnpb.CLASS_C,
				IsApplicationTime: true,
			},
			ExpectedOk: true,
		},
		{
			Name:       "unicast/class C/no uplink/expired absolute-time application downlink",
			EarliestAt: absTime.Add(time.Nanosecond),
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
					DeviceClass:    ttnpb.CLASS_C,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: &absTime,
								Gateways: []ttnpb.GatewayAntennaIdentifiers{
									{
										GatewayIdentifiers: ttnpb.GatewayIdentifiers{
											GatewayID: "test-gtw",
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
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DeviceClass:    ttnpb.CLASS_C,
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: &absTime,
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
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DeviceClass:    ttnpb.CLASS_C,
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								AbsoluteTime: &absTime,
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
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: rxDelay,
					},
					DeviceClass:    ttnpb.CLASS_C,
					LoRaWANVersion: ttnpb.MAC_V1_0_3,
				},
				Session: &ttnpb.Session{
					DevAddr: DevAddr,
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							ClassBC: &ttnpb.ApplicationDownlink_ClassBC{
								Gateways: []ttnpb.GatewayAntennaIdentifiers{
									{
										GatewayIdentifiers: ttnpb.GatewayIdentifiers{
											GatewayID: "test-gtw",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedSlot: &networkInitiatedDownlinkSlot{
				Class: ttnpb.CLASS_C,
			},
			ExpectedOk: true,
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				ret, ok := nextDataDownlinkSlot(ctx, tc.Device, LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_1_REV_B], ttnpb.MACSettings{}, tc.EarliestAt)
				if a.So(ok, should.Equal, tc.ExpectedOk) {
					a.So(ret, should.Resemble, tc.ExpectedSlot)
				}
			},
		})
	}
}
