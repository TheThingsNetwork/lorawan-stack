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

func TestHandleDLChannelAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DLChannelAns
		Error            error
		Events           []events.DefinitionDataClosure
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
			Name: "frequency ack/chanel index ack/no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDLChannelAccept.BindData(&ttnpb.MACCommand_DLChannelAns{
					FrequencyAck:    true,
					ChannelIndexAck: true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "frequency nack/channel index ack/no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				ChannelIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDLChannelReject.BindData(&ttnpb.MACCommand_DLChannelAns{
					ChannelIndexAck: true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "frequency nack/channel index ack/valid request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								UplinkFrequency:   41,
								DownlinkFrequency: 41,
							},
						},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								UplinkFrequency:   41,
								DownlinkFrequency: 41,
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				ChannelIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDLChannelReject.BindData(&ttnpb.MACCommand_DLChannelAns{
					ChannelIndexAck: true,
				}),
			},
		},
		{
			Name: "frequency ack/channel index ack/no channel",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDLChannelAccept.BindData(&ttnpb.MACCommand_DLChannelAns{
					FrequencyAck:    true,
					ChannelIndexAck: true,
				}),
			},
			Error: errCorruptedMACState.WithCause(errUnknownChannel),
		},
		{
			Name: "frequency ack/channel index ack/channel exists",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								UplinkFrequency:   41,
								DownlinkFrequency: 41,
							},
						},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								UplinkFrequency:   41,
								DownlinkFrequency: 42,
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveDLChannelAccept.BindData(&ttnpb.MACCommand_DLChannelAns{
					FrequencyAck:    true,
					ChannelIndexAck: true,
				}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleDLChannelAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
