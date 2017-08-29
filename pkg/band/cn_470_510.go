// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var cn_470_510 Band

const (
	CN_470_510 BandID = "CN_470_510"
)

func init() {
	uplinkChannels := make([]Channel, 0)
	for i := 0; i < 96; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency: 470300000 + 200000*i,
			DataRates: []int{0, 1, 2, 3, 4, 5},
		})
	}

	downlinkChannels := make([]Channel, 0)
	for i := 0; i < 48; i++ {
		downlinkChannels = append(downlinkChannels, Channel{
			Frequency: 500300000 + 200000*i,
			DataRates: []int{0, 1, 2, 3, 4, 5},
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

		DataRates: []DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{123, 115}, NoRepeaterMaxSize: maxPayloadSize{123, 115}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
		},

		ImplementsCFList: false,

		ReceiveDelay1:    defaultReceiveDelay1,
		ReceiveDelay2:    defaultReceiveDelay2,
		JoinAcceptDelay1: defaultJoinAcceptDelay2,
		JoinAcceptDelay2: defaultJoinAcceptDelay2,
		MaxFCntGap:       defaultMaxFCntGap,
		AdrAckLimit:      defaultAdrAckLimit,
		AdrAckDelay:      defaultAdrAckDelay,
		MinAckTimeout:    defaultAckTimeout - defaultAckTimeoutMargin,
		MaxAckTimeout:    defaultAckTimeout + defaultAckTimeoutMargin,

		DefaultMaxEIRP: 19.15,
		TXOffset:       []float32{0, -2, -4, -6, -8, -10, -12, -14},

		RX1Parameters: func(dataRateIndex, frequency, RX1DROffset int, _ bool) (int, int) {
			outDataRateIndex := dataRateIndex - RX1DROffset
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

		DefaultRX2Parameters: func(_, _, _ int) (int, int) {
			return 0, 505300000
		},
	}
	All = append(All, cn_470_510)
}
