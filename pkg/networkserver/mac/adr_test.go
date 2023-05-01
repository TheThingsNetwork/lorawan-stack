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
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestAdaptDataRate(t *testing.T) {
	t.Parallel()

	semtechPaperUplinks := ADRMatrixToUplinks([]ADRMatrixRow{
		{FCnt: 10, MaxSNR: -6, GtwDiversity: 2},
		{FCnt: 11, MaxSNR: -7, GtwDiversity: 2},
		{FCnt: 12, MaxSNR: -25, GtwDiversity: 1},
		{FCnt: 13, MaxSNR: -25, GtwDiversity: 1},
		{FCnt: 14, MaxSNR: -10, GtwDiversity: 2},
		{FCnt: 16, MaxSNR: -25, GtwDiversity: 1},
		{FCnt: 17, MaxSNR: -10, GtwDiversity: 2},
		{FCnt: 19, MaxSNR: -10, GtwDiversity: 3},
		{FCnt: 20, MaxSNR: -6, GtwDiversity: 2},
		{FCnt: 21, MaxSNR: -7, GtwDiversity: 2},
		{FCnt: 22, MaxSNR: -25, GtwDiversity: 0},
		{FCnt: 23, MaxSNR: -25, GtwDiversity: 1},
		{FCnt: 24, MaxSNR: -10, GtwDiversity: 2},
		{FCnt: 25, MaxSNR: -10, GtwDiversity: 2},
		{FCnt: 26, MaxSNR: -25, GtwDiversity: 1},
		{FCnt: 27, MaxSNR: -8, GtwDiversity: 2},
		{FCnt: 28, MaxSNR: -10, GtwDiversity: 2},
		{FCnt: 29, MaxSNR: -10, GtwDiversity: 3},
		{FCnt: 30, MaxSNR: -9, GtwDiversity: 3},
		{
			FCnt: 31, MaxSNR: -7, GtwDiversity: 2,
			TxSettings: &ttnpb.TxSettings{
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							SpreadingFactor: 12,
							Bandwidth:       125000,
							CodingRate:      band.Cr4_5,
						},
					},
				},
			},
		},
	})
	for _, tc := range []struct {
		Name       string
		Device     *ttnpb.EndDevice
		DeviceDiff func(*ttnpb.EndDevice)
		Error      error
	}{
		{
			Name: "adapted example from Semtech paper/no rejections",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrNbTrans:      1,
						AdrTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
						AdrNbTrans:       3,
						AdrTxPowerIndex:  2,
						Channels:         MakeDefaultEU868CurrentChannels(),
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MacSettings: &ttnpb.MACSettings{
					AdrMargin: &wrapperspb.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_4
				dev.MacState.DesiredParameters.AdrTxPowerIndex = 1
				dev.MacState.DesiredParameters.AdrNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected DR:(1,4)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrNbTrans:      1,
						AdrTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedAdrDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DataRateIndex_DATA_RATE_1, ttnpb.DataRateIndex_DATA_RATE_4,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MacSettings: &ttnpb.MACSettings{
					AdrMargin: &wrapperspb.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_3
				dev.MacState.DesiredParameters.AdrTxPowerIndex = 2
				dev.MacState.DesiredParameters.AdrNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected TXPower:(1)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrNbTrans:      1,
						AdrTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedAdrTxPowerIndexes: []uint32{
						1,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MacSettings: &ttnpb.MACSettings{
					AdrMargin: &wrapperspb.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_4
				dev.MacState.DesiredParameters.AdrTxPowerIndex = 0
				dev.MacState.DesiredParameters.AdrNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected TXPower:(0,1)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrNbTrans:      1,
						AdrTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedAdrTxPowerIndexes: []uint32{
						0, 1,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MacSettings: &ttnpb.MACSettings{
					AdrMargin: &wrapperspb.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_3
				dev.MacState.DesiredParameters.AdrTxPowerIndex = 2
				dev.MacState.DesiredParameters.AdrNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected DR:(1,4), rejected TXPower:(0,2,3)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrNbTrans:      1,
						AdrTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedAdrTxPowerIndexes: []uint32{
						0, 2, 3,
					},
					RejectedAdrDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DataRateIndex_DATA_RATE_1, ttnpb.DataRateIndex_DATA_RATE_4,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MacSettings: &ttnpb.MACSettings{
					AdrMargin: &wrapperspb.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_3
				dev.MacState.DesiredParameters.AdrTxPowerIndex = 1
				dev.MacState.DesiredParameters.AdrNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected DR:(3), rejected TXPower:(0,1)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.EUFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrNbTrans:      1,
						AdrTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedAdrTxPowerIndexes: []uint32{
						0, 1,
					},
					RejectedAdrDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DataRateIndex_DATA_RATE_3,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MacSettings: &ttnpb.MACSettings{
					AdrMargin: &wrapperspb.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_2
				dev.MacState.DesiredParameters.AdrTxPowerIndex = 3
				dev.MacState.DesiredParameters.AdrNbTrans = 1
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)
				fp := test.FrequencyPlan(dev.FrequencyPlanId)
				err := AdaptDataRate(ctx, dev, LoRaWANBands[fp.BandID][dev.LorawanPhyVersion], nil)
				if !a.So(err, should.Equal, tc.Error) {
					t.Fatalf("ADR failed with: %s", err)
				}
				expected := ttnpb.Clone(tc.Device)
				if tc.DeviceDiff != nil {
					tc.DeviceDiff(expected)
				}
				a.So(dev, should.Resemble, expected)
			},
		})
	}
}

func TestIssue458(t *testing.T) {
	t.Parallel()

	txSettings := &ttnpb.TxSettings{
		DataRate: &ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Lora{
				Lora: &ttnpb.LoRaDataRate{
					SpreadingFactor: 10,
					Bandwidth:       125000,
					CodingRate:      band.Cr4_5,
				},
			},
		},
	}
	issue458Uplinks := ADRMatrixToUplinks([]ADRMatrixRow{
		{FCnt: 1, MaxSNR: -7.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 8, MaxSNR: -3, GtwDiversity: 2, TxSettings: txSettings},
		{FCnt: 11, MaxSNR: -7, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 13, MaxSNR: -13.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 16, MaxSNR: -6.8, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 17, MaxSNR: -3, GtwDiversity: 2, TxSettings: txSettings},
		{FCnt: 18, MaxSNR: -4, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 26, MaxSNR: -5.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 27, MaxSNR: -7.8, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 28, MaxSNR: -6.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 33, MaxSNR: -9.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 36, MaxSNR: -6.8, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 114, MaxSNR: -1.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 141, MaxSNR: -4, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 203, MaxSNR: -7.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 204, MaxSNR: -4.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 208, MaxSNR: -5.8, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 209, MaxSNR: -5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 210, MaxSNR: -6, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 211, MaxSNR: -7.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 212, MaxSNR: -7.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 213, MaxSNR: -7.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 215, MaxSNR: -6.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 216, MaxSNR: -6.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 217, MaxSNR: -3, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 219, MaxSNR: -5.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 220, MaxSNR: -5.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 222, MaxSNR: -4.5, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 224, MaxSNR: -9.2, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 225, MaxSNR: -7.8, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 226, MaxSNR: -7.8, GtwDiversity: 1, TxSettings: txSettings},
		{FCnt: 228, MaxSNR: -8.8, GtwDiversity: 1, TxSettings: txSettings},
	})
	for _, tc := range []struct {
		Name       string
		Device     *ttnpb.EndDevice
		DeviceDiff func(*ttnpb.EndDevice)
		Error      error
	}{
		{
			Name: "initial uplinks, no change",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.USFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					DesiredParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					RecentUplinks:     issue458Uplinks[:9],
				},
			},
		},
		{
			Name: "all uplinks, increase nbTrans",
			Device: &ttnpb.EndDevice{
				FrequencyPlanId:   test.USFrequencyPlanID,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
				MacState: &ttnpb.MACState{
					CurrentParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					DesiredParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHYVersion_RP001_V1_0_2_REV_B),
					RecentUplinks:     issue458Uplinks[:],
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MacState.DesiredParameters.AdrNbTrans = 3
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)
				fp := test.FrequencyPlan(dev.FrequencyPlanId)
				err := AdaptDataRate(ctx, dev, LoRaWANBands[fp.BandID][dev.LorawanPhyVersion], &ttnpb.MACSettings{})
				if !a.So(err, should.Equal, tc.Error) {
					t.Fatalf("ADR failed with: %s", err)
				}
				expected := ttnpb.Clone(tc.Device)
				if tc.DeviceDiff != nil {
					tc.DeviceDiff(expected)
				}
				a.So(dev, should.Resemble, expected)
			},
		})
	}
}

