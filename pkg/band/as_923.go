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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

//revive:disable:var-naming

var as_923 Band

// AS_923 is the ID of the Asian 923Mhz band
const AS_923 = "AS_923"

//revive:enable:var-naming

func init() {
	defaultChannels := []Channel{
		{Frequency: 923200000, MinDataRate: 0, MaxDataRate: 5},
		{Frequency: 923400000, MinDataRate: 0, MaxDataRate: 5},
	}
	asBeaconChannel := uint32(923400000)
	as_923 = Band{
		ID: AS_923,

		MaxUplinkChannels: 16,
		UplinkChannels:    defaultChannels,

		MaxDownlinkChannels: 16,
		DownlinkChannels:    defaultChannels,

		SubBands: []SubBandParameters{
			{
				MinFrequency: 923000000,
				MaxFrequency: 923500000,
				DutyCycle:    0.01,
				MaxEIRP:      14.0 + eirpDelta,
			},
		},

		DataRates: [16]DataRate{
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 12,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{59, 0}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 11,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{59, 0}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 10,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{59, 19}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 9,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{123, 61}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 8,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{230, 133}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 7,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{230, 250}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 7,
				Bandwidth:       250000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{230, 250}},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_FSK{FSK: &ttnpb.FSKDataRate{
				BitRate: 50000,
			}}}, DefaultMaxSize: dwellTimePayloadSizer{230, 250}},
			{}, {}, {}, {}, {}, {}, {}, // RFU
			{}, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},
		MaxADRDataRateIndex: 5,

		ReceiveDelay1:    defaultReceiveDelay1,
		ReceiveDelay2:    defaultReceiveDelay2,
		JoinAcceptDelay1: defaultJoinAcceptDelay1,
		JoinAcceptDelay2: defaultJoinAcceptDelay2,
		MaxFCntGap:       defaultMaxFCntGap,
		ADRAckLimit:      defaultADRAckLimit,
		ADRAckDelay:      defaultADRAckDelay,
		MinAckTimeout:    defaultAckTimeout - defaultAckTimeoutMargin,
		MaxAckTimeout:    defaultAckTimeout + defaultAckTimeoutMargin,

		DefaultMaxEIRP: 16,
		TxOffset: [16]float32{0, -2, -4, -6, -8, -10, -12, -14,
			0, 0, 0, 0, 0, 0, 0, // RFU
			0, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},
		MaxTxPowerIndex: 7,

		LoRaCodingRate: "4/5",

		FreqMultiplier:   100,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		Rx1Channel: channelIndexIdentity,
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset uint32, dwellTime bool) (ttnpb.DataRateIndex, error) {
			if offset > 7 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 7)
			}

			so := int(offset)
			if so > 5 {
				so = 5 - so
			}
			si := int(idx) - so

			minDR := uint32(0)
			if dwellTime {
				minDR = 2
			}

			switch {
			case si <= int(minDR):
				return ttnpb.DataRateIndex(minDR), nil
			case si >= 5:
				return 5, nil
			}
			return ttnpb.DataRateIndex(si), nil
		},

		GenerateChMasks: generateChMask16,
		ParseChMask:     parseChMask16,

		DefaultRx2Parameters: Rx2Parameters{2, 923200000},

		Beacon: Beacon{
			DataRateIndex:    3,
			CodingRate:       "4/5",
			PingSlotChannels: []uint32{asBeaconChannel},
			BroadcastChannel: func(_ float64) uint32 { return asBeaconChannel },
		},

		TxParamSetupReqSupport: true,

		// No LoRaWAN Regional Parameters 1.0
		// No LoRaWAN Regional Parameters 1.0.1
		regionalParameters1_0_2RevA: bandIdentity,
		regionalParameters1_0_2RevB: bandIdentity,
		regionalParameters1_0_3RevA: bandIdentity,
		regionalParameters1_1RevA:   bandIdentity,
	}
	All[AS_923] = as_923
}
