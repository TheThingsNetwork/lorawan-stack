// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var us Band

const (
	US BandID = "US"
)

func init() {
	uplinkChannels := make([]Channel, 0)
	for i := 0; i < 64; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:       902300000 + 200000*i,
			DataRateIndexes: []int{0, 1, 2, 3},
		})
	}
	for i := 0; i < 8; i++ {
		uplinkChannels = append(uplinkChannels, Channel{
			Frequency:       903000000 + 1600000*i,
			DataRateIndexes: []int{4},
		})
	}

	downlinkChannels := make([]Channel, 0)
	for i := 0; i < 8; i++ {
		downlinkChannels = append(downlinkChannels, Channel{
			Frequency:       923300000 + 600000*i,
			DataRateIndexes: []int{8, 9, 10, 11, 12, 13},
		})
	}

	us = Band{
		ID: US,

		UplinkChannels:   uplinkChannels,
		DownlinkChannels: downlinkChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 902000000,
				MaxFrequency: 928000000,
				DutyCycle:    1,
			},
		},

		DataRates: []DataRate{
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{19, 11}, NoRepeaterMaxSize: maxPayloadSize{19, 11}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{61, 53}, NoRepeaterMaxSize: maxPayloadSize{61, 53}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{133, 125}, NoRepeaterMaxSize: maxPayloadSize{133, 125}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{250, 242}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF8BW500"}, DefaultMaxSize: maxPayloadSize{250, 242}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{}, // Reserved for future use
			{}, // Reserved for future use
			{}, // Reserved for future use
			{Rate: types.DataRate{LoRa: "SF12BW500"}, DefaultMaxSize: maxPayloadSize{41, 33}, NoRepeaterMaxSize: maxPayloadSize{61, 53}},
			{Rate: types.DataRate{LoRa: "SF11BW500"}, DefaultMaxSize: maxPayloadSize{117, 109}, NoRepeaterMaxSize: maxPayloadSize{137, 129}},
			{Rate: types.DataRate{LoRa: "SF10BW500"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF9BW500"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF8BW500"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW500"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
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

		DefaultMaxEIRP: 30,
		TXOffset: func() []float32 {
			offset := []float32{}
			for i := 0; i < 11; i++ {
				offset = append(offset, float32(0-2*i))
			}
			return offset
		}(),

		RX1Parameters: func(dataRateIndex, frequency, RX1DROffset int, _ bool) (int, int) {
			outDataRateIndex := dataRateIndex + 10 - RX1DROffset
			if outDataRateIndex < 8 {
				outDataRateIndex = 8
			} else if outDataRateIndex > 13 {
				outDataRateIndex = 13
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

		DefaultRX2Parameters: Rx2Parameters{8, 923300000},
	}
	All = append(All, us)
}
