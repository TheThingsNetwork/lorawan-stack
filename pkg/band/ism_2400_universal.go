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

// ISM_2400_Universal is the band definition for universal LoRa 2.4 GHz.
var ISM_2400_Universal = Band{
	ID: ISM_2400,

	MaxUplinkChannels: 16,
	UplinkChannels:    ism2400DefaultChannels,

	MaxDownlinkChannels: 16,
	DownlinkChannels:    ism2400DefaultChannels,

	SubBands: []SubBandParameters{
		{
			MinFrequency: 2400000000,
			MaxFrequency: 2500000000,
			DutyCycle:    1,
			MaxEIRP:      8.0 + eirpDelta,
		},
	},

	DataRates: map[ttnpb.DataRateIndex]DataRate{
		ttnpb.DataRateIndex_DATA_RATE_0: makeLoRaDataRate(12, 812000, makeConstMaxMACPayloadSizeFunc(59)),
		ttnpb.DataRateIndex_DATA_RATE_1: makeLoRaDataRate(11, 812000, makeConstMaxMACPayloadSizeFunc(123)),
		ttnpb.DataRateIndex_DATA_RATE_2: makeLoRaDataRate(10, 812000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_3: makeLoRaDataRate(9, 812000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_4: makeLoRaDataRate(8, 812000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_5: makeLoRaDataRate(7, 812000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_6: makeLoRaDataRate(6, 812000, makeConstMaxMACPayloadSizeFunc(230)),
		ttnpb.DataRateIndex_DATA_RATE_7: makeLoRaDataRate(5, 812000, makeConstMaxMACPayloadSizeFunc(230)),
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
		if idx > ttnpb.DataRateIndex_DATA_RATE_7 {
			return 0, errDataRateIndexTooHigh.WithAttributes("max", 7)
		}
		if offset > 5 {
			return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
		}
		return ism2400DownlinkDRTable[idx][offset], nil
	},

	GenerateChMasks: generateChMask16,
	ParseChMask:     parseChMask16,

	LoRaCodingRate: "4/8LI",

	FreqMultiplier:   200,
	ImplementsCFList: true,
	CFListType:       ttnpb.CFListType_FREQUENCIES,

	DefaultRx2Parameters: Rx2Parameters{ttnpb.DataRateIndex_DATA_RATE_0, 2423000000},

	Beacon: Beacon{
		DataRateIndex:    ttnpb.DataRateIndex_DATA_RATE_3,
		CodingRate:       "4/8LI",
		ComputeFrequency: func(_ float64) uint64 { return ism2400BeaconFrequency },
	},
	PingSlotFrequency: uint64Ptr(ism2400BeaconFrequency),
}
