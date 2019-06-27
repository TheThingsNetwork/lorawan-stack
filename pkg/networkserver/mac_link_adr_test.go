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
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleLinkADRAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_LinkADRAns
		DupCount         uint
		Events           []events.DefinitionDataClosure
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 42,
									},
								},
							},
						},
					},
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 43,
									},
								},
							},
						},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 42,
									},
								},
							},
						},
					},
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 43,
									},
								},
							},
						},
					},
				},
			},
			Error: errNoPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 42,
									},
								},
							},
						},
					},
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 43,
									},
								},
							},
						},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 42,
									},
								},
							},
						},
					},
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{
									FHDR: ttnpb.FHDR{
										FCtrl: ttnpb.FCtrl{
											ADR: true,
										},
										FCnt: 43,
									},
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveLinkADRAccept.BindData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				}),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "1 request/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							nil,
							{UplinkFrequency: 42},
							{DownlinkFrequency: 23},
							nil,
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_4,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								false, true, false, false,
								false, false, false, false,
								false, false, false, false,
								false, false, false, false,
							},
						}).MACCommand(),
					},
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					newADRUplink(42, -2, 2, true, ttnpb.TxSettings{}),
					newADRUplink(43, -2, 2, false, ttnpb.TxSettings{}),
					newADRUplink(44, -3, 2, false, ttnpb.TxSettings{}),
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_4,
						ADRTxPowerIndex:  42,
						Channels: []*ttnpb.MACParameters_Channel{
							nil,
							{
								EnableUplink:    true,
								UplinkFrequency: 42,
							},
							{
								EnableUplink:      false,
								DownlinkFrequency: 23,
							},
							nil,
						},
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveLinkADRAccept.BindData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				}),
			},
		},
		{
			Name: "1.1/2 requests/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							nil,
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_10,
							TxPowerIndex:  43,
							ChannelMask: []bool{
								false, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
					},
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					newADRUplink(42, -2, 2, true, ttnpb.TxSettings{}),
					newADRUplink(43, -2, 2, false, ttnpb.TxSettings{}),
					newADRUplink(44, -3, 2, false, ttnpb.TxSettings{}),
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_10,
						ADRTxPowerIndex:  43,
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: false},
							{EnableUplink: true},
							{EnableUplink: true},
							nil,
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: false},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveLinkADRAccept.BindData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				}),
			},
		},
		{
			Name:     "1.0.2/2 requests/all ack",
			DupCount: 1,
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_0_2,
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							nil,
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_10,
							TxPowerIndex:  43,
							ChannelMask: []bool{
								false, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
					},
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					newADRUplink(42, -2, 2, true, ttnpb.TxSettings{}),
					newADRUplink(43, -2, 2, false, ttnpb.TxSettings{}),
					newADRUplink(44, -3, 2, false, ttnpb.TxSettings{}),
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_0_2,
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_10,
						ADRTxPowerIndex:  43,
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: false},
							{EnableUplink: true},
							{EnableUplink: true},
							nil,
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: false},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveLinkADRAccept.BindData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				}),
			},
		},
		{
			Name: "1.0/2 requests/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					CurrentParameters: ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							nil,
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_10,
							TxPowerIndex:  43,
							ChannelMask: []bool{
								false, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
					},
				},
				RecentADRUplinks: []*ttnpb.UplinkMessage{
					newADRUplink(42, -2, 2, true, ttnpb.TxSettings{}),
					newADRUplink(43, -2, 2, false, ttnpb.TxSettings{}),
					newADRUplink(44, -3, 2, false, ttnpb.TxSettings{}),
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_5,
						ADRTxPowerIndex:  42,
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							nil,
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: false},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_10,
							TxPowerIndex:  43,
							ChannelMask: []bool{
								false, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
					},
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveLinkADRAccept.BindData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				}),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleLinkADRAns(test.Context(), dev, tc.Payload, tc.DupCount, frequencyplans.NewStore(test.FrequencyPlansFetcher))
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
