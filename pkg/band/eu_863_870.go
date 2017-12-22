// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var eu_863_870 Band

const (
	// EU_863_870 is the ID of the European 863-870Mhz band
	EU_863_870 ID = "EU_863_870"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 868100000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 868300000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 868500000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	eu_863_870 = Band{
		ID: EU_863_870,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 863000000,
				MaxFrequency: 865000000,
				DutyCycle:    0.001,
			},
			{
				MinFrequency: 865000000,
				MaxFrequency: 868000000,
				DutyCycle:    0.01,
			},
			{
				MinFrequency: 868000000,
				MaxFrequency: 868600000,
				DutyCycle:    0.01,
			},
			{
				MinFrequency: 868700000,
				MaxFrequency: 869200000,
				DutyCycle:    0.001,
			},
			{
				MinFrequency: 869400000,
				MaxFrequency: 869650000,
				DutyCycle:    0.1,
			},
			{
				MinFrequency: 869700000,
				MaxFrequency: 870000000,
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

		DefaultMaxEIRP: 16,
		TxOffset:       []float32{0, -2, -4, -6, -8, -10, -12, -14},

		Rx1Parameters: func(frequency uint64, dataRateIndex, Rx1DROffset int, _ bool) (int, uint64) {
			outDataRateIndex := dataRateIndex - Rx1DROffset
			if outDataRateIndex < 0 {
				outDataRateIndex = 0
			}
			return outDataRateIndex, frequency
		},

		DefaultRx2Parameters: Rx2Parameters{0, 869525000},
	}
	All = append(All, eu_863_870)
}
