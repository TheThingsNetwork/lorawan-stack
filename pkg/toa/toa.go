// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// Package toa provides methods for computing a LoRaWAN packet's time-on-air.
package toa

import (
	"math"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Compute computes the time-on-air for the given payload size and the TxSettings.
// This function takes into account PHYPayload.
func Compute(payloadSize int, settings ttnpb.TxSettings) (d time.Duration, err error) {
	switch dr := settings.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_LoRa:
		return computeLoRa(payloadSize, settings.Frequency, uint8(dr.LoRa.SpreadingFactor), dr.LoRa.Bandwidth, settings.CodingRate)
	case *ttnpb.DataRate_FSK:
		return computeFSK(payloadSize, settings.Frequency, dr.FSK.BitRate)
	default:
		panic("invalid modulation")
	}
}

var (
	errBandwidth       = errors.DefineInvalidArgument("bandwidth", "invalid bandwidth")
	errSpreadingFactor = errors.DefineInvalidArgument("spreading_factor", "invalid spreading factor")
	errCodingRate      = errors.DefineInvalidArgument("coding_rate", "invalid coding rate")
	errFrequency       = errors.DefineInvalidArgument("frequency", "invalid frequency")
)

func computeLoRa(payloadSize int, frequency uint64, spreadingFactor uint8, bandwidth uint32, codingRate string) (time.Duration, error) {
	if spreadingFactor < 5 || spreadingFactor > 12 {
		return 0, errSpreadingFactor
	}
	if bandwidth == 0 {
		return 0, errBandwidth
	}

	switch {
	case frequency < 1000000000:
		// See http://www.semtech.com/images/datasheet/LoraDesignGuide_STD.pdf, page 7.
		var cr float64
		switch codingRate {
		case "4/5":
			cr = 1
		case "4/6":
			cr = 2
		case "4/7":
			cr = 3
		case "4/8":
			cr = 4
		default:
			return 0, errCodingRate
		}
		var de float64
		if bandwidth == 125000 && (spreadingFactor == 11 || spreadingFactor == 12) {
			de = 1.0
		}
		pl := float64(payloadSize)
		floatSF := float64(spreadingFactor)
		floatBW := float64(bandwidth) / 1000
		h := 0.0 // 0 means header is enabled
		tSym := math.Pow(2, floatSF) / floatBW
		payloadNb := 8.0 + math.Max(0.0, math.Ceil((8.0*pl-4.0*floatSF+28.0+16.0-20.0*h)/(4.0*(floatSF-2.0*de)))*(cr+4.0))
		timeOnAir := (payloadNb + 12.25) * tSym * 1000000 // in nanoseconds
		return time.Duration(timeOnAir), nil

	default:
		return 0, errFrequency
	}
}

func computeFSK(payloadSize int, frequency uint64, bitRate uint32) (time.Duration, error) {
	switch {
	case frequency < 1000000000:
		timeOnAir := int64(time.Second) * (int64(payloadSize) + 5 + 3 + 1 + 2) * 8 / int64(bitRate)
		return time.Duration(timeOnAir), nil

	default:
		return 0, errFrequency
	}
}
