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

package validate

import "go.thethings.network/lorawan-stack/pkg/errors"

// ErrInvalidBandwidth indicates an invalid bandwidth.
var ErrInvalidBandwidth = errors.DefineInvalidArgument("bandwidth", "invalid bandwidth")

// LoRaBandwidth validates a LoRa bandwidth.
func LoRaBandwidth(bw int) error {
	switch bw {
	case 125, 250, 500:
		return nil
	default:
		return ErrInvalidBandwidth
	}
}

// ErrInvalidSpreadingFactor indicates an invalid spreading factor.
var ErrInvalidSpreadingFactor = errors.DefineInvalidArgument("spreading_factor", "invalid spreading factor")

// LoRaSpreadingFactor validates a LoRa spreading factor.
func LoRaSpreadingFactor(sf int) error {
	switch sf {
	case 6, 7, 8, 9, 10, 11, 12:
		return nil
	default:
		return ErrInvalidSpreadingFactor
	}
}

// ErrInvalidCodingRate indicates an invalid LoRa coding rate.
var ErrInvalidCodingRate = errors.DefineInvalidArgument("coding_rate", "invalid coding rate")

// LoRaCodingRate validatess a LoRa coding rate.
func LoRaCodingRate(cr int) error {
	if cr >= 1 && cr <= 4 {
		return nil
	}
	return ErrInvalidCodingRate
}

// LoRaCodingRateString validates a LoRa coding rate string.
func LoRaCodingRateString(cr string) error {
	switch cr {
	case "4/5", "4/6", "4/7", "4/8":
		return nil
	default:
		return ErrInvalidCodingRate
	}
}
