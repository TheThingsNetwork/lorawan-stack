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

package messages

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	pfconfig "go.thethings.network/lorawan-stack/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestType(t *testing.T) {
	a := assertions.New(t)
	msg := Version{
		Station:  "test",
		Firmware: "2.0.0",
		Package:  "test",
		Model:    "test",
		Protocol: 2,
	}

	data, err := json.Marshal(msg)
	a.So(err, should.BeNil)

	mt, err := Type(data)
	a.So(err, should.BeNil)
	a.So(mt, should.Equal, TypeUpstreamVersion)
}

func TestIsProduction(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Message          Version
		ExpectedResponse bool
	}{
		{
			Name:             "EmptyMessage",
			Message:          Version{},
			ExpectedResponse: false,
		},
		{
			Name: "EmptyMessage1",
			Message: Version{
				Features: "",
			},
			ExpectedResponse: false,
		},
		{
			Name: "NonProduction",
			Message: Version{
				Features: "gps rmtsh",
			},
			ExpectedResponse: false,
		},
		{
			Name: "Production",
			Message: Version{
				Features: "prod",
			},
			ExpectedResponse: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(tc.Message.IsProduction(), should.Equal, tc.ExpectedResponse)
		})
	}
}

func TestGetRouterConfig(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		FrequencyPlan  frequencyplans.FrequencyPlan
		Cfg            RouterConfig
		IsProd         bool
		ErrorAssertion func(err error) bool
	}{
		{
			Name:          "NilFrequencyPlan",
			FrequencyPlan: frequencyplans.FrequencyPlan{},
			Cfg:           RouterConfig{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errFrequencyPlan)
			},
		},
		{
			Name: "InvalidFrequencyPlan",
			FrequencyPlan: frequencyplans.FrequencyPlan{
				BandID: "PinkFloyd",
			},
			Cfg: RouterConfig{},
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
				SX1301Config: []pfconfig.SX1301Config{
					{
						LoRaWANPublic: true,
						ClockSource:   0,
						AntennaGain:   0,
						Radios: []pfconfig.RFConfig{
							{
								Enable:    true,
								Type:      "SX1257",
								Frequency: 922300000,
								TxEnable:  true,
							},
							{
								Enable: false, Type: "SX1257",
								Frequency: 923000000,
								TxEnable:  false,
							},
						},
						Channels: []pfconfig.IFConfig{
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &pfconfig.IFConfig{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						FSKChannel:          &pfconfig.IFConfig{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						TxLUTConfigs: []pfconfig.TxLUTConfig{
							{PAGain: 0, MixGain: 8, RFPower: -6, DigGain: 0},
							{PAGain: 0, MixGain: 10, RFPower: -3, DigGain: 0},
							{PAGain: 0, MixGain: 12, RFPower: 0, DigGain: 0},
							{PAGain: 1, MixGain: 8, RFPower: 3, DigGain: 0},
							{PAGain: 1, MixGain: 10, RFPower: 6, DigGain: 0},
							{PAGain: 1, MixGain: 12, RFPower: 10, DigGain: 0},
							{PAGain: 1, MixGain: 13, RFPower: 11, DigGain: 0},
							{PAGain: 2, MixGain: 9, RFPower: 12, DigGain: 0},
							{PAGain: 1, MixGain: 15, RFPower: 13, DigGain: 0},
							{PAGain: 2, MixGain: 10, RFPower: 14, DigGain: 0},
							{PAGain: 2, MixGain: 11, RFPower: 16, DigGain: 0},
							{PAGain: 3, MixGain: 9, RFPower: 20, DigGain: 0},
							{PAGain: 3, MixGain: 10, RFPower: 23, DigGain: 0},
							{PAGain: 3, MixGain: 11, RFPower: 25, DigGain: 0},
							{PAGain: 3, MixGain: 12, RFPower: 26, DigGain: 0},
							{PAGain: 3, MixGain: 14, RFPower: 27, DigGain: 0},
						},
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
				SX1301Config: []pfconfig.SX1301Config{
					{
						LoRaWANPublic: true,
						ClockSource:   0,
						AntennaGain:   0,
						Radios: []pfconfig.RFConfig{
							{
								Enable:    true,
								Type:      "SX1257",
								Frequency: 922300000,
								TxEnable:  true,
							},
							{
								Enable: false, Type: "SX1257",
								Frequency: 923000000,
								TxEnable:  false,
							},
						},
						Channels: []pfconfig.IFConfig{
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
							{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						},
						LoRaStandardChannel: &pfconfig.IFConfig{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						FSKChannel:          &pfconfig.IFConfig{Enable: false, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
						TxLUTConfigs: []pfconfig.TxLUTConfig{
							{PAGain: 0, MixGain: 8, RFPower: -6, DigGain: 0},
							{PAGain: 0, MixGain: 10, RFPower: -3, DigGain: 0},
							{PAGain: 0, MixGain: 12, RFPower: 0, DigGain: 0},
							{PAGain: 1, MixGain: 8, RFPower: 3, DigGain: 0},
							{PAGain: 1, MixGain: 10, RFPower: 6, DigGain: 0},
							{PAGain: 1, MixGain: 12, RFPower: 10, DigGain: 0},
							{PAGain: 1, MixGain: 13, RFPower: 11, DigGain: 0},
							{PAGain: 2, MixGain: 9, RFPower: 12, DigGain: 0},
							{PAGain: 1, MixGain: 15, RFPower: 13, DigGain: 0},
							{PAGain: 2, MixGain: 10, RFPower: 14, DigGain: 0},
							{PAGain: 2, MixGain: 11, RFPower: 16, DigGain: 0},
							{PAGain: 3, MixGain: 9, RFPower: 20, DigGain: 0},
							{PAGain: 3, MixGain: 10, RFPower: 23, DigGain: 0},
							{PAGain: 3, MixGain: 11, RFPower: 25, DigGain: 0},
							{PAGain: 3, MixGain: 12, RFPower: 26, DigGain: 0},
							{PAGain: 3, MixGain: 14, RFPower: 27, DigGain: 0},
						},
					},
				},
			},
			IsProd: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			cfg, err := GetRouterConfig(tc.FrequencyPlan, tc.IsProd, time.Now())
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
