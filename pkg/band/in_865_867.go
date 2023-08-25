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
	// IN_865_867 is the ID of the Indian frequency plan
	IN_865_867 = "IN_865_867"
)

var (
	in865867BeaconFrequencies = []uint64{866500000}

	in865867DefaultChannels = []Channel{
		{
			Frequency:   865062500,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Frequency:   865402500,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Frequency:   865985000,
			MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
		},
	}

	in865867RelayParameters = RelayParameters{
		WORChannels: []RelayWORChannel{
			{
				Frequency:     866000000,
				ACKFrequency:  866200000,
				DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
			},
			{
				Frequency:     866700000,
				ACKFrequency:  866900000,
				DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
			},
		},
	}
)
