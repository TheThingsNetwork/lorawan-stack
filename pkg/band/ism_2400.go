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
	// ISM_2400 is the ID of the LoRa 2.4 GHz band.
	ISM_2400 = "ISM_2400"

	ism2400BeaconFrequency = 2424000000
)

var (
	ism2400DefaultChannels = []Channel{
		{
			Frequency:   2403000000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_7,
		},
		{
			Frequency:   2425000000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_7,
		},
		{
			Frequency:   2479000000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_7,
		},
	}

	ism2400DownlinkDRTable = [8][6]ttnpb.DataRateIndex{
		{0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0},
		{2, 1, 0, 0, 0, 0},
		{3, 2, 1, 0, 0, 0},
		{4, 3, 2, 1, 0, 0},
		{5, 4, 3, 2, 1, 0},
		{6, 5, 4, 3, 2, 1},
		{7, 6, 5, 4, 3, 2},
	}
)
