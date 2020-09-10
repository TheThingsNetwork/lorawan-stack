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
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
				RecentADRUplinks: semtechPaperUplinks,
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
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
				RecentADRUplinks: semtechPaperUplinks,
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
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
				RecentADRUplinks: semtechPaperUplinks,
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
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
				RecentADRUplinks: semtechPaperUplinks,
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
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
				RecentADRUplinks: semtechPaperUplinks,
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
				},
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{
						Value: 2,
					},
				},
				RecentADRUplinks: semtechPaperUplinks,
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
				fp := FrequencyPlan(dev.FrequencyPlanID)
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
