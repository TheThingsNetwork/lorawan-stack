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
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsLinkADRReq(t *testing.T) {
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
			Name: "current(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[]),desired(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
					},
				},
			},
		},
		{
			Name: "current(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[on,on,off]),desired(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[on,on,off])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{},
						},
					},
				},
			},
		},
		{
			Name: "current(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[]),desired(data-rate-index:1,nb-trans:2,tx-power-index:4,channels:[])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  4,
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[]),desired(data-rate-index:1,nb-trans:3,tx-power-index:3,channels:[])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       3,
						ADRTxPowerIndex:  3,
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[on,on,on]),desired(data-rate-index:1,nb-trans:2,tx-power-index:3,channels:[off,on,off])",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
						Channels: []*ttnpb.MACParameters_Channel{
							{EnableUplink: true},
							{EnableUplink: true},
							{EnableUplink: true},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_1,
						ADRNbTrans:       2,
						ADRTxPowerIndex:  3,
						Channels: []*ttnpb.MACParameters_Channel{
							{},
							{EnableUplink: true},
							{},
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
			res := deviceNeedsLinkADRReq(dev, DefaultConfig.DefaultMACSettings.Parse(), LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_0_3_REV_A])
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestEnqueueLinkADRReq(t *testing.T) {
	for _, tc := range []struct {
		Name                        string
		Band                        *band.Band
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       macCommandEnqueueState
		ErrorAssertion              func(*testing.T, error) bool
	}{
		{
			Name: "payload fits/US915 FSB2/MAC:1.0.3,PHY:1.0.3a",
			Band: LoRaWANBands[band.US_902_928][ttnpb.PHY_V1_0_3_REV_A],
			InputDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState:        MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A),
			},
			ExpectedDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState: func() *ttnpb.MACState {
					macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A)
					macState.PendingRequests = []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							ChannelMask: []bool{
								false, false, false, false, false, false, false, false,
								false, false, false, false, false, false, false, false,
							},
							ChannelMaskControl: 7,
							NbTrans:            1,
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							ChannelMask: []bool{
								false, false, false, false, false, false, false, false,
								true, true, true, true, true, true, true, true,
							},
							NbTrans: 1,
						}).MACCommand(),
					}
					return macState
				}(),
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 32,
				MaxUpLen:   20,
				Ok:         true,
				QueuedEvents: events.Builders{
					evtEnqueueLinkADRRequest.With(events.WithData(&ttnpb.MACCommand_LinkADRReq{
						ChannelMask: []bool{
							false, false, false, false, false, false, false, false,
							false, false, false, false, false, false, false, false,
						},
						ChannelMaskControl: 7,
						NbTrans:            1,
					})),
					evtEnqueueLinkADRRequest.With(events.WithData(&ttnpb.MACCommand_LinkADRReq{
						ChannelMask: []bool{
							false, false, false, false, false, false, false, false,
							true, true, true, true, true, true, true, true,
						},
						NbTrans: 1,
					})),
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "payload fits/US915 FSB2/MAC:1.0.3,PHY:1.0.3a/ADR/rejected desired data rate and TX power",
			Band: LoRaWANBands[band.US_902_928][ttnpb.PHY_V1_0_3_REV_A],
			InputDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState: func() *ttnpb.MACState {
					macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A)
					macState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_3
					macState.DesiredParameters.ADRTxPowerIndex = 1
					macState.RejectedADRDataRateIndexes = []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_2,
					}
					macState.RejectedADRTxPowerIndexes = []uint32{
						0,
						1,
					}
					return macState
				}(),
			},
			ExpectedDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState: func() *ttnpb.MACState {
					macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A)
					macState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_3
					macState.DesiredParameters.ADRTxPowerIndex = 1
					macState.RejectedADRDataRateIndexes = []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_2,
					}
					macState.RejectedADRTxPowerIndexes = []uint32{
						0,
						1,
					}
					macState.PendingRequests = []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							ChannelMask: []bool{
								false, false, false, false, false, false, false, false,
								false, false, false, false, false, false, false, false,
							},
							ChannelMaskControl: 7,
							NbTrans:            1,
							DataRateIndex:      ttnpb.DATA_RATE_1,
							TxPowerIndex:       15,
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							ChannelMask: []bool{
								false, false, false, false, false, false, false, false,
								true, true, true, true, true, true, true, true,
							},
							NbTrans:       1,
							DataRateIndex: ttnpb.DATA_RATE_1,
							TxPowerIndex:  15,
						}).MACCommand(),
					}
					return macState
				}(),
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 32,
				MaxUpLen:   20,
				Ok:         true,
				QueuedEvents: events.Builders{
					evtEnqueueLinkADRRequest.With(events.WithData(&ttnpb.MACCommand_LinkADRReq{
						ChannelMask: []bool{
							false, false, false, false, false, false, false, false,
							false, false, false, false, false, false, false, false,
						},
						ChannelMaskControl: 7,
						NbTrans:            1,
						DataRateIndex:      ttnpb.DATA_RATE_1,
						TxPowerIndex:       15,
					})),
					evtEnqueueLinkADRRequest.With(events.WithData(&ttnpb.MACCommand_LinkADRReq{
						ChannelMask: []bool{
							false, false, false, false, false, false, false, false,
							true, true, true, true, true, true, true, true,
						},
						NbTrans:       1,
						DataRateIndex: ttnpb.DATA_RATE_1,
						TxPowerIndex:  15,
					})),
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "payload fits/EU868/MAC:1.0.1,PHY:1.0.1/ADR/rejected all possible data rates",
			Band: LoRaWANBands[band.EU_863_870][ttnpb.PHY_V1_0_1],
			InputDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState: func() *ttnpb.MACState {
					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_1, ttnpb.PHY_V1_0_1)
					macState.CurrentParameters.ADRDataRateIndex = ttnpb.DATA_RATE_1
					macState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_5
					macState.DesiredParameters.ADRTxPowerIndex = 3
					macState.RejectedADRDataRateIndexes = []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_1,
						ttnpb.DATA_RATE_2,
						ttnpb.DATA_RATE_3,
						ttnpb.DATA_RATE_4,
						ttnpb.DATA_RATE_5,
					}
					return macState
				}(),
			},
			ExpectedDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState: func() *ttnpb.MACState {
					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_1, ttnpb.PHY_V1_0_1)
					macState.CurrentParameters.ADRDataRateIndex = ttnpb.DATA_RATE_1
					macState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_5
					macState.DesiredParameters.ADRTxPowerIndex = 3
					macState.RejectedADRDataRateIndexes = []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_1,
						ttnpb.DATA_RATE_2,
						ttnpb.DATA_RATE_3,
						ttnpb.DATA_RATE_4,
						ttnpb.DATA_RATE_5,
					}
					return macState
				}(),
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 42,
				MaxUpLen:   24,
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "payload fits/US915 FSB2/MAC:1.0.4,PHY:1.0.3a/no data rate change",
			Band: LoRaWANBands[band.US_902_928][ttnpb.PHY_V1_0_3_REV_A],
			InputDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState:        MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_4, ttnpb.PHY_V1_0_3_REV_A),
			},
			ExpectedDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState: func() *ttnpb.MACState {
					macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_4, ttnpb.PHY_V1_0_3_REV_A)
					macState.PendingRequests = []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							ChannelMask: []bool{
								false, false, false, false, false, false, false, false,
								false, false, false, false, false, false, false, false,
							},
							ChannelMaskControl: 7,
							NbTrans:            1,
							DataRateIndex:      ttnpb.DATA_RATE_15,
							TxPowerIndex:       15,
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							ChannelMask: []bool{
								false, false, false, false, false, false, false, false,
								true, true, true, true, true, true, true, true,
							},
							NbTrans:       1,
							DataRateIndex: ttnpb.DATA_RATE_15,
							TxPowerIndex:  15,
						}).MACCommand(),
					}
					return macState
				}(),
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 32,
				MaxUpLen:   20,
				Ok:         true,
				QueuedEvents: events.Builders{
					evtEnqueueLinkADRRequest.With(events.WithData(&ttnpb.MACCommand_LinkADRReq{
						ChannelMask: []bool{
							false, false, false, false, false, false, false, false,
							false, false, false, false, false, false, false, false,
						},
						ChannelMaskControl: 7,
						NbTrans:            1,
						DataRateIndex:      ttnpb.DATA_RATE_15,
						TxPowerIndex:       15,
					})),
					evtEnqueueLinkADRRequest.With(events.WithData(&ttnpb.MACCommand_LinkADRReq{
						ChannelMask: []bool{
							false, false, false, false, false, false, false, false,
							true, true, true, true, true, true, true, true,
						},
						NbTrans:       1,
						DataRateIndex: ttnpb.DATA_RATE_15,
						TxPowerIndex:  15,
					})),
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "downlink does not fit/US915 FSB2/MAC:1.0.3,PHY:1.0.3a",
			Band: LoRaWANBands[band.US_902_928][ttnpb.PHY_V1_0_3_REV_A],
			InputDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState:        MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A),
			},
			ExpectedDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState:        MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A),
			},
			MaxDownlinkLength: 7,
			MaxUplinkLength:   24,
			State: macCommandEnqueueState{
				MaxDownLen: 7,
				MaxUpLen:   24,
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
		{
			Name: "uplink does not fit/US915 FSB2/MAC:1.1,PHY:1.1b",
			Band: LoRaWANBands[band.US_902_928][ttnpb.PHY_V1_1_REV_B],
			InputDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState:        MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B),
			},
			ExpectedDevice: &ttnpb.EndDevice{
				FrequencyPlanID: test.USFrequencyPlanID,
				MACState:        MakeDefaultUS915FSB2MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B),
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   1,
			State: macCommandEnqueueState{
				MaxDownLen: 42,
				MaxUpLen:   1,
			},
			ErrorAssertion: func(t *testing.T, err error) bool { return assertions.New(t).So(err, should.BeNil) },
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)

			st, err := enqueueLinkADRReq(log.NewContext(test.Context(), test.GetLogger(t)), dev, tc.MaxDownlinkLength, tc.MaxUplinkLength, ttnpb.MACSettings{}, tc.Band)
			if !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.ExpectedDevice)
			a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
			st.QueuedEvents = tc.State.QueuedEvents
			a.So(st, should.Resemble, tc.State)
		})
	}
}

