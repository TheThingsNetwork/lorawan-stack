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
	CR_4_5   = "4/5"
	CR_4_6   = "4/6"
	CR_4_7   = "4/7"
	CR_4_8   = "4/8"
	CR_4_8LI = "4/8LI"
)

// MaxMACPayloadSizeFunc is a function that returns the maximum payload size
// depending on whether dwell time restrictions apply or not.
type MaxMACPayloadSizeFunc func(dwellTime bool) uint16

func makeConstMaxMACPayloadSizeFunc(v uint16) MaxMACPayloadSizeFunc {
	return func(_ bool) uint16 {
		return v
	}
}

func makeDwellTimeMaxMACPayloadSizeFunc(noDwellTimeSize, dwellTimeSize uint16) MaxMACPayloadSizeFunc {
	return func(dwellTime bool) uint16 {
		if dwellTime {
			return dwellTimeSize
		}
		return noDwellTimeSize
	}
}

// DataRate indicates the properties of a band's data rate.
type DataRate struct {
	Rate              *ttnpb.DataRate
	MaxMACPayloadSize MaxMACPayloadSizeFunc
}

func makeLoRaDataRate(spreadingFactor uint8, bandwidth uint32, codingRate string, maximumMACPayloadSize MaxMACPayloadSizeFunc) DataRate {
	return DataRate{
		Rate: (&ttnpb.LoRaDataRate{
			SpreadingFactor: uint32(spreadingFactor),
			Bandwidth:       bandwidth,
			CodingRate:      codingRate,
		}).DataRate(),
		MaxMACPayloadSize: maximumMACPayloadSize,
	}
}

func makeFSKDataRate(bitRate uint32, maximumMACPayloadSize MaxMACPayloadSizeFunc) DataRate {
	return DataRate{
		Rate: (&ttnpb.FSKDataRate{
			BitRate: bitRate,
		}).DataRate(),
		MaxMACPayloadSize: maximumMACPayloadSize,
	}
}

func makeLRFHSSDataRate(modulationType uint32, operatingChannelWidth uint32, codingRate string, maximumMACPayloadSize MaxMACPayloadSizeFunc) DataRate {
	return DataRate{
		Rate: (&ttnpb.LRFHSSDataRate{
			ModulationType:        modulationType,
			OperatingChannelWidth: operatingChannelWidth,
			CodingRate:            codingRate,
		}).DataRate(),
		MaxMACPayloadSize: maximumMACPayloadSize,
	}
}
