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

package lbslns

import (
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type TestFeatures struct {
	Production bool
}

func (f TestFeatures) IsProduction() bool { return f.Production }

func TestGetRouterConfig(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name            string
		FrequencyPlan   frequencyplans.FrequencyPlan
		FrequencyPlanID string
		Cfg             RouterConfig
		Features        TestFeatures
		ErrorAssertion  func(err error) bool
	}{
		{
			Name:            "NilFrequencyPlan",
			FrequencyPlan:   frequencyplans.FrequencyPlan{},
			FrequencyPlanID: "dummy",
			Cfg:             RouterConfig{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errFrequencyPlan)
			},
		},
		{
			Name: "InvalidFrequencyPlan",
			FrequencyPlan: frequencyplans.FrequencyPlan{
				BandID: "PinkFloyd",
			},
			FrequencyPlanID: "dummy",
			Cfg:             RouterConfig{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errFrequencyPlan)
			},
		},
		{
			Name: "ValidFrequencyPlan",
			FrequencyPlan: frequencyplans.FrequencyPlan{
				BandID: "US_902_928",
				Radios: []frequencyplans.Radio{
					{
						Enable:    true,
						ChipType:  "SX1257",
						Frequency: 922300000,
						TxConfiguration: &frequencyplans.RadioTxConfiguration{
							MinFrequency: 909000000,
							MaxFrequency: 925000000,
						},
					},
					{
						Enable:    false,
						ChipType:  "SX1257",
						Frequency: 923000000,
					},
				},
			},
			FrequencyPlanID: "US_902_928",
			Cfg: RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{909000000, 925000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
					[3]int{0, 0, 0},
					[3]int{0, 0, 0},
					[3]int{0, 0, 0},
					[3]int{12, 500, 0},
					[3]int{11, 500, 0},
					[3]int{10, 500, 0},
					[3]int{9, 500, 0},
					[3]int{8, 500, 0},
					[3]int{7, 500, 0},
				},
				NoCCA:       true,
				NoDutyCycle: true,
				NoDwellTime: true,
				SX1301Config: []LBSSX1301Config{
					{
						Radios: []LBSRFConfig{
							{
								Enable:    true,
								Frequency: 922300000,
							},
							{
								Enable:    false,
								Frequency: 923000000,
							},
						},
						Channels: []shared.IFConfig{
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0, Datarate: 0,
						},
						FSKChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
					},
				},
				Beacon: &BeaconingConfig{
					DR:     ttnpb.DataRateIndex_DATA_RATE_8,
					Layout: [3]int{5, 11, 23},
					Freqs: []uint64{
						923300000,
						923900000,
						924500000,
						925100000,
						925700000,
						926300000,
						926900000,
						927500000,
					},
				},
			},
		},
		{
			Name: "ValidFrequencyPlanProd",
			FrequencyPlan: frequencyplans.FrequencyPlan{
				BandID: "US_902_928",
				Radios: []frequencyplans.Radio{
					{
						Enable:    true,
						ChipType:  "SX1257",
						Frequency: 922300000,
						TxConfiguration: &frequencyplans.RadioTxConfiguration{
							MinFrequency: 909000000,
							MaxFrequency: 925000000,
						},
					},
					{
						Enable:    false,
						ChipType:  "SX1257",
						Frequency: 923000000,
					},
				},
			},
			FrequencyPlanID: "US_902_928",
			Cfg: RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{909000000, 925000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
					[3]int{0, 0, 0},
					[3]int{0, 0, 0},
					[3]int{0, 0, 0},
					[3]int{12, 500, 0},
					[3]int{11, 500, 0},
					[3]int{10, 500, 0},
					[3]int{9, 500, 0},
					[3]int{8, 500, 0},
					[3]int{7, 500, 0},
				},
				SX1301Config: []LBSSX1301Config{
					{
						Radios: []LBSRFConfig{
							{
								Enable:    true,
								Frequency: 922300000,
							},
							{
								Enable:    false,
								Frequency: 923000000,
							},
						},
						Channels: []shared.IFConfig{
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
						FSKChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
					},
				},
				Beacon: &BeaconingConfig{
					DR:     ttnpb.DataRateIndex_DATA_RATE_8,
					Layout: [3]int{5, 11, 23},
					Freqs: []uint64{
						923300000,
						923900000,
						924500000,
						925100000,
						925700000,
						926300000,
						926900000,
						927500000,
					},
				},
			},
			Features: TestFeatures{
				Production: true,
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, ctx := test.New(t)
			fps := []*frequencyplans.FrequencyPlan{&tc.FrequencyPlan}
			cfg, err := GetRouterConfig(ctx, tc.FrequencyPlan.BandID, fps, tc.Features, time.Now(), 0)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				cfg.MuxTime = 0
				if !a.So(cfg, should.Resemble, tc.Cfg) {
					t.Fatalf("Invalid config: %v", cfg)
				}
			}
		})
	}
}

