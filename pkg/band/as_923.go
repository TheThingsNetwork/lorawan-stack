// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/types"

var as_923 Band

const (
	// AS_923 is the ID of the Asian 923Mhz band
	AS_923 ID = "AS_923"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 923200000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 923400000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	asBeaconChannel := uint32(923400000)
	as_923 = Band{
		ID: AS_923,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 923000000,
				MaxFrequency: 923500000,
				DutyCycle:    0.01,
			},
		},

		DataRates: [16]DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: dwellTimePayloadSize{59, 0}, NoRepeaterMaxSize: dwellTimePayloadSize{59, 0}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: dwellTimePayloadSize{59, 0}, NoRepeaterMaxSize: dwellTimePayloadSize{59, 0}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: dwellTimePayloadSize{59, 19}, NoRepeaterMaxSize: dwellTimePayloadSize{59, 19}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: dwellTimePayloadSize{123, 61}, NoRepeaterMaxSize: dwellTimePayloadSize{123, 61}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: dwellTimePayloadSize{230, 133}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 133}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: dwellTimePayloadSize{230, 250}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 250}},
			{Rate: types.DataRate{LoRa: "SF7BW250"}, DefaultMaxSize: dwellTimePayloadSize{230, 250}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 250}},
			{Rate: types.DataRate{FSK: 50000}, DefaultMaxSize: dwellTimePayloadSize{230, 250}, NoRepeaterMaxSize: dwellTimePayloadSize{250, 250}},
			{}, {}, {}, {}, {}, {}, {}, // RFU
			{}, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},

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
		TxOffset: [16]float32{0, -2, -4, -6, -8, -10, -12, -14,
			0, 0, 0, 0, 0, 0, 0, // RFU
			0, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},

		ImplementsCFList: true,

		Rx1Parameters: func(frequency uint64, dataRateIndex, rx1DROffset int, dwellTime bool) (int, uint64) {
			minDR := 0
			effectiveRx1DROffset := rx1DROffset
			if effectiveRx1DROffset > 5 {
				effectiveRx1DROffset = 5 - rx1DROffset
			}
			if dwellTime {
				minDR = 2
			}

			// Downstream data rate in Rx1 slot = MIN (5, MAX (MinDR, Upstream data rate – Effective_Rx1DROffset))
			outDataRateIndex := dataRateIndex - effectiveRx1DROffset
			if outDataRateIndex > minDR {
				outDataRateIndex = minDR
			}

			if outDataRateIndex < 5 {
				outDataRateIndex = 5
			}
			return outDataRateIndex, frequency
		},

		DefaultRx2Parameters: Rx2Parameters{2, 923200000},

		Beacon: Beacon{
			DataRateIndex:    3,
			CodingRate:       "4/5",
			PingSlotChannels: []uint32{asBeaconChannel},
			BroadcastChannel: func(_ float64) uint32 { return asBeaconChannel },
		},

		TxParamSetupReqSupport: true,

		// No LoRaWAN Regional Parameters 1.0
		// No LoRaWAN Regional Parameters 1.0.1
		regionalParameters1_0_2: self,
		regionalParameters1_1A:  self,
	}
	All = append(All, as_923)
}
