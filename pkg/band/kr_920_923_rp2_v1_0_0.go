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

// KR_920_923_RP2_V1_0_0 is the band definition for KR920-923 in the RP002-1.0.0 specification.
var KR_920_923_RP2_V1_0_0 = Band{
	ID: KR_920_923,

	EnableADR: true,

	MaxUplinkChannels: 16,
	UplinkChannels:    kr920923DefaultChannels,

	MaxDownlinkChannels: 16,
	DownlinkChannels:    kr920923DefaultChannels,

	SubBands: []SubBandParameters{
		{
			MinFrequency: 920000000,
			MaxFrequency: 923000000,
			DutyCycle:    1,
			MaxEIRP:      14.0 + eirpDelta,
		},
	},

	DataRates: map[ttnpb.DataRateIndex]DataRate{
		ttnpb.DataRateIndex_DATA_RATE_0: makeLoRaDataRate(12, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_1: makeLoRaDataRate(11, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_2: makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_3: makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(123)),
		ttnpb.DataRateIndex_DATA_RATE_4: makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_5: makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(230)),
	},
	MaxADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,

	ReceiveDelay1:        defaultReceiveDelay1,
	ReceiveDelay2:        defaultReceiveDelay2,
	JoinAcceptDelay1:     defaultJoinAcceptDelay1,
	JoinAcceptDelay2:     defaultJoinAcceptDelay2,
	MaxFCntGap:           defaultMaxFCntGap,
	ADRAckLimit:          defaultADRAckLimit,
	ADRAckDelay:          defaultADRAckDelay,
	MinRetransmitTimeout: defaultRetransmitTimeout - defaultRetransmitTimeoutMargin,
	MaxRetransmitTimeout: defaultRetransmitTimeout + defaultRetransmitTimeoutMargin,

	DefaultMaxEIRP: 14,
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
		if idx > ttnpb.DataRateIndex_DATA_RATE_5 {
			return 0, errDataRateIndexTooHigh.WithAttributes("max", 5)
		}
		if offset > 5 {
			return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
		}
		return kr920923DownlinkDRTable[idx][offset], nil
	},

	GenerateChMasks: generateChMask16,
	ParseChMask:     parseChMask16,

	LoRaCodingRate: "4/5",

	FreqMultiplier:   100,
	ImplementsCFList: true,
	CFListType:       ttnpb.CFListType_FREQUENCIES,

	DefaultRx2Parameters: Rx2Parameters{ttnpb.DataRateIndex_DATA_RATE_0, 921900000},

	Beacon: Beacon{
		DataRateIndex:    ttnpb.DataRateIndex_DATA_RATE_3,
		CodingRate:       "4/5",
		ComputeFrequency: func(_ float64) uint64 { return kr920923BeaconFrequency },
	},
	PingSlotFrequency: uint64Ptr(kr920923BeaconFrequency),
}
