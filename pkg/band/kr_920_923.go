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

import (
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var kr_920_923 Band

// KR_920_923 is the ID of the Korean frequency plan
const KR_920_923 = "KR_920_923"

func init() {
	defaultChannels := []Channel{
		{Frequency: 922100000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 922300000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 922500000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	krBeaconChannel := uint32(923100000)
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

		DataRates: [16]DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{123, 115}, NoRepeaterMaxSize: maxPayloadSize{123, 115}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{}, {}, {}, {}, {}, {}, {}, {}, {},
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

		DefaultMaxEIRP: 14,
		TxOffset: [16]float32{0, -2, -4, -6, -8, -10, -12, -14,
			0, 0, 0, 0, 0, 0, 0, // RFU
			0, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},

		Rx1Channel: channelIndexIdentity,
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset uint32, _ bool) (ttnpb.DataRateIndex, error) {
			if offset > 5 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
			}

			si := int(uint32(idx) - offset)
			switch {
			case si <= 0:
				return 0, nil
			case si >= 5:
				return 5, nil
			}
			return ttnpb.DataRateIndex(si), nil
		},
		ParseChMask: parseChMask16,

		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		DefaultRx2Parameters: Rx2Parameters{0, 921900000},

		Beacon: Beacon{
			DataRateIndex:    3,
			CodingRate:       "4/5",
			PingSlotChannels: []uint32{krBeaconChannel},
			BroadcastChannel: func(_ float64) uint32 { return krBeaconChannel },
		},

		// No LoRaWAN 1.0
		// No LoRaWAN 1.0.1
		regionalParameters1_0_2RevA: bandIdentity,
		regionalParameters1_0_2RevB: bandIdentity,
		regionalParameters1_1RevA:   bandIdentity,
	}
	All[KR_920_923] = kr_920_923
}
