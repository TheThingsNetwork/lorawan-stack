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

var au_915_928 Band

// AU_915_928 is the ID of the Australian band
const AU_915_928 = "AU_915_928"

//revive:enable:var-naming

func init() {
	uplinkChannels := make([]Channel, 0, 72)
	for i := 0; i < 64; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:   uint64(915200000 + 200000*i),
			MaxDataRate: ttnpb.DATA_RATE_5,
		})
	}
	for i := 0; i < 8; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:   uint64(915900000 + 1600000*i),
			MinDataRate: ttnpb.DATA_RATE_6,
			MaxDataRate: ttnpb.DATA_RATE_6,
		})
	}

	downlinkChannels := make([]Channel, 0, 8)
	for i := 0; i < 8; i++ {
		downlinkChannels = append(downlinkChannels, Channel{
			Frequency:   uint64(923300000 + 600000*i),
			MinDataRate: ttnpb.DATA_RATE_8,
			MaxDataRate: ttnpb.DATA_RATE_13,
		})
	}

	downlinkDRTable := [7][6]ttnpb.DataRateIndex{
		{8, 8, 8, 8, 8, 8},
		{9, 8, 8, 8, 8, 8},
		{10, 9, 8, 8, 8, 8},
		{11, 10, 9, 8, 8, 8},
		{12, 11, 10, 9, 8, 8},
		{13, 12, 11, 10, 9, 8},
		{13, 13, 12, 11, 10, 9},
	}

	au_915_928 = Band{
		ID: AU_915_928,

		EnableADR: true,

		MaxUplinkChannels: 72,
		UplinkChannels:    uplinkChannels,

		MaxDownlinkChannels: 8,
		DownlinkChannels:    downlinkChannels,

		// See Radiocommunications (Low Interference Potential Devices) Class Licence 2015
		SubBands: []SubBandParameters{
			{
				MinFrequency: 915000000,
				MaxFrequency: 928000000,
				DutyCycle:    1,
				MaxEIRP:      30,
			},
		},

		DataRates: map[ttnpb.DataRateIndex]DataRate{
			ttnpb.DATA_RATE_0: makeLoRaDataRate(12, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_1: makeLoRaDataRate(11, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_2: makeLoRaDataRate(10, 125000, makeConstMaxMACPayloadSizeFunc(59)),
			ttnpb.DATA_RATE_3: makeLoRaDataRate(9, 125000, makeConstMaxMACPayloadSizeFunc(123)),
			ttnpb.DATA_RATE_4: makeLoRaDataRate(8, 125000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_5: makeLoRaDataRate(7, 125000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_6: makeLoRaDataRate(8, 500000, makeConstMaxMACPayloadSizeFunc(230)),

			ttnpb.DATA_RATE_8:  makeLoRaDataRate(12, 500000, makeConstMaxMACPayloadSizeFunc(61)),
			ttnpb.DATA_RATE_9:  makeLoRaDataRate(11, 500000, makeConstMaxMACPayloadSizeFunc(137)),
			ttnpb.DATA_RATE_10: makeLoRaDataRate(10, 500000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_11: makeLoRaDataRate(9, 500000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_12: makeLoRaDataRate(8, 500000, makeConstMaxMACPayloadSizeFunc(230)),
			ttnpb.DATA_RATE_13: makeLoRaDataRate(7, 500000, makeConstMaxMACPayloadSizeFunc(230)),
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
			-22,
			-24,
			-26,
			-28,
		},

		LoRaCodingRate: "4/5",

		FreqMultiplier:   100,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_CHANNEL_MASKS,

		Rx1Channel: channelIndexModulo(8),
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset ttnpb.DataRateOffset, _ bool) (ttnpb.DataRateIndex, error) {
			if idx > ttnpb.DATA_RATE_6 {
				return 0, errDataRateIndexTooHigh.WithAttributes("max", 6)
			}
			if offset > 5 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
			}
			return downlinkDRTable[idx][offset], nil
		},

		GenerateChMasks: makeGenerateChMask72(true),
		ParseChMask:     parseChMask72,

		DefaultRx2Parameters: Rx2Parameters{ttnpb.DATA_RATE_8, 923300000},

		Beacon: Beacon{
			DataRateIndex:    ttnpb.DATA_RATE_8,
			CodingRate:       "4/5",
			ComputeFrequency: makeBeaconFrequencyFunc(usAuBeaconFrequencies),
		},

		TxParamSetupReqSupport: true,

		// No LoRaWAN Regional Parameters 1.0
		regionalParameters1_v1_0_1: bandIdentity,
		regionalParameters1_v1_0_2: func(b Band) Band {
			dataRates := make(map[ttnpb.DataRateIndex]DataRate, len(b.DataRates)-2)
			for drIdx := ttnpb.DATA_RATE_0; drIdx <= ttnpb.DATA_RATE_4; drIdx++ {
				dataRates[drIdx] = b.DataRates[drIdx+2]
			}
			for drIdx := ttnpb.DATA_RATE_8; drIdx <= ttnpb.DATA_RATE_13; drIdx++ {
				dataRates[drIdx] = b.DataRates[drIdx]
			}
			b.DataRates = dataRates

			b.UplinkChannels = append(b.UplinkChannels[:0:0], b.UplinkChannels...)
			for i := 0; i < 64; i++ {
				b.UplinkChannels[i].MaxDataRate = ttnpb.DATA_RATE_3
			}
			for i := 0; i < 8; i++ {
				b.UplinkChannels[64+i].MinDataRate = ttnpb.DATA_RATE_4
				b.UplinkChannels[64+i].MaxDataRate = ttnpb.DATA_RATE_4
			}

			b.MaxADRDataRateIndex = ttnpb.DATA_RATE_3
			return b
		},
		regionalParameters1_v1_0_2RevB: composeSwaps(
			disableCFList,
			disableChMaskCntl5,
			disableTxParamSetupReq,
			makeSetMaxTxPowerIndexFunc(10),
		),
		regionalParameters1_v1_0_3RevA: composeSwaps(
			enableTxParamSetupReq,
			makeAddTxPowerFunc(-22),
			makeAddTxPowerFunc(-24),
			makeAddTxPowerFunc(-26),
			makeAddTxPowerFunc(-28),
			makeAddTxPowerFunc(-30),
		),
		regionalParameters1_v1_1RevA: composeSwaps(
			disableTxParamSetupReq,
			makeSetBeaconDataRateIndex(ttnpb.DATA_RATE_10),
			makeSetMaxTxPowerIndexFunc(10),
		),
	}
	All[AU_915_928] = au_915_928
}
