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
	"fmt"
	"math"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNeedsDLChannelReq(t *testing.T) {
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
			Suffix: "current([]),desired([])",
		},
		{
			Suffix: "current([123,123,123]),desired([123,123,123])",
			CurrentParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
				},
			},
			DesiredParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
				},
			},
		},
		{
			Suffix: "current([123,123,123]),desired([123,123])",
			CurrentParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
				},
			},
			DesiredParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
				},
			},
		},
		{
			Suffix: "current([123,123,123]),desired([123,124])",
			CurrentParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 123},
				},
			},
			DesiredParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{DownlinkFrequency: 123},
					{DownlinkFrequency: 124},
				},
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
					Needs: conf.Needs && macVersion.Compare(ttnpb.MAC_V1_0_2) >= 0,
				},
			)
		})
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)
			res := deviceNeedsDLChannelReq(dev)
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestEnqueueDLChannelReq(t *testing.T) {
	for _, tc := range []struct {
		Name                                 string
		CurrentParameters, DesiredParameters ttnpb.MACParameters
		ExpectedRequests                     []*ttnpb.MACCommand_DLChannelReq
	}{
		{
			Name: "no DLChannelReq necessary",
			CurrentParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{
						UplinkFrequency:   124,
						DownlinkFrequency: 124,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_3,
					},
					nil,
					{
						UplinkFrequency:   123,
						DownlinkFrequency: 123,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
					{
						UplinkFrequency:   129,
						DownlinkFrequency: 129,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_4,
					},
					{
						UplinkFrequency:   130,
						DownlinkFrequency: 131,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
				},
			},
			DesiredParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					nil,
					{
						UplinkFrequency:   128,
						DownlinkFrequency: 128,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_4,
					},
					{
						UplinkFrequency:   123,
						DownlinkFrequency: 123,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
					{
						UplinkFrequency:   130,
						DownlinkFrequency: 131,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
				},
			},
		},
		{
			Name: "4 DLChannelReq necessary",
			CurrentParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{
						UplinkFrequency:   124,
						DownlinkFrequency: 124,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_3,
					},
					{
						UplinkFrequency:   123,
						DownlinkFrequency: 123,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
					{
						UplinkFrequency:   129,
						DownlinkFrequency: 129,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_4,
					},
					{
						UplinkFrequency:   130,
						DownlinkFrequency: 131,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
					{
						UplinkFrequency:   130,
						DownlinkFrequency: 134,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
				},
			},
			DesiredParameters: ttnpb.MACParameters{
				Channels: []*ttnpb.MACParameters_Channel{
					{
						UplinkFrequency:   124,
						DownlinkFrequency: 128,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_3,
					},
					{
						UplinkFrequency:   123,
						DownlinkFrequency: 123,
						MinDataRateIndex:  ttnpb.DATA_RATE_1,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
					{
						UplinkFrequency:   129,
						DownlinkFrequency: 100,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_4,
					},
					{
						UplinkFrequency:   130,
						DownlinkFrequency: 125,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
					{
						UplinkFrequency:   130,
						DownlinkFrequency: 140,
						MinDataRateIndex:  ttnpb.DATA_RATE_2,
						MaxDataRateIndex:  ttnpb.DATA_RATE_5,
					},
				},
			},
			ExpectedRequests: []*ttnpb.MACCommand_DLChannelReq{
				{
					Frequency: 128,
				},
				{
					ChannelIndex: 2,
					Frequency:    100,
				},
				{
					ChannelIndex: 3,
					Frequency:    125,
				},
				{
					ChannelIndex: 4,
					Frequency:    140,
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			downlinkLength := 1 + lorawan.DefaultMACCommands[ttnpb.CID_DL_CHANNEL].DownlinkLength
			uplinkLength := 1 + lorawan.DefaultMACCommands[ttnpb.CID_DL_CHANNEL].UplinkLength

			type TestConf struct {
				MaxDownlinkLength, MaxUplinkLength uint16
				ExpectedCount                      int
			}
			confs := []TestConf{
				{},
				{
					MaxUplinkLength: math.MaxUint16,
				},
				{
					MaxDownlinkLength: math.MaxUint16,
				},
				{
					MaxDownlinkLength: math.MaxUint16,
					MaxUplinkLength:   math.MaxUint16,
					ExpectedCount:     len(tc.ExpectedRequests),
				},
			}
			for i := range tc.ExpectedRequests {
				for j := 0; j <= i; j++ {
					confs = append(confs, TestConf{
						MaxDownlinkLength: uint16(i+1) * downlinkLength,
						MaxUplinkLength:   uint16(j+1) * uplinkLength,
						ExpectedCount:     j + 1,
					})
				}
			}

			for _, conf := range confs {
				for _, pendingReqs := range [][]*ttnpb.MACCommand{
					nil,
					{
						{},
					},
				} {
					t.Run(fmt.Sprintf("max_downlink_len:%d,max_uplink_len:%d,pending_requests:%d", conf.MaxDownlinkLength, conf.MaxUplinkLength, len(pendingReqs)), func(t *testing.T) {
						a := assertions.New(t)

						dev := &ttnpb.EndDevice{
							MACState: &ttnpb.MACState{
								LoRaWANVersion:    ttnpb.MAC_V1_0_2,
								CurrentParameters: tc.CurrentParameters,
								DesiredParameters: tc.DesiredParameters,
								PendingRequests:   pendingReqs,
							},
						}
						reqs := tc.ExpectedRequests[:conf.ExpectedCount]
						expectedDev := CopyEndDevice(dev)
						var expectedEvs []events.DefinitionDataClosure
						for _, req := range reqs {
							expectedDev.MACState.PendingRequests = append(expectedDev.MACState.PendingRequests, req.MACCommand())
							expectedEvs = append(expectedEvs, evtEnqueueDLChannelRequest.BindData(req))
						}

						st := enqueueDLChannelReq(test.Context(), dev, conf.MaxDownlinkLength, conf.MaxUplinkLength)
						a.So(dev, should.Resemble, expectedDev)
						a.So(st.QueuedEvents, should.ResembleEventDefinitionDataClosures, expectedEvs)
						a.So(st, should.Resemble, macCommandEnqueueState{
							MaxDownLen:   conf.MaxDownlinkLength - uint16(conf.ExpectedCount)*downlinkLength,
							MaxUpLen:     conf.MaxUplinkLength - uint16(conf.ExpectedCount)*uplinkLength,
							Ok:           len(tc.ExpectedRequests) == conf.ExpectedCount,
							QueuedEvents: st.QueuedEvents,
						})
					})
				}
			}
		})
	}
}

func TestHandleDLChannelAns(t *testing.T) {
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		Payload                     *ttnpb.MACCommand_DLChannelAns
		Error                       error
		Events                      []events.DefinitionDataClosure
	}{
		{
			Name: "nil payload",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Error: errNoPayload,
		},
		{
			Name: "frequency ack/chanel index ack/no request",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			ExpectedDevice: &ttnpb.EndDevice{
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
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			ExpectedDevice: &ttnpb.EndDevice{
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
			InputDevice: &ttnpb.EndDevice{
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
			ExpectedDevice: &ttnpb.EndDevice{
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
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
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
			InputDevice: &ttnpb.EndDevice{
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
			ExpectedDevice: &ttnpb.EndDevice{
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

			dev := CopyEndDevice(tc.InputDevice)

			evs, err := handleDLChannelAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.ExpectedDevice)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
