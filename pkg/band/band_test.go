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
	downlinkDatarateIndex, downlinkFrequency := euBand.Rx1Parameters(uint64(uplinkTxSettings.DataRateIndex), int(uplinkTxSettings.Frequency), 0, false)
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
