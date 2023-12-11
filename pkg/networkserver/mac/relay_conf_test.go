// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDeviceNeedsRelayConfReq(t *testing.T) {
	t.Parallel()
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
			Name: "no relay",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
		},
		{
			Name: "disable serving",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Needs: true,
		},
		{
			Name: "enable serving",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{},
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "enable second channel",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
								},
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "no change",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.InputDevice)
				res := mac.DeviceNeedsRelayConfReq(dev)
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

func TestEnqueueRelayConfReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       mac.EnqueueState
	}{
		{
			Name: "enable serving",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								SecondChannel: &ttnpb.RelaySecondChannel{
									AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
									DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
									Frequency:     123,
								},
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
							},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 44,
				MaxUpLen:   48,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayConfRequest.With(events.WithData(&ttnpb.MACCommand_RelayConfReq{
						Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
							SecondChannel: &ttnpb.RelaySecondChannel{
								AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
								DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
								Frequency:     123,
							},
							DefaultChannelIndex: 1,
							CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
						},
					})),
				},
				Ok: true,
			},
		},
		{
			Name: "disable serving",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: nil,
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 44,
				MaxUpLen:   48,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayConfRequest.With(events.WithData(&ttnpb.MACCommand_RelayConfReq{
						Configuration: nil,
					})),
				},
				Ok: true,
			},
		},
		{
			Name: "disable second channel",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								SecondChannel:       nil,
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
							},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 44,
				MaxUpLen:   48,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayConfRequest.With(events.WithData(&ttnpb.MACCommand_RelayConfReq{
						Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
							SecondChannel:       nil,
							DefaultChannelIndex: 1,
							CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
						},
					})),
				},
				Ok: true,
			},
		},
		{
			Name: "enable second channel",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								SecondChannel: &ttnpb.RelaySecondChannel{
									AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
									DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
									Frequency:     123,
								},
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
							},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 44,
				MaxUpLen:   48,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayConfRequest.With(events.WithData(&ttnpb.MACCommand_RelayConfReq{
						Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
							SecondChannel: &ttnpb.RelaySecondChannel{
								AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_1600,
								DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
								Frequency:     123,
							},
							DefaultChannelIndex: 1,
							CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_100_MILLISECONDS,
						},
					})),
				},
				Ok: true,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.InputDevice)
				st := mac.EnqueueRelayConfReq(ctx, dev, tc.MaxDownlinkLength, tc.MaxUplinkLength)
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
				st.QueuedEvents = tc.State.QueuedEvents
				a.So(st, should.Resemble, tc.State)
			},
		})
	}
}

func TestHandleRelayConfAns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RelayConfAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "reject serving",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								SecondChannel: &ttnpb.RelaySecondChannel{
									AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
									DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
									Frequency:     123,
								},
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfAns{},
			Events: events.Builders{
				mac.EvtReceiveRelayConfReject.With(events.WithData(&ttnpb.MACCommand_RelayConfAns{})),
			},
		},
		{
			Name: "enable serving",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								SecondChannel: &ttnpb.RelaySecondChannel{
									AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
									DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
									Frequency:     123,
								},
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfAns{
				SecondChannelFrequencyAck:     true,
				SecondChannelAckOffsetAck:     true,
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         true,
				DefaultChannelIndexAck:        true,
				CadPeriodicityAck:             true,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayConfAccept.With(events.WithData(&ttnpb.MACCommand_RelayConfAns{
					SecondChannelFrequencyAck:     true,
					SecondChannelAckOffsetAck:     true,
					SecondChannelDataRateIndexAck: true,
					SecondChannelIndexAck:         true,
					DefaultChannelIndexAck:        true,
					CadPeriodicityAck:             true,
				})),
			},
		},
		{
			Name: "disable serving",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests:   []*ttnpb.MACCommand{(&ttnpb.MACCommand_RelayConfReq{}).MACCommand()},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfAns{
				SecondChannelFrequencyAck:     true,
				SecondChannelAckOffsetAck:     true,
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         true,
				DefaultChannelIndexAck:        true,
				CadPeriodicityAck:             true,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayConfAccept.With(events.WithData(&ttnpb.MACCommand_RelayConfAns{
					SecondChannelFrequencyAck:     true,
					SecondChannelAckOffsetAck:     true,
					SecondChannelDataRateIndexAck: true,
					SecondChannelIndexAck:         true,
					DefaultChannelIndexAck:        true,
					CadPeriodicityAck:             true,
				})),
			},
		},
		{
			Name: "enable second channel",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								SecondChannel: &ttnpb.RelaySecondChannel{
									AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
									DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
									Frequency:     123,
								},
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfAns{
				SecondChannelFrequencyAck:     true,
				SecondChannelAckOffsetAck:     true,
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         true,
				DefaultChannelIndexAck:        true,
				CadPeriodicityAck:             true,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayConfAccept.With(events.WithData(&ttnpb.MACCommand_RelayConfAns{
					SecondChannelFrequencyAck:     true,
					SecondChannelAckOffsetAck:     true,
					SecondChannelDataRateIndexAck: true,
					SecondChannelIndexAck:         true,
					DefaultChannelIndexAck:        true,
					CadPeriodicityAck:             true,
				})),
			},
		},
		{
			Name: "disable second channel",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									SecondChannel: &ttnpb.RelaySecondChannel{
										AckOffset:     ttnpb.RelaySecondChAckOffset_RELAY_SECOND_CH_ACK_OFFSET_3200,
										DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
										Frequency:     123,
									},
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfReq{
							Configuration: &ttnpb.MACCommand_RelayConfReq_Configuration{
								DefaultChannelIndex: 1,
								CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									DefaultChannelIndex: 1,
									CadPeriodicity:      ttnpb.RelayCADPeriodicity_RELAY_CAD_PERIODICITY_1_SECOND,
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfAns{
				SecondChannelFrequencyAck:     true,
				SecondChannelAckOffsetAck:     true,
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         true,
				DefaultChannelIndexAck:        true,
				CadPeriodicityAck:             true,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayConfAccept.With(events.WithData(&ttnpb.MACCommand_RelayConfAns{
					SecondChannelFrequencyAck:     true,
					SecondChannelAckOffsetAck:     true,
					SecondChannelDataRateIndexAck: true,
					SecondChannelIndexAck:         true,
					DefaultChannelIndexAck:        true,
					CadPeriodicityAck:             true,
				})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.Device)
				evs, err := mac.HandleRelayConfAns(ctx, dev, tc.Payload)
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
