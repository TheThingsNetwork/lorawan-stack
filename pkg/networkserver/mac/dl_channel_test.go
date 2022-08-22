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
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDLChannelReq(t *testing.T) {
	for _, tc := range []struct {
		CurrentChannels, DesiredChannels []*ttnpb.MACParameters_Channel
		RejectedFrequencies              []uint64
		Commands                         []*ttnpb.MACCommand_DLChannelReq
	}{
		{},
		{
			CurrentChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				nil,
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
			},
			DesiredChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				nil,
				{
					UplinkFrequency:   128,
					DownlinkFrequency: 128,
				},
			},
		},
		{
			CurrentChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   124,
					DownlinkFrequency: 124,
				},
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
			},
			DesiredChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   124,
					DownlinkFrequency: 124,
				},
				{
					UplinkFrequency:   128,
					DownlinkFrequency: 128,
				},
			},
		},
		{
			CurrentChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   124,
					DownlinkFrequency: 124,
				},
				nil,
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   129,
					DownlinkFrequency: 129,
				},
				{
					UplinkFrequency:   150,
					DownlinkFrequency: 150,
				},
			},
			DesiredChannels: []*ttnpb.MACParameters_Channel{
				nil,
				{
					UplinkFrequency:   124,
					DownlinkFrequency: 128,
				},
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   129,
					DownlinkFrequency: 130,
				},
			},
			Commands: []*ttnpb.MACCommand_DLChannelReq{
				{
					ChannelIndex: 1,
					Frequency:    128,
				},
				{
					ChannelIndex: 3,
					Frequency:    130,
				},
			},
		},
		{
			CurrentChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   124,
					DownlinkFrequency: 124,
				},
				nil,
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   129,
					DownlinkFrequency: 129,
				},
				{
					UplinkFrequency:   150,
					DownlinkFrequency: 150,
				},
			},
			DesiredChannels: []*ttnpb.MACParameters_Channel{
				nil,
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 128,
				},
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   130,
					DownlinkFrequency: 130,
				},
			},
			Commands: []*ttnpb.MACCommand_DLChannelReq{
				{
					ChannelIndex: 1,
					Frequency:    128,
				},
			},
			RejectedFrequencies: []uint64{130},
		},

		// https://github.com/TheThingsIndustries/lorawan-stack/issues/2525
		{
			CurrentChannels: []*ttnpb.MACParameters_Channel{
				{
					UplinkFrequency:   124,
					DownlinkFrequency: 124,
				},
				nil,
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
			},
			DesiredChannels: []*ttnpb.MACParameters_Channel{
				nil,
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 128,
				},
				{
					UplinkFrequency:   123,
					DownlinkFrequency: 123,
				},
				{
					UplinkFrequency:   130,
					DownlinkFrequency: 131,
				},
			},
			Commands: []*ttnpb.MACCommand_DLChannelReq{
				{
					ChannelIndex: 1,
					Frequency:    128,
				},
			},
			RejectedFrequencies: []uint64{130},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name: func() string {
				formatChannels := func(chs ...*ttnpb.MACParameters_Channel) string {
					return fmt.Sprintf("[%s]", test.JoinStringsMap(func(_, v interface{}) string {
						ch := v.(*ttnpb.MACParameters_Channel)
						if ch == nil {
							return "nil"
						}
						return fmt.Sprintf("%d", ch.DownlinkFrequency)
					}, ",", chs))
				}
				return fmt.Sprintf("channels:%s->%s,rejected_freqs:[%s]",
					formatChannels(tc.CurrentChannels...),
					formatChannels(tc.DesiredChannels...),
					test.JoinStringsf("%d", ",", false, tc.RejectedFrequencies),
				)
			}(),
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				makeDevice := func() *ttnpb.EndDevice {
					return ttnpb.Clone(&ttnpb.EndDevice{
						MacState: &ttnpb.MACState{
							CurrentParameters: &ttnpb.MACParameters{
								Channels: tc.CurrentChannels,
							},
							DesiredParameters: &ttnpb.MACParameters{
								Channels: tc.DesiredChannels,
							},
							RejectedFrequencies: tc.RejectedFrequencies,
							LorawanVersion:      ttnpb.MACVersion_MAC_V1_0_3,
						},
					})
				}

				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name:     "DeviceNeedsDLChannelReqAtIndex",
					Parallel: true,
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						dev := makeDevice()
						max := len(dev.MacState.CurrentParameters.Channels)
						if n := len(dev.MacState.DesiredParameters.Channels); n > max {
							max = n
						}
						needs := make(map[int]struct{}, max)
						for _, cmd := range tc.Commands {
							needs[int(cmd.ChannelIndex)] = struct{}{}
						}
						for i := 0; i <= max+1; i++ {
							i := i
							assert := should.BeFalse
							if _, ok := needs[i]; ok {
								assert = should.BeTrue
							}
							test.RunSubtestFromContext(ctx, test.SubtestConfig{
								Name: fmt.Sprintf("idx:%d", i),
								Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
									a.So(DeviceNeedsDLChannelReqAtIndex(dev, i), assert)
								},
							})
						}
					},
				})

				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name:     "DeviceNeedsDLChannelReq",
					Parallel: true,
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						dev := makeDevice()
						a.So(DeviceNeedsDLChannelReq(dev), func() func(interface{}, ...interface{}) string {
							if len(tc.Commands) > 0 {
								return should.BeTrue
							}
							return should.BeFalse
						}())
						a.So(dev, should.Resemble, makeDevice())
					},
				})

				for _, n := range func() []int {
					switch len(tc.Commands) {
					case 0:
						return []int{0}
					case 1:
						return []int{0, 1}
					default:
						return []int{0, len(tc.Commands) / 2, len(tc.Commands)}
					}
				}() {
					cmdsFit := n >= len(tc.Commands)
					cmdLen := (1 + lorawan.DefaultMACCommands[ttnpb.MACCommandIdentifier_CID_DL_CHANNEL].DownlinkLength) * uint16(n)
					cmds := tc.Commands[:n]
					answerLen := (1 + lorawan.DefaultMACCommands[ttnpb.MACCommandIdentifier_CID_DL_CHANNEL].UplinkLength) * uint16(n)
					test.RunSubtestFromContext(ctx, test.SubtestConfig{
						Name:     fmt.Sprintf("EnqueueDLChannelReq/max_down_len:%d", cmdLen),
						Parallel: true,
						Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
							dev := makeDevice()
							st := EnqueueDLChannelReq(ctx, dev, cmdLen, answerLen)
							expectedDevice := makeDevice()
							var expectedEventBuilders []events.Builder
							for _, cmd := range cmds {
								expectedDevice.MacState.PendingRequests = append(expectedDevice.MacState.PendingRequests, cmd.MACCommand())
								expectedEventBuilders = append(expectedEventBuilders, EvtEnqueueDLChannelRequest.BindData(cmd))
							}
							a.So(st.QueuedEvents, should.ResembleEventBuilders, events.Builders(expectedEventBuilders))
							if a.So(st, should.Resemble, EnqueueState{
								QueuedEvents: st.QueuedEvents,
								Ok:           cmdsFit,
							}) {
								a.So(dev, should.Resemble, expectedDevice)
							}
						},
					})
				}
			},
		})
	}
}

