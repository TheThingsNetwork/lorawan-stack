// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var in_865_867 Band

const (
	IN_865_867 BandID = "IN_865_867"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 865062500, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 865402500, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 865985000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	in_865_867 = Band{
		ID: IN_865_867,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 865000000,
				MaxFrequency: 867000000,
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
			{}, // RFU
			{Rate: types.DataRate{FSK: 50000}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
		},

		ImplementsCFList: true,

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
			effectiveRX1DROffset := RX1DROffset
			if effectiveRX1DROffset > 5 {
				effectiveRX1DROffset = 5 - RX1DROffset
			}

			outDataRateIndex := dataRateIndex - effectiveRX1DROffset
			if outDataRateIndex < 5 {
				outDataRateIndex = 5
			}
			return outDataRateIndex, frequency
		},

		DefaultRX2Parameters: Rx2Parameters{2, 866550000},
	}
	All = append(All, in_865_867)
}
