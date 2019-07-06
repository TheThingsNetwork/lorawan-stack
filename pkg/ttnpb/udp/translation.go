// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"strconv"
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/gpstime"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/version"
)

const (
	delta = 0.001 // For GPS comparisons
	lora  = "LORA"
	fsk   = "FSK"

	// eirpDelta is the delta between EIRP and ERP.
	eirpDelta = 2.15
)

var (
	ttnVersions = map[string]string{
		"ttn-lw-gateway-server": version.TTN,
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

// ToGatewayUp converts the UDP message to a gateway upstream message.
func ToGatewayUp(data Data, md UpstreamMetadata) (*ttnpb.GatewayUp, error) {
	up := &ttnpb.GatewayUp{}
	up.UplinkMessages = make([]*ttnpb.UplinkMessage, 0)
	for _, rx := range data.RxPacket {
		if rx == nil {
			continue
		}
		convertedRx, err := convertUplink(*rx, md)
		if err != nil {
			return nil, err
		}
		up.UplinkMessages = append(up.UplinkMessages, &convertedRx)
	}
	if data.Stat != nil {
		up.GatewayStatus = convertStatus(*data.Stat, md)
	}
	if data.TxPacketAck != nil {
		result, ok := ttnAckError[data.TxPacketAck.Error]
		if !ok {
			result = ttnpb.TxAcknowledgment_UNKNOWN_ERROR
		}
		up.TxAcknowledgment = &ttnpb.TxAcknowledgment{
			Result: result,
		}
	}
	return up, nil
}

var (
	ttnAckError = map[TxError]ttnpb.TxAcknowledgment_Result{
		TxErrNone:            ttnpb.TxAcknowledgment_SUCCESS,
		TxErrTooLate:         ttnpb.TxAcknowledgment_TOO_LATE,
		TxErrTooEarly:        ttnpb.TxAcknowledgment_TOO_EARLY,
		TxErrCollisionBeacon: ttnpb.TxAcknowledgment_COLLISION_BEACON,
		TxErrCollisionPacket: ttnpb.TxAcknowledgment_COLLISION_PACKET,
		TxErrTxFreq:          ttnpb.TxAcknowledgment_TX_FREQ,
		TxErrTxPower:         ttnpb.TxAcknowledgment_TX_POWER,
		TxErrGPSUnlocked:     ttnpb.TxAcknowledgment_GPS_UNLOCKED,
	}
	semtechAckError = map[ttnpb.TxAcknowledgment_Result]TxError{
		ttnpb.TxAcknowledgment_SUCCESS:          TxErrNone,
		ttnpb.TxAcknowledgment_TOO_LATE:         TxErrTooLate,
		ttnpb.TxAcknowledgment_TOO_EARLY:        TxErrTooEarly,
		ttnpb.TxAcknowledgment_COLLISION_BEACON: TxErrCollisionBeacon,
		ttnpb.TxAcknowledgment_COLLISION_PACKET: TxErrCollisionPacket,
		ttnpb.TxAcknowledgment_TX_FREQ:          TxErrTxFreq,
		ttnpb.TxAcknowledgment_TX_POWER:         TxErrTxPower,
		ttnpb.TxAcknowledgment_GPS_UNLOCKED:     TxErrGPSUnlocked,
	}
)

func metadata(rx RxPacket, gatewayID ttnpb.GatewayIdentifiers) []*ttnpb.RxMetadata {
	return []*ttnpb.RxMetadata{
		{
			GatewayIdentifiers: gatewayID,
			AntennaIndex:       0,
			ChannelIndex:       uint32(rx.Chan),
			Timestamp:          rx.Tmst,
			RSSI:               float32(rx.RSSI),
			ChannelRSSI:        float32(rx.RSSI),
			SNR:                float32(rx.LSNR),
		},
	}
}

func fineTimestampMetadata(rx RxPacket, gatewayID ttnpb.GatewayIdentifiers) []*ttnpb.RxMetadata {
	md := make([]*ttnpb.RxMetadata, 0)
	for _, signal := range rx.RSig {
		signalMetadata := &ttnpb.RxMetadata{
			GatewayIdentifiers: gatewayID,
			AntennaIndex:       uint32(signal.Ant),
			ChannelIndex:       uint32(signal.Chan),
			Timestamp:          rx.Tmst,
			RSSI:               float32(signal.RSSIC),
			ChannelRSSI:        float32(signal.RSSIC),
			SNR:                float32(signal.LSNR),
			FrequencyOffset:    int64(signal.FOff),
		}
		if signal.RSSIS != nil {
			signalMetadata.SignalRSSI = &pbtypes.FloatValue{
				Value: float32(*signal.RSSIS),
			}
		}
		if signal.RSSISD != nil {
			signalMetadata.RSSIStandardDeviation = float32(*signal.RSSISD)
		}
		if signal.ETime != "" {
			if etime, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(signal.ETime, "=")); err == nil {
				signalMetadata.EncryptedFineTimestampKeyID = strconv.Itoa(int(rx.Aesk))
				signalMetadata.EncryptedFineTimestamp = etime
			}
		}
		if signal.FTime != nil {
			signalMetadata.FineTimestamp = uint64(*signal.FTime)
		}
		md = append(md, signalMetadata)
	}
	return md
}

func convertUplink(rx RxPacket, md UpstreamMetadata) (ttnpb.UplinkMessage, error) {
	up := ttnpb.UplinkMessage{
		Settings: ttnpb.TxSettings{
			Frequency: uint64(rx.Freq * 1000000),
		},
	}

	rawPayload, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(rx.Data, "="))
	if err != nil {
		return up, errPayload.WithCause(err)
	}
	up.RawPayload = rawPayload

	if len(rx.RSig) > 0 {
		up.RxMetadata = fineTimestampMetadata(rx, md.ID)
	} else {
		up.RxMetadata = metadata(rx, md.ID)
	}
	for _, md := range up.RxMetadata {
		if up.Settings.Timestamp == 0 || up.Settings.Timestamp > md.Timestamp {
			up.Settings.Timestamp = md.Timestamp
		}
	}

	if rx.Time != nil {
		goTime := time.Time(*rx.Time)
		for _, md := range up.RxMetadata {
			md.Time = &goTime
		}
	}

	up.Settings.DataRate = rx.DatR.DataRate
	if lora := up.Settings.DataRate.GetLoRa(); lora != nil {
		up.Settings.CodingRate = rx.CodR
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
	status.Metrics["rxin"] = float32(stat.RxNb)
	status.Metrics["rxok"] = float32(stat.RxOk)
	status.Metrics["rxfw"] = float32(stat.RxFW)
	status.Metrics["ackr"] = float32(stat.ACKR)
	status.Metrics["txin"] = float32(stat.DWNb)
	status.Metrics["txok"] = float32(stat.TxNb)
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
		loc := &ttnpb.Location{Latitude: *stat.Lati, Longitude: *stat.Long}
		if stat.Alti != nil {
			loc.Altitude = *stat.Alti
		}
		if validLocation(*loc) {
			status.AntennaLocations = []*ttnpb.Location{loc}
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

// FromGatewayUp converts the upstream message to the UDP format.
func FromGatewayUp(up *ttnpb.GatewayUp) (rxs []*RxPacket, stat *Stat, ack *TxPacketAck) {
	rxs = make([]*RxPacket, 0, len(up.UplinkMessages))
	var modulation, codr string
	for _, msg := range up.UplinkMessages {
		switch msg.Settings.DataRate.Modulation.(type) {
		case *ttnpb.DataRate_LoRa:
			modulation = lora
			codr = msg.Settings.CodingRate
		case *ttnpb.DataRate_FSK:
			modulation = fsk
		}
		rxs = append(rxs, &RxPacket{
			Freq: float64(msg.Settings.Frequency) / 1000000,
			Chan: uint8(msg.RxMetadata[0].ChannelIndex),
			Modu: modulation,
			DatR: DataRate{msg.Settings.DataRate},
			CodR: codr,
			Size: uint16(len(msg.RawPayload)),
			Data: base64.StdEncoding.EncodeToString(msg.RawPayload),
			Tmst: msg.RxMetadata[0].Timestamp,
			RSSI: int16(msg.RxMetadata[0].RSSI),
			LSNR: float64(msg.RxMetadata[0].SNR),
		})
	}
	if up.GatewayStatus != nil {
		stat = &Stat{
			Time: ExpandedTime(up.GatewayStatus.Time),
		}
	}
	if up.TxAcknowledgment != nil {
		ack = &TxPacketAck{
			Error: semtechAckError[up.TxAcknowledgment.Result],
		}
	}
	return
}

// ToDownlinkMessage converts the UDP format to a downlink message.
func ToDownlinkMessage(tx *TxPacket) (*ttnpb.DownlinkMessage, error) {
	scheduled := &ttnpb.TxSettings{
		DataRate:  tx.DatR.DataRate,
		Frequency: uint64(tx.Freq * 1000000),
		Downlink: &ttnpb.TxSettings_Downlink{
			InvertPolarization: tx.IPol,
			TxPower:            float32(tx.Powe) + eirpDelta,
		},
		Timestamp: tx.Tmst,
	}
	if lora := scheduled.DataRate.GetLoRa(); lora != nil {
		scheduled.CodingRate = tx.CodR
	}
	if tx.Time != nil {
		gpsTime := gpstime.Parse(int64(*tx.Tmms))
		scheduled.Time = &gpsTime
	}
	buf, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(tx.Data, "="))
	if err != nil {
		return nil, err
	}
	return &ttnpb.DownlinkMessage{
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: scheduled,
		},
		RawPayload: buf,
	}, nil
}

// FromDownlinkMessage converts to the downlink message to the UDP format.
func FromDownlinkMessage(msg *ttnpb.DownlinkMessage) (*TxPacket, error) {
	payload := msg.GetRawPayload()
	scheduled := msg.GetScheduled()
	if scheduled == nil {
		return nil, errNotScheduled
	}
	tx := &TxPacket{
		Freq: float64(scheduled.Frequency) / 1000000,
		IPol: scheduled.Downlink.InvertPolarization,
		Powe: uint8(scheduled.Downlink.TxPower - eirpDelta),
		Size: uint16(len(payload)),
		Data: base64.StdEncoding.EncodeToString(payload),
		Tmst: scheduled.Timestamp,
	}
	if scheduled.Time != nil {
		gpsTime := uint64(gpstime.ToGPS(*scheduled.Time))
		tx.Tmms = &gpsTime
	} else if scheduled.Timestamp == 0 {
		tx.Imme = true
	}

	tx.DatR.DataRate = scheduled.DataRate
	switch scheduled.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_LoRa:
		tx.CodR = scheduled.CodingRate
		tx.NCRC = !scheduled.EnableCRC
		tx.Modu = lora
	case *ttnpb.DataRate_FSK:
		tx.Modu = fsk
	default:
		return tx, errModulation.WithAttributes("modulation", scheduled.DataRate)
	}
	return tx, nil
}
