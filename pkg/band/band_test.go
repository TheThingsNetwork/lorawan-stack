// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band_test

import (
	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

func GetUplink() ttnpb.UplinkMessage                  { return ttnpb.UplinkMessage{} }
func SendDownlink(ttnpb.DownlinkMessage)              { return }
func ParseSpreadingFactor(types.DataRate) uint32      { return 7 }
func ParseBandwidth(types.DataRate) uint32            { return 125000 }
func ParseBitRate(types.DataRate) uint32              { return 0 }
func ParseModulation(types.DataRate) ttnpb.Modulation { return ttnpb.Modulation_LORA }

func Example() {
	euBand, err := band.GetByID(band.EU_863_870)
	if err != nil {
		panic(err)
	}

	uplink := GetUplink()
	uplinkTxSettings := uplink.TxSettings
	downlinkDatarateIndex, downlinkFrequency := euBand.RX1Parameters(int(uplinkTxSettings.DataRateIndex), int(uplinkTxSettings.Frequency), 0, false)
	downlinkDatarate := euBand.DataRates[downlinkDatarateIndex]

	minimumTimestamp := uint64(0xFFFFFFFFFFFFFFFF)
	for _, rxMetadata := range uplink.RxMetadata {
		if rxMetadata.Timestamp < minimumTimestamp {
			minimumTimestamp = rxMetadata.Timestamp
		}
	}

	downlink := ttnpb.DownlinkMessage{
		TxSettings: ttnpb.TxSettings{
			DataRateIndex:   int32(downlinkDatarateIndex),
			Frequency:       uint64(downlinkFrequency),
			Modulation:      ParseModulation(downlinkDatarate.Rate),
			SpreadingFactor: ParseSpreadingFactor(downlinkDatarate.Rate),
			BitRate:         ParseBitRate(downlinkDatarate.Rate),
			Bandwidth:       ParseBandwidth(downlinkDatarate.Rate),
		},
		TxMetadata: ttnpb.TxMetadata{
			Timestamp: minimumTimestamp + 1000000000*uint64(euBand.ReceiveDelay1),
		},
	}
	SendDownlink(downlink)
}
