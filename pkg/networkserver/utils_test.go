// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package networkserver_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestEnableLoRaStandardChannel(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		Frequency     uint64
		MACState      *ttnpb.MACState
		Band          *band.Band
		FrequencyPlan *frequencyplans.FrequencyPlan

		ExpectedMACState *ttnpb.MACState
	}{
		{
			Name: "dynamic channel plan",

			Band: &band.EU_863_870_RP1_V1_0_2_Rev_B,
		},
		{
			Name: "no standard channel",

			Band:          &band.US_902_928_RP1_V1_0_2_Rev_B,
			FrequencyPlan: &frequencyplans.FrequencyPlan{},
		},
		{
			Name: "frequency mismatch",

			Frequency:     903000000, // FSB1 standard channel frequency
			Band:          &band.US_902_928_RP1_V1_0_2_Rev_B,
			FrequencyPlan: test.FrequencyPlan(test.USFrequencyPlanID), // FSB2 frequency plan
		},
		{
			Name: "enable channel",

			Frequency: 904600000,
			MACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					Channels: []*ttnpb.MACParameters_Channel{
						{
							UplinkFrequency: 903900000,
							EnableUplink:    true,
						},
						{
							UplinkFrequency: 904600000,
						},
					},
				},
			},
			Band:          &band.US_902_928_RP1_V1_0_2_Rev_B,
			FrequencyPlan: test.FrequencyPlan(test.USFrequencyPlanID),

			ExpectedMACState: &ttnpb.MACState{
				DesiredParameters: &ttnpb.MACParameters{
					Channels: []*ttnpb.MACParameters_Channel{
						{
							UplinkFrequency: 903900000,
							EnableUplink:    true,
						},
						{
							UplinkFrequency: 904600000,
							EnableUplink:    true,
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
			networkserver.EnableLoRaStandardChannel(tc.Frequency, tc.MACState, tc.Band, tc.FrequencyPlan)
			a.So(tc.MACState, should.Resemble, tc.ExpectedMACState)
		})
	}
}
