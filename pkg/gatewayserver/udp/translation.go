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

package udp

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/version"
)

const delta = 0.001 // For GPS comparisons

var (
	ttnVersions = map[string]string{
		"ttn-gateway-server": version.TTN,
	}
	invalidLocations = []ttnpb.Location{
		{Latitude: 0.0, Longitude: 0.0},
		{Latitude: 10.0, Longitude: 20.0},
	}
)

func validLocation(loc ttnpb.Location) bool {
	for _, invalidLoc := range invalidLocations {
		if (loc.Latitude > invalidLoc.Latitude-delta && loc.Latitude < invalidLoc.Latitude+delta) &&
			(loc.Longitude > invalidLoc.Longitude-delta && loc.Longitude < invalidLoc.Longitude+delta) {
			return false
		}
	}

	return true
}

// UpstreamMetadata related to an uplink.
type UpstreamMetadata struct {
	ID ttnpb.GatewayIdentifiers
	IP string
}

// TranslateUpstream message from the UDP format to the protobuf format.
func TranslateUpstream(data Data, md UpstreamMetadata) (*ttnpb.GatewayUp, error) {
	up := &ttnpb.GatewayUp{}
	up.UplinkMessages = make([]*ttnpb.UplinkMessage, 0)
	for rxIndex, rx := range data.RxPacket {
		if rx == nil {
			continue
		}
		convertedRx, err := convertUplink(data, rxIndex, md)
		if err != nil {
			continue
		}
		up.UplinkMessages = append(up.UplinkMessages, &convertedRx)
	}

	if data.Stat != nil {
		up.GatewayStatus = convertStatus(*data.Stat, md)
	}

	if len(up.UplinkMessages) == 0 && up.GatewayStatus == nil {
		return nil, errors.New("Message could not be converted to TTN format")
	}

	return up, nil
}

func metadata(rx RxPacket, gatewayID ttnpb.GatewayIdentifiers) []*ttnpb.RxMetadata {
	return []*ttnpb.RxMetadata{
		{
			GatewayIdentifiers: gatewayID,

			AntennaIndex: 0,
			Timestamp:    uint64(rx.Tmst) * 1000,
			RSSI:         float32(rx.RSSI),
			SNR:          float32(rx.LSNR),
		},
	}
}

