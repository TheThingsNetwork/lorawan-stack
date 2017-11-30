// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band_test

import (
	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

func GetUplink() ttnpb.UplinkMessage                   { return ttnpb.UplinkMessage{} }
func SendDownlink(ttnpb.DownlinkMessage)               {}
func ParseSpreadingFactor(types.DataRate) uint32       { return 7 }
func ParseBandwidth(types.DataRate) uint32             { return 125000 }
func ParseBitRate(types.DataRate) uint32               { return 0 }
func ParseModulation(types.DataRate) ttnpb.Modulation  { return ttnpb.Modulation_LORA }
func GetReceptionTimestamp(ttnpb.UplinkMessage) uint64 { return 0 }

func Example() {
	euBand, err := band.GetByID(band.EU_863_870)
	if err != nil {
		panic(err)
	}

	uplink := GetUplink()
	uplinkTxSettings := uplink.Settings
	downlinkDatarateIndex, downlinkFrequency := euBand.Rx1Parameters(int(uplinkTxSettings.DataRateIndex), int(uplinkTxSettings.Frequency), 0, false)
	downlinkDatarate := euBand.DataRates[downlinkDatarateIndex]

	downlink := ttnpb.DownlinkMessage{
		Settings: ttnpb.TxSettings{
			DataRateIndex:   uint32(downlinkDatarateIndex),
			Frequency:       uint64(downlinkFrequency),
			Modulation:      ParseModulation(downlinkDatarate.Rate),
			SpreadingFactor: ParseSpreadingFactor(downlinkDatarate.Rate),
			BitRate:         ParseBitRate(downlinkDatarate.Rate),
			Bandwidth:       ParseBandwidth(downlinkDatarate.Rate),
		},
		TxMetadata: ttnpb.TxMetadata{
			Timestamp: GetReceptionTimestamp(uplink) + 1000000000*uint64(euBand.ReceiveDelay1),
		},
	}
	SendDownlink(downlink)
}
