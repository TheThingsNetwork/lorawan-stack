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
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsRejoinParamSetupReq(t *testing.T) {
	type TestCase struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Needs       bool
	}
	tcs := []TestCase{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
	}
	for _, conf := range []struct {
		Suffix                               string
		CurrentParameters, DesiredParameters *ttnpb.MACParameters
		Needs                                bool
	}{
		{
			Suffix: "current(count:128,time:10),desired(count:128,time:10)",
			CurrentParameters: &ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
			},
			DesiredParameters: &ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
			},
		},
		{
			Suffix: "current(count:128,time:10),desired(count:128,time:12)",
			CurrentParameters: &ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
			},
			DesiredParameters: &ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_12,
			},
			Needs: true,
		},
		{
			Suffix: "current(count:128,time:10),desired(count:256,time:10)",
			CurrentParameters: &ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
			},
			DesiredParameters: &ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_256,
				RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
			},
			Needs: true,
		},
	} {
		ForEachMACVersion(t, func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
			tcs = append(tcs,
				TestCase{
					Name: makeMACName(conf.Suffix),
					InputDevice: &ttnpb.EndDevice{
						MacState: &ttnpb.MACState{
							LorawanVersion:    macVersion,
							CurrentParameters: conf.CurrentParameters,
							DesiredParameters: conf.DesiredParameters,
						},
					},
					Needs: conf.Needs && macspec.UseRejoinParamSetupReq(macVersion),
				},
			)
		})
	}

	for _, tc := range tcs {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)
				res := DeviceNeedsRejoinParamSetupReq(dev)
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

func TestHandleRejoinParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RejoinParamSetupAns
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
			Name: "no request",
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
			Payload: &ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			Events: events.Builders{
				EvtReceiveRejoinParamSetupAnswer.With(events.WithData(&ttnpb.MACCommand_RejoinParamSetupAns{
					MaxTimeExponentAck: true,
				})),
			},
			Error: ErrRequestNotFound,
		},
		{
			Name: "ack",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RejoinParamSetupReq{
							MaxCountExponent: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
							MaxTimeExponent:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_128,
						RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_10,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests:   []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			Events: events.Builders{
				EvtReceiveRejoinParamSetupAnswer.With(events.WithData(&ttnpb.MACCommand_RejoinParamSetupAns{
					MaxTimeExponentAck: true,
				})),
			},
		},
		{
			Name: "nack",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						RejoinTimePeriodicity: ttnpb.RejoinTimeExponent_REJOIN_TIME_1,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RejoinParamSetupReq{
							MaxCountExponent: ttnpb.RejoinCountExponent_REJOIN_COUNT_1024,
							MaxTimeExponent:  ttnpb.RejoinTimeExponent_REJOIN_TIME_11,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						RejoinCountPeriodicity: ttnpb.RejoinCountExponent_REJOIN_COUNT_1024,
						RejoinTimePeriodicity:  ttnpb.RejoinTimeExponent_REJOIN_TIME_1,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests:   []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: false,
			},
			Events: events.Builders{
				EvtReceiveRejoinParamSetupAnswer.With(events.WithData(&ttnpb.MACCommand_RejoinParamSetupAns{})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)

				evs, err := HandleRejoinParamSetupAns(ctx, dev, tc.Payload)
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
