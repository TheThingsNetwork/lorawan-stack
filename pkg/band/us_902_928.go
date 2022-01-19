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
	// US_902_928 is the ID of the US frequency plan
	US_902_928 = "US_902_928"
)

var (
	us902928UplinkChannels = func() []Channel {
		uplinkChannels := make([]Channel, 0, 72)
		for i := 0; i < 64; i++ {
			uplinkChannels = append(uplinkChannels, Channel{
				Frequency:   uint64(902300000 + 200000*i),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_3,
			})
		}
		for i := 0; i < 8; i++ {
			uplinkChannels = append(uplinkChannels, Channel{
				Frequency:   uint64(903000000 + 1600000*i),
				MinDataRate: ttnpb.DataRateIndex_DATA_RATE_4,
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_4,
			})
		}
		return uplinkChannels
	}()

	us902928DownlinkChannels = func() []Channel {
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

	us902928DownlinkDRTable = [5][4]ttnpb.DataRateIndex{
		{10, 9, 8, 8},
		{11, 10, 9, 8},
		{12, 11, 10, 9},
		{13, 12, 11, 10},
		{13, 13, 12, 11},
	}
)
