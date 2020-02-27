// Copyright © 2019 The Things Industries B.V.

package band

import (
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

//revive:disable:var-naming

var ism_2400 Band

// ISM_2400 is the ID of the LoRa 2.4 GHz band.
const ISM_2400 = "ISM_2400"

//revive:enable:var-naming

func init() {
	defaultChannels := []Channel{
		{Frequency: 2403000000, MinDataRate: 0, MaxDataRate: 7},
		{Frequency: 2425000000, MinDataRate: 0, MaxDataRate: 7},
		{Frequency: 2479000000, MinDataRate: 0, MaxDataRate: 7},
	}
	ism2400BeaconChannel := uint32(2422000000)
	ism_2400 = Band{
		ID: ISM_2400,

		MaxUplinkChannels: 16,
		UplinkChannels:    defaultChannels,

		MaxDownlinkChannels: 16,
		DownlinkChannels:    defaultChannels,

		SubBands: []SubBandParameters{
			{
				MinFrequency: 2400000000,
				MaxFrequency: 2500000000,
				DutyCycle:    1,
				MaxEIRP:      8.0 + eirpDelta,
			},
		},

		DataRates: [16]DataRate{
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 12,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(59)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 11,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(123)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 10,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 9,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 8,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 7,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 6,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{Rate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 5,
				Bandwidth:       812000,
			}}}, DefaultMaxSize: constPayloadSizer(230)},
			{}, {}, {}, {}, {}, {}, {}, {}, // RFU
		},
		MaxADRDataRateIndex: 5,

		ReceiveDelay1:    defaultReceiveDelay1,
		ReceiveDelay2:    defaultReceiveDelay2,
		JoinAcceptDelay1: defaultJoinAcceptDelay1,
		JoinAcceptDelay2: defaultJoinAcceptDelay2,
		MaxFCntGap:       defaultMaxFCntGap,
		ADRAckLimit:      defaultADRAckLimit,
		ADRAckDelay:      defaultADRAckDelay,
		MinAckTimeout:    defaultAckTimeout - defaultAckTimeoutMargin,
		MaxAckTimeout:    defaultAckTimeout + defaultAckTimeoutMargin,

		DefaultMaxEIRP: 16,
		TxOffset: []float32{
			0,
			-2,
			-4,
			-6,
			-8,
			-10,
			-12,
			-14,
		},

		Rx1Channel: channelIndexIdentity,
		Rx1DataRate: func(idx ttnpb.DataRateIndex, offset uint32, _ bool) (ttnpb.DataRateIndex, error) {
			if idx > 7 {
				return 0, errDataRateIndexTooHigh.WithAttributes("max", 7)
			}
			if offset > 5 {
				return 0, errDataRateOffsetTooHigh.WithAttributes("max", 5)
			}

			si := int(uint32(idx) - offset)
			switch {
			case si <= 0:
				return 0, nil
			case si >= 7:
				return 7, nil
			}
			return ttnpb.DataRateIndex(si), nil
		},

		GenerateChMasks: generateChMask16,
		ParseChMask:     parseChMask16,

		LoRaCodingRate: "4/8",

		FreqMultiplier:   200,
		ImplementsCFList: true,
		CFListType:       ttnpb.CFListType_FREQUENCIES,

		DefaultRx2Parameters: Rx2Parameters{0, 2422000000},

		Beacon: Beacon{
			DataRateIndex:    3,
			CodingRate:       "4/8",
			BroadcastChannel: func(_ float64) uint32 { return ism2400BeaconChannel },
			PingSlotChannels: []uint32{ism2400BeaconChannel},
		},

		regionalParameters1_0:       bandIdentity,
		regionalParameters1_0_1:     bandIdentity,
		regionalParameters1_0_2RevA: bandIdentity,
		regionalParameters1_0_2RevB: bandIdentity,
		regionalParameters1_0_3RevA: bandIdentity,
		regionalParameters1_1RevA:   bandIdentity,
	}
	All[ISM_2400] = ism_2400
}