func TestADRLossRate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name    string
		Uplinks []*ttnpb.MACState_UplinkMessage
		Rate    float32
	}{
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 13},
			}),
			Rate: 1. / 3.,
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 14},
			}),
			Rate: 2. / 4.,
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 13},
				{FCnt: 15},
			}),
			Rate: 2. / 5.,
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 1},
				{FCnt: 1},
				{FCnt: 3},
				{FCnt: 3},
			}),
			Rate: 1. / 3.,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name: strings.Join(func() (ss []string) {
				for _, up := range tc.Uplinks {
					ss = append(ss, fmt.Sprintf("%d", up.Payload.GetMacPayload().FHdr.FCnt))
				}
				return ss
			}(), ","),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				a.So(ADRLossRate(tc.Uplinks...), should.Equal, tc.Rate)
			},
		})
	}
}

func TestADRInstability(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)

	makeUplink := func(i int, penalty float32) *ttnpb.MACState_UplinkMessage {
		up := &ttnpb.MACState_UplinkMessage{
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_UP,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FullFCnt: uint32(i),
					},
				},
			},
			Settings: &ttnpb.MACState_UplinkMessage_TxSettings{
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							Bandwidth:       125_000,
							SpreadingFactor: 7,
							CodingRate:      band.Cr4_5,
						},
					},
				},
			},
			RxMetadata: []*ttnpb.MACState_UplinkMessage_RxMetadata{
				{
					Snr: 14.0 - penalty,
				},
			},
		}
		return up
	}

	channels := []*ttnpb.MACParameters_Channel{
		{
			EnableUplink:     true,
			MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
			MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
		},
	}
	currentParameters, desiredParameters := &ttnpb.MACParameters{
		AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
		AdrTxPowerIndex:  0,
		AdrNbTrans:       1,
		Channels:         channels,
	}, &ttnpb.MACParameters{
		AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
		AdrTxPowerIndex:  0,
		AdrNbTrans:       1,
		Channels:         channels,
	}
	macState := &ttnpb.MACState{
		CurrentParameters: currentParameters,
		DesiredParameters: desiredParameters,
	}
	dev := &ttnpb.EndDevice{
		MacState: macState,
	}
	phy := &band.EU_863_870_RP1_V1_0_2_Rev_B

	expectedTxPowers := []uint32{
		// Jump from transmission power 0 to 2 due to extra budget of 4 dB.
		2,
		// Do nothing while the ADR safety margin is in place.
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		// Jump from transmission power 2 to 3 due to extra budget of 2 dB (removed safety margin).
		3,
		// Do nothing as the transmission power change is under the safety margin, or we have enough
		// uplinks to trust our SNR estimate.
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	}

	penalty := float32(0)
	for i := 0; i < 50; i++ {
		macState.RecentUplinks = append(macState.RecentUplinks, makeUplink(i, penalty))
		err := AdaptDataRate(ctx, dev, phy, nil)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(desiredParameters.AdrDataRateIndex, should.Equal, ttnpb.DataRateIndex_DATA_RATE_5)
		a.So(desiredParameters.AdrTxPowerIndex, should.Equal, expectedTxPowers[i])
		a.So(desiredParameters.AdrNbTrans, should.Equal, 1)

		if currentParameters.AdrDataRateIndex != desiredParameters.AdrDataRateIndex ||
			currentParameters.AdrTxPowerIndex != desiredParameters.AdrTxPowerIndex ||
			currentParameters.AdrNbTrans != desiredParameters.AdrNbTrans {
			diff := TxPowerStep(phy, currentParameters.AdrTxPowerIndex, desiredParameters.AdrTxPowerIndex)
			penalty += diff

			currentParameters.AdrDataRateIndex = desiredParameters.AdrDataRateIndex
			currentParameters.AdrTxPowerIndex = desiredParameters.AdrTxPowerIndex
			currentParameters.AdrNbTrans = desiredParameters.AdrNbTrans
			macState.LastAdrChangeFCntUp = uint32(i + 1)
		}
	}
}

