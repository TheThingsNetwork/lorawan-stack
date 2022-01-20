// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package band

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

const (
	// AU_915_928 is the ID of the Australian band
	AU_915_928 = "AU_915_928"
)

var (
	au915928DownlinkDRTable = [7][6]ttnpb.DataRateIndex{
		{8, 8, 8, 8, 8, 8},
		{9, 8, 8, 8, 8, 8},
		{10, 9, 8, 8, 8, 8},
		{11, 10, 9, 8, 8, 8},
		{12, 11, 10, 9, 8, 8},
		{13, 12, 11, 10, 9, 8},
		{13, 13, 12, 11, 10, 9},
	}

	au915928UplinkChannels = func(delta ttnpb.DataRateIndex) []Channel {
		uplinkChannels := make([]Channel, 0, 72)
		for i := 0; i < 64; i++ {
			uplinkChannels = append(uplinkChannels, Channel{
				Frequency:   uint64(915200000 + 200000*i),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5 + delta,
			})
		}
		for i := 0; i < 8; i++ {
			uplinkChannels = append(uplinkChannels, Channel{
				Frequency:   uint64(915900000 + 1600000*i),
				MinDataRate: ttnpb.DataRateIndex_DATA_RATE_6 + delta,
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_6 + delta,
			})
		}
		return uplinkChannels
	}

	au915928DownlinkChannels = func() []Channel {
		downlinkChannels := make([]Channel, 0, 8)
		for i := 0; i < 8; i++ {
			downlinkChannels = append(downlinkChannels, Channel{
				Frequency:   uint64(923300000 + 600000*i),
				MinDataRate: ttnpb.DataRateIndex_DATA_RATE_8,
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_13,
			})
		}
		return downlinkChannels
	}()
)
