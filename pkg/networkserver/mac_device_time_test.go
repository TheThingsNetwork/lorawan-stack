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
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleDeviceTimeReq(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Message          *ttnpb.UplinkMessage
		Events           []events.DefinitionDataClosure
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
							Time: time.Unix(42, 42),
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				RxMetadata: []*ttnpb.RxMetadata{
					{
						Time: timePtr(time.Unix(42, 42)),
					},
				},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDeviceTimeRequest.BindData(nil),
				evtEnqueueDeviceTimeAnswer.BindData(&ttnpb.MACCommand_DeviceTimeAns{
					Time: time.Unix(42, 42),
				}),
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
							Time: time.Unix(42, 42),
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				RxMetadata: []*ttnpb.RxMetadata{
					{
						Time: timePtr(time.Unix(42, 41)),
					},
					{
						Time: timePtr(time.Unix(42, 42)),
					},
					{
						Time: timePtr(time.Unix(420, 435353)),
					},
				},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDeviceTimeRequest.BindData(nil),
				evtEnqueueDeviceTimeAnswer.BindData(&ttnpb.MACCommand_DeviceTimeAns{
					Time: time.Unix(42, 42),
				}),
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
							Time: time.Unix(42, 45),
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				RxMetadata: []*ttnpb.RxMetadata{
					{
						Time: timePtr(time.Unix(42, 44)),
					},
					{
						Time: timePtr(time.Unix(42, 44)),
					},
					{
						Time: timePtr(time.Unix(42, 46)),
					},
					{
						Time: timePtr(time.Unix(42, 47)),
					},
				},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDeviceTimeRequest.BindData(nil),
				evtEnqueueDeviceTimeAnswer.BindData(&ttnpb.MACCommand_DeviceTimeAns{
					Time: time.Unix(42, 45),
				}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleDeviceTimeReq(test.Context(), dev, tc.Message)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
