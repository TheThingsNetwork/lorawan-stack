// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var kr_920_923 Band

const (
	// KR_920_923 is the ID of the Korean frequency plan
	KR_920_923 ID = "KR_920_923"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 922100000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 922300000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 922500000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	kr_920_923 = Band{
		ID: KR_920_923,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 920000000,
				MaxFrequency: 923000000,
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

		DefaultMaxEIRP: 14,
		TxOffset:       []float32{0, -2, -4, -6, -8, -10, -12, -14},

		Rx1Parameters: func(frequency uint64, dataRateIndex, Rx1DROffset int, _ bool) (int, uint64) {
			outDataRateIndex := dataRateIndex - Rx1DROffset
			if outDataRateIndex < 0 {
				outDataRateIndex = 0
			}
			return outDataRateIndex, frequency
		},

		DefaultRx2Parameters: Rx2Parameters{0, 921900000},
	}
	All = append(All, kr_920_923)
}
