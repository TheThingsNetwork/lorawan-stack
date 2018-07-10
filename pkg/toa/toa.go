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

// Package toa provides methods for computing a LoRaWAN packet's time-on-air
package toa

import (
	"math"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

// Compute the time-on-air from the payload and RF parameters. This function only takes into account the PHY payload.
//
// See http://www.semtech.com/images/datasheet/LoraDesignGuide_STD.pdf, page 7
func Compute(rawPayload []byte, settings ttnpb.TxSettings) (d time.Duration, err error) {
	switch settings.Modulation {
	case ttnpb.Modulation_LORA:
		return computeLoRa(rawPayload, settings)
	case ttnpb.Modulation_FSK:
		return computeFSK(rawPayload, settings), nil
	default:
		panic("invalid modulation")
	}
}

var codingRate = map[string]float64{
	"4/5": 1,
	"4/6": 2,
	"4/7": 3,
	"4/8": 4,
}

func computeLoRa(rawPayload []byte, settings ttnpb.TxSettings) (time.Duration, error) {
	err := validate.All(
		validate.LoRaBandwidth(int(settings.Bandwidth/1000)),
		validate.LoRaSpreadingFactor(int(settings.SpreadingFactor)),
		validate.LoRaCodingRateString(settings.CodingRate),
	)
	if err != nil {
		return 0, err
	}

	cr := codingRate[settings.CodingRate]
	bandwidth := settings.Bandwidth / 1000 // Bandwidth in KHz
	spreadingFactor := settings.SpreadingFactor

	var de float64
	if bandwidth == 125 && (spreadingFactor == 11 || spreadingFactor == 12) {
		de = 1.0
	}

	pl := float64(len(rawPayload))
	floatBW := float64(bandwidth)
	floatSF := float64(spreadingFactor)
	h := 0.0 // 0 means header is enabled

	tSym := math.Pow(2, floatSF) / floatBW

	payloadNb := 8.0 + math.Max(0.0, math.Ceil((8.0*pl-4.0*floatSF+28.0+16.0-20.0*h)/(4.0*(floatSF-2.0*de)))*(cr+4.0))
	timeOnAir := (payloadNb + 12.25) * tSym * 1000000 // in nanoseconds

	return time.Duration(timeOnAir), nil
}

func computeFSK(rawPayload []byte, settings ttnpb.TxSettings) time.Duration {
	payloadSize := len(rawPayload)
	bitRate := settings.BitRate

	tPkt := int64(time.Second) * (int64(payloadSize) + 5 + 3 + 1 + 2) * 8 / int64(bitRate)

	return time.Duration(tPkt)
}
