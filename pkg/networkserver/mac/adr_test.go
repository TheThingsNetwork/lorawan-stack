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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestAdaptDataRate(t *testing.T) {
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
			TxSettings: ttnpb.TxSettings{
				DataRate: ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_LoRa{
						LoRa: &ttnpb.LoRaDataRate{
							SpreadingFactor: 12,
							Bandwidth:       125000,
						},
					},
				},
				DataRateIndex: 0,
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
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRNbTrans:      1,
						ADRTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRDataRateIndex: ttnpb.DATA_RATE_4,
						ADRNbTrans:       3,
						ADRTxPowerIndex:  2,
						Channels:         MakeDefaultEU868CurrentChannels(),
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_4
				dev.MACState.DesiredParameters.ADRTxPowerIndex = 1
				dev.MACState.DesiredParameters.ADRNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected DR:(1,4)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRNbTrans:      1,
						ADRTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_1, ttnpb.DATA_RATE_4,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_3
				dev.MACState.DesiredParameters.ADRTxPowerIndex = 2
				dev.MACState.DesiredParameters.ADRNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected TXPower:(1)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRNbTrans:      1,
						ADRTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedADRTxPowerIndexes: []uint32{
						1,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_4
				dev.MACState.DesiredParameters.ADRTxPowerIndex = 0
				dev.MACState.DesiredParameters.ADRNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected TXPower:(0,1)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRNbTrans:      1,
						ADRTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedADRTxPowerIndexes: []uint32{
						0, 1,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_3
				dev.MACState.DesiredParameters.ADRTxPowerIndex = 2
				dev.MACState.DesiredParameters.ADRNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected DR:(1,4), rejected TXPower:(0,2,3)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRNbTrans:      1,
						ADRTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedADRTxPowerIndexes: []uint32{
						0, 2, 3,
					},
					RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_1, ttnpb.DATA_RATE_4,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_3
				dev.MACState.DesiredParameters.ADRTxPowerIndex = 1
				dev.MACState.DesiredParameters.ADRNbTrans = 1
			},
		},
		{
			Name: "adapted example from Semtech paper/rejected DR:(3), rejected TXPower:(0,1)",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRNbTrans:      1,
						ADRTxPowerIndex: 1,
						Channels:        MakeDefaultEU868CurrentChannels(),
					},
					DesiredParameters: ttnpb.MACParameters{
						Channels: MakeDefaultEU868CurrentChannels(),
					},
					RejectedADRTxPowerIndexes: []uint32{
						0, 1,
					},
					RejectedADRDataRateIndexes: []ttnpb.DataRateIndex{
						ttnpb.DATA_RATE_3,
					},
					RecentUplinks: semtechPaperUplinks,
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRDataRateIndex = ttnpb.DATA_RATE_2
				dev.MACState.DesiredParameters.ADRTxPowerIndex = 3
				dev.MACState.DesiredParameters.ADRNbTrans = 1
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)
				fp := test.FrequencyPlan(dev.FrequencyPlanID)
				err := AdaptDataRate(ctx, dev, LoRaWANBands[fp.BandID][dev.LoRaWANPHYVersion], ttnpb.MACSettings{})
				if !a.So(err, should.Equal, tc.Error) {
					t.Fatalf("ADR failed with: %s", err)
				}
				expected := CopyEndDevice(tc.Device)
				if tc.DeviceDiff != nil {
					tc.DeviceDiff(expected)
				}
				a.So(dev, should.Resemble, expected)
			},
		})
	}
}

func TestIssue458(t *testing.T) {
	issue458Uplinks := ADRMatrixToUplinks([]ADRMatrixRow{
		{FCnt: 1, MaxSNR: -7.2, GtwDiversity: 1,
			TxSettings: ttnpb.TxSettings{
				DataRate: ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_LoRa{
						LoRa: &ttnpb.LoRaDataRate{
							SpreadingFactor: 10,
							Bandwidth:       125000,
						},
					},
				},
				DataRateIndex: 0,
			},
		},
		{FCnt: 8, MaxSNR: -3, GtwDiversity: 2},
		{FCnt: 11, MaxSNR: -7, GtwDiversity: 1},
		{FCnt: 13, MaxSNR: -13.5, GtwDiversity: 1},
		{FCnt: 16, MaxSNR: -6.8, GtwDiversity: 1},
		{FCnt: 17, MaxSNR: -3, GtwDiversity: 2},
		{FCnt: 18, MaxSNR: -4, GtwDiversity: 1},
		{FCnt: 26, MaxSNR: -5.5, GtwDiversity: 1},
		{FCnt: 27, MaxSNR: -7.8, GtwDiversity: 1},
		{FCnt: 28, MaxSNR: -6.5, GtwDiversity: 1},
		{FCnt: 33, MaxSNR: -9.5, GtwDiversity: 1},
		{FCnt: 36, MaxSNR: -6.8, GtwDiversity: 1},
		{FCnt: 114, MaxSNR: -1.2, GtwDiversity: 1},
		{FCnt: 141, MaxSNR: -4, GtwDiversity: 1},
		{FCnt: 203, MaxSNR: -7.5, GtwDiversity: 1},
		{FCnt: 204, MaxSNR: -4.2, GtwDiversity: 1},
		{FCnt: 208, MaxSNR: -5.8, GtwDiversity: 1},
		{FCnt: 209, MaxSNR: -5, GtwDiversity: 1},
		{FCnt: 210, MaxSNR: -6, GtwDiversity: 1},
		{FCnt: 211, MaxSNR: -7.5, GtwDiversity: 1},
		{FCnt: 212, MaxSNR: -7.5, GtwDiversity: 1},
		{FCnt: 213, MaxSNR: -7.2, GtwDiversity: 1},
		{FCnt: 215, MaxSNR: -6.2, GtwDiversity: 1},
		{FCnt: 216, MaxSNR: -6.5, GtwDiversity: 1},
		{FCnt: 217, MaxSNR: -3, GtwDiversity: 1},
		{FCnt: 219, MaxSNR: -5.2, GtwDiversity: 1},
		{FCnt: 220, MaxSNR: -5.5, GtwDiversity: 1},
		{FCnt: 222, MaxSNR: -4.5, GtwDiversity: 1},
		{FCnt: 224, MaxSNR: -9.2, GtwDiversity: 1},
		{FCnt: 225, MaxSNR: -7.8, GtwDiversity: 1},
		{FCnt: 226, MaxSNR: -7.8, GtwDiversity: 1},
		{FCnt: 228, MaxSNR: -8.8, GtwDiversity: 1},
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
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					DesiredParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					RecentUplinks:     issue458Uplinks[:9],
				},
			},
		},
		{
			Name: "all uplinks, increase nbTrans",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
				MACState: &ttnpb.MACState{
					CurrentParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					DesiredParameters: MakeDefaultUS915CurrentMACParameters(ttnpb.PHY_V1_0_2_REV_B),
					RecentUplinks:     issue458Uplinks[:],
				},
			},
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.DesiredParameters.ADRNbTrans = 3
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)
				fp := test.FrequencyPlan(dev.FrequencyPlanID)
				err := AdaptDataRate(ctx, dev, LoRaWANBands[fp.BandID][dev.LoRaWANPHYVersion], ttnpb.MACSettings{})
				if !a.So(err, should.Equal, tc.Error) {
					t.Fatalf("ADR failed with: %s", err)
				}
				expected := CopyEndDevice(tc.Device)
				if tc.DeviceDiff != nil {
					tc.DeviceDiff(expected)
				}
				a.So(dev, should.Resemble, expected)
			},
		})
	}
}
