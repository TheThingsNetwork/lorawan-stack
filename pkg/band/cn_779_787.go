// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var cn_779_787 Band

const (
	// CN_779_787 is the ID of the Chinese 779-787Mhz band
	CN_779_787 ID = "CN_779_787"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 779500000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 779500000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 779900000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 780500000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 780700000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 780900000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	cn_779_787 = Band{
		ID: CN_779_787,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 779000000,
				MaxFrequency: 787000000,
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
			{Rate: types.DataRate{LoRa: "SF7BW250"}, DefaultMaxSize: maxPayloadSize{250, 242}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{FSK: 50000}, DefaultMaxSize: maxPayloadSize{250, 242}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
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
		TxOffset:       []float32{0, -2, -4, -6, -8, -10},

		Rx1Parameters: func(dataRateIndex, frequency, rx1DROffset int, _ bool) (int, int) {
			outDataRateIndex := dataRateIndex - rx1DROffset
			if outDataRateIndex < 0 {
				outDataRateIndex = 0
			}
			return outDataRateIndex, frequency
		},

		DefaultRx2Parameters: Rx2Parameters{0, 786000000},
	}
	All = append(All, cn_779_787)
}
