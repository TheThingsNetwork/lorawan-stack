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
	fp, err := band.GetByID(band.EU_863_870)
	if err != nil {
		panic(err)
	}

	up := GetUplink()
	sets := up.GetSettings()
	drIdx, err := fp.Rx1DataRate(sets.GetDataRateIndex(), 0, false)
	if err != nil {
		panic(err)
	}

	chIdx, err := fp.Rx1Channel(sets.GetChannelIndex())
	if err != nil {
		panic(err)
	}

	dr := fp.DataRates[drIdx]

	downlink := ttnpb.DownlinkMessage{
		Settings: ttnpb.TxSettings{
			DataRateIndex:   uint32(drIdx),
			Frequency:       fp.DownlinkChannels[chIdx].Frequency,
			ChannelIndex:    chIdx,
			Modulation:      ParseModulation(dr.Rate),
			SpreadingFactor: ParseSpreadingFactor(dr.Rate),
			BitRate:         ParseBitRate(dr.Rate),
			Bandwidth:       ParseBandwidth(dr.Rate),
		},
		TxMetadata: ttnpb.TxMetadata{
			Timestamp: GetReceptionTimestamp(up) + 1000000000*uint64(fp.ReceiveDelay1),
		},
	}
	SendDownlink(downlink)
}
