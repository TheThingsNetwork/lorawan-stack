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
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
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
	ForEachClass(t, func(makeClassName func(parts ...string) string, class ttnpb.Class) {
		for _, conf := range []struct {
			Suffix                               string
			CurrentParameters, DesiredParameters *ttnpb.MACParameters
			Needs                                bool
		}{
			{
				Suffix: "current(frequency:42),desired(frequency:42)",
				CurrentParameters: &ttnpb.MACParameters{
					BeaconFrequency: 42,
				},
				DesiredParameters: &ttnpb.MACParameters{
					BeaconFrequency: 42,
				},
			},
			{
				Suffix: "current(frequency:24),desired(frequency:42)",
				CurrentParameters: &ttnpb.MACParameters{
					BeaconFrequency: 24,
				},
				DesiredParameters: &ttnpb.MACParameters{
					BeaconFrequency: 42,
				},
				Needs: true,
			},
		} {
			tcs = append(tcs,
				TestCase{
					Name: makeClassName(conf.Suffix),
					InputDevice: &ttnpb.EndDevice{
						MacState: &ttnpb.MACState{
							DeviceClass:       class,
							CurrentParameters: conf.CurrentParameters,
							DesiredParameters: conf.DesiredParameters,
						},
					},
					Needs: conf.Needs && class == ttnpb.Class_CLASS_B,
				},
			)
		}
	})

	for _, tc := range tcs {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)
				res := DeviceNeedsBeaconFreqReq(dev)
				if tc.Needs {
					a.So(res, should.BeTrue)
				} else {
					a.So(res, should.BeFalse)
				}
				a.So(dev, should.Resemble, tc.InputDevice)
			},
		})
	}
}

func TestHandleBeaconFreqAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_BeaconFreqAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Error: ErrNoPayload,
		},
		{
			Name: "ack/no request",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			Events: events.Builders{
				EvtReceiveBeaconFreqAccept.With(events.WithData(&ttnpb.MACCommand_BeaconFreqAns{
					FrequencyAck: true,
				})),
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_BEACON_FREQ),
		},
		{
			Name: "nack/no request",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{},
			Events: events.Builders{
				EvtReceiveBeaconFreqReject.With(events.WithData(&ttnpb.MACCommand_BeaconFreqAns{})),
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_BEACON_FREQ),
		},
		{
			Name: "ack/valid request",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_BeaconFreqReq{
							Frequency: 42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: &ttnpb.MACParameters{
						BeaconFrequency: 42,
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{
				FrequencyAck: true,
			},
			Events: events.Builders{
				EvtReceiveBeaconFreqAccept.With(events.WithData(&ttnpb.MACCommand_BeaconFreqAns{
					FrequencyAck: true,
				})),
			},
		},
		{
			Name: "nack/valid request",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_BeaconFreqReq{
							Frequency: 42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests:   []*ttnpb.MACCommand{},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_BeaconFreqAns{},
			Events: events.Builders{
				EvtReceiveBeaconFreqReject.With(events.WithData(&ttnpb.MACCommand_BeaconFreqAns{})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)

				var err error
				evs, err := HandleBeaconFreqAns(ctx, dev, tc.Payload)
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
