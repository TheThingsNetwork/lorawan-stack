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

func TestNeedsPingSlotChannelReq(t *testing.T) {
	for _, tc := range []struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Needs       bool
	}{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
		{
			Name: "current(data-rate-index:1,frequency:123),desired(data-rate-index:1,frequency:123)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_1},
						PingSlotFrequency:     123,
					},
					DesiredParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_1},
						PingSlotFrequency:     123,
					},
				},
			},
		},
		{
			Name: "current(data-rate-index:1,frequency:123),desired(data-rate-index:2,frequency:123)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_1},
						PingSlotFrequency:     123,
					},
					DesiredParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_2},
						PingSlotFrequency:     123,
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(data-rate-index:1,frequency:123),desired(data-rate-index:1,frequency:124)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_1},
						PingSlotFrequency:     123,
					},
					DesiredParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: ttnpb.DATA_RATE_1},
						PingSlotFrequency:     124,
					},
				},
			},
			Needs: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)
			res := deviceNeedsPingSlotChannelReq(dev)
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestHandlePingSlotChannelAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_PingSlotChannelAns
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
			Payload: &ttnpb.MACCommand_PingSlotChannelAns{
				FrequencyAck:     true,
				DataRateIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceivePingSlotChannelAnswer.BindData(&ttnpb.MACCommand_PingSlotChannelAns{
					FrequencyAck:     true,
					DataRateIndexAck: true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "both ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_PingSlotChannelReq{
							Frequency:     42,
							DataRateIndex: 43,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: ttnpb.MACParameters{
						PingSlotDataRateIndex: &ttnpb.DataRateIndexValue{Value: 43},
						PingSlotFrequency:     42,
					},
				},
			},
			Payload: &ttnpb.MACCommand_PingSlotChannelAns{
				FrequencyAck:     true,
				DataRateIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceivePingSlotChannelAnswer.BindData(&ttnpb.MACCommand_PingSlotChannelAns{
					FrequencyAck:     true,
					DataRateIndexAck: true,
				}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handlePingSlotChannelAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
