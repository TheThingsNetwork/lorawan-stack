// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleDevStatusAns(t *testing.T) {
	events := collectEvents("ns.mac.device_status")

	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DevStatusAns
		ReceivedAt       time.Time
		Error            error
		ExpectedEvents   int
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload:    nil,
			ReceivedAt: time.Unix(42, 0),
			Error:      errNoPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload:    ttnpb.NewPopulatedMACCommand_DevStatusAns(test.Randy, false),
			ReceivedAt: time.Unix(42, 0),
			Error:      errMACRequestNotFound,
		},
		{
			Name: "battery 42/margin 4",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: 44,
				PowerState:        ttnpb.PowerState_POWER_EXTERNAL,
			},
			Expected: &ttnpb.EndDevice{
				LastStatusReceivedAt: timePtr(time.Unix(42, 0)),
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
				},
				BatteryPercentage: float32(42-2) / float32(253),
				DownlinkMargin:    4,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 42,
				Margin:  4,
			},
			ReceivedAt:     time.Unix(42, 0),
			ExpectedEvents: 1,
		},
		{
			Name: "external power/margin 20",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: 44,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Expected: &ttnpb.EndDevice{
				LastStatusReceivedAt: timePtr(time.Unix(42, 0)),
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
				},
				BatteryPercentage: -1,
				DownlinkMargin:    20,
				PowerState:        ttnpb.PowerState_POWER_EXTERNAL,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 0,
				Margin:  20,
			},
			ReceivedAt:     time.Unix(42, 0),
			ExpectedEvents: 1,
		},
		{
			Name: "unknown power/margin -5",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: 44,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Expected: &ttnpb.EndDevice{
				LastStatusReceivedAt: timePtr(time.Unix(42, 0)),
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
				},
				BatteryPercentage: -1,
				DownlinkMargin:    -5,
				PowerState:        ttnpb.PowerState_POWER_UNKNOWN,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 255,
				Margin:  -5,
			},
			ReceivedAt:     time.Unix(42, 0),
			ExpectedEvents: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleDevStatusAns(test.Context(), dev, tc.Payload, tc.ReceivedAt)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}

			if tc.ExpectedEvents > 0 {
				events.expect(t, tc.ExpectedEvents)
			}
			a.So(dev, should.Resemble, tc.Expected)
		})
	}
}
