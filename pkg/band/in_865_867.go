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

var in_865_867 Band

// IN_865_867 is the ID of the Indian frequency plan
const IN_865_867 = "IN_865_867"

//revive:enable:var-naming

func init() {
	defaultChannels := []Channel{
		{
			Frequency:   865062500,
			MaxDataRate: ttnpb.DATA_RATE_5,
		},
		{
			Frequency:   865402500,
			MaxDataRate: ttnpb.DATA_RATE_5,
		},
		{
			Frequency:   865985000,
			MaxDataRate: ttnpb.DATA_RATE_5,
		},
	}
	const beaconFrequency = 866500000

	in_865_867 = Band{
		ID: IN_865_867,

		EnableADR: true,

		MaxUplinkChannels: 16,
		UplinkChannels:    defaultChannels,

		MaxDownlinkChannels: 16,
		DownlinkChannels:    defaultChannels,

		SubBands: []SubBandParameters{
			{
				MinFrequency: 865000000,
				MaxFrequency: 867000000,
				DutyCycle:    1,
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

		Rx1Channel: channelIndexIdentity,
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset ttnpb.DataRateOffset, _ bool) (ttnpb.DataRateIndex, error) {
			so := int8(offset)
			if so > 5 {
				so = 5 - so
			}
			si := int8(idx) - so

			switch {
			case si <= 0:
				return ttnpb.DATA_RATE_0, nil
			case si >= 5:
				return ttnpb.DATA_RATE_5, nil
			}
			return ttnpb.DataRateIndex(si), nil
		},

		GenerateChMasks: generateChMask16,
		ParseChMask:     parseChMask16,

		LoRaCodingRate: "4/5",

		FreqMultiplier:   100,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		DefaultRx2Parameters: Rx2Parameters{ttnpb.DATA_RATE_2, 866550000},

		Beacon: Beacon{
			DataRateIndex:    ttnpb.DATA_RATE_4,
			CodingRate:       "4/5",
			ComputeFrequency: func(_ float64) uint64 { return beaconFrequency },
		},
		PingSlotFrequency: uint64Ptr(beaconFrequency),

		// No LoRaWAN 1.0
		// No LoRaWAN 1.0.1
		// No LoRaWAN 1.0.2rA
		regionalParameters1_v1_0_2RevB: bandIdentity,
		regionalParameters1_v1_0_3RevA: bandIdentity,
		regionalParameters1_v1_1RevA:   bandIdentity,
	}
	All[IN_865_867] = in_865_867
}
