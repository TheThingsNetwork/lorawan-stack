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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Compute computes the time-on-air for the given payload size and the TxSettings.
// This function takes into account PHYPayload.
func Compute(payloadSize int, settings *ttnpb.TxSettings) (d time.Duration, err error) {
	switch dr := settings.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_Lora:
		return computeLoRa(payloadSize, settings.Frequency, uint8(dr.Lora.SpreadingFactor), dr.Lora.Bandwidth, settings.CodingRate, settings.EnableCrc)
	case *ttnpb.DataRate_Fsk:
		return computeFSK(payloadSize, settings.Frequency, dr.Fsk.BitRate, settings.EnableCrc)
	case *ttnpb.DataRate_Lrfhss:
		return computeLRFHSS(payloadSize, settings.CodingRate, settings.EnableCrc)
	default:
		panic("invalid modulation")
	}
}

var (
	errBandwidth       = errors.DefineInvalidArgument("bandwidth", "invalid bandwidth `{bandwidth}`")
	errSpreadingFactor = errors.DefineInvalidArgument("spreading_factor", "invalid spreading factor `{spreading_factor}`")
	errCodingRate      = errors.DefineInvalidArgument("coding_rate", "invalid coding rate `{coding_rate}`")
	errFrequency       = errors.DefineInvalidArgument("frequency", "invalid frequency `{frequency}`")
)

func computeLoRa(payloadSize int, frequency uint64, spreadingFactor uint8, bandwidth uint32, codingRate string, crc bool) (time.Duration, error) {
	if spreadingFactor < 5 || spreadingFactor > 12 {
		return 0, errSpreadingFactor.WithAttributes("spreading_factor", spreadingFactor)
	}
	if bandwidth == 0 {
		return 0, errBandwidth.WithAttributes("bandwidth", bandwidth)
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
			return 0, errCodingRate.WithAttributes("coding_rate", codingRate)
		}
		var de float64
		if bandwidth == 125000 && (spreadingFactor == 11 || spreadingFactor == 12) {
			de = 1.0
		}
		pl := float64(payloadSize)
		sf := float64(spreadingFactor)
		bw := float64(bandwidth) / 1000
		h := 0.0 // 0 means header is enabled
		tSym := math.Pow(2, sf) / bw
		payloadNb := 8.0 + math.Max(0.0, math.Ceil((8.0*pl-4.0*sf+28.0+16.0-20.0*h)/(4.0*(sf-2.0*de)))*(cr+4.0))
		timeOnAir := (payloadNb + 12.25) * tSym * 1000000
		return time.Duration(timeOnAir), nil

	case frequency >= 2400000000 && frequency < 2500000000:
		// See Semtech SX1280/SX1281/SX1282 Data Sheet Rev 3.0, 7.4.4
		nBitCRC := 0.0
		if crc {
			nBitCRC = 16.0
		}
		sf := float64(spreadingFactor)
		bw := float64(bandwidth) / 1000
		var cr float64
		switch codingRate {
		case "4/5LI":
			cr = 5
		case "4/6LI":
			cr = 6
		case "4/7LI", "4/8LI": // 4/7LI is wrongly defined; it is in fact 4/8LI.
			cr = 8
		default:
			return 0, errCodingRate.New()
		}
		var nBitHeaderSpace float64
		var denominator float64
		nPreamble := 8.0
		if spreadingFactor < 7 {
			nBitHeaderSpace = math.Floor((sf-5)/2) * 8
			nPreamble += 6.25
			denominator = 4 * sf
		} else if spreadingFactor >= 7 && spreadingFactor <= 10 {
			nBitHeaderSpace = math.Floor((sf-7)/2) * 8
			nPreamble += 4.25
			denominator = 4 * sf
		} else {
			nBitHeaderSpace = math.Floor((sf-7)/2) * 8
			nPreamble += 4.25
			denominator = 4 * (sf - 2)
		}
		var nSymbol float64
		nBytePayload := float64(payloadSize)
		if 8.0*nBytePayload+nBitCRC > nBitHeaderSpace {
			nSymbol = nPreamble + 8.0 + math.Ceil(math.Max(0, 8*nBytePayload+nBitCRC-math.Min(nBitHeaderSpace, 8.0*nBytePayload))/denominator*cr)
		} else {
			nSymbol = nPreamble + 8.0 + math.Ceil(math.Max(0, 8*nBytePayload+nBitCRC-nBitHeaderSpace)/denominator*cr)
		}
		timeOnAir := math.Pow(2, sf) / bw * nSymbol * 1000000
		return time.Duration(timeOnAir), nil

	default:
		return 0, errFrequency.WithAttributes("frequency", frequency)
	}
}

func computeFSK(payloadSize int, frequency uint64, bitRate uint32, crc bool) (time.Duration, error) {
	switch {
	case frequency < 1000000000:
		timeOnAir := int64(time.Second) * (int64(payloadSize) + 5 + 3 + 1 + 2) * 8 / int64(bitRate)
		return time.Duration(timeOnAir), nil

	default:
		return 0, errFrequency.WithAttributes("frequency", frequency)
	}
}

func computeLRFHSS(phyPayloadLength int, codingRate string, crc bool) (time.Duration, error) {
	var n int
	switch codingRate {
	case "1/3":
		n = 3
	case "2/3":
		n = 2
	default:
		return 0, errCodingRate.WithAttributes("coding_rate", codingRate)
	}

	timeOnAir := time.Duration(n) * 233472 * time.Microsecond
	switch codingRate {
	case "1/3":
		timeOnAir += time.Duration(math.Ceil((float64(phyPayloadLength+3) / 2))) * 102400 * time.Microsecond
	case "2/3":
		timeOnAir += time.Duration(math.Ceil((float64(phyPayloadLength+3) / 4))) * 102400 * time.Microsecond
	}
	return timeOnAir, nil
}
