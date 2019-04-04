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

var au_915_928 Band

// AU_915_928 is the ID of the Australian band
const AU_915_928 = "AU_915_928"

//revive:enable:var-naming

func init() {
	uplinkChannels := make([]Channel, 0, 72)
	for i := 0; i < 64; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:   uint64(915200000 + 200000*i),
			MinDataRate: 0,
			MaxDataRate: 3,
		})
	}
	for i := 0; i < 8; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:   uint64(915900000 + 1600000*i),
			MinDataRate: 4,
			MaxDataRate: 4,
		})
	}

	downlinkChannels := make([]Channel, 0, 8)
	for i := 0; i < 8; i++ {
		downlinkChannels = append(downlinkChannels, Channel{
			Frequency:   uint64(923300000 + 600000*i),
			MinDataRate: 8,
			MaxDataRate: 13,
		})
	}

	au_915_928 = Band{
		ID: AU_915_928,

		MaxUplinkChannels: 72,
		UplinkChannels:    uplinkChannels,

		MaxDownlinkChannels: 8,
		DownlinkChannels:    downlinkChannels,

		// See Radiocommunications (Low Interference Potential Devices) Class Licence 2015
		SubBands: []SubBandParameters{
			{
				MinFrequency: 902000000,
				MaxFrequency: 928000000,
				DutyCycle:    1,
				MaxEIRP:      14.0 + eirpDelta,
			},
		},

		DataRates: [16]DataRate{
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 12,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: constPayloadSizer(59)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 11,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: constPayloadSizer(59)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 10,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: constPayloadSizer(59)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 9,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: constPayloadSizer(123)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 8,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 7,
				Bandwidth:       125000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 8,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{}, // RFU
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 12,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(41)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 11,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(117)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 10,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 9,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 8,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 7,
				Bandwidth:       500000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{}, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU for previous versions
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

		DefaultMaxEIRP: 30,
		TxOffset: func() [16]float32 {
			offset := [16]float32{}
			for i := 0; i < 15; i++ {
				offset[i] = float32(0 - 2*i)
			}
			return offset
		}(),
		MaxTxPowerIndex: 14,

		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_CHANNEL_MASKS,

		Rx1Channel: channelIndexModulo(8),
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset uint32, _ bool) (ttnpb.DataRateIndex, error) {
			if idx > 6 {
				return 0, errDataRateIndexTooHigh.WithAttributes("max", 6)
			}
			if offset > 5 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
			}

			si := int(uint32(idx) + 8 - offset)
			switch {
			case si <= 8:
				return 8, nil
			case si >= 13:
				return 13, nil
			}
			return ttnpb.DataRateIndex(si), nil
		},

		GenerateChMasks: makeGenerateChMask72(true),
		ParseChMask:     parseChMask72,

		DefaultRx2Parameters: Rx2Parameters{8, 923300000},

		Beacon: Beacon{
			DataRateIndex:    8,
			CodingRate:       "4/5",
			BroadcastChannel: beaconChannelFromFrequencies(usAuBeaconFrequencies),
			PingSlotChannels: usAuBeaconFrequencies[:],
		},

		// No LoRaWAN Regional Parameters 1.0
		regionalParameters1_0_1:     bandIdentity,
		regionalParameters1_0_2RevA: auDataRates1_0_2,
		regionalParameters1_0_2RevB: disableChMaskCntl51_0_2,
		regionalParameters1_0_3RevA: bandIdentity,
		regionalParameters1_1RevA:   bandIdentity,
	}
	All[AU_915_928] = au_915_928
}
