// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var as923 Band

const (
	AS923 BandID = "AS923"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 923200000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 923400000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	as923 = Band{
		ID: AS923,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 923000000,
				MaxFrequency: 923500000,
				DutyCycle:    0.01,
			},
		},

		DataRates: []DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: dwellTimePayloadSize{59, 0}, NoRepeaterMaxSize: dwellTimePayloadSize{59, 0}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: dwellTimePayloadSize{59, 0}, NoRepeaterMaxSize: dwellTimePayloadSize{59, 0}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: dwellTimePayloadSize{59, 19}, NoRepeaterMaxSize: dwellTimePayloadSize{59, 19}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: dwellTimePayloadSize{123, 61}, NoRepeaterMaxSize: dwellTimePayloadSize{123, 61}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: dwellTimePayloadSize{230, 133}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 133}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: dwellTimePayloadSize{230, 250}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 250}},
			{Rate: types.DataRate{LoRa: "SF7BW250"}, DefaultMaxSize: dwellTimePayloadSize{230, 250}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 250}},
			{Rate: types.DataRate{FSK: 50000}, DefaultMaxSize: dwellTimePayloadSize{230, 250}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 250}},
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

		DefaultMaxEIRP: 16,
		TXOffset:       []float32{0, -2, -4, -6, -8, -10, -12, -14},

		RX1Parameters: func(dataRateIndex, frequency, RX1DROffset int, dwellTime bool) (int, int) {
			minDR := 0
			effectiveRX1DROffset := RX1DROffset
			if effectiveRX1DROffset > 5 {
				effectiveRX1DROffset = 5 - RX1DROffset
			}
			if dwellTime {
				minDR = 2
			}

			// Downstream data rate in RX1 slot = MIN (5, MAX (MinDR, Upstream data rate – Effective_RX1DROffset))
			outDataRateIndex := dataRateIndex - effectiveRX1DROffset
			if outDataRateIndex > minDR {
				outDataRateIndex = minDR
			}

			if outDataRateIndex < 5 {
				outDataRateIndex = 5
			}
			return outDataRateIndex, frequency
		},

		DefaultRX2Parameters: Rx2Parameters{2, 923200000},
	}
	All = append(All, as923)
}
