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
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       macCommandEnqueueState
	}{
		{
			Name: "payload fits/EIRP 26/dwell time both",
			InputDevice: &ttnpb.EndDevice{
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
			ExpectedDevice: &ttnpb.EndDevice{
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
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 40,
				MaxUpLen:   23,
				Ok:         true,
				QueuedEvents: []events.DefinitionDataClosure{
					evtEnqueueTxParamSetupRequest.BindData(&ttnpb.MACCommand_TxParamSetupReq{
						MaxEIRPIndex:      ttnpb.DEVICE_EIRP_26,
						DownlinkDwellTime: true,
						UplinkDwellTime:   true,
					}),
				},
			},
		},
		{
			Name: "payload fits/EIRP 26/no dwell time limitations",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
					DesiredParameters: ttnpb.MACParameters{
						MaxEIRP: 26,
					},
				},
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 42,
				MaxUpLen:   24,
				Ok:         true,
			},
		},
		{
			Name: "downlink does not fit/EIRP 26/dwell time both",
			InputDevice: &ttnpb.EndDevice{
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
			ExpectedDevice: &ttnpb.EndDevice{
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
			MaxDownlinkLength: 1,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 1,
				MaxUpLen:   24,
			},
		},
		{
			Name: "uplink does not fit/EIRP 26/dwell time both",
			InputDevice: &ttnpb.EndDevice{
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
			ExpectedDevice: &ttnpb.EndDevice{
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
			MaxDownlinkLength: 42,
			State: macCommandEnqueueState{
				MaxDownLen: 42,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.InputDevice).(*ttnpb.EndDevice)

			st := enqueueTxParamSetupReq(test.Context(), dev, tc.MaxDownlinkLength, tc.MaxUplinkLength)
			a.So(dev, should.Resemble, tc.ExpectedDevice)
			a.So(st.QueuedEvents, should.ResembleEventDefinitionDataClosures, tc.State.QueuedEvents)
			st.QueuedEvents = tc.State.QueuedEvents
			a.So(st, should.Resemble, tc.State)
		})
	}
}

func TestHandleTxParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		Events                      []events.DefinitionDataClosure
		Error                       error
	}{
		{
			Name: "no request",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveTxParamSetupAnswer.BindData(nil),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "EIRP 26, dwell time both",
			InputDevice: &ttnpb.EndDevice{
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
			ExpectedDevice: &ttnpb.EndDevice{
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

			dev := deepcopy.Copy(tc.InputDevice).(*ttnpb.EndDevice)

			evs, err := handleTxParamSetupAns(test.Context(), dev)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.ExpectedDevice)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
