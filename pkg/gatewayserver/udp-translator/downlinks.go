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

package translator

import (
	"encoding/base64"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func (t translator) Downlink(d *ttnpb.GatewayDown) (udp.Data, error) {
	data := udp.Data{}

	if d != nil && d.DownlinkMessage != nil {
		if err := t.insertDownlink(&data, *d.DownlinkMessage); err != nil {
			return data, errors.NewWithCause(err, "Could not process received downlink")
		}
	}

	return data, nil
}

func (t translator) insertDownlink(data *udp.Data, downlink ttnpb.DownlinkMessage) (err error) {
	data.TxPacket = &udp.TxPacket{
		CodR: downlink.Settings.CodingRate,
		Freq: float64(downlink.Settings.Frequency) / 1000000,
		Imme: downlink.TxMetadata.Timestamp == 0,
		IPol: downlink.Settings.PolarizationInversion,
		Powe: uint8(downlink.Settings.TxPower),
		Size: uint16(len(downlink.RawPayload)),
		Tmst: uint32(downlink.TxMetadata.Timestamp / 1000), // nano->microseconds conversion
		Data: base64.StdEncoding.EncodeToString(downlink.RawPayload),
	}
	gpsTime := udp.CompactTime(downlink.TxMetadata.Time)
	data.TxPacket.Time = &gpsTime

	switch downlink.Settings.Modulation {
	case ttnpb.Modulation_LORA:
		data.TxPacket.Modu = "LORA"
		data.TxPacket.NCRC = true
		data.TxPacket.DatR.LoRa = fmt.Sprintf("SF%dBW%d", downlink.Settings.SpreadingFactor, downlink.Settings.Bandwidth/1000)
	case ttnpb.Modulation_FSK:
		data.TxPacket.Modu = "FSK"
		data.TxPacket.DatR.FSK = downlink.Settings.BitRate
	default:
		return errors.New("Unknown modulation")
	}

	return nil
}