func TestClampDataRateRange(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name     string
		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings

		InputMinDataRateIndex ttnpb.DataRateIndex
		InputMaxDataRateIndex ttnpb.DataRateIndex

		ExpectedMinDataRateIndex ttnpb.DataRateIndex
		ExpectedMaxDataRateIndex ttnpb.DataRateIndex
	}{
		{
			Name: "no device",

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_5,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_12,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_12,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "maximum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
		},
		{
			Name: "maximum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_5,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Name: "maximum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_12,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum+maximum;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_2,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
		},
		{
			Name: "minimum+maximum;left-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_6,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
		},
		{
			Name: "minimum+maximum;inside of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_7,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_9,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_7,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_9,
		},
		{
			Name: "minimum+maximum;right-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_7,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_11,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_7,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum+maximum;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_12,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_15,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_12,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			min, max := ClampDataRateRange(tc.Device, tc.Defaults, tc.InputMinDataRateIndex, tc.InputMaxDataRateIndex)
			a.So(min, should.Equal, tc.ExpectedMinDataRateIndex)
			a.So(max, should.Equal, tc.ExpectedMaxDataRateIndex)
		})
	}
}

func TestClampTxPowerRange(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name     string
		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings

		InputMinTxPowerIndex uint32
		InputMaxTxPowerIndex uint32

		ExpectedMinTxPowerIndex uint32
		ExpectedMaxTxPowerIndex uint32
	}{
		{
			Name: "no device",

			InputMinTxPowerIndex: 1,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 1,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 5,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 1,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 12,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 12,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "maximum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 3,
		},
		{
			Name: "maximum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 5,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 1,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 1,
			ExpectedMaxTxPowerIndex: 5,
		},
		{
			Name: "maximum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 12,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum+maximum;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 2,
								},
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 3,
		},
		{
			Name: "minimum+maximum;left-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 3,
								},
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 6,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 6,
		},
		{
			Name: "minimum+maximum;inside of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 7,
								},
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 9,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 7,
			ExpectedMaxTxPowerIndex: 9,
		},
		{
			Name: "minimum+maximum;right-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 7,
								},
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 11,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 7,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum+maximum;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 12,
								},
								MaxTxPowerIndex: &wrapperspb.UInt32Value{
									Value: 15,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 12,
			ExpectedMaxTxPowerIndex: 10,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			min, max := ClampTxPowerRange(tc.Device, tc.Defaults, tc.InputMinTxPowerIndex, tc.InputMaxTxPowerIndex)
			a.So(min, should.Equal, tc.ExpectedMinTxPowerIndex)
			a.So(max, should.Equal, tc.ExpectedMaxTxPowerIndex)
		})
	}
}

func TestClampNbTrans(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name     string
		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings

		InputNbTrans uint32

		ExpectedNbTrans uint32
	}{
		{
			Name: "no device",

			InputNbTrans: 1,

			ExpectedNbTrans: 1,
		},
		{
			Name: "minimum only;left of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &wrapperspb.UInt32Value{
									Value: 1,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 2,

			ExpectedNbTrans: 2,
		},
		{
			Name: "minimum only;right of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &wrapperspb.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 2,

			ExpectedNbTrans: 3,
		},
		{
			Name: "maximum only;left of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxNbTrans: &wrapperspb.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 5,

			ExpectedNbTrans: 3,
		},
		{
			Name: "maximum only;right of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxNbTrans: &wrapperspb.UInt32Value{
									Value: 7,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 5,

			ExpectedNbTrans: 5,
		},
		{
			Name: "minimum+maximum;left of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &wrapperspb.UInt32Value{
									Value: 2,
								},
								MaxNbTrans: &wrapperspb.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 5,

			ExpectedNbTrans: 3,
		},
		{
			Name: "minimum+maximum;inside of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &wrapperspb.UInt32Value{
									Value: 7,
								},
								MaxNbTrans: &wrapperspb.UInt32Value{
									Value: 9,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 8,

			ExpectedNbTrans: 8,
		},
		{
			Name: "minimum+maximum;right of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &wrapperspb.UInt32Value{
									Value: 12,
								},
								MaxNbTrans: &wrapperspb.UInt32Value{
									Value: 15,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 8,

			ExpectedNbTrans: 12,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			value := ClampNbTrans(tc.Device, tc.Defaults, tc.InputNbTrans)
			a.So(value, should.Equal, tc.ExpectedNbTrans)
		})
	}
}

