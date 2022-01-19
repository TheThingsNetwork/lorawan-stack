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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHandlePingSlotInfoReq(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_PingSlotInfoReq
		Events           events.Builders
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_B,
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_B,
				},
			},
			Error: ErrNoPayload,
		},
		{
			Name: "class B device",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_B,
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_B,
				},
			},
			Payload: &ttnpb.MACCommand_PingSlotInfoReq{
				Period: 42,
			},
			Events: events.Builders{
				EvtReceivePingSlotInfoRequest.With(events.WithData(&ttnpb.MACCommand_PingSlotInfoReq{
					Period: 42,
				})),
			},
		},
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_A,
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass:         ttnpb.CLASS_A,
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: 42},
					QueuedResponses: []*ttnpb.MACCommand{
						ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO.MACCommand(),
					},
				},
			},
			Payload: &ttnpb.MACCommand_PingSlotInfoReq{
				Period: 42,
			},
			Events: events.Builders{
				EvtReceivePingSlotInfoRequest.With(events.WithData(&ttnpb.MACCommand_PingSlotInfoReq{
					Period: 42,
				})),
				EvtEnqueuePingSlotInfoAnswer,
			},
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_A,
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DeviceClass:         ttnpb.CLASS_A,
					PingSlotPeriodicity: &ttnpb.PingSlotPeriodValue{Value: 42},
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO.MACCommand(),
					},
				},
			},
			Payload: &ttnpb.MACCommand_PingSlotInfoReq{
				Period: 42,
			},
			Events: events.Builders{
				EvtReceivePingSlotInfoRequest.With(events.WithData(&ttnpb.MACCommand_PingSlotInfoReq{
					Period: 42,
				})),
				EvtEnqueuePingSlotInfoAnswer,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)

				evs, err := HandlePingSlotInfoReq(ctx, dev, tc.Payload)
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
