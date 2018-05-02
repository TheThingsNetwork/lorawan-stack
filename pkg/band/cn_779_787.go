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
	cnBeaconChannel := uint32(785000000)
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

		DataRates: [16]DataRate{
			{Rate: types.DataRate{LoRa: "SF12BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF11BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF10BW125"}, DefaultMaxSize: maxPayloadSize{59, 51}, NoRepeaterMaxSize: maxPayloadSize{59, 51}},
			{Rate: types.DataRate{LoRa: "SF9BW125"}, DefaultMaxSize: maxPayloadSize{123, 115}, NoRepeaterMaxSize: maxPayloadSize{123, 115}},
			{Rate: types.DataRate{LoRa: "SF8BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW125"}, DefaultMaxSize: maxPayloadSize{230, 222}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{LoRa: "SF7BW250"}, DefaultMaxSize: maxPayloadSize{250, 242}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{Rate: types.DataRate{FSK: 50000}, DefaultMaxSize: maxPayloadSize{250, 242}, NoRepeaterMaxSize: maxPayloadSize{250, 242}},
			{}, {}, {}, {}, {}, {}, {}, // RFU
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

		DefaultMaxEIRP: 12.15,
		TxOffset: [16]float32{0, -2, -4, -6, -8, -10,
			0, 0, 0, 0, 0, 0, 0, 0, 0, // RFU
			0, // Used by LinkADRReq starting from LoRaWAN Regional Parameters 1.1, RFU before
		},

		Rx1Channel: channelIndexIdentity,
		Rx1DataRate: func(idx, offset uint32, _ bool) (uint32, error) {
			if idx > 7 {
				return 0, ErrLoRaWANParametersInvalid.NewWithCause(nil, errors.New("Data rate index must be lower or equal to 7"))
			}
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

		DefaultRx2Parameters: Rx2Parameters{0, 786000000},

		Beacon: Beacon{
			DataRateIndex:    3,
			CodingRate:       "4/5",
			PingSlotChannels: []uint32{cnBeaconChannel},
			BroadcastChannel: func(_ float64) uint32 { return cnBeaconChannel },
		},

		regionalParameters1_0:   bandIdentity,
		regionalParameters1_0_1: bandIdentity,
		regionalParameters1_0_2: bandIdentity,
	}
	All = append(All, cn_779_787)
}
