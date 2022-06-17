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
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHandleLinkCheckReq(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Message          *ttnpb.UplinkMessage
		Events           events.Builders
		Error            error
	}{
		{
			Name: "SF13BW250",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 13,
								Bandwidth:       250000,
							},
						},
					},
				},
			},
			Events: events.Builders{
				EvtReceiveLinkCheckRequest,
			},
			Error: ErrInvalidDataRate,
		},
		{
			Name: "SF12BW250/no gateways",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 12,
								Bandwidth:       250000,
							},
						},
					},
				},
			},
			Events: events.Builders{
				EvtReceiveLinkCheckRequest,
			},
		},
		{
			Name: "SF12BW250/1 gateway/empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkCheckAns{
							Margin:       42, // 25-(-17)
							GatewayCount: 1,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 12,
								Bandwidth:       250000,
							},
						},
					},
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test",
						},
						Snr: 25,
					},
				},
			},
			Events: events.Builders{
				EvtReceiveLinkCheckRequest,
				EvtEnqueueLinkCheckAnswer.With(events.WithData(&ttnpb.MACCommand_LinkCheckAns{
					Margin:       42,
					GatewayCount: 1,
				})),
			},
		},
		{
			Name: "SF12BW250/1 gateway/non-empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						(&ttnpb.MACCommand_LinkCheckAns{
							Margin:       42, // 25-(-17)
							GatewayCount: 1,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 12,
								Bandwidth:       250000,
							},
						},
					},
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test",
						},
						Snr: 25,
					},
				},
			},
			Events: events.Builders{
				EvtReceiveLinkCheckRequest,
				EvtEnqueueLinkCheckAnswer.With(events.WithData(&ttnpb.MACCommand_LinkCheckAns{
					Margin:       42,
					GatewayCount: 1,
				})),
			},
		},
		{
			Name: "SF12BW250/3 gateways/non-empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						(&ttnpb.MACCommand_LinkCheckAns{
							Margin:       42, // 25-(-17)
							GatewayCount: 3,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 12,
								Bandwidth:       250000,
							},
						},
					},
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test",
						},
						Snr: 24,
					},
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test2",
						},
						Snr: 25,
					},
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test3",
						},
						Snr: 2,
					},
				},
			},
			Events: events.Builders{
				EvtReceiveLinkCheckRequest,
				EvtEnqueueLinkCheckAnswer.With(events.WithData(&ttnpb.MACCommand_LinkCheckAns{
					Margin:       42,
					GatewayCount: 3,
				})),
			},
		},
		{
			Name: "SF12BW250/3 gateways + Packet Broker/non-empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						(&ttnpb.MACCommand_LinkCheckAns{
							Margin:       43, // 26-(-17)
							GatewayCount: 4,
						}).MACCommand(),
					},
				},
			},
			Message: &ttnpb.UplinkMessage{
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 12,
								Bandwidth:       250000,
							},
						},
					},
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test",
						},
						Snr: 24,
					},
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test2",
						},
						Snr: 25,
					},
					{
						GatewayIds: cluster.PacketBrokerGatewayID,
						PacketBroker: &ttnpb.PacketBrokerMetadata{
							ForwarderNetId:     types.NetID{0x0, 0x0, 0x42}.Bytes(),
							ForwarderTenantId:  "test",
							ForwarderClusterId: "test",
						},
						Snr: 26,
					},
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test3",
						},
						Snr: 2,
					},
				},
			},
			Events: events.Builders{
				EvtReceiveLinkCheckRequest,
				EvtEnqueueLinkCheckAnswer.With(events.WithData(&ttnpb.MACCommand_LinkCheckAns{
					Margin:       43,
					GatewayCount: 4,
				})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)

				evs, err := HandleLinkCheckReq(ctx, dev, tc.Message)
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
