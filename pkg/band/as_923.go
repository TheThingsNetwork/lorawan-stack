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

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type as923GroupOffset int64

const (
	// AS_923 is the ID of the Asian 923Mhz Group 1 band
	AS_923 = "AS_923"
	// AS_923_2 is the ID of the Asian 923Mhz Group 2 band
	AS_923_2 = "AS_923_2"
	// AS_923_3 is the ID of the Asian 923Mhz Group 3 band
	AS_923_3 = "AS_923_3"

	as923Group1Offset as923GroupOffset = 0
	as923Group2Offset as923GroupOffset = -1.8 * 1e6
	as923Group3Offset as923GroupOffset = -6.6 * 1e6
)

var (
	as923BeaconFrequency = func(offset as923GroupOffset) uint64 {
		return uint64(923400000 + offset)
	}

	as923DefaultChannels = func(offset as923GroupOffset) []Channel {
		return []Channel{
			{
				Frequency:   uint64(923200000 + offset),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
			},
			{
				Frequency:   uint64(923400000 + offset),
				MaxDataRate: ttnpb.DataRateIndex_DATA_RATE_5,
			},
		}
	}

	as923DefaultRX2Frequency = func(offset as923GroupOffset) uint64 {
		return uint64(923200000 + offset)
	}

	as923SubBandParameters = func(offset as923GroupOffset) []SubBandParameters {
		var minFrequency, maxFrequency uint64
		switch offset {
		case as923Group1Offset:
			minFrequency = 923000000
			maxFrequency = 923500000
		case as923Group2Offset:
			minFrequency = 921400000
			maxFrequency = 922000000
		case as923Group3Offset:
			minFrequency = 916500000
			maxFrequency = 917000000
		default:
			panic(fmt.Sprintf("unknown offset %v", offset))
		}
		return []SubBandParameters{
			{
				MinFrequency: minFrequency,
				MaxFrequency: maxFrequency,
				DutyCycle:    0.01,
				MaxEIRP:      16,
			},
		}
	}
)
