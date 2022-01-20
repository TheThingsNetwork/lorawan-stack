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

package mac_test

import (
	"context"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsDevStatusReq(t *testing.T) {
	scheduleAt := time.Unix(424242, 42)

	for _, tc := range []struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Defaults    ttnpb.MACSettings
		Needs       bool
	}{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
		{
			Name: "device-settings(count-periodicity:5,time-periodicity:nil),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt,last-status-fcnt:1,last-fcnt:5",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt),
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{
						Value: 5,
					},
				},
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 5,
				},
			},
		},
		{
			Name: "device-settings(count-periodicity:5,time-periodicity:nil),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt,last-status-fcnt:1,last-fcnt:6",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt),
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{
						Value: 5,
					},
				},
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 6,
				},
			},
			Needs: true,
		},
		{
			Name: "device-settings(count-periodicity:1000,time-periodicity:1hr),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt,last-status-fcnt:1,last-fcnt:2",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt),
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{
						Value: 1000,
					},
					StatusTimePeriodicity: ttnpb.ProtoDurationPtr(time.Hour),
				},
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 2,
				},
			},
		},
		{
			Name: "device-settings(count-periodicity:1000,time-periodicity:1hr),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt-1hr+1ns,last-status-fcnt:1,last-fcnt:2",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt.Add(-time.Hour + time.Nanosecond)),
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{
						Value: 1000,
					},
					StatusTimePeriodicity: ttnpb.ProtoDurationPtr(time.Hour),
				},
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 2,
				},
			},
		},
		{
			Name: "device-settings(count-periodicity:1000,time-periodicity:1hr),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt-1hr,last-status-fcnt:1,last-fcnt:2",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt.Add(-time.Hour)),
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{
						Value: 1000,
					},
					StatusTimePeriodicity: ttnpb.ProtoDurationPtr(time.Hour),
				},
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 2,
				},
			},
			Needs: true,
		},
		{
			Name: "device-settings(count-periodicity:1000,time-periodicity:1hr),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt-1hr1ns,last-status-fcnt:1,last-fcnt:2",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt.Add(-time.Hour - time.Nanosecond)),
				MacSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: &pbtypes.UInt32Value{
						Value: 1000,
					},
					StatusTimePeriodicity: ttnpb.ProtoDurationPtr(time.Hour),
				},
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 2,
				},
			},
			Needs: true,
		},
		{
			Name: "device-settings(nil),ns-settings(count-periodicity:nil,time-periodicity:nil),last-status-at:scheduleAt,last-status-fcnt:1,last-fcnt:1000",
			InputDevice: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(scheduleAt),
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 1,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 1000,
				},
			},
			Needs: 1000-1 >= DefaultStatusCountPeriodicity,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.InputDevice)
				defaults := deepcopy.Copy(tc.Defaults).(ttnpb.MACSettings)
				res := DeviceNeedsDevStatusReq(dev, tc.Defaults, scheduleAt)
				if tc.Needs {
					a.So(res, should.BeTrue)
				} else {
					a.So(res, should.BeFalse)
				}
				a.So(dev, should.Resemble, tc.InputDevice)
				a.So(defaults, should.Resemble, tc.Defaults)
			},
		})
	}
}

func TestHandleDevStatusAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DevStatusAns
		FCntUp           uint32
		ReceivedAt       time.Time
		Events           events.Builders
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			ReceivedAt: time.Unix(42, 0),
			Error:      ErrNoPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 42,
				Margin:  4,
			},
			ReceivedAt: time.Unix(42, 0),
			Events: events.Builders{
				EvtReceiveDevStatusAnswer.With(events.WithData(&ttnpb.MACCommand_DevStatusAns{
					Battery: 42,
					Margin:  4,
				})),
			},
			Error: ErrRequestNotFound,
		},
		{
			Name: "battery 42%/margin 4",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 2,
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.MACCommandIdentifier_CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: &pbtypes.FloatValue{Value: 0.44},
				PowerState:        ttnpb.PowerState_POWER_EXTERNAL,
			},
			Expected: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(time.Unix(42, 0)),
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 43,
					PendingRequests:     []*ttnpb.MACCommand{},
				},
				BatteryPercentage: &pbtypes.FloatValue{Value: float32(42-1) / float32(253)},
				DownlinkMargin:    4,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 42,
				Margin:  4,
			},
			FCntUp:     43,
			ReceivedAt: time.Unix(42, 0),
			Events: events.Builders{
				EvtReceiveDevStatusAnswer.With(events.WithData(&ttnpb.MACCommand_DevStatusAns{
					Battery: 42,
					Margin:  4,
				})),
			},
		},
		{
			Name: "external power/margin 20",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 2,
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.MACCommandIdentifier_CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: &pbtypes.FloatValue{Value: 0.44},
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Expected: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(time.Unix(42, 0)),
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 43,
					PendingRequests:     []*ttnpb.MACCommand{},
				},
				DownlinkMargin: 20,
				PowerState:     ttnpb.PowerState_POWER_EXTERNAL,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 0,
				Margin:  20,
			},
			FCntUp:     43,
			ReceivedAt: time.Unix(42, 0),
			Events: events.Builders{
				EvtReceiveDevStatusAnswer.With(events.WithData(&ttnpb.MACCommand_DevStatusAns{
					Margin: 20,
				})),
			},
		},
		{
			Name: "nil power/margin -5",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 2,
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.MACCommandIdentifier_CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: &pbtypes.FloatValue{Value: 0.44},
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Expected: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: ttnpb.ProtoTimePtr(time.Unix(42, 0)),
				MacState: &ttnpb.MACState{
					LastDevStatusFCntUp: 43,
					PendingRequests:     []*ttnpb.MACCommand{},
				},
				DownlinkMargin: -5,
				PowerState:     ttnpb.PowerState_POWER_UNKNOWN,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 255,
				Margin:  -5,
			},
			FCntUp:     43,
			ReceivedAt: time.Unix(42, 0),
			Events: events.Builders{
				EvtReceiveDevStatusAnswer.With(events.WithData(&ttnpb.MACCommand_DevStatusAns{
					Battery: 255,
					Margin:  -5,
				})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)

				evs, err := HandleDevStatusAns(ctx, dev, tc.Payload, tc.FCntUp, tc.ReceivedAt)
				if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
					tc.Error == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(dev, should.Resemble, tc.Expected)
				a.So(evs, should.ResembleEventBuilders, tc.Events)
			},
		})
	}
}
