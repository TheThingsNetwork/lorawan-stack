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
	"strconv"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

const upstreamBufferSize = 32 * 1024

var upstreamBuffer = make([]byte, upstreamBufferSize)

func (t *translator) Upstream(data udp.Data, md Metadata) (up *ttnpb.GatewayUp, err error) {
	up = &ttnpb.GatewayUp{}
	up.UplinkMessages = make([]*ttnpb.UplinkMessage, 0)
	for rxIndex, rx := range data.RxPacket {
		if rx == nil {
			continue
		}
		convertedRx, err := t.convertUplink(data, rxIndex, md.ID)
		if err != nil {
			t.Logger.WithError(err).Warn("Could not convert uplink to TTN format, ignoring")
			continue
		}
		up.UplinkMessages = append(up.UplinkMessages, &convertedRx)
	}

	if data.Stat != nil {
		up.GatewayStatus = t.convertStatus(*data.Stat, md)
		if up.GatewayStatus.AntennasLocation != nil && len(up.GatewayStatus.AntennasLocation) > 0 {
			t.Location = up.GatewayStatus.AntennasLocation[0]
		}
	}

	if len(up.UplinkMessages) == 0 && up.GatewayStatus == nil {
		return up, errors.New("Message could not be converted to TTN format")
	}

	return
}

func (t translator) metadata(rx udp.RxPacket, gatewayID ttnpb.GatewayIdentifiers) []*ttnpb.RxMetadata {
	return []*ttnpb.RxMetadata{
		{
			GatewayIdentifiers: gatewayID,
			Location:           t.Location,

			AntennaIndex: 0,

			Timestamp: uint64(rx.Tmst) * 1000,

			RSSI: float32(rx.RSSI),
			SNR:  float32(rx.LSNR),
		},
	}
}

func (t translator) fineTimestampMetadata(rx udp.RxPacket, gatewayID ttnpb.GatewayIdentifiers) []*ttnpb.RxMetadata {
	md := make([]*ttnpb.RxMetadata, 0)
	for _, signal := range rx.RSig {
		signalMetadata := &ttnpb.RxMetadata{
			GatewayIdentifiers: gatewayID,
			Location:           t.Location,

			AntennaIndex: uint32(signal.Ant),

			Timestamp: uint64(rx.Tmst) * 1000,

			RSSI:                  float32(signal.RSSIS),
			ChannelRSSI:           float32(signal.RSSIC),
			RSSIStandardDeviation: float32(signal.RSSISD),

			SNR:             float32(signal.LSNR),
			FrequencyOffset: int64(signal.FOff),
		}
		if signal.ETime != "" {
			signalMetadata.AESKeyID = strconv.Itoa(int(rx.Aesk))
			signalMetadata.EncryptedFineTimestamp = signal.ETime
		}
		md = append(md, signalMetadata)
	}
	return md
}

func (t translator) convertUplink(data udp.Data, rxIndex int, gatewayID ttnpb.GatewayIdentifiers) (up ttnpb.UplinkMessage, err error) {
	rx := *data.RxPacket[rxIndex]
	up.Settings = ttnpb.TxSettings{
		CodingRate: rx.CodR,
		Frequency:  uint64(rx.Freq * 1000000),
	}

	up.RawPayload, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(rx.Data, "="))
	if err != nil {
		return up, errors.NewWithCause(err, "Could not decode RX packet payload from base64 format")
	}

	if rx.RSig != nil && len(rx.RSig) > 0 {
		up.RxMetadata = t.fineTimestampMetadata(rx, gatewayID)
	} else {
		up.RxMetadata = t.metadata(rx, gatewayID)
	}

	if rx.Time != nil {
		goTime := time.Time(*rx.Time)
		for mdIndex := range up.RxMetadata {
			up.RxMetadata[mdIndex].Time = goTime
		}
		up.RxMetadata[0].Time = goTime
	}

	switch rx.Modu {
	case "LORA":
		up.Settings.Modulation = ttnpb.Modulation_LORA

		sf, err := rx.DatR.SpreadingFactor()
		if err != nil {
			return up, errors.NewWithCause(err, "Could not parse spreading factor")
		}
		up.Settings.SpreadingFactor = uint32(sf)
		if up.Settings.Bandwidth, err = rx.DatR.Bandwidth(); err != nil {
			return up, errors.NewWithCause(err, "Could not parse bandwidth")
		}
	case "FSK":
		up.Settings.Modulation = ttnpb.Modulation_FSK
		up.Settings.BitRate = rx.DatR.FSK
	default:
		return up, errors.New("Unknown modulation")
	}

	return
}

func addVersions(status *ttnpb.GatewayStatus, stat udp.Stat) {
	if stat.FPGA != nil {
		status.Versions["fpga"] = strconv.Itoa(int(*stat.FPGA))
	}
	if stat.DSP != nil {
		status.Versions["dsp"] = strconv.Itoa(int(*stat.DSP))
	}
	if stat.HAL != nil {
		status.Versions["hal"] = *stat.HAL
	}
}

func addMetrics(status *ttnpb.GatewayStatus, stat udp.Stat) {
	status.Metrics["rxnb"] = float32(stat.RXNb)
	status.Metrics["rxok"] = float32(stat.RXOK)
	status.Metrics["rxfw"] = float32(stat.RXFW)
	status.Metrics["ackr"] = float32(stat.ACKR)
	status.Metrics["dwnb"] = float32(stat.DWNb)
	status.Metrics["txnb"] = float32(stat.TXNb)
	if stat.Temp != nil {
		status.Metrics["temp"] = float32(*stat.Temp)
	}
	if stat.LPPS != nil {
		status.Metrics["lpps"] = float32(*stat.LPPS)
	}
	if stat.LMNW != nil {
		status.Metrics["lmnw"] = float32(*stat.LMNW)
	}
	if stat.LMST != nil {
		status.Metrics["lmst"] = float32(*stat.LMST)
	}
	if stat.LMOK != nil {
		status.Metrics["lmok"] = float32(*stat.LMOK)
	}
}

func (t translator) convertStatus(stat udp.Stat, md Metadata) *ttnpb.GatewayStatus {
	status := &ttnpb.GatewayStatus{
		Metrics:  map[string]float32{},
		Versions: map[string]string{},
	}

	if md.IP != "" {
		status.IP = []string{md.IP}
	}

	if t.locationFromAS {
		status.AntennasLocation = []*ttnpb.Location{t.Location}
	} else if stat.Lati != nil && stat.Long != nil {
		status.AntennasLocation = []*ttnpb.Location{
			{Latitude: float32(*stat.Lati), Longitude: float32(*stat.Long)},
		}
		if stat.Alti != nil {
			status.AntennasLocation[0].Altitude = *stat.Alti
		}
	}

	currentTime := time.Time(stat.Time)
	status.Time = currentTime
	if stat.Boot != nil {
		bootTime := time.Time(*stat.Boot)
		status.BootTime = bootTime
	}

	addVersions(status, stat)
	for versionName, version := range md.Versions {
		status.Versions[versionName] = version
	}
	addMetrics(status, stat)
	return status
}
