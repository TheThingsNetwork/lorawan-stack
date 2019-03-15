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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
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
	a := assertions.New(t)
	for _, tc := range []struct {
		Name             string
		Message          Version
		ExpectedResponse bool
	}{
		{
			"EmptyMessage",
			Version{},
			false,
		},
		{
			"EmptyMessage1",
			Version{
				Features: []string{""},
			},
			false,
		},
		{
			"NonProduction",
			Version{
				Features: []string{"gps", "rmtsh"},
			},
			false,
		},
		{
			"Production",
			Version{
				Features: []string{"prod"},
			},
			true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a.So(tc.Message.IsProduction(), should.Equal, tc.ExpectedResponse)
		})
	}

}

func TestGetRouterConfig(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name          string
		FP            frequencyplans.FrequencyPlan
		Cfg           RouterConfig
		IsProd        bool
		ExpectedError error
	}{
		{
			"NilFrequencyPlan",
			frequencyplans.FrequencyPlan{},
			RouterConfig{},
			false,
			errFrequencyPlan,
		},
		{
			"InvalidFrequencyPlan",
			frequencyplans.FrequencyPlan{
				BandID: "PinkFloyd",
			},
			RouterConfig{},
			false,
			errFrequencyPlan,
		},
		{
			"ValidFrequencyPlan",
			frequencyplans.FrequencyPlan{
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
			RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{909000000, 925000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
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
			},
			false,
			nil,
		},
		{
			"ValidFrequencyPlanProd",
			frequencyplans.FrequencyPlan{
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
			RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{909000000, 925000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
					[3]int{12, 500, 0},
					[3]int{11, 500, 0},
					[3]int{10, 500, 0},
					[3]int{9, 500, 0},
					[3]int{8, 500, 0},
					[3]int{7, 500, 0},
				},
			},
			true,
			nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			cfg, err := GetRouterConfig(tc.FP, tc.IsProd)
			if !(a.So(err, should.Resemble, tc.ExpectedError)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(cfg, should.Resemble, tc.Cfg)) {
				t.Fatalf("Invalid config: %v", cfg)
			}
		})
	}
}