func TestHandleDLChannelAns(t *testing.T) {
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		Payload                     *ttnpb.MACCommand_DLChannelAns
		Error                       error
		Events                      events.Builders
	}{
		{
			Name: "nil payload",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Error: ErrNoPayload,
		},
		{
			Name: "frequency ack/chanel index ack/no request",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
			Events: events.Builders{
				EvtReceiveDLChannelAccept.With(events.WithData(&ttnpb.MACCommand_DLChannelAns{
					FrequencyAck:    true,
					ChannelIndexAck: true,
				})),
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_DL_CHANNEL),
		},
		{
			Name: "frequency nack/channel index ack/no request",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				ChannelIndexAck: true,
			},
			Events: events.Builders{
				EvtReceiveDLChannelReject.With(events.WithData(&ttnpb.MACCommand_DLChannelAns{
					ChannelIndexAck: true,
				})),
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_DL_CHANNEL),
		},
		{
			Name: "frequency nack/channel index nack/valid request/no rejections",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								DownlinkFrequency: 41,
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								DownlinkFrequency: 41,
							},
						},
					},
					DesiredParameters:   &ttnpb.MACParameters{},
					RejectedFrequencies: []uint64{42},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{},
			Events: events.Builders{
				EvtReceiveDLChannelReject.With(events.WithData(&ttnpb.MACCommand_DLChannelAns{})),
			},
		},
		{
			Name: "frequency nack/channel index ack/valid request/rejected frequencies:(1,2,100)",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								DownlinkFrequency: 41,
							},
						},
					},
					DesiredParameters:   &ttnpb.MACParameters{},
					RejectedFrequencies: []uint64{1, 2, 100},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								DownlinkFrequency: 41,
							},
						},
					},
					DesiredParameters:   &ttnpb.MACParameters{},
					RejectedFrequencies: []uint64{1, 2, 42, 100},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				ChannelIndexAck: true,
			},
			Events: events.Builders{
				EvtReceiveDLChannelReject.With(events.WithData(&ttnpb.MACCommand_DLChannelAns{
					ChannelIndexAck: true,
				})),
			},
		},
		{
			Name: "frequency ack/channel index ack/no channel",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
			Events: events.Builders{
				EvtReceiveDLChannelAccept.With(events.WithData(&ttnpb.MACCommand_DLChannelAns{
					FrequencyAck:    true,
					ChannelIndexAck: true,
				})),
			},
			Error: ErrCorruptedMACState.
				WithAttributes(
					"channels_len", 0,
					"request_channel_id", uint32(2),
				).
				WithCause(ErrUnknownChannel),
		},
		{
			Name: "frequency ack/channel index ack/channel exists",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 2,
							Frequency:    42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								DownlinkFrequency: 41,
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{},
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: true,
							},
							nil,
							{
								DownlinkFrequency: 42,
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_DLChannelAns{
				FrequencyAck:    true,
				ChannelIndexAck: true,
			},
			Events: events.Builders{
				EvtReceiveDLChannelAccept.With(events.WithData(&ttnpb.MACCommand_DLChannelAns{
					FrequencyAck:    true,
					ChannelIndexAck: true,
				})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)

				evs, err := HandleDLChannelAns(ctx, dev, tc.Payload)
				if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
					tc.Error == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(evs, should.ResembleEventBuilders, tc.Events)
			},
		})
	}
}
