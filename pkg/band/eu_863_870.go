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

var eu_863_870 Band

// EU_863_870 is the ID of the European 863-870Mhz band
const EU_863_870 = "EU_863_870"

//revive:enable:var-naming

func init() {
	defaultChannels := []Channel{
		{
			Frequency:   868100000,
			MaxDataRate: ttnpb.DATA_RATE_5,
		},
		{
			Frequency:   868300000,
			MaxDataRate: ttnpb.DATA_RATE_5,
		},
		{
			Frequency:   868500000,
			MaxDataRate: ttnpb.DATA_RATE_5,
		},
	}
	const beaconFrequency = 869525000

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

	eu_863_870 = Band{
		ID: EU_863_870,

		EnableADR: true,

		MaxUplinkChannels: 16,
		UplinkChannels:    defaultChannels,

		MaxDownlinkChannels: 16,
		DownlinkChannels:    defaultChannels,

		// See ETSI EN 300.220-2 V3.1.1 (2017-02)
		SubBands: []SubBandParameters{
			{
				// Band K
				MinFrequency: 863000000,
				MaxFrequency: 865000000,
				DutyCycle:    0.001,
				MaxEIRP:      14.0 + eirpDelta,
			},
			{
				// Band L
				MinFrequency: 865000000,
				MaxFrequency: 868000000,
				DutyCycle:    0.01,
				MaxEIRP:      14.0 + eirpDelta,
			},
			{
				// Band M
				MinFrequency: 868000000,
				MaxFrequency: 868600000,
				DutyCycle:    0.01,
				MaxEIRP:      14.0 + eirpDelta,
			},
			{
				// Band N
				MinFrequency: 868700000,
				MaxFrequency: 869200000,
				DutyCycle:    0.001,
				MaxEIRP:      14.0 + eirpDelta,
			},
			// Band O is skipped intentionally
			{
				// Band P
				MinFrequency: 869400000,
				MaxFrequency: 869650000,
				DutyCycle:    0.1,
				MaxEIRP:      27.0 + eirpDelta,
			},
			{
				// Band R
				MinFrequency: 869700000,
				MaxFrequency: 870000000,
				DutyCycle:    0.01,
				MaxEIRP:      14.0 + eirpDelta,
			},
		},

		DataRates: map[ttnpb.DataRateIndex]DataRate{
			ttnpb.DATA_RATE_0: makeLoRaDataRate(12, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_1: makeLoRaDataRate(11, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_2: makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_3: makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(123)),
			ttnpb.DATA_RATE_4: makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_5: makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_6: makeLoRaDataRate(7, 250000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_7: makeFSKDataRate(50000, makeConstMaxMACPayloadSizeFunc(230)),
		},
		MaxADRDataRateIndex: ttnpb.DATA_RATE_5,

		ReceiveDelay1:        defaultReceiveDelay1,
		ReceiveDelay2:        defaultReceiveDelay2,
		JoinAcceptDelay1:     defaultJoinAcceptDelay1,
		JoinAcceptDelay2:     defaultJoinAcceptDelay2,
		MaxFCntGap:           defaultMaxFCntGap,
		ADRAckLimit:          defaultADRAckLimit,
		ADRAckDelay:          defaultADRAckDelay,
		MinRetransmitTimeout: defaultRetransmitTimeout - defaultRetransmitTimeoutMargin,
		MaxRetransmitTimeout: defaultRetransmitTimeout + defaultRetransmitTimeoutMargin,

		DefaultMaxEIRP: 16,
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

		LoRaCodingRate: "4/5",

		FreqMultiplier:   100,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		DefaultRx2Parameters: Rx2Parameters{ttnpb.DATA_RATE_0, 869525000},

		Beacon: Beacon{
			DataRateIndex:    ttnpb.DATA_RATE_3,
			CodingRate:       "4/5",
			ComputeFrequency: func(_ float64) uint64 { return beaconFrequency },
		},
		PingSlotFrequency: uint64Ptr(beaconFrequency),

		regionalParameters1_v1_0:   bandIdentity,
		regionalParameters1_v1_0_1: bandIdentity,
		regionalParameters1_v1_0_2: composeSwaps(
			makeSetMaxTxPowerIndexFunc(5),
		),
		regionalParameters1_v1_0_2RevB: bandIdentity,
		regionalParameters1_v1_0_3RevA: bandIdentity,
		regionalParameters1_v1_1RevA:   bandIdentity,
	}
	All[EU_863_870] = eu_863_870
}
