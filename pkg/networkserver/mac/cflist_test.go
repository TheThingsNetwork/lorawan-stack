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

package mac_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestFrequenciesCFList(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	euChannels := []*ttnpb.MACParameters_Channel{
		// 3 band default channels.
		{EnableUplink: true, UplinkFrequency: 868100000},
		{EnableUplink: true, UplinkFrequency: 868300000},
		{EnableUplink: true, UplinkFrequency: 868500000},
		// 8 custom channels.
		{EnableUplink: false, UplinkFrequency: 866500000},
		{EnableUplink: true, UplinkFrequency: 866700000},
		{EnableUplink: true, UplinkFrequency: 866900000},
		{EnableUplink: true, UplinkFrequency: 867100000},
		{EnableUplink: true, UplinkFrequency: 867300000},
		{EnableUplink: true, UplinkFrequency: 867500000},
		{EnableUplink: true, UplinkFrequency: 867700000},
		{EnableUplink: true, UplinkFrequency: 867900000},
	}

	phy := &band.EU_863_870_RP1_V1_1_Rev_B

	cfList := mac.CFList(phy, euChannels...)
	a.So(cfList.Type, should.Equal, ttnpb.CFListType_FREQUENCIES)
	a.So(cfList.Freq, should.HaveLength, 5)
	var seen int
outer:
	for _, channel := range euChannels {
		for _, bandChannel := range phy.UplinkChannels {
			if bandChannel.Frequency == channel.UplinkFrequency {
				continue outer
			}
		}

		var found bool
		for _, cfListFrequency := range cfList.Freq {
			if uint64(cfListFrequency)*phy.FreqMultiplier == channel.UplinkFrequency {
				found = true
			}
		}
		if seen < 5 && channel.EnableUplink {
			a.So(found, should.BeTrue)
			seen++
		} else {
			a.So(found, should.BeFalse)
		}
	}
}

func TestChannelMasksCFList(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	usChannels := []*ttnpb.MACParameters_Channel{
		{EnableUplink: false, UplinkFrequency: 903700000},
		{EnableUplink: true, UplinkFrequency: 903900000},
		{EnableUplink: true, UplinkFrequency: 904100000},
		{EnableUplink: true, UplinkFrequency: 904300000},
		{EnableUplink: true, UplinkFrequency: 904500000},
		{EnableUplink: true, UplinkFrequency: 904700000},
		{EnableUplink: true, UplinkFrequency: 904900000},
		{EnableUplink: true, UplinkFrequency: 905100000},
		{EnableUplink: true, UplinkFrequency: 905300000},
		{EnableUplink: false, UplinkFrequency: 905500000},
	}

	phy := &band.US_902_928_RP1_V1_1_Rev_B
	cfList := mac.CFList(phy, usChannels...)

	enabledChannels := []int{8, 9, 10, 11, 12, 13, 14, 15}
outer:
	for index, chMaskEntry := range cfList.ChMasks {
		for _, enabledChannel := range enabledChannels {
			if enabledChannel == index {
				a.So(chMaskEntry, should.BeTrue)
				continue outer
			}
		}
		a.So(chMaskEntry, should.BeFalse)
	}
}

func TestUnimplementedCFList(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	cfList := mac.CFList(&band.US_902_928_TS1_V1_0_1)
	a.So(cfList, should.BeNil)
}
