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
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleDevStatusAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DevStatusAns
		ReceivedAt       time.Time
		Error            error
		EventAssertion   func(*testing.T, ...events.Event) bool
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
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
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
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
		},
		{
			Name: "battery 42/margin 4",
			Device: &ttnpb.EndDevice{
				Session: &ttnpb.Session{
					LastFCntUp: 43,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 2,
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: 44,
				PowerState:        ttnpb.PowerState_POWER_EXTERNAL,
			},
			Expected: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: timePtr(time.Unix(42, 0)),
				Session: &ttnpb.Session{
					LastFCntUp: 43,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 43,
					PendingRequests:     []*ttnpb.MACCommand{},
				},
				BatteryPercentage: float32(42-2) / float32(253),
				DownlinkMargin:    4,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 42,
				Margin:  4,
			},
			ReceivedAt: time.Unix(42, 0),
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 1) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.dev_status.answer") &&
					a.So(evs[0].Data(), should.Resemble, &ttnpb.MACCommand_DevStatusAns{
						Battery: 42,
						Margin:  4,
					})
			},
		},
		{
			Name: "external power/margin 20",
			Device: &ttnpb.EndDevice{
				Session: &ttnpb.Session{
					LastFCntUp: 43,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 2,
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: 44,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Expected: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: timePtr(time.Unix(42, 0)),
				Session: &ttnpb.Session{
					LastFCntUp: 43,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 43,
					PendingRequests:     []*ttnpb.MACCommand{},
				},
				BatteryPercentage: -1,
				DownlinkMargin:    20,
				PowerState:        ttnpb.PowerState_POWER_EXTERNAL,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 0,
				Margin:  20,
			},
			ReceivedAt: time.Unix(42, 0),
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 1) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.dev_status.answer") &&
					a.So(evs[0].Data(), should.Resemble, &ttnpb.MACCommand_DevStatusAns{
						Battery: 0,
						Margin:  20,
					})
			},
		},
		{
			Name: "unknown power/margin -5",
			Device: &ttnpb.EndDevice{
				Session: &ttnpb.Session{
					LastFCntUp: 43,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 2,
					PendingRequests: []*ttnpb.MACCommand{
						ttnpb.CID_DEV_STATUS.MACCommand(),
					},
				},
				BatteryPercentage: 44,
				PowerState:        ttnpb.PowerState_POWER_BATTERY,
			},
			Expected: &ttnpb.EndDevice{
				LastDevStatusReceivedAt: timePtr(time.Unix(42, 0)),
				Session: &ttnpb.Session{
					LastFCntUp: 43,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 43,
					PendingRequests:     []*ttnpb.MACCommand{},
				},
				BatteryPercentage: -1,
				DownlinkMargin:    -5,
				PowerState:        ttnpb.PowerState_POWER_UNKNOWN,
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 255,
				Margin:  -5,
			},
			ReceivedAt: time.Unix(42, 0),
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 1) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.dev_status.answer") &&
					a.So(evs[0].Data(), should.Resemble, &ttnpb.MACCommand_DevStatusAns{
						Battery: 255,
						Margin:  -5,
					})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			var err error
			evs := collectEvents(func() {
				err = handleDevStatusAns(test.Context(), dev, tc.Payload, tc.ReceivedAt)
			})
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(tc.EventAssertion(t, evs...), should.BeTrue)
		})
	}
}