func TestADRUplinks(t *testing.T) {
	t.Parallel()

	newUplink := func(mType ttnpb.MType, spreadingFactor uint32, fCnt uint32) *ttnpb.MACState_UplinkMessage {
		up := &ttnpb.MACState_UplinkMessage{
			Payload: &ttnpb.Message{
				MHdr: &ttnpb.MHDR{
					MType: mType,
					Major: ttnpb.Major_LORAWAN_R1,
				},
			},
			Settings: &ttnpb.MACState_UplinkMessage_TxSettings{
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							Bandwidth:       125_000,
							SpreadingFactor: spreadingFactor,
							CodingRate:      band.Cr4_5,
						},
					},
				},
			},
		}
		switch mType {
		case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
			up.Payload.Payload = &ttnpb.Message_MacPayload{
				MacPayload: &ttnpb.MACPayload{
					FullFCnt: fCnt,
				},
			}
		default:
		}
		return up
	}

	for _, tc := range []struct {
		Name string

		MACState *ttnpb.MACState
		Band     *band.Band

		ExpectedUplinks []*ttnpb.MACState_UplinkMessage
	}{
		{
			Name: "no uplinks",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{},
			},
		},
		{
			Name: "from join request",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{},
				RecentUplinks: []*ttnpb.MACState_UplinkMessage{
					newUplink(ttnpb.MType_JOIN_REQUEST, 12, 0),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 0),
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			ExpectedUplinks: []*ttnpb.MACState_UplinkMessage{
				newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 0),
			},
		},
		{
			Name: "from last change",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{},
				RecentUplinks: []*ttnpb.MACState_UplinkMessage{
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 0),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 1),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 2),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 3),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 4),
				},
				LastAdrChangeFCntUp: 2,
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			ExpectedUplinks: []*ttnpb.MACState_UplinkMessage{
				newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 2),
				newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 3),
				newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 4),
			},
		},
		{
			Name: "from data rate change",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
				},
				RecentUplinks: []*ttnpb.MACState_UplinkMessage{
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 0),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 12, 1),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 10, 2),
					newUplink(ttnpb.MType_UNCONFIRMED_UP, 10, 3),
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			ExpectedUplinks: []*ttnpb.MACState_UplinkMessage{
				newUplink(ttnpb.MType_UNCONFIRMED_UP, 10, 2),
				newUplink(ttnpb.MType_UNCONFIRMED_UP, 10, 3),
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			ups := ADRUplinks(tc.MACState, tc.Band)
			a.So(ups, should.Resemble, tc.ExpectedUplinks)
		})
	}
}

func TestADRDataRange(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		Device   *ttnpb.EndDevice
		Band     *band.Band
		Defaults *ttnpb.MACSettings

		AssertError func(error) bool

		Min, Max ttnpb.DataRateIndex
		Allowed  map[ttnpb.DataRateIndex]struct{}
		Ok       bool
	}{
		{
			Name: "invalid desired channels",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DesiredParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								EnableUplink: false,
							},
						},
					},
				},
			},

			AssertError: errors.IsDataLoss,
		},
		{
			Name: "clamping mismatch",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					DesiredParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
							},
						},
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_0,
								},
							},
						},
					},
				},
			},

			AssertError: errors.IsDataLoss,
		},
		{
			Name: "clamp to max",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
								EnableUplink:     true,
							},
						},
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			Min:     ttnpb.DataRateIndex_DATA_RATE_1,
			Max:     ttnpb.DataRateIndex_DATA_RATE_5,
			Allowed: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_1, ttnpb.DataRateIndex_DATA_RATE_6),
			Ok:      true,
		},
		{
			Name: "clamp to current",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_2,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
								EnableUplink:     true,
							},
						},
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			Min:     ttnpb.DataRateIndex_DATA_RATE_2,
			Max:     ttnpb.DataRateIndex_DATA_RATE_5,
			Allowed: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_1, ttnpb.DataRateIndex_DATA_RATE_6),
			Ok:      true,
		},
		{
			Name: "rejected; ok",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
								EnableUplink:     true,
							},
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
								EnableUplink:     true,
							},
						},
					},
					RejectedAdrDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DataRateIndex_DATA_RATE_0,
						ttnpb.DataRateIndex_DATA_RATE_3,
						ttnpb.DataRateIndex_DATA_RATE_6,
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			Min: ttnpb.DataRateIndex_DATA_RATE_1,
			Max: ttnpb.DataRateIndex_DATA_RATE_5,
			Allowed: map[ttnpb.DataRateIndex]struct{}{
				ttnpb.DataRateIndex_DATA_RATE_1: {},
				ttnpb.DataRateIndex_DATA_RATE_2: {},
				ttnpb.DataRateIndex_DATA_RATE_4: {},
				ttnpb.DataRateIndex_DATA_RATE_5: {},
			},
			Ok: true,
		},
		{
			Name: "rejected; no overlap",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Channels: []*ttnpb.MACParameters_Channel{
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
								EnableUplink:     true,
							},
							{
								MinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
								MaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
								EnableUplink:     true,
							},
						},
					},
					RejectedAdrDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DataRateIndex_DATA_RATE_0,
						ttnpb.DataRateIndex_DATA_RATE_1,
						ttnpb.DataRateIndex_DATA_RATE_2,
						ttnpb.DataRateIndex_DATA_RATE_3,
						ttnpb.DataRateIndex_DATA_RATE_4,
						ttnpb.DataRateIndex_DATA_RATE_5,
						ttnpb.DataRateIndex_DATA_RATE_6,
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, ctx := test.New(t)
			min, max, allowed, ok, err := ADRDataRateRange(ctx, tc.Device, tc.Band, tc.Defaults)
			if assertError := tc.AssertError; assertError != nil {
				a.So(assertError(err), should.BeTrue)
			} else {
				a.So(min, should.Equal, tc.Min)
				a.So(max, should.Equal, tc.Max)
				a.So(allowed, should.Resemble, tc.Allowed)
				a.So(ok, should.Equal, tc.Ok)
				a.So(err, should.BeNil)
			}
		})
	}
}