func fineTimestampMetadata(rx RxPacket, gatewayID ttnpb.GatewayIdentifiers) []*ttnpb.RxMetadata {
	md := make([]*ttnpb.RxMetadata, 0)
	for _, signal := range rx.RSig {
		signalMetadata := &ttnpb.RxMetadata{
			GatewayIdentifiers: gatewayID,

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

func convertUplink(data Data, rxIndex int, md UpstreamMetadata) (ttnpb.UplinkMessage, error) {
	up := ttnpb.UplinkMessage{}
	rx := *data.RxPacket[rxIndex]
	up.Settings = ttnpb.TxSettings{
		CodingRate: rx.CodR,
		Frequency:  uint64(rx.Freq * 1000000),
	}

	rawPayload, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(rx.Data, "="))
	if err != nil {
		return up, ErrDecodingPayloadFromBase64.NewWithCause(nil, err)
	}

	up.RawPayload = rawPayload
	if err := up.Payload.UnmarshalLoRaWAN(rawPayload); err != nil {
		return up, ErrUnmarshalFailed.NewWithCause(nil, err)
	}
	if macPayload := up.Payload.GetMACPayload(); macPayload != nil {
		up.DevAddr = &macPayload.FHDR.DevAddr
	}
	if jrPayload := up.Payload.GetJoinRequestPayload(); jrPayload != nil {
		up.DevEUI = &jrPayload.DevEUI
		up.JoinEUI = &jrPayload.JoinEUI
	}
	if rrPayload := up.Payload.GetRejoinRequestPayload(); rrPayload != nil {
		up.DevEUI = &rrPayload.DevEUI
		up.JoinEUI = &rrPayload.JoinEUI
	}

	if rx.RSig != nil && len(rx.RSig) > 0 {
		up.RxMetadata = fineTimestampMetadata(rx, md.ID)
	} else {
		up.RxMetadata = metadata(rx, md.ID)
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
			return up, ErrParsingSpreadingFactor.NewWithCause(nil, err)
		}
		up.Settings.SpreadingFactor = uint32(sf)
		if up.Settings.Bandwidth, err = rx.DatR.Bandwidth(); err != nil {
			return up, ErrParsingBandwidth.NewWithCause(nil, err)
		}
	case "FSK":
		up.Settings.Modulation = ttnpb.Modulation_FSK
		up.Settings.BitRate = rx.DatR.FSK
	default:
		return up, ErrUnknownModulation.New(nil)
	}

	return up, nil
}

func addVersions(status *ttnpb.GatewayStatus, stat Stat) {
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

func addMetrics(status *ttnpb.GatewayStatus, stat Stat) {
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

func convertStatus(stat Stat, md UpstreamMetadata) *ttnpb.GatewayStatus {
	status := &ttnpb.GatewayStatus{
		Metrics:  map[string]float32{},
		Versions: map[string]string{},
		IP:       []string{md.IP},
	}

	if stat.Lati != nil && stat.Long != nil {
		loc := &ttnpb.Location{Latitude: float32(*stat.Lati), Longitude: float32(*stat.Long)}
		if stat.Alti != nil {
			loc.Altitude = *stat.Alti
		}
		if validLocation(*loc) {
			status.AntennasLocation = []*ttnpb.Location{loc}
		}
	}

	currentTime := time.Time(stat.Time)
	status.Time = currentTime
	if stat.Boot != nil {
		bootTime := time.Time(*stat.Boot)
		status.BootTime = bootTime
	}

	addVersions(status, stat)
	for versionName, version := range ttnVersions {
		status.Versions[versionName] = version
	}
	addMetrics(status, stat)
	return status
}

// TranslateDownstream message from the protobuf format to the UDP format.
func TranslateDownstream(d *ttnpb.GatewayDown) (Data, error) {
	data := Data{}

	if d != nil && d.DownlinkMessage != nil {
		if err := insertDownlink(&data, *d.DownlinkMessage); err != nil {
			return data, errors.NewWithCause(err, "Could not process received downlink")
		}
	}

	return data, nil
}

func insertDownlink(data *Data, downlink ttnpb.DownlinkMessage) (err error) {
	payload := downlink.GetRawPayload()
	if payload == nil {
		var err error
		if payload, err = downlink.GetPayload().MarshalLoRaWAN(); err != nil {
			return ErrMarshalFailed.NewWithCause(nil, err)
		}
	}

	data.TxPacket = &TxPacket{
		CodR: downlink.Settings.CodingRate,
		Freq: float64(downlink.Settings.Frequency) / 1000000,
		Imme: downlink.TxMetadata.Timestamp == 0,
		IPol: downlink.Settings.PolarizationInversion,
		Powe: uint8(downlink.Settings.TxPower),
		Size: uint16(len(payload)),
		Tmst: uint32(downlink.TxMetadata.Timestamp / 1000), // nano->microseconds conversion
		Data: base64.StdEncoding.EncodeToString(payload),
	}
	gpsTime := CompactTime(downlink.TxMetadata.Time)
	data.TxPacket.Time = &gpsTime

	switch downlink.Settings.Modulation {
	case ttnpb.Modulation_LORA:
		data.TxPacket.Modu = "LORA"
		data.TxPacket.NCRC = !downlink.TxMetadata.EnableCRC
		data.TxPacket.DatR.LoRa = fmt.Sprintf("SF%dBW%d", downlink.Settings.SpreadingFactor, downlink.Settings.Bandwidth/1000)
	case ttnpb.Modulation_FSK:
		data.TxPacket.Modu = "FSK"
		data.TxPacket.DatR.FSK = downlink.Settings.BitRate
	default:
		return ErrUnknownModulation.New(nil)
	}

	return nil
}
