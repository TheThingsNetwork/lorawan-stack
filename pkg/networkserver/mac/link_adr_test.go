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
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestLinkADRReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name                                             string
		BandID                                           string
		LoRaWANVersion                                   ttnpb.MACVersion
		LoRaWANPHYVersion                                ttnpb.PHYVersion
		CurrentChannels, DesiredChannels                 []*ttnpb.MACParameters_Channel
		CurrentADRDataRateIndex, DesiredADRDataRateIndex ttnpb.DataRateIndex
		CurrentADRTxPowerIndex, DesiredADRTxPowerIndex   uint32
		CurrentADRNbTrans, DesiredADRNbTrans             uint32
		RejectedADRDataRateIndexes                       []ttnpb.DataRateIndex
		RejectedADRTxPowerIndexes                        []uint32
		Commands                                         []*ttnpb.MACCommand_LinkADRReq
		EventBuildersAssertion                           func(*testing.T, events.Builders) bool
	}{
		{
			Name:              "no channels",
			BandID:            band.US_902_928,
			LoRaWANVersion:    ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion: ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentADRNbTrans: 1,
			DesiredADRNbTrans: 1,
		},
		{
			Name:              "invalid channel",
			BandID:            band.US_902_928,
			LoRaWANVersion:    ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion: ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:   MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans: 1,
			DesiredChannels: []*ttnpb.MACParameters_Channel{
				{EnableUplink: true},
			},
			DesiredADRNbTrans: 1,
			EventBuildersAssertion: func(t *testing.T, bs events.Builders) bool {
				t.Helper()
				a, _ := test.New(t)
				return a.So(bs, should.ResembleEventBuilders, events.Builders{
					EvtGenerateLinkADRFail.BindData(ErrNoUplinkFrequency.WithAttributes(
						"parameters", "desired",
						"i", 0,
					)),
				})
			},
		},
		{
			Name:              "invalid channel count",
			BandID:            band.EU_863_870,
			LoRaWANVersion:    ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion: ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:   MakeDefaultEU868CurrentChannels(),
			CurrentADRNbTrans: 1,
			DesiredChannels:   MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans: 1,
			EventBuildersAssertion: func(t *testing.T, bs events.Builders) bool {
				t.Helper()
				a, _ := test.New(t)
				return a.So(bs, should.ResembleEventBuilders, events.Builders{
					EvtGenerateLinkADRFail.BindData(ErrTooManyChannels.WithAttributes(
						"parameters", "desired",
						"channels_len", 72,
						"phy_max_uplink_channels", uint8(16),
					)),
				})
			},
		},
		{
			Name:              "invalid band channels",
			BandID:            band.EU_863_870,
			LoRaWANVersion:    ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion: ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:   MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans: 1,
			DesiredChannels:   MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans: 1,
			EventBuildersAssertion: func(t *testing.T, bs events.Builders) bool {
				t.Helper()
				a, _ := test.New(t)
				return a.So(bs, should.ResembleEventBuilders, events.Builders{
					EvtGenerateLinkADRFail.BindData(ErrTooManyChannels.WithAttributes(
						"parameters", "current",
						"channels_len", 72,
						"phy_max_uplink_channels", uint8(16),
					)),
				})
			},
		},
		{
			Name:                    "non-existent data rate",
			BandID:                  band.EU_863_870,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:         MakeDefaultEU868DesiredChannels(),
			CurrentADRNbTrans:       1,
			DesiredChannels:         MakeDefaultEU868DesiredChannels(),
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_15,
			DesiredADRNbTrans:       1,
			EventBuildersAssertion: func(t *testing.T, bs events.Builders) bool {
				t.Helper()
				a, _ := test.New(t)
				return a.So(bs, should.ResembleEventBuilders, events.Builders{
					EvtGenerateLinkADRFail.BindData(ErrInvalidDataRateIndex.WithAttributes(
						"desired_adr_data_rate_index", ttnpb.DataRateIndex(15),
					)),
				})
			},
		},
		{
			Name:                   "TX power too high",
			BandID:                 band.EU_863_870,
			LoRaWANVersion:         ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion:      ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:        MakeDefaultEU868DesiredChannels(),
			CurrentADRNbTrans:      1,
			DesiredChannels:        MakeDefaultEU868DesiredChannels(),
			DesiredADRTxPowerIndex: 14,
			DesiredADRNbTrans:      1,
			EventBuildersAssertion: func(t *testing.T, bs events.Builders) bool {
				t.Helper()
				a, _ := test.New(t)
				return a.So(bs, should.ResembleEventBuilders, events.Builders{
					EvtGenerateLinkADRFail.BindData(ErrTxPowerIndexTooHigh.WithAttributes(
						"desired_adr_tx_power_index", uint32(14),
						"phy_max_tx_power_index", uint8(7),
					)),
				})
			},
		},
		{
			Name:              "ABP channel setup",
			BandID:            band.US_902_928,
			LoRaWANVersion:    ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion: ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:   MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans: 1,
			DesiredChannels:   MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans: 1,
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					NbTrans:            1,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					NbTrans: 1,
				},
			},
		},
		{
			Name:              "ABP channel setup",
			BandID:            band.US_902_928,
			LoRaWANVersion:    ttnpb.MACVersion_MAC_V1_1,
			LoRaWANPHYVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
			CurrentChannels:   MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans: 1,
			DesiredChannels:   MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans: 1,
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_15,
					TxPowerIndex:       15,
					NbTrans:            1,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_15,
					TxPowerIndex:  15,
					NbTrans:       1,
				},
			},
		},
		{
			Name:                    "ADR",
			BandID:                  band.US_902_928,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			CurrentADRNbTrans:       1,
			DesiredChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
			DesiredADRTxPowerIndex:  1,
			DesiredADRNbTrans:       2,
			RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
				ttnpb.DataRateIndex_DATA_RATE_2,
			},
			RejectedADRTxPowerIndexes: []uint32{
				0,
				1,
			},
			EventBuildersAssertion: func(t *testing.T, bs events.Builders) bool {
				t.Helper()
				a, _ := test.New(t)
				return a.So(bs, should.ResembleEventBuilders, events.Builders{
					EvtGenerateLinkADRFail.BindData(ErrRejectedParameters.WithAttributes(
						"parameters", "current",
						"data_rate_index", ttnpb.DataRateIndex_DATA_RATE_0,
						"tx_power_index", uint32(0),
					)),
				})
			},
		},
		{
			Name:                    "ADR",
			BandID:                  band.EU_863_870,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_1,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_TS001_V1_0_1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			CurrentADRNbTrans:       1,
			CurrentChannels:         MakeDefaultEU868DesiredChannels(),
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			DesiredADRNbTrans:       2,
			DesiredADRTxPowerIndex:  3,
			DesiredChannels:         MakeDefaultEU868DesiredChannels(),
			RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
				ttnpb.DataRateIndex_DATA_RATE_1,
				ttnpb.DataRateIndex_DATA_RATE_2,
				ttnpb.DataRateIndex_DATA_RATE_3,
				ttnpb.DataRateIndex_DATA_RATE_4,
			},
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
					TxPowerIndex:  3,
					NbTrans:       2,
				},
			},
		},
		{
			Name:                    "ADR",
			BandID:                  band.EU_863_870,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_1,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_TS001_V1_0_1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			CurrentADRNbTrans:       1,
			CurrentChannels:         MakeDefaultEU868DesiredChannels(),
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			DesiredADRNbTrans:       1,
			DesiredADRTxPowerIndex:  3,
			DesiredChannels:         MakeDefaultEU868DesiredChannels(),
			RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
				ttnpb.DataRateIndex_DATA_RATE_2,
				ttnpb.DataRateIndex_DATA_RATE_3,
				ttnpb.DataRateIndex_DATA_RATE_4,
			},
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					TxPowerIndex:  3,
					NbTrans:       1,
				},
			},
		},
		{
			Name:                    "ABP channel setup + ADR",
			BandID:                  band.US_902_928,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_3,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:         MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans:       1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			DesiredChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans:       2,
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
			DesiredADRTxPowerIndex:  3,
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_2,
					TxPowerIndex:       3,
					NbTrans:            2,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
					TxPowerIndex:  3,
					NbTrans:       2,
				},
			},
		},
		{
			Name:                    "fallback to current indices on data rate rejection",
			BandID:                  band.US_902_928,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_2,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
			CurrentChannels:         MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans:       1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			DesiredChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans:       2,
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
			DesiredADRTxPowerIndex:  3,
			RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
				ttnpb.DataRateIndex_DATA_RATE_2,
			},
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_1,
					NbTrans:            2,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					NbTrans:       2,
				},
			},
		},
		{
			Name:                    "fallback to current indices on TX power rejection",
			BandID:                  band.US_902_928,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_2,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
			CurrentChannels:         MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans:       1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			DesiredChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans:       2,
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
			DesiredADRTxPowerIndex:  3,
			RejectedADRTxPowerIndexes: []uint32{
				3,
			},
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_1,
					NbTrans:            2,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					NbTrans:       2,
				},
			},
		},
		{
			Name:                    "fallback to no-change indices on data rate rejection",
			BandID:                  band.US_902_928,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_4,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:         MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans:       1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			DesiredChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans:       2,
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
			DesiredADRTxPowerIndex:  3,
			RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
				ttnpb.DataRateIndex_DATA_RATE_2,
			},
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_15,
					TxPowerIndex:       15,
					NbTrans:            2,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_15,
					TxPowerIndex:  15,
					NbTrans:       2,
				},
			},
		},
		{
			Name:                    "fallback to no-change indices on TX power rejection",
			BandID:                  band.US_902_928,
			LoRaWANVersion:          ttnpb.MACVersion_MAC_V1_0_4,
			LoRaWANPHYVersion:       ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
			CurrentChannels:         MakeDefaultUS915CurrentChannels(),
			CurrentADRNbTrans:       1,
			CurrentADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			DesiredChannels:         MakeDefaultUS915FSB2DesiredChannels(),
			DesiredADRNbTrans:       2,
			DesiredADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
			DesiredADRTxPowerIndex:  3,
			RejectedADRTxPowerIndexes: []uint32{
				3,
			},
			Commands: []*ttnpb.MACCommand_LinkADRReq{
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
					ChannelMaskControl: 7,
					DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_15,
					TxPowerIndex:       15,
					NbTrans:            2,
				},
				{
					ChannelMask: []bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_15,
					TxPowerIndex:  15,
					NbTrans:       2,
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name: fmt.Sprintf("%s/band:%s,MAC:%s,PHY:%s,DR:%d->%d,TX:%d->%d,NB:%d->%d,rejected_DR:%v,rejected_TX:%v",
				tc.Name,
				tc.BandID,
				tc.LoRaWANVersion,
				tc.LoRaWANPHYVersion,
				tc.CurrentADRDataRateIndex, tc.DesiredADRDataRateIndex,
				tc.CurrentADRTxPowerIndex, tc.DesiredADRTxPowerIndex,
				tc.CurrentADRNbTrans, tc.DesiredADRNbTrans,
				fmt.Sprintf("[%s]", test.JoinStringsf("%d", ",", false, tc.RejectedADRDataRateIndexes)),
				fmt.Sprintf("[%s]", test.JoinStringsf("%d", ",", false, tc.RejectedADRTxPowerIndexes)),
			),
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				makeDevice := func() *ttnpb.EndDevice {
					return ttnpb.Clone(&ttnpb.EndDevice{
						MacState: &ttnpb.MACState{
							LorawanVersion: tc.LoRaWANVersion,
							CurrentParameters: &ttnpb.MACParameters{
								Channels:         tc.CurrentChannels,
								AdrDataRateIndex: tc.CurrentADRDataRateIndex,
								AdrTxPowerIndex:  tc.CurrentADRTxPowerIndex,
								AdrNbTrans:       tc.CurrentADRNbTrans,
							},
							DesiredParameters: &ttnpb.MACParameters{
								Channels:         tc.DesiredChannels,
								AdrDataRateIndex: tc.DesiredADRDataRateIndex,
								AdrTxPowerIndex:  tc.DesiredADRTxPowerIndex,
								AdrNbTrans:       tc.DesiredADRNbTrans,
							},
							RejectedAdrDataRateIndexes: tc.RejectedADRDataRateIndexes,
							RejectedAdrTxPowerIndexes:  tc.RejectedADRTxPowerIndexes,
						},
					})
				}
				phy := LoRaWANBands[tc.BandID][tc.LoRaWANPHYVersion]

				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name:     "DeviceNeedsLinkADRReq",
					Parallel: true,
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						dev := makeDevice()
						a.So(DeviceNeedsLinkADRReq(ctx, dev, phy), func() func(interface{}, ...interface{}) string {
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
					default:
						return []int{0, len(tc.Commands)}
					}
				}() {
					cmdsFit := n >= len(tc.Commands)
					cmdLen := (1 + lorawan.DefaultMACCommands[ttnpb.MACCommandIdentifier_CID_LINK_ADR].DownlinkLength) * uint16(n)
					cmds := tc.Commands[:n]
					answerLen := (1 + lorawan.DefaultMACCommands[ttnpb.MACCommandIdentifier_CID_LINK_ADR].UplinkLength) * func() uint16 {
						switch {
						case n == 0:
							return 0
						case macspec.SingularLinkADRAns(tc.LoRaWANVersion):
							return 1
						default:
							return uint16(n)
						}
					}()
					test.RunSubtestFromContext(ctx, test.SubtestConfig{
						Name:     fmt.Sprintf("EnqueueLinkADRReq/max_down_len:%d", cmdLen),
						Parallel: true,
						Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
							dev := makeDevice()
							st := EnqueueLinkADRReq(ctx, dev, cmdLen, answerLen, phy)
							if tc.EventBuildersAssertion != nil {
								if !a.So(tc.EventBuildersAssertion(t, st.QueuedEvents), should.BeTrue) {
									t.FailNow()
								}
								a.So(
									EnqueueState{
										MaxDownLen: st.MaxDownLen,
										MaxUpLen:   st.MaxUpLen,
										Ok:         st.Ok,
									},
									should.Resemble,
									EnqueueState{
										MaxDownLen: cmdLen,
										MaxUpLen:   answerLen,
										Ok:         true,
									})
								return
							}
							expectedDevice := makeDevice()
							var expectedEventBuilders []events.Builder
							for _, cmd := range cmds {
								expectedDevice.MacState.PendingRequests = append(expectedDevice.MacState.PendingRequests, cmd.MACCommand())
								expectedEventBuilders = append(expectedEventBuilders, EvtEnqueueLinkADRRequest.BindData(cmd))
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

func TestHandleLinkADRAns(t *testing.T) {
	t.Parallel()
	const fCntUp = 42
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
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
				},
			},
			Error: ErrNoPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				EvtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_LINK_ADR),
		},
		{
			Name: "1 request/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							nil,
							{UplinkFrequency: 42},
							{DownlinkFrequency: 23},
							nil,
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
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
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
						AdrTxPowerIndex:  42,
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
					PendingRequests:     []*ttnpb.MACCommand{},
					LastAdrChangeFCntUp: fCntUp,
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				EvtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
			},
		},
		{
			Name: "1.1/2 requests/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
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
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
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
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
						AdrTxPowerIndex:  43,
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
					PendingRequests:     []*ttnpb.MACCommand{},
					LastAdrChangeFCntUp: fCntUp,
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				EvtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
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
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0_2,
					CurrentParameters: &ttnpb.MACParameters{
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
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
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
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0_2,
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
						AdrTxPowerIndex:  43,
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
					PendingRequests:     []*ttnpb.MACCommand{},
					LastAdrChangeFCntUp: fCntUp,
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				EvtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
			},
		},
		{
			Name: "1.0/2 requests/all ack",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0,
					CurrentParameters: &ttnpb.MACParameters{
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
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
							TxPowerIndex:  42,
							ChannelMask: []bool{
								true, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
							TxPowerIndex:  43,
							ChannelMask: []bool{
								false, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
					},
					LastAdrChangeFCntUp: fCntUp - 3,
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0,
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
						AdrTxPowerIndex:  42,
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
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
							TxPowerIndex:  43,
							ChannelMask: []bool{
								false, true, true, false,
								true, true, true, true,
								true, true, true, true,
								true, true, false, false,
							},
						}).MACCommand(),
					},
					LastAdrChangeFCntUp: fCntUp,
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				EvtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
			},
		},
		{
			Name: "1.0.2/2 requests/US915 FSB2",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.USFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_2,
					CurrentParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					DesiredParameters: MakeDefaultUS915FSB2DesiredMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      ttnpb.DataRateIndex_DATA_RATE_3,
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
							DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
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
					LastAdrChangeFCntUp: fCntUp - 2,
				},
			},
			Expected: &ttnpb.EndDevice{
				FrequencyPlanId:   test.USFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_0_2,
					CurrentParameters: func() *ttnpb.MACParameters {
						params := MakeDefaultUS915FSB2DesiredMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B)
						params.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_3
						params.AdrTxPowerIndex = 1
						params.AdrNbTrans = 3
						return params
					}(),
					DesiredParameters:   MakeDefaultUS915FSB2DesiredMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					PendingRequests:     []*ttnpb.MACCommand{},
					LastAdrChangeFCntUp: fCntUp,
				},
			},
			Payload: &ttnpb.MACCommand_LinkADRAns{
				ChannelMaskAck:   true,
				DataRateIndexAck: true,
				TxPowerIndexAck:  true,
			},
			Events: events.Builders{
				EvtReceiveLinkADRAccept.With(events.WithData(&ttnpb.MACCommand_LinkADRAns{
					ChannelMaskAck:   true,
					DataRateIndexAck: true,
					TxPowerIndexAck:  true,
				})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)

				evs, err := HandleLinkADRAns(ctx, dev, tc.Payload, tc.DupCount, fCntUp, frequencyplans.NewStore(test.FrequencyPlansFetcher))
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
