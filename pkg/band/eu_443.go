// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var eu_443 Band

const (
	EU_443 BandID = "EU_443"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 433175000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 433375000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 433575000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	eu_443 = Band{
		ID: EU_443,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 433175000,
				MaxFrequency: 434665000,
				DutyCycle:    0.01,
			},
		},

		DataRates: []DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{123, 115}, NoRepeaterMaxSize: maxPayloadSize{123, 115}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW250"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
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

		DefaultMaxEIRP: 12.15,
		TXOffset:       []float32{0, -2, -4, -6, -8, -10},

		RX1Parameters: func(dataRateIndex, frequency, RX1DROffset int, _ bool) (int, int) {
			outDataRateIndex := dataRateIndex - RX1DROffset
			if outDataRateIndex < 0 {
				outDataRateIndex = 0
			}
			return outDataRateIndex, frequency
		},

		DefaultRX2Parameters: func(_, _, _ int) (int, int) {
			return 0, 434665000
		},
	}
	All = append(All, eu_443)
}
