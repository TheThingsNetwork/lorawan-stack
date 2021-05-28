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

package band

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

//revive:disable:var-naming

var ism_2400 Band

// ISM_2400 is the ID of the LoRa 2.4 GHz band.
const ISM_2400 = "ISM_2400"

//revive:enable:var-naming

func init() {
	defaultChannels := []Channel{
		{
			Frequency:   2403000000,
			MaxDataRate: ttnpb.DATA_RATE_7,
		},
		{
			Frequency:   2425000000,
			MaxDataRate: ttnpb.DATA_RATE_7,
		},
		{
			Frequency:   2479000000,
			MaxDataRate: ttnpb.DATA_RATE_7,
		},
	}
	const beaconFrequency = 2424000000

	downlinkDRTable := [8][6]ttnpb.DataRateIndex{
		{0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0},
		{2, 1, 0, 0, 0, 0},
		{3, 2, 1, 0, 0, 0},
		{4, 3, 2, 1, 0, 0},
		{5, 4, 3, 2, 1, 0},
		{6, 5, 4, 3, 2, 1},
		{7, 6, 5, 4, 3, 2},
	}

	ism_2400 = Band{
		ID: ISM_2400,

		MaxUplinkChannels: 16,
		UplinkChannels:    defaultChannels,

		MaxDownlinkChannels: 16,
		DownlinkChannels:    defaultChannels,

		SubBands: []SubBandParameters{
			{
				MinFrequency: 2400000000,
				MaxFrequency: 2500000000,
				DutyCycle:    1,
				MaxEIRP:      8.0 + eirpDelta,
			},
		},

		DataRates: map[ttnpb.DataRateIndex]DataRate{
			ttnpb.DATA_RATE_0: makeLoRaDataRate(12, 812000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_1: makeLoRaDataRate(11, 812000, makeConstMaxMACPayloadSizeFunc(123)),
			ttnpb.DATA_RATE_2: makeLoRaDataRate(10, 812000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_3: makeLoRaDataRate(9, 812000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_4: makeLoRaDataRate(8, 812000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_5: makeLoRaDataRate(7, 812000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_6: makeLoRaDataRate(6, 812000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_7: makeLoRaDataRate(5, 812000, makeConstMaxMACPayloadSizeFunc(230)),
		},

		ReceiveDelay1:        defaultReceiveDelay1,
		ReceiveDelay2:        defaultReceiveDelay2,
		JoinAcceptDelay1:     defaultJoinAcceptDelay1,
		JoinAcceptDelay2:     defaultJoinAcceptDelay2,
		MaxFCntGap:           defaultMaxFCntGap,
		ADRAckLimit:          defaultADRAckLimit,
		ADRAckDelay:          defaultADRAckDelay,
		MinRetransmitTimeout: defaultRetransmitTimeout - defaultRetransmitTimeoutMargin,
		MaxRetransmitTimeout: defaultRetransmitTimeout + defaultRetransmitTimeoutMargin,

		DefaultMaxEIRP: 10,
		TxOffset: []float32{
			0,
			-2,
			-4,
			-6,
			-8,
			-10,
			-12,
			-14,
		},

		Rx1Channel: channelIndexIdentity,
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset ttnpb.DataRateOffset, _ bool) (ttnpb.DataRateIndex, error) {
			if idx > ttnpb.DATA_RATE_7 {
				return 0, errDataRateIndexTooHigh.WithAttributes("max", 7)
			}
			if offset > 5 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
			}
			return downlinkDRTable[idx][offset], nil
		},

		GenerateChMasks: generateChMask16,
		ParseChMask:     parseChMask16,

		LoRaCodingRate: "4/8LI",

		FreqMultiplier:   200,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		DefaultRx2Parameters: Rx2Parameters{ttnpb.DATA_RATE_0, 2423000000},

		Beacon: Beacon{
			DataRateIndex:    ttnpb.DATA_RATE_3,
			CodingRate:       "4/8LI",
			ComputeFrequency: func(_ float64) uint64 { return beaconFrequency },
		},
		PingSlotFrequency: uint64Ptr(beaconFrequency),

		regionalParameters1_v1_0:       bandIdentity,
		regionalParameters1_v1_0_1:     bandIdentity,
		regionalParameters1_v1_0_2:     bandIdentity,
		regionalParameters1_v1_0_2RevB: bandIdentity,
		regionalParameters1_v1_0_3RevA: bandIdentity,
		regionalParameters1_v1_1RevA:   bandIdentity,
	}
	All[ISM_2400] = ism_2400
}