func TestHandleLinkADRAns(t *testing.T) {
	recentADRUplinks := []*ttnpb.UplinkMessage{
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
	}

	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_LinkADRAns
		DupCount         uint
		Events           events.Builders
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
				RecentADRUplinks: recentADRUplinks,
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				RecentADRUplinks: recentADRUplinks,
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
				RecentADRUplinks: recentADRUplinks,
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				RecentADRUplinks: recentADRUplinks,
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				evtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
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
				RecentADRUplinks: recentADRUplinks,
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
			Events: events.Builders{
				evtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
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
				RecentADRUplinks: recentADRUplinks,
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
			Events: events.Builders{
				evtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
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
				RecentADRUplinks: recentADRUplinks,
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
			Events: events.Builders{
				evtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
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
				RecentADRUplinks: recentADRUplinks,
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
			Events: events.Builders{
				evtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
			},
		},
		{
			Name: "1.0.2/2 requests/US915 FSB2",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					CurrentParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					DesiredParameters: MakeDefaultUS915FSB2DesiredMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      ttnpb.DATA_RATE_3,
							TxPowerIndex:       1,
							ChannelMaskControl: 7,
							NbTrans:            3,
							ChannelMask: []bool{
								false, false, false, false,
								false, false, false, false,
								false, false, false, false,
								false, false, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_3,
							TxPowerIndex:  1,
							NbTrans:       3,
							ChannelMask: []bool{
								false, false, false, false,
								false, false, false, false,
								true, true, true, true,
								true, true, true, true,
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_0_2,
					CurrentParameters: func() ttnpb.MACParameters {
						params := MakeDefaultUS915FSB2DesiredMACParameters(ttnpb.PHY_V1_0_2_REV_B)
						params.ADRDataRateIndex = ttnpb.DATA_RATE_3
						params.ADRTxPowerIndex = 1
						params.ADRNbTrans = 3
						return params
					}(),
					DesiredParameters: MakeDefaultUS915FSB2DesiredMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					PendingRequests:   []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				evtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
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
			a.So(evs, should.ResembleEventBuilders, tc.Events)
		})
	}
}
