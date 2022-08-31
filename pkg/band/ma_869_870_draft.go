// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	// MA_869_870_DRAFT is the ID of the draft Morocco 869-870Mhz band.
	MA_869_870_DRAFT = "MA_869_870_DRAFT"
)

var (
	ma869870DraftBeaconFrequencies = []uint64{869525000}

	ma869870DraftDefaultChannels = []Channel{
		{
			Frequency:   869100000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Frequency:   869300000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Frequency:   869700000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
	}

	ma869870DraftDownlinkDRTable = [8][6]ttnpb.DataRateIndex{
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
