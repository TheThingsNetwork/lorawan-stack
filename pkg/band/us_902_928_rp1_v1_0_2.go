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

// US_902_928_RP1_V1_0_2 is the band definition for US902-928 in the RP1 v1.0.2 specification.
var US_902_928_RP1_V1_0_2 = Band{
	ID: US_902_928,

	EnableADR: true,

	MaxUplinkChannels: 72,
	UplinkChannels:    us902928UplinkChannels,

	MaxDownlinkChannels: 8,
	DownlinkChannels:    us902928DownlinkChannels,

	// As per FCC Rules for Unlicensed Wireless Equipment operating in the ISM bands
	SubBands: []SubBandParameters{
		{
			MinFrequency: 902300000,
			MaxFrequency: 914900000,
			DutyCycle:    1,
			MaxEIRP:      21.0 + eirpDelta,
		},
		{
			MinFrequency: 923300000,
			MaxFrequency: 927500000,
			DutyCycle:    1,
			MaxEIRP:      26.0 + eirpDelta,
		},
	},

	DataRates: map[ttnpb.DataRateIndex]DataRate{
		ttnpb.DataRateIndex_DATA_RATE_0: makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(19)),
		ttnpb.DataRateIndex_DATA_RATE_1: makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(61)),
		ttnpb.DataRateIndex_DATA_RATE_2: makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(133)),
		ttnpb.DataRateIndex_DATA_RATE_3: makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(250)),
		ttnpb.DataRateIndex_DATA_RATE_4: makeLoRaDataRate(8, 500000, makeConstMaxMACPayloadSizeFunc(250)),

		ttnpb.DataRateIndex_DATA_RATE_8:  makeLoRaDataRate(12, 500000, makeConstMaxMACPayloadSizeFunc(41)),
		ttnpb.DataRateIndex_DATA_RATE_9:  makeLoRaDataRate(11, 500000, makeConstMaxMACPayloadSizeFunc(117)),
		ttnpb.DataRateIndex_DATA_RATE_10: makeLoRaDataRate(10, 500000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_11: makeLoRaDataRate(9, 500000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_12: makeLoRaDataRate(8, 500000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_13: makeLoRaDataRate(7, 500000, makeConstMaxMACPayloadSizeFunc(230)),
	},
	MaxADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,

	ReceiveDelay1:        defaultReceiveDelay1,
	ReceiveDelay2:        defaultReceiveDelay2,
	JoinAcceptDelay1:     defaultJoinAcceptDelay1,
	JoinAcceptDelay2:     defaultJoinAcceptDelay2,
	MaxFCntGap:           defaultMaxFCntGap,
	ADRAckLimit:          defaultADRAckLimit,
	ADRAckDelay:          defaultADRAckDelay,
	MinRetransmitTimeout: defaultRetransmitTimeout - defaultRetransmitTimeoutMargin,
	MaxRetransmitTimeout: defaultRetransmitTimeout + defaultRetransmitTimeoutMargin,

	DefaultMaxEIRP: 30,
	TxOffset: []float32{
		0,
		-2,
		-4,
		-6,
		-8,
		-10,
		-12,
		-14,
		-16,
		-18,
		-20,
	},

	Rx1Channel: channelIndexModulo(8),
	Rx1DataRate: func(idx ttnpb.DataRateIndex, offset ttnpb.DataRateOffset, _ bool) (ttnpb.DataRateIndex, error) {
		if idx > ttnpb.DataRateIndex_DATA_RATE_4 {
			return 0, errDataRateIndexTooHigh.WithAttributes("max", 4)
		}
		if offset > 3 {
			return 0, errDataRateOffsetTooHigh.WithAttributes("max", 3)
		}
		return us902928DownlinkDRTable[idx][offset], nil
	},

	GenerateChMasks: makeGenerateChMask72(false),
	ParseChMask:     parseChMask72,

	LoRaCodingRate: "4/5",

	FreqMultiplier:   100,
	ImplementsCFList: false,
	CFListType:       ttnpb.CFListType_CHANNEL_MASKS,

	DefaultRx2Parameters: Rx2Parameters{ttnpb.DataRateIndex_DATA_RATE_8, 923300000},

	Beacon: Beacon{
		DataRateIndex:    ttnpb.DataRateIndex_DATA_RATE_3,
		CodingRate:       "4/5",
		ComputeFrequency: makeBeaconFrequencyFunc(usAuBeaconFrequencies),
	},
}
