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
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHandleDeviceModeInd(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DeviceModeInd
		Events           events.Builders
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Error: errNoPayload,
		},
		{
			Name: "does not support class C/empty queue",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_A,
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DeviceModeConf{
							Class: ttnpb.CLASS_A,
						}).MACCommand(),
					},
					DeviceClass: ttnpb.CLASS_A,
				},
			},
			Payload: &ttnpb.MACCommand_DeviceModeInd{
				Class: ttnpb.CLASS_C,
			},
			Events: events.Builders{
				evtReceiveDeviceModeIndication.With(events.WithData(&ttnpb.MACCommand_DeviceModeInd{
					Class: ttnpb.CLASS_C,
				})),
				evtEnqueueDeviceModeConfirmation.With(events.WithData(&ttnpb.MACCommand_DeviceModeConf{
					Class: ttnpb.CLASS_A,
				})),
			},
		},
		{
			Name: "supports class C/empty queue",
			Device: &ttnpb.EndDevice{
				SupportsClassC: true,
				MACState: &ttnpb.MACState{
					DeviceClass: ttnpb.CLASS_A,
				},
			},
			Expected: &ttnpb.EndDevice{
				SupportsClassC: true,
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DeviceModeConf{
							Class: ttnpb.CLASS_C,
						}).MACCommand(),
					},
					DeviceClass: ttnpb.CLASS_C,
				},
			},
			Payload: &ttnpb.MACCommand_DeviceModeInd{
				Class: ttnpb.CLASS_C,
			},
			Events: events.Builders{
				evtReceiveDeviceModeIndication.With(events.WithData(&ttnpb.MACCommand_DeviceModeInd{
					Class: ttnpb.CLASS_C,
				})),
				evtClassCSwitch.With(events.WithData(ttnpb.CLASS_A)),
				evtEnqueueDeviceModeConfirmation.With(events.WithData(&ttnpb.MACCommand_DeviceModeConf{
					Class: ttnpb.CLASS_C,
				})),
			},
		},
		{
			Name: "supports class C/non-empty queue",
			Device: &ttnpb.EndDevice{
				SupportsClassC: true,
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
					DeviceClass: ttnpb.CLASS_C,
				},
			},
			Expected: &ttnpb.EndDevice{
				SupportsClassC: true,
				MACState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						(&ttnpb.MACCommand_DeviceModeConf{
							Class: ttnpb.CLASS_A,
						}).MACCommand(),
					},
					DeviceClass: ttnpb.CLASS_A,
				},
			},
			Payload: &ttnpb.MACCommand_DeviceModeInd{
				Class: ttnpb.CLASS_A,
			},
			Events: events.Builders{
				evtReceiveDeviceModeIndication.With(events.WithData(&ttnpb.MACCommand_DeviceModeInd{
					Class: ttnpb.CLASS_A,
				})),
				evtClassASwitch.With(events.WithData(ttnpb.CLASS_C)),
				evtEnqueueDeviceModeConfirmation.With(events.WithData(&ttnpb.MACCommand_DeviceModeConf{
					Class: ttnpb.CLASS_A,
				})),
			},
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)

				evs, err := handleDeviceModeInd(ctx, dev, tc.Payload)
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
