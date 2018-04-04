// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import "github.com/TheThingsNetwork/ttn/pkg/types"

var cn_470_510 Band

const (
	// CN_470_510 is the ID of the Chinese 470-510Mhz band
	CN_470_510 ID = "CN_470_510"
)

func init() {
	uplinkChannels := make([]Channel, 0)
	for i := 0; i < 96; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:       uint64(470300000 + 200000*i),
			DataRateIndexes: []int{0, 1, 2, 3, 4, 5},
		})
	}

	downlinkChannels := make([]Channel, 0)
	for i := 0; i < 48; i++ {
		downlinkChannels = append(downlinkChannels, Channel{
			Frequency:       uint64(500300000 + 200000*i),
			DataRateIndexes: []int{0, 1, 2, 3, 4, 5},
		})
	}

	cn_470_510 = Band{
		ID: CN_470_510,

		UplinkChannels:   uplinkChannels,
		DownlinkChannels: downlinkChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 470000000,
				MaxFrequency: 510000000,
				DutyCycle:    1,
			},
		},

		DataRates: [16]DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{123, 115}, NoRepeaterMaxSize: maxPayloadSize{123, 115}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{}, {}, {}, {}, {}, {}, {}, {}, {}, // RFU
			{}, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},

		ReceiveDelay1:    defaultReceiveDelay1,
		ReceiveDelay2:    defaultReceiveDelay2,
		JoinAcceptDelay1: defaultJoinAcceptDelay2,
		JoinAcceptDelay2: defaultJoinAcceptDelay2,
		MaxFCntGap:       defaultMaxFCntGap,
		ADRAckLimit:      defaultADRAckLimit,
		ADRAckDelay:      defaultADRAckDelay,
		MinAckTimeout:    defaultAckTimeout - defaultAckTimeoutMargin,
		MaxAckTimeout:    defaultAckTimeout + defaultAckTimeoutMargin,

		DefaultMaxEIRP: 19.15,
		TxOffset: [16]float32{0, -2, -4, -6, -8, -10, -12, -14,
			0, 0, 0, 0, 0, 0, 0, // RFU
			0, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},

		Rx1Parameters: func(frequency uint64, dataRateIndex, rx1DROffset int, _ bool) (int, uint64) {
			outDataRateIndex := dataRateIndex - rx1DROffset
			if outDataRateIndex < 0 {
				outDataRateIndex = 0
			}

			frequencyIndex := 0
			for channelIndex, uplinkChannel := range uplinkChannels {
				if frequency == uplinkChannel.Frequency {
					frequencyIndex = channelIndex
				}
			}
			frequencyIndex = frequencyIndex % 8

			return outDataRateIndex, downlinkChannels[frequencyIndex].Frequency
		},

		DefaultRx2Parameters: Rx2Parameters{0, 505300000},

		Beacon: Beacon{
			DataRateIndex:    2,
			CodingRate:       "4/5",
			BroadcastChannel: beaconChannelFromFrequencies(cn470BeaconFrequencies),
			PingSlotChannels: cn470BeaconFrequencies[:],
		},

		ImplementsCFList: true,

		// No LoRaWAN Regional Parameters 1.0
		regionalParameters1_0_1: self,
		regionalParameters1_0_2: disableCFList1_0_2,
		regionalParameters1_1A:  self,
	}
	All = append(All, cn_470_510)
}

var cn470BeaconFrequencies = func() [8]uint32 {
	freqs := [8]uint32{}
	for i := 0; i < 8; i++ {
		freqs[i] = 508300000 + uint32(i*200000)
	}
	return freqs
}()
