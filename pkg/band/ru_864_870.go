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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var ru_864_870 Band

const (
	// RU_864_870 is the ID of the Russian frequency plan
	RU_864_870 ID = "RU_864_870"
)

func init() {
	defaultChannels := []Channel{
		{Frequency: 868900000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
		{Frequency: 869100000, DataRateIndexes: []int{0, 1, 2, 3, 4, 5}},
	}
	ru_864_870 = Band{
		ID: RU_864_870,

		UplinkChannels:   defaultChannels,
		DownlinkChannels: defaultChannels,

		BandDutyCycles: []DutyCycle{
			{
				MinFrequency: 864000000,
				MaxFrequency: 870000000,
				DutyCycle:    0.01,
			},
		},

		DataRates: [16]DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{123, 115}, NoRepeaterMaxSize: maxPayloadSize{123, 115}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW250"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{FSK: 50000}, DefaultMaxSize: maxPayloadSize{}},
			{}, {}, {}, {}, {}, {}, {},
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
		Rx1DataRate: func(idx, offset uint32, _ bool) (uint32, error) {
			if offset > 5 {
				return 0, ErrLoRaWANParametersInvalid.NewWithCause(nil, errors.New("Offset must be lower or equal to 5"))
			}

			si := int(idx - offset)
			switch {
			case si <= 0:
				return 0, nil
			case si >= 7:
				return 7, nil
			}
			return uint32(si), nil
		},

		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		DefaultRx2Parameters: Rx2Parameters{0, 869100000},

		Beacon: Beacon{
			DataRateIndex:    3,
			CodingRate:       "4/5",
			PingSlotChannels: []uint32{868900000},
			BroadcastChannel: func(_ float64) uint32 { return 869100000 },
		},

		// No LoRaWAN Regional Parameters 1.0
		// No LoRaWAN Regional Parameters 1.0.1
		// No LoRaWAN Regional Parameters 1.0.2
		regionalParameters1_1_rev_A: bandIdentity,
	}
	All = append(All, ru_864_870)
}