func TestADRTxPowerRange(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		Device   *ttnpb.EndDevice
		Band     *band.Band
		Defaults *ttnpb.MACSettings

		Min, Max uint32
		Rejected map[uint32]struct{}
		Ok       bool
	}{
		{
			Name: "clamping mismatch",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: wrapperspb.UInt32(8),
								MaxTxPowerIndex: wrapperspb.UInt32(10),
							},
						},
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,
		},
		{
			Name: "rejected; ok",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					RejectedAdrTxPowerIndexes: []uint32{
						0,
						4,
						7,
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,

			Min: 1,
			Max: 6,
			Rejected: map[uint32]struct{}{
				0: {},
				4: {},
				7: {},
			},
			Ok: true,
		},
		{
			Name: "rejected; no overlap",

			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					RejectedAdrTxPowerIndexes: []uint32{
						0,
						1,
						2,
						3,
						4,
						5,
						6,
						7,
					},
				},
			},
			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, ctx := test.New(t)
			min, max, rejected, ok := ADRTxPowerRange(ctx, tc.Device, tc.Band, tc.Defaults)
			a.So(min, should.Equal, tc.Min)
			a.So(max, should.Equal, tc.Max)
			a.So(rejected, should.Resemble, tc.Rejected)
			a.So(ok, should.Equal, tc.Ok)
		})
	}
}

func TestADRMargin(t *testing.T) {
	t.Parallel()

	float32Ptr := func(f float32) *float32 { return &f }
	newUplink := func(maxSNR *float32, spreadingFactor, bandwidth uint32) *ttnpb.MACState_UplinkMessage {
		up := &ttnpb.MACState_UplinkMessage{
			Settings: &ttnpb.MACState_UplinkMessage_TxSettings{
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							Bandwidth:       bandwidth,
							SpreadingFactor: spreadingFactor,
							CodingRate:      band.Cr4_5,
						},
					},
				},
			},
		}
		if maxSNR != nil {
			up.RxMetadata = []*ttnpb.MACState_UplinkMessage_RxMetadata{
				{
					Snr: *maxSNR,
				},
			}
		}
		return up
	}
	repeatUplink := func(up *ttnpb.MACState_UplinkMessage, count int) []*ttnpb.MACState_UplinkMessage {
		ups := make([]*ttnpb.MACState_UplinkMessage, 0, count)
		for i := 0; i != count; i++ {
			ups = append(ups, up)
		}
		return ups
	}

	for _, tc := range []struct {
		Name string

		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings
		Uplinks  []*ttnpb.MACState_UplinkMessage

		AssertError func(error) bool

		Margin  float32
		Optimal bool
		Ok      bool
	}{
		{
			Name: "no max SNR",

			Uplinks: []*ttnpb.MACState_UplinkMessage{
				newUplink(nil, 7, 125_000),
			},
		},
		{
			Name: "unknown demodulation floor",

			Uplinks: []*ttnpb.MACState_UplinkMessage{
				newUplink(float32Ptr(7.125), 5, 125_000),
			},

			AssertError: errors.IsInvalidArgument,
		},
		{
			Name: "SF7BW125 suboptimal",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								Margin: wrapperspb.Float(15),
							},
						},
					},
				},
			},
			Uplinks: repeatUplink(
				newUplink(float32Ptr(7.125), 7, 125_000),
				5,
			),

			// Best SNR of 7.125 dB, demodulation floor of -7.5 dB, margin of 15 dB
			// and a safety margin of 2.5 dB. 7.125 - (-7.5) - 15 - 2.5 = -2.875.
			Margin:  7.125 - (-7.5) - 15 - 2.5,
			Optimal: false,
			Ok:      true,
		},
		{
			Name: "SF7BW125 optimal",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								Margin: wrapperspb.Float(15),
							},
						},
					},
				},
			},
			Uplinks: repeatUplink(
				newUplink(float32Ptr(7.125), 7, 125_000),
				20,
			),

			// Best SNR of 7.125 dB, demodulation floor of -7.5 dB, margin of 15 dB.
			// 7.125 - (-7.5) - 15 = -0.375.
			Margin:  7.125 - (-7.5) - 15,
			Optimal: true,
			Ok:      true,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, ctx := test.New(t)
			margin, optimal, ok, err := ADRMargin(ctx, tc.Device, tc.Defaults, tc.Uplinks...)
			if assertError := tc.AssertError; assertError != nil {
				a.So(assertError(err), should.BeTrue)
			} else {
				a.So(margin, should.Equal, tc.Margin)
				a.So(optimal, should.Equal, tc.Optimal)
				a.So(ok, should.Equal, tc.Ok)
				a.So(err, should.BeNil)
			}
		})
	}
}

