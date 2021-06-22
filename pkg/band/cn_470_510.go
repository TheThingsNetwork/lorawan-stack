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

var cn_470_510 Band

// CN_470_510 is the ID of the Chinese 470-510Mhz band
const CN_470_510 = "CN_470_510"

//revive:enable:var-naming

func init() {
	uplinkChannels := make([]Channel, 0, 96)
	for i := 0; i < 96; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:   uint64(470300000 + 200000*i),
			MaxDataRate: ttnpb.DATA_RATE_5,
		})
	}

	downlinkChannels := make([]Channel, 0, 48)
	for i := 0; i < 48; i++ {
		downlinkChannels = append(downlinkChannels, Channel{
			Frequency:   uint64(500300000 + 200000*i),
			MaxDataRate: ttnpb.DATA_RATE_5,
		})
	}

	var beaconFrequencies [8]uint64
	for i := 0; i < 8; i++ {
		beaconFrequencies[i] = 508300000 + uint64(i*200000)
	}

	downlinkDRTable := [6][6]ttnpb.DataRateIndex{
		{0, 0, 0, 0, 0, 0},
		{1, 0, 0, 0, 0, 0},
		{2, 1, 0, 0, 0, 0},
		{3, 2, 1, 0, 0, 0},
		{4, 3, 2, 1, 0, 0},
		{5, 4, 3, 2, 1, 0},
	}

	cn_470_510 = Band{
		ID: CN_470_510,

		EnableADR: true,

		MaxUplinkChannels: 96,
		UplinkChannels:    uplinkChannels,

		MaxDownlinkChannels: 48,
		DownlinkChannels:    downlinkChannels,

		// See IEEE 11-11/0972r0
		SubBands: []SubBandParameters{
			{
				MinFrequency: 470000000,
				MaxFrequency: 510000000,
				DutyCycle:    1,
				MaxEIRP:      17.0 + eirpDelta,
			},
		},

		DataRates: map[ttnpb.DataRateIndex]DataRate{
			ttnpb.DATA_RATE_0: makeLoRaDataRate(12, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_1: makeLoRaDataRate(11, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_2: makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_3: makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(123)),
			ttnpb.DATA_RATE_4: makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_5: makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(230)),
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

		DefaultMaxEIRP: 19.15,
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

		Rx1Channel: channelIndexModulo(48),
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset ttnpb.DataRateOffset, _ bool) (ttnpb.DataRateIndex, error) {
			if idx > ttnpb.DATA_RATE_5 {
				return 0, errDataRateIndexTooHigh.WithAttributes("max", 5)
			}
			if offset > 5 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
			}
			return downlinkDRTable[idx][offset], nil
		},

		GenerateChMasks: generateChMask96,
		ParseChMask:     parseChMask96,

		DefaultRx2Parameters: Rx2Parameters{ttnpb.DATA_RATE_0, 505300000},

		Beacon: Beacon{
			DataRateIndex:    ttnpb.DATA_RATE_2,
			CodingRate:       "4/5",
			ComputeFrequency: makeBeaconFrequencyFunc(beaconFrequencies),
		},

		LoRaCodingRate: "4/5",

		FreqMultiplier:   100,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_CHANNEL_MASKS,

		// No LoRaWAN Regional Parameters 1.0
		regionalParameters1_v1_0_1: bandIdentity,
		regionalParameters1_v1_0_2: func(b Band) Band {
			b.DefaultMaxEIRP = 17
			b.TxOffset = []float32{
				0,
				-1,
				-3,
				-5,
				-7,
				-10,
				-12,
				-15,
			}
			return b
		},
		regionalParameters1_v1_0_2RevB: disableCFList,
		regionalParameters1_v1_0_3RevA: bandIdentity,
		regionalParameters1_v1_1RevA:   bandIdentity,
	}
	All[CN_470_510] = cn_470_510
}
