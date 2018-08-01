// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleLinkADRAns(t *testing.T) {
	events := collectEvents("ns.mac.adr.accept")

	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_LinkADRAns
		Error            error
		ExpectedEvents   int
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState:        &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState:        &ttnpb.MACState{},
			},
			Payload: nil,
			Error:   errMissingPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState:        &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState:        &ttnpb.MACState{},
			},
			Payload: ttnpb.NewPopulatedMACCommand_LinkADRAns(test.Randy, false),
			Error:   errMACRequestNotFound,
		},
		{
			Name: "1 request/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					MACParameters: ttnpb.MACParameters{
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
								true, true, false, true,
								false, true, true, true,
								true, false, true, true,
								true, true, false, true,
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					MACParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_4,
						ADRTxPowerIndex:  42,
						Channels: []*ttnpb.MACParameters_Channel{
							{UplinkEnabled: true},
							{
								UplinkEnabled:   true,
								UplinkFrequency: 42,
							},
							{
								UplinkEnabled:     false,
								DownlinkFrequency: 23,
							},
							{UplinkEnabled: true},
							{UplinkEnabled: false},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: false},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: false},
							{UplinkEnabled: true},
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
			ExpectedEvents: 1,
		},
		{
			Name: "2 requests/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, true,
								true, true, true, true,
								true, true, true, true,
								true, true, false, true,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: 5,
							TxPowerIndex:  43,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					MACParameters: ttnpb.MACParameters{
						ADRDataRateIndex: 5,
						ADRTxPowerIndex:  42,
						Channels: []*ttnpb.MACParameters_Channel{
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: true},
							{UplinkEnabled: false},
							{UplinkEnabled: true},
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
			ExpectedEvents: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleLinkADRAns(test.Context(), dev, tc.Payload, frequencyPlansStore)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}

			if tc.ExpectedEvents > 0 {
				events.expect(t, tc.ExpectedEvents)
			}
			a.So(dev, should.Resemble, tc.Expected)
		})
	}
}