func TestADRAdaptDataRate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		MACState                           *ttnpb.MACState
		Band                               *band.Band
		MinDataRateIndex, MaxDataRateIndex ttnpb.DataRateIndex
		AllowedDataRateIndices             map[ttnpb.DataRateIndex]struct{}
		MinTxPowerIndex                    uint32
		InitialMargin                      float32

		OutputMACState *ttnpb.MACState
		OutputMargin   float32
	}{
		{
			Name: "below min data rate index",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
			},
			Band:                   &band.EU_863_870_RP1_V1_0_2_Rev_B,
			MinDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_1,
			MaxDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_5,
			AllowedDataRateIndices: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_1, ttnpb.DataRateIndex_DATA_RATE_5),
			InitialMargin:          -5.0,

			OutputMACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					AdrTxPowerIndex:  0,
				},
			},
			OutputMargin: -7.5,
		},
		{
			Name: "positive steps",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
			},
			Band:                   &band.EU_863_870_RP1_V1_0_2_Rev_B,
			MaxDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_5,
			AllowedDataRateIndices: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_5),
			InitialMargin:          15.0,

			OutputMACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
					AdrTxPowerIndex:  0,
				},
			},
			OutputMargin: 2.5,
		},
		{
			Name: "negative steps",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
					AdrTxPowerIndex:  1,
				},
			},
			Band:                   &band.EU_863_870_RP1_V1_0_2_Rev_B,
			MinDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_3,
			MaxDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_5,
			AllowedDataRateIndices: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_3, ttnpb.DataRateIndex_DATA_RATE_5),
			InitialMargin:          -7.5,

			OutputMACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
					AdrTxPowerIndex:  1,
				},
			},
			OutputMargin: -7.5,
		},
		{
			Name: "rejected min",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
			},
			Band:                   &band.EU_863_870_RP1_V1_0_2_Rev_B,
			MaxDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_5,
			AllowedDataRateIndices: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_1, ttnpb.DataRateIndex_DATA_RATE_5),
			InitialMargin:          2.5,

			OutputMACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
					AdrTxPowerIndex:  0,
				},
			},
			OutputMargin: 0.0,
		},
		{
			Name: "rejected max",

			MACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
			},
			Band:                   &band.EU_863_870_RP1_V1_0_2_Rev_B,
			MaxDataRateIndex:       ttnpb.DataRateIndex_DATA_RATE_5,
			AllowedDataRateIndices: newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_4),
			InitialMargin:          15.0,

			OutputMACState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					AdrTxPowerIndex:  1,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
					AdrTxPowerIndex:  0,
				},
			},
			OutputMargin: 5.0,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			margin := ADRAdaptDataRate(
				tc.MACState,
				tc.Band,
				tc.MinDataRateIndex, tc.MaxDataRateIndex,
				tc.AllowedDataRateIndices,
				tc.MinTxPowerIndex,
				tc.InitialMargin,
			)
			a.So(margin, should.Equal, tc.OutputMargin)
			a.So(tc.MACState, should.Resemble, tc.OutputMACState)
		})
	}
}

func TestADRAdaptTxPowerIndex(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		MACState      *ttnpb.MACState
		Band          *band.Band
		Min, Max      uint32
		Rejected      map[uint32]struct{}
		InitialMargin float32
		Optimal       bool

		OutputMACState *ttnpb.MACState
		OutputMargin   float32
	}{
		{
			Name: "min clamping",

			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 0,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Min:           1,
			Max:           7,
			InitialMargin: 0.0,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 1,
				},
			},
			OutputMargin: -2.0,
		},
		{
			Name: "max clamping",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 7,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           6,
			InitialMargin: 0.0,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 6,
				},
			},
			OutputMargin: 2.0,
		},
		{
			Name: "one step forward",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 2,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           7,
			InitialMargin: 3.0,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 3,
				},
			},
			OutputMargin: 1.0,
		},
		{
			Name: "two steps forward",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 2,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           7,
			InitialMargin: 5.0,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 4,
				},
			},
			OutputMargin: 1.0,
		},
		{
			Name: "one step backward; suboptimal",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 3,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           7,
			InitialMargin: -1.5,
			Optimal:       false,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 3,
				},
			},
			OutputMargin: -1.5,
		},
		{
			Name: "one step backward; optimal",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 3,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           7,
			InitialMargin: -1.5,
			Optimal:       true,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 2,
				},
			},
			OutputMargin: 0.5,
		},
		{
			Name: "two steps backward",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 3,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           7,
			InitialMargin: -3.5,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 1,
				},
			},
			OutputMargin: 0.5,
		},
		{
			Name: "backward to zero",
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 3,
				},
			},
			Band:          &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Max:           7,
			InitialMargin: -20.5,

			OutputMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					AdrTxPowerIndex: 0,
				},
			},
			OutputMargin: -14.5,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			margin := ADRAdaptTxPowerIndex(tc.MACState, tc.Band, tc.Min, tc.Max, tc.Rejected, tc.InitialMargin, tc.Optimal)
			a.So(margin, should.Equal, tc.OutputMargin)
			a.So(tc.MACState, should.Resemble, tc.OutputMACState)
		})
	}
}

