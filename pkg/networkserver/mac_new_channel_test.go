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

func TestNeedsNewChannelReq(t *testing.T) {
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
			Name: "current(channels:[(123,1-5),(124,1-3),(128,2-4)]),desired(channels:[(123,1-5),(124,1-3),(128,2-4)])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  128,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_4,
							},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  128,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_4,
							},
						},
					},
				},
			},
		},
		{
			// TODO: Disable channels using NewChannelReq. (https://github.com/TheThingsNetwork/lorawan-stack/issues/1499)
			Name: "current(channels:[(123,1-5),(124,1-3),(128,2-4)]),desired(channels:[(123,1-5),(124,1-3)])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  128,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_4,
							},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
						},
					},
				},
			},
		},
		{
			Name: "current(channels:[(123,1-5),(124,1-3),(128,2-4)]),desired(channels:[(123,1-5),(124,1-3),(128,2-3)])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  128,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_4,
							},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  128,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(channels:[(123,1-5),(124,1-3),(128,2-4)]),desired(channels:[(123,1-5),(124,1-3),(127,2-4)])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  128,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_4,
							},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:  123,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_5,
							},
							{
								UplinkFrequency:  124,
								MinDataRateIndex: ttnpb.DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DATA_RATE_3,
							},
							{
								UplinkFrequency:  127,
								MinDataRateIndex: ttnpb.DATA_RATE_2,
								MaxDataRateIndex: ttnpb.DATA_RATE_4,
							},
						},
					},
				},
			},
			Needs: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)
			res := deviceNeedsNewChannelReq(dev)
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestHandleNewChannelAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_NewChannelAns
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
			Payload: nil,
			Error:   errNoPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_NewChannelAns{
				FrequencyAck: true,
				DataRateAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveNewChannelAccept.BindData(&ttnpb.MACCommand_NewChannelAns{
					FrequencyAck: true,
					DataRateAck:  true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "both ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_NewChannelReq{
							ChannelIndex:     4,
							Frequency:        42,
							MinDataRateIndex: 2,
							MaxDataRateIndex: 3,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							nil,
							nil,
							nil,
							nil,
							{
								DownlinkFrequency: 42,
								UplinkFrequency:   42,
								MinDataRateIndex:  2,
								MaxDataRateIndex:  3,
								EnableUplink:      true,
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_NewChannelAns{
				FrequencyAck: true,
				DataRateAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveNewChannelAccept.BindData(&ttnpb.MACCommand_NewChannelAns{
					FrequencyAck: true,
					DataRateAck:  true,
				}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleNewChannelAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
