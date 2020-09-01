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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHandleDeviceTimeReq(t *testing.T) {
	recvAt := time.Unix(42, 42).UTC()
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Message          *ttnpb.UplinkMessage
		Events           events.Builders
		Error            error
	}{
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DeviceTimeAns{
							Time: recvAt,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				ReceivedAt: recvAt,
			},
			Events: events.Builders{
				EvtReceiveDeviceTimeRequest,
				EvtEnqueueDeviceTimeAnswer.With(events.WithData(&ttnpb.MACCommand_DeviceTimeAns{
					Time: recvAt,
				})),
			},
		},
		{
			Name: "non-empty queue/odd",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						(&ttnpb.MACCommand_DeviceTimeAns{
							Time: recvAt,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				ReceivedAt: recvAt,
			},
			Events: events.Builders{
				EvtReceiveDeviceTimeRequest,
				EvtEnqueueDeviceTimeAnswer.With(events.WithData(&ttnpb.MACCommand_DeviceTimeAns{
					Time: recvAt,
				})),
			},
		},
		{
			Name: "non-empty queue/even",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						(&ttnpb.MACCommand_DeviceTimeAns{
							Time: recvAt,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				ReceivedAt: recvAt,
			},
			Events: events.Builders{
				EvtReceiveDeviceTimeRequest,
				EvtEnqueueDeviceTimeAnswer.With(events.WithData(&ttnpb.MACCommand_DeviceTimeAns{
					Time: recvAt,
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

				evs, err := HandleDeviceTimeReq(ctx, dev, tc.Message)
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
