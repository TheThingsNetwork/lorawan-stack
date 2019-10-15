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

func TestNeedsRejoinParamSetupReq(t *testing.T) {
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
	for _, conf := range []struct {
		Suffix                               string
		CurrentParameters, DesiredParameters ttnpb.MACParameters
		Needs                                bool
	}{
		{
			Suffix: "current(count:128,time:10),desired(count:128,time:10)",
			CurrentParameters: ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_10,
			},
			DesiredParameters: ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_10,
			},
		},
		{
			Suffix: "current(count:128,time:10),desired(count:128,time:12)",
			CurrentParameters: ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_10,
			},
			DesiredParameters: ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_12,
			},
			Needs: true,
		},
		{
			Suffix: "current(count:128,time:10),desired(count:256,time:10)",
			CurrentParameters: ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_128,
				RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_10,
			},
			DesiredParameters: ttnpb.MACParameters{
				RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_256,
				RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_10,
			},
			Needs: true,
		},
	} {
		ForEachMACVersion(func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
			tcs = append(tcs,
				TestCase{
					Name: makeMACName(conf.Suffix),
					InputDevice: &ttnpb.EndDevice{
						MACState: &ttnpb.MACState{
							LoRaWANVersion:    macVersion,
							CurrentParameters: conf.CurrentParameters,
							DesiredParameters: conf.DesiredParameters,
						},
					},
					Needs: conf.Needs && macVersion.Compare(ttnpb.MAC_V1_1) >= 0,
				},
			)
		})
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)
			res := needsRejoinParamSetupReq(dev)
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestHandleRejoinParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RejoinParamSetupAns
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
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveRejoinParamSetupAnswer.BindData(&ttnpb.MACCommand_RejoinParamSetupAns{
					MaxTimeExponentAck: true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RejoinParamSetupReq{
							MaxCountExponent: ttnpb.REJOIN_COUNT_128,
							MaxTimeExponent:  ttnpb.REJOIN_TIME_10,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_128,
						RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_10,
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveRejoinParamSetupAnswer.BindData(&ttnpb.MACCommand_RejoinParamSetupAns{
					MaxTimeExponentAck: true,
				}),
			},
		},
		{
			Name: "nack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						RejoinTimePeriodicity: ttnpb.REJOIN_TIME_1,
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RejoinParamSetupReq{
							MaxCountExponent: ttnpb.REJOIN_COUNT_1024,
							MaxTimeExponent:  ttnpb.REJOIN_TIME_11,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_1024,
						RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_1,
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RejoinParamSetupAns{
				MaxTimeExponentAck: false,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveRejoinParamSetupAnswer.BindData(&ttnpb.MACCommand_RejoinParamSetupAns{}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleRejoinParamSetupAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
