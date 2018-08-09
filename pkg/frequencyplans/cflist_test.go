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

package frequencyplans_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestFrequenciesCFList(t *testing.T) {
	a := assertions.New(t)

	euFP := ttnpb.FrequencyPlan{
		BandID: "EU_863_870",
		Channels: []*ttnpb.FrequencyPlan_Channel{
			{Frequency: 867100000},
			{Frequency: 867300000},
			{Frequency: 867500000},
			{Frequency: 867700000},
			{Frequency: 867900000},
			{Frequency: 868100000},
			{Frequency: 868300000},
			{Frequency: 868500000},
		},
	}

	cfList := frequencyplans.CFList(euFP, ttnpb.PHY_V1_1_REV_B)

	a.So(cfList.Type, should.Equal, ttnpb.CFListType_FREQUENCIES)

	euBand, err := band.GetByID(euFP.BandID)
	a.So(err, should.BeNil)

fpChannels:
	for _, channel := range euFP.Channels {
		for _, bandChannel := range euBand.UplinkChannels {
			if bandChannel.Frequency == channel.Frequency {
				continue fpChannels
			}
		}

		var found bool
		for _, cfListFrequency := range cfList.Freq {
			if uint64(cfListFrequency)*100 == channel.Frequency {
				found = true
			}
		}
		a.So(found, should.BeTrue)
	}
}

func TestChannelMasksCFList(t *testing.T) {
	a := assertions.New(t)

	usFP := ttnpb.FrequencyPlan{
		BandID: "US_902_928",
		Channels: []*ttnpb.FrequencyPlan_Channel{
			{Frequency: 903900000},
			{Frequency: 904100000},
			{Frequency: 904300000},
			{Frequency: 904500000},
			{Frequency: 904700000},
			{Frequency: 904900000},
			{Frequency: 905100000},
			{Frequency: 905300000},
		},
	}

	cfList := frequencyplans.CFList(usFP, ttnpb.PHY_V1_1_REV_B)

	enabledChannels := []int{8, 9, 10, 11, 12, 13, 14, 15}
chMaskLoop:
	for index, chMaskEntry := range cfList.ChMasks {
		for _, enabledChannel := range enabledChannels {
			if enabledChannel == index {
				a.So(chMaskEntry, should.BeTrue)
				continue chMaskLoop
			}
		}
		a.So(chMaskEntry, should.BeFalse)
	}
}

func TestUnimplementedCFList(t *testing.T) {
	a := assertions.New(t)

	usFP := ttnpb.FrequencyPlan{
		BandID: "US_902_928",
		Channels: []*ttnpb.FrequencyPlan_Channel{
			{Frequency: 903900000},
			{Frequency: 904100000},
			{Frequency: 904300000},
			{Frequency: 904500000},
			{Frequency: 904700000},
			{Frequency: 904900000},
			{Frequency: 905100000},
			{Frequency: 905300000},
		},
	}

	cfList := frequencyplans.CFList(usFP, ttnpb.PHY_V1_0)
	a.So(cfList, should.BeNil)
}