func TestGetRouterConfigWithMultipleFP(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name           string
		BandID         string
		FrequencyPlans []*frequencyplans.FrequencyPlan
		Cfg            RouterConfig
		Features       TestFeatures
		ErrorAssertion func(err error) bool
	}{
		{
			Name:   "ValidFrequencyPlan",
			BandID: "US_902_928",
			FrequencyPlans: []*frequencyplans.FrequencyPlan{
				{
					BandID: "US_902_928",
					Radios: []frequencyplans.Radio{
						{
							Enable:    true,
							ChipType:  "SX1257",
							Frequency: 924300000,
							TxConfiguration: &frequencyplans.RadioTxConfiguration{
								MinFrequency: 909000000,
								MaxFrequency: 927000000,
							},
						},
						{
							Enable:    false,
							ChipType:  "SX1257",
							Frequency: 925000000,
						},
					},
				},
				{
					BandID: "US_902_928",
					Radios: []frequencyplans.Radio{
						{
							Enable:    true,
							ChipType:  "SX1257",
							Frequency: 924300000,
							TxConfiguration: &frequencyplans.RadioTxConfiguration{
								MinFrequency: 900000000,
								MaxFrequency: 925000000,
							},
						},
						{
							Enable:    false,
							ChipType:  "SX1257",
							Frequency: 925000000,
						},
					},
				},
			},
			Cfg: RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/2",
				FrequencyRange: []int{900000000, 927000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
					[3]int{0, 0, 0},
					[3]int{0, 0, 0},
					[3]int{0, 0, 0},
					[3]int{12, 500, 0},
					[3]int{11, 500, 0},
					[3]int{10, 500, 0},
					[3]int{9, 500, 0},
					[3]int{8, 500, 0},
					[3]int{7, 500, 0},
				},
				NoCCA:       true,
				NoDutyCycle: true,
				NoDwellTime: true,
				SX1301Config: []LBSSX1301Config{
					{
						Radios: []LBSRFConfig{
							{
								Enable:      true,
								Frequency:   924300000,
								AntennaGain: 3,
							},
							{
								Enable:      false,
								Frequency:   925000000,
								AntennaGain: 3,
							},
						},
						Channels: []shared.IFConfig{
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
						FSKChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
					},
					{
						Radios: []LBSRFConfig{
							{
								Enable:      true,
								Frequency:   924300000,
								AntennaGain: 3,
							},
							{
								Enable:      false,
								Frequency:   925000000,
								AntennaGain: 3,
							},
						},
						Channels: []shared.IFConfig{
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
						FSKChannel: &shared.IFConfig{
							Enable:       false,
							Radio:        0,
							IFValue:      0,
							Bandwidth:    0,
							SpreadFactor: 0,
							Datarate:     0,
						},
					},
				},
				Beacon: &BeaconingConfig{
					DR:     ttnpb.DataRateIndex_DATA_RATE_8,
					Layout: [3]int{5, 11, 23},
					Freqs: []uint64{
						923300000,
						923900000,
						924500000,
						925100000,
						925700000,
						926300000,
						926900000,
						927500000,
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, ctx := test.New(t)
			cfg, err := GetRouterConfig(ctx, tc.BandID, tc.FrequencyPlans, tc.Features, time.Now(), 3)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				cfg.MuxTime = 0
				if !a.So(cfg, should.Resemble, tc.Cfg) {
					t.Fatalf("Invalid config: %v", cfg)
				}
			}
		})
	}
}

func TestGetDataRatesFromFrequencyPlan(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)
	for _, tc := range []struct {
		Name           string
		BandID         string
		DataRates      DataRates
		ErrorAssertion func(error) bool
	}{
		{
			Name:           "InvalidBandID",
			BandID:         "EU",
			DataRates:      DataRates{},
			ErrorAssertion: errors.IsNotFound,
		},
		{
			Name:   "ValidBAndID",
			BandID: "EU_433",
			DataRates: DataRates{
				[3]int{12, 125, 0},
				[3]int{11, 125, 0},
				[3]int{10, 125, 0},
				[3]int{9, 125, 0},
				[3]int{8, 125, 0},
				[3]int{7, 125, 0},
				[3]int{7, 250, 0},
			},
		},
		{
			Name:   "ValidBAndIDUS",
			BandID: "US_902_928",
			DataRates: DataRates{
				[3]int{10, 125, 0},
				[3]int{9, 125, 0},
				[3]int{8, 125, 0},
				[3]int{7, 125, 0},
				[3]int{8, 500, 0},
				[3]int{0, 0, 0},
				[3]int{0, 0, 0},
				[3]int{0, 0, 0},
				[3]int{12, 500, 0},
				[3]int{11, 500, 0},
				[3]int{10, 500, 0},
				[3]int{9, 500, 0},
				[3]int{8, 500, 0},
				[3]int{7, 500, 0},
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			drs, err := getDataRatesFromBandID(tc.BandID)
			if err != nil && (tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !a.So(drs, should.Resemble, tc.DataRates) {
				t.Fatalf("Invalid datarates: %v", drs)
			}
		})
	}
}