func TestADRAdaptNbTrans(t *testing.T) {
	t.Parallel()

	newUplink := func(fCnt uint32) *ttnpb.MACState_UplinkMessage {
		up := &ttnpb.MACState_UplinkMessage{
			Payload: &ttnpb.Message{
				Payload: &ttnpb.Message_MacPayload{
					MacPayload: &ttnpb.MACPayload{
						FullFCnt: fCnt,
					},
				},
			},
		}
		return up
	}
	newUplinkSeries := func(n uint32) []*ttnpb.MACState_UplinkMessage {
		ups := make([]*ttnpb.MACState_UplinkMessage, 0, n)
		for i := uint32(0); i != n; i++ {
			ups = append(ups, newUplink(i))
		}
		return ups
	}
	newEndDevice := func(current, desired uint32) *ttnpb.EndDevice {
		return &ttnpb.EndDevice{
			MacState: &ttnpb.MACState{
				CurrentParameters: &ttnpb.MACParameters{
					AdrNbTrans: current,
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrNbTrans: desired,
				},
			},
		}
	}

	for _, tc := range []struct {
		Name string

		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings
		Uplinks  []*ttnpb.MACState_UplinkMessage

		ExpectedDevice *ttnpb.EndDevice
	}{
		{
			Name: "not enough uplinks",

			Device: newEndDevice(2, 2),

			ExpectedDevice: newEndDevice(2, 2),
		},
		{
			Name: "under 5% loss rate; 3 to 2",

			Device:  newEndDevice(3, 3),
			Uplinks: newUplinkSeries(10),

			ExpectedDevice: newEndDevice(3, 2),
		},
		{
			Name: "under 5% loss rate; 2 to 1",

			Device:  newEndDevice(2, 2),
			Uplinks: newUplinkSeries(10),

			ExpectedDevice: newEndDevice(2, 1),
		},
		{
			Name: "under 5% loss rate; 1 to 1",

			Device:  newEndDevice(1, 1),
			Uplinks: newUplinkSeries(10),

			ExpectedDevice: newEndDevice(1, 1),
		},
		{
			Name: "under 10% loss rate; 3 to 3",

			Device:  newEndDevice(3, 3),
			Uplinks: append(newUplinkSeries(20), newUplink(22)), // 2 / 22 = 9% loss rate

			ExpectedDevice: newEndDevice(3, 3),
		},
		{
			Name: "under 10% loss rate; 2 to 2",

			Device:  newEndDevice(2, 2),
			Uplinks: append(newUplinkSeries(20), newUplink(22)), // 2 / 22 = 9% loss rate

			ExpectedDevice: newEndDevice(2, 2),
		},
		{
			Name: "under 10% loss rate; 1 to 1",

			Device:  newEndDevice(1, 1),
			Uplinks: append(newUplinkSeries(20), newUplink(22)), // 3 / 22 = 13% loss rate

			ExpectedDevice: newEndDevice(1, 1),
		},
		{
			Name: "under 30% loss rate; 3 to 3",

			Device:  newEndDevice(3, 3),
			Uplinks: append(newUplinkSeries(20), newUplink(23)), // 3 / 22 = 13% loss rate

			ExpectedDevice: newEndDevice(3, 3),
		},
		{
			Name: "under 30% loss rate; 2 to 3",

			Device:  newEndDevice(2, 2),
			Uplinks: append(newUplinkSeries(20), newUplink(23)), // 3 / 22 = 13% loss rate

			ExpectedDevice: newEndDevice(2, 3),
		},
		{
			Name: "under 30% loss rate; 1 to 2",

			Device:  newEndDevice(1, 1),
			Uplinks: append(newUplinkSeries(20), newUplink(23)), // 3 / 22 = 13% loss rate

			ExpectedDevice: newEndDevice(1, 2),
		},
		{
			Name: "over 30% loss rate; 1 to 3",

			Device:  newEndDevice(1, 1),
			Uplinks: append(newUplinkSeries(10), newUplink(15)), // 5 / 16 = 31% loss rate

			ExpectedDevice: newEndDevice(1, 3),
		},
		{
			Name: "over 30% loss rate; 2 to 3",

			Device:  newEndDevice(2, 2),
			Uplinks: append(newUplinkSeries(10), newUplink(15)), // 5 / 16 = 31% loss rate

			ExpectedDevice: newEndDevice(2, 3),
		},
		{
			Name: "over 30% loss rate; 3 to 3",

			Device:  newEndDevice(3, 3),
			Uplinks: append(newUplinkSeries(10), newUplink(15)), // 5 / 16 = 31% loss rate

			ExpectedDevice: newEndDevice(3, 3),
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			ADRAdaptNbTrans(tc.Device, tc.Defaults, tc.Uplinks)
			a.So(tc.Device, should.Resemble, tc.ExpectedDevice)
		})
	}
}

func TestDemodulationFloorStep(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		Band     *band.Band
		From, To ttnpb.DataRateIndex

		Step float32
	}{
		{
			Name: "EU868; SF7BW250 to SF7BW125",

			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,
			From: ttnpb.DataRateIndex_DATA_RATE_6,
			To:   ttnpb.DataRateIndex_DATA_RATE_5,

			Step: 3.0,
		},
		{
			Name: "EU868; SF7BW250 to SF8BW125",

			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,
			From: ttnpb.DataRateIndex_DATA_RATE_6,
			To:   ttnpb.DataRateIndex_DATA_RATE_4,

			Step: 5.5,
		},
		{
			Name: "US915; SF8BW500 to SF7BW125",

			Band: &band.US_902_928_RP1_V1_0_2_Rev_B,
			From: ttnpb.DataRateIndex_DATA_RATE_4,
			To:   ttnpb.DataRateIndex_DATA_RATE_3,

			Step: 3.5,
		},
		{
			Name: "US915; SF8BW500 to SF8BW125",

			Band: &band.US_902_928_RP1_V1_0_2_Rev_B,
			From: ttnpb.DataRateIndex_DATA_RATE_4,
			To:   ttnpb.DataRateIndex_DATA_RATE_2,

			Step: 6.0,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			step := DemodulationFloorStep(tc.Band, tc.From, tc.To)
			a.So(step, should.Equal, tc.Step)
			reverseStep := DemodulationFloorStep(tc.Band, tc.To, tc.From)
			a.So(reverseStep, should.Equal, -step)
		})
	}
}

