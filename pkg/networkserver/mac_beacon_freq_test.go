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

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNeedsBeaconFreqReq(t *testing.T) {
	type TestCase struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Needs       bool
	}
	var tcs []TestCase

	tcs = append(tcs,
		TestCase{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
	)
	ForEachClass(func(makeClassName func(parts ...string) string, class ttnpb.Class) {
		for _, conf := range []struct {
			Suffix                               string
			CurrentParameters, DesiredParameters ttnpb.MACParameters
			Needs                                bool
		}{
			{
				Suffix: "current(frequency:42),desired(frequency:42)",
				CurrentParameters: ttnpb.MACParameters{
					BeaconFrequency: 42,
				},
				DesiredParameters: ttnpb.MACParameters{
					BeaconFrequency: 42,
				},
			},
			{
				Suffix: "current(frequency:24),desired(frequency:42)",
				CurrentParameters: ttnpb.MACParameters{
					BeaconFrequency: 24,
				},
				DesiredParameters: ttnpb.MACParameters{
					BeaconFrequency: 42,
				},
				Needs: true,
			},
		} {
			tcs = append(tcs,
				TestCase{
					Name: makeClassName(conf.Suffix),
					InputDevice: &ttnpb.EndDevice{
						MACState: &ttnpb.MACState{
							DeviceClass:       class,
							CurrentParameters: conf.CurrentParameters,
							DesiredParameters: conf.DesiredParameters,
						},
					},
					Needs: conf.Needs && class == ttnpb.CLASS_B,
				},
			)
		}
	})

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)
			res := deviceNeedsBeaconFreqReq(dev)
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestHandleBeaconFreqAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_BeaconFreqAns
		Events           []events.DefinitionDataClosure
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
			Name: "ack/no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveBeaconFreqAccept.BindData(&ttnpb.MACCommand_BeaconFreqAns{
					FrequencyAck: true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "nack/no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{},
			Events: []events.DefinitionDataClosure{
				evtReceiveBeaconFreqReject.BindData(&ttnpb.MACCommand_BeaconFreqAns{}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "ack/valid request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_BeaconFreqReq{
							Frequency: 42,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: ttnpb.MACParameters{
						BeaconFrequency: 42,
					},
				},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveBeaconFreqAccept.BindData(&ttnpb.MACCommand_BeaconFreqAns{
					FrequencyAck: true,
				}),
			},
		},
		{
			Name: "nack/valid request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_BeaconFreqReq{
							Frequency: 42,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{},
			Events: []events.DefinitionDataClosure{
				evtReceiveBeaconFreqReject.BindData(&ttnpb.MACCommand_BeaconFreqAns{}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			var err error
			evs, err := handleBeaconFreqAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
