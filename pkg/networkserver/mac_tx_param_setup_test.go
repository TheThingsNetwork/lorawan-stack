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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEnqueueTxParamSetupReq(t *testing.T) {
	for _, tc := range []struct {
		Name                                              string
		Device, Expected                                  *ttnpb.EndDevice
		AssertEvents                                      func(*testing.T, ...events.Event) bool
		InputMaxDownlinkLength, ExpectedMaxDownlinkLength uint16
		InputMaxUplinkLength, ExpectedMaxUplinkLength     uint16
		Ok                                                bool
	}{
		{
			Name: "payload fits/EIRP 26/dwell time both",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_TxParamSetupReq{
							MaxEIRPIndex:      ttnpb.DEVICE_EIRP_26,
							DownlinkDwellTime: true,
							UplinkDwellTime:   true,
						}).MACCommand(),
					},
				},
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 1) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.tx_param_setup.request") &&
					a.So(evs[0].Data(), should.Resemble, &ttnpb.MACCommand_TxParamSetupReq{
						MaxEIRPIndex:      ttnpb.DEVICE_EIRP_26,
						DownlinkDwellTime: true,
						UplinkDwellTime:   true,
					})
			},
			InputMaxDownlinkLength:    42,
			InputMaxUplinkLength:      24,
			ExpectedMaxDownlinkLength: 40,
			ExpectedMaxUplinkLength:   23,
			Ok:                        true,
		},
		{
			Name: "payload fits/EIRP 26/no dwell time limitations",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
				},
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
			InputMaxDownlinkLength:    42,
			InputMaxUplinkLength:      24,
			ExpectedMaxDownlinkLength: 42,
			ExpectedMaxUplinkLength:   24,
			Ok:                        true,
		},
		{
			Name: "downlink does not fit/EIRP 26/dwell time both",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
				},
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
			InputMaxDownlinkLength:    1,
			InputMaxUplinkLength:      24,
			ExpectedMaxDownlinkLength: 1,
			ExpectedMaxUplinkLength:   24,
			Ok:                        false,
		},
		{
			Name: "uplink does not fit/EIRP 26/dwell time both",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
				},
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
			InputMaxDownlinkLength:    42,
			InputMaxUplinkLength:      0,
			ExpectedMaxDownlinkLength: 42,
			ExpectedMaxUplinkLength:   0,
			Ok:                        false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			var maxDownLen, maxUpLen uint16
			var ok bool
			evs := test.CollectEvents(func() {
				maxDownLen, maxUpLen, ok = enqueueTxParamSetupReq(test.Context(), dev, tc.InputMaxDownlinkLength, tc.InputMaxUplinkLength)
			})
			a.So(dev, should.Resemble, tc.Expected)
			a.So(maxDownLen, should.Equal, tc.ExpectedMaxDownlinkLength)
			a.So(maxUpLen, should.Equal, tc.ExpectedMaxUplinkLength)
			a.So(ok, should.Resemble, tc.Ok)
			a.So(tc.AssertEvents(t, evs...), should.BeTrue)
		})
	}
}

func TestHandleTxParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Events           []events.DefinitionDataClosure
		Error            error
	}{
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveTxParamSetupAnswer.BindData(nil),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "EIRP 26, dwell time both",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_TxParamSetupReq{
							MaxEIRPIndex:      ttnpb.DEVICE_EIRP_26,
							DownlinkDwellTime: true,
							UplinkDwellTime:   true,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP:           26,
						DownlinkDwellTime: &pbtypes.BoolValue{Value: true},
						UplinkDwellTime:   &pbtypes.BoolValue{Value: true},
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveTxParamSetupAnswer.BindData(nil),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleTxParamSetupAns(test.Context(), dev)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
