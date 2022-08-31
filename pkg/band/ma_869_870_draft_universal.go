// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// MA_869_870_Draft_Universal is the band definition for MA869-870.
var MA_869_870_Draft_Universal = Band{
	ID: MA_869_870_DRAFT,

	SupportsDynamicADR: true,

	MaxUplinkChannels: 16,
	UplinkChannels:    ma869870DraftDefaultChannels,

	MaxDownlinkChannels: 16,
	DownlinkChannels:    ma869870DraftDefaultChannels,

	// ANRT DG No 07/2021 of May 2021.
	SubBands: []SubBandParameters{
		{
			MinFrequency: 869000000,
			MaxFrequency: 869400000,
			DutyCycle:    0.001,
			MaxEIRP:      14.0 + eirpDelta,
		},
		{
			MinFrequency: 869400000,
			MaxFrequency: 869650000,
			DutyCycle:    0.1,
			MaxEIRP:      27.0 + eirpDelta,
		},
		{
			MinFrequency: 869650000,
			MaxFrequency: 870000000,
			DutyCycle:    0.01,
			MaxEIRP:      14.0 + eirpDelta,
		},
	},

	DataRates: map[ttnpb.DataRateIndex]DataRate{
		ttnpb.DataRateIndex_DATA_RATE_0:  makeLoRaDataRate(12, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_1:  makeLoRaDataRate(11, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_2:  makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_3:  makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(123)),
		ttnpb.DataRateIndex_DATA_RATE_4:  makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(250)),
		ttnpb.DataRateIndex_DATA_RATE_5:  makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(250)),
		ttnpb.DataRateIndex_DATA_RATE_6:  makeLoRaDataRate(7, 250000, makeConstMaxMACPayloadSizeFunc(250)),
		ttnpb.DataRateIndex_DATA_RATE_7:  makeFSKDataRate(50000, makeConstMaxMACPayloadSizeFunc(250)),
		ttnpb.DataRateIndex_DATA_RATE_8:  makeLRFHSSDataRate(0, 137000, "1/3", makeConstMaxMACPayloadSizeFunc(58)),
		ttnpb.DataRateIndex_DATA_RATE_9:  makeLRFHSSDataRate(0, 137000, "2/3", makeConstMaxMACPayloadSizeFunc(123)),
		ttnpb.DataRateIndex_DATA_RATE_10: makeLRFHSSDataRate(0, 336000, "1/3", makeConstMaxMACPayloadSizeFunc(58)),
		ttnpb.DataRateIndex_DATA_RATE_11: makeLRFHSSDataRate(0, 336000, "2/3", makeConstMaxMACPayloadSizeFunc(123)),
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
		if idx > ttnpb.DataRateIndex_DATA_RATE_7 {
			return 0, errDataRateIndexTooHigh.WithAttributes("max", 7)
		}
		if offset > 5 {
			return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
		}
		return ma869870DraftDownlinkDRTable[idx][offset], nil
	},

	GenerateChMasks: generateChMask16,
	ParseChMask:     parseChMask16,

	LoRaCodingRate: "4/5",

	FreqMultiplier:   100,
	ImplementsCFList: true,
	CFListType:       ttnpb.CFListType_FREQUENCIES,

	DefaultRx2Parameters: Rx2Parameters{ttnpb.DataRateIndex_DATA_RATE_0, 869525000},

	Beacon: Beacon{
		DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
		CodingRate:    "4/5",
		Frequencies:   ma869870DraftBeaconFrequencies,
	},
	PingSlotFrequencies: ma869870DraftBeaconFrequencies,
}