func TestIsNarrowDataRateIndex(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		Band  *band.Band
		Index ttnpb.DataRateIndex

		LoRa, Ok bool
	}{
		{
			Name: "EU868; SF7BW125",

			Band:  &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Index: ttnpb.DataRateIndex_DATA_RATE_5,

			LoRa: true,
			Ok:   true,
		},
		{
			Name: "EU868; SF7BW250",

			Band:  &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Index: ttnpb.DataRateIndex_DATA_RATE_6,

			LoRa: true,
			Ok:   false,
		},
		{
			Name: "EU868; 50000",

			Band:  &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Index: ttnpb.DataRateIndex_DATA_RATE_7,

			LoRa: false,
			Ok:   false,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			lora, ok := IsNarrowDataRateIndex(tc.Band, tc.Index)
			a.So(lora, should.Equal, tc.LoRa)
			a.So(ok, should.Equal, tc.Ok)
		})
	}
}

func TestADRSteerDeviceChannels(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		InputDevice *ttnpb.EndDevice
		Defaults    *ttnpb.MACSettings
		Band        *band.Band
		Allowed     map[ttnpb.DataRateIndex]struct{}
		InputMargin float32

		OutputMargin float32
		Ok           bool
		OutputDevice *ttnpb.EndDevice
	}{
		{
			Name: "no mode",

			InputDevice: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{},
						},
					},
				},
			},
			InputMargin: 5.0,

			OutputMargin: 5.0,
			Ok:           false,
		},
		{
			Name: "disabled",
			InputDevice: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Disabled{
										Disabled: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_DisabledMode{},
									},
								},
							},
						},
					},
				},
			},
			InputMargin: 5.0,

			OutputMargin: 5.0,
			Ok:           false,
		},
		{
			Name: "already narrow",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
						AdrTxPowerIndex:  1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
						AdrTxPowerIndex:  1,
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Narrow{
										Narrow: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_NarrowMode{},
									},
								},
							},
						},
					},
				},
			},
			Band:        &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Allowed:     newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_6),
			InputMargin: 5.0,

			OutputMargin: 5.0,
			Ok:           false,
		},
		{
			Name: "not LoRa modulated",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_7,
						AdrTxPowerIndex:  1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_7,
						AdrTxPowerIndex:  1,
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Narrow{
										Narrow: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_NarrowMode{},
									},
								},
							},
						},
					},
				},
			},
			Band:        &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Allowed:     newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_7),
			InputMargin: 5.0,

			OutputMargin: 5.0,
			Ok:           false,
		},
		{
			Name: "EU868; SF7BW250 to SF7BW125",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
						AdrTxPowerIndex:  1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
						AdrTxPowerIndex:  1,
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Narrow{
										Narrow: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_NarrowMode{},
									},
								},
							},
						},
					},
				},
			},
			Band:        &band.EU_863_870_RP1_V1_0_2_Rev_B,
			Allowed:     newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_6),
			InputMargin: 5.0,

			OutputMargin: 2.0,
			Ok:           true,
			OutputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
						AdrTxPowerIndex:  1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
						AdrTxPowerIndex:  0,
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Narrow{
										Narrow: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_NarrowMode{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "US915; SF8BW500 to SF7BW125",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
						AdrTxPowerIndex:  1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
						AdrTxPowerIndex:  1,
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Narrow{
										Narrow: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_NarrowMode{},
									},
								},
							},
						},
					},
				},
			},
			Band:        &band.US_902_928_RP1_V1_0_2_Rev_B,
			Allowed:     newDataRateIndexRange(ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_4),
			InputMargin: 5.0,

			OutputMargin: 1.5,
			Ok:           true,
			OutputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_4,
						AdrTxPowerIndex:  1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						AdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
						AdrTxPowerIndex:  0,
					},
				},
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								ChannelSteering: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings{
									Mode: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_Narrow{
										Narrow: &ttnpb.ADRSettings_DynamicMode_ChannelSteeringSettings_NarrowMode{},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			device := ttnpb.Clone(tc.InputDevice)
			margin, ok := ADRSteerDeviceChannels(device, tc.Defaults, tc.Band, tc.Allowed, tc.InputMargin)
			a.So(margin, should.Equal, tc.OutputMargin)
			a.So(ok, should.Equal, tc.Ok)
			if tc.OutputDevice != nil {
				a.So(device, should.Resemble, tc.OutputDevice)
			} else {
				a.So(device, should.Resemble, tc.InputDevice)
			}
		})
	}
}

func newDataRateIndexRange(min, max ttnpb.DataRateIndex) map[ttnpb.DataRateIndex]struct{} {
	m := make(map[ttnpb.DataRateIndex]struct{})
	for idx := min; idx <= max; idx++ {
		m[idx] = struct{}{}
	}
	return m
}
