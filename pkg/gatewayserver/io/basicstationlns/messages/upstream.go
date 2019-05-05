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

package messages

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"time"

	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	errJoinRequestMessage = errors.Define("join_request_message", "invalid join-request message received")
	errUplinkDataFrame    = errors.Define("uplink_data_Frame", "invalid uplink data frame received")
	errUplinkMessage      = errors.Define("uplink_message", "invalid uplink message received")
	errGatewayID          = errors.Define("gateway_id", "invalid gateway ID `{id}`")
)

// UpInfo provides additional metadata on each upstream message.
type UpInfo struct {
	RxTime  float64 `json:"rxtime"`
	RCtx    int64   `json:"rtcx"`
	XTime   int64   `json:"xtime"`
	GPSTime int64   `json:"gpstime"`
	RSSI    float32 `json:"rssi"`
	SNR     float32 `json:"snr"`
}

// RadioMetaData is a the metadata that is received as part of all upstream messages (except Tx Confirmation).
type RadioMetaData struct {
	DataRate  int    `json:"DR"`
	Frequency uint64 `json:"Freq"`
	UpInfo    UpInfo `json:"upinfo"`
}

// JoinRequest is the LoRaWAN Join Request message from the BasicStation.
type JoinRequest struct {
	MHdr          uint             `json:"MHdr"`
	JoinEUI       basicstation.EUI `json:"JoinEui"`
	DevEUI        basicstation.EUI `json:"DevEui"`
	DevNonce      uint             `json:"DevNonce"`
	MIC           int32            `json:"MIC"`
	RefTime       float64          `json:"RefTime"`
	RadioMetaData RadioMetaData
}

// MarshalJSON implements json.Marshaler.
func (req JoinRequest) MarshalJSON() ([]byte, error) {
	type Alias JoinRequest
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamJoinRequest,
		Alias: Alias(req),
	})
}

// UplinkDataFrame is the LoRaWAN Uplink message from the BasicStation.
type UplinkDataFrame struct {
	MHdr          uint    `json:"MHdr"`
	DevAddr       int32   `json:"DevAddr"`
	FCtrl         uint    `json:"FCtrl"`
	FCnt          uint    `json:"Fcnt"`
	FOpts         string  `json:"FOpts"`
	FPort         int     `json:"FPort"`
	FRMPayload    string  `json:"FRMPayload"`
	MIC           int32   `json:"MIC"`
	RefTime       float64 `json:"RefTime"`
	RadioMetaData RadioMetaData
}

// MarshalJSON implements json.Marshaler.
func (updf UplinkDataFrame) MarshalJSON() ([]byte, error) {
	type Alias UplinkDataFrame
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamUplinkDataFrame,
		Alias: Alias(updf),
	})
}

// TxConfirmation is the LoRaWAN Join Request message from the BasicStation.
type TxConfirmation struct {
	Diid    int64            `json:"diid"`
	DevEUI  basicstation.EUI `json:"DevEui"`
	RCtx    int64            `json:"rctx"`
	XTime   int64            `json:"xtime"`
	TxTime  float64          `json:"txtime"`
	GpsTime int64            `json:"gpstime"`
}

// MarshalJSON implements json.Marshaler.
func (conf TxConfirmation) MarshalJSON() ([]byte, error) {
	type Alias TxConfirmation
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamTxConfirmation,
		Alias: Alias(conf),
	})
}

// ToUplinkMessage extracts fields from the basic station Join Request "jreq" message and converts them into an UplinkMessage for the network server.
func (req *JoinRequest) ToUplinkMessage(ids ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	var up ttnpb.UplinkMessage
	up.ReceivedAt = receivedAt

	var parsedMHDR ttnpb.MHDR
	if err := lorawan.UnmarshalMHDR([]byte{byte(req.MHdr)}, &parsedMHDR); err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	micBytes, err := getInt32AsByteSlice(req.MIC)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}
	up.Payload = &ttnpb.Message{
		MHDR: parsedMHDR,
		MIC:  micBytes,
		Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
			JoinEUI:  req.JoinEUI.EUI64,
			DevEUI:   req.DevEUI.EUI64,
			DevNonce: [2]byte{byte(req.DevNonce), byte(req.DevNonce >> 8)},
		}},
	}

	up.RawPayload, err = lorawan.MarshalMessage(*up.Payload)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	timestamp := uint32(req.RadioMetaData.UpInfo.XTime & 0xFFFFFFFF)

	ulToken := ttnpb.UplinkToken{
		GatewayAntennaIdentifiers: ttnpb.GatewayAntennaIdentifiers{
			GatewayIdentifiers: ids,
			AntennaIndex:       uint32(req.RadioMetaData.UpInfo.RCtx),
		},
		Timestamp: timestamp,
	}
	ulTokenBytes, err := ulToken.Marshal()
	if err != nil {
		return nil, err
	}

	var rxTime *time.Time
	sec, nsec := math.Modf(req.RadioMetaData.UpInfo.RxTime)
	if sec != 0 {
		val := time.Unix(int64(sec), int64(nsec*(1e9)))
		rxTime = &val
	}

	ids.EUI, err = GetEUIfromUID(ids.GatewayID)
	if err != nil {
		return nil, errGatewayID.WithAttributes("id", ids.GatewayID).WithCause(err)
	}

	rxMetadata := &ttnpb.RxMetadata{
		GatewayIdentifiers: ids,
		Time:               rxTime,
		Timestamp:          timestamp,
		RSSI:               req.RadioMetaData.UpInfo.RSSI,
		SNR:                req.RadioMetaData.UpInfo.SNR,
		UplinkToken:        ulTokenBytes,
	}
	up.RxMetadata = append(up.RxMetadata, rxMetadata)

	dataRate, isLora, err := getDataRateFromIndex(bandID, req.RadioMetaData.DataRate)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	var codingRate string
	if isLora {
		codingRate = "4/5"
	}

	up.Settings = ttnpb.TxSettings{
		Frequency:  req.RadioMetaData.Frequency,
		DataRate:   dataRate,
		CodingRate: codingRate,
		Timestamp:  timestamp,
		Time:       rxTime,
	}

	return &up, nil
}

// FromUplinkMessage extracts fields from ttnpb.UplinkMessage and creates the Basic Station Join Request Frame.
func (req *JoinRequest) FromUplinkMessage(up *ttnpb.UplinkMessage, bandID string) error {
	var payload ttnpb.Message
	err := lorawan.UnmarshalMessage(up.RawPayload, &payload)
	if err != nil {
		return errUplinkMessage
	}
	req.MHdr = (uint(payload.MHDR.GetMType()) << 5) | uint(payload.MHDR.GetMajor())
	req.MIC = int32(binary.LittleEndian.Uint32(payload.MIC[:]))
	jreqPayload := payload.GetJoinRequestPayload()
	if jreqPayload == nil {
		return errUplinkMessage
	}

	req.DevEUI = basicstation.EUI{
		EUI64: jreqPayload.DevEUI,
	}

	req.JoinEUI = basicstation.EUI{
		EUI64: jreqPayload.JoinEUI,
	}

	devNonce, err := jreqPayload.DevNonce.Marshal()
	if err != nil {
		return err
	}
	req.DevNonce = uint(binary.LittleEndian.Uint16(devNonce[:]))

	dr, err := getDataRateIndexFromDataRate(bandID, up.Settings.GetDataRate())
	if err != nil {
		return err
	}
	rxMetadata := up.RxMetadata[0]
	antennaIDs, _, err := io.ParseUplinkToken(rxMetadata.UplinkToken)
	if err != nil {
		return err
	}

	var rxTime float64
	if rxMetadata.Time != nil {
		rxTime = float64(rxMetadata.Time.Unix()) + float64(rxMetadata.Time.Nanosecond())/(1e9)
	}
	req.RadioMetaData = RadioMetaData{
		DataRate:  dr,
		Frequency: up.Settings.GetFrequency(),
		UpInfo: UpInfo{
			RCtx:   int64(antennaIDs.AntennaIndex),
			XTime:  int64(rxMetadata.Timestamp),
			RSSI:   rxMetadata.RSSI,
			SNR:    rxMetadata.SNR,
			RxTime: rxTime,
		},
	}
	return nil
}

// ToUplinkMessage extracts fields from the basic station Uplink Data Frame "updf" message and converts them into an UplinkMessage for the network server.
func (updf *UplinkDataFrame) ToUplinkMessage(ids ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	var up ttnpb.UplinkMessage
	up.ReceivedAt = receivedAt

	var parsedMHDR ttnpb.MHDR
	if err := lorawan.UnmarshalMHDR([]byte{byte(updf.MHdr)}, &parsedMHDR); err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}
	if parsedMHDR.MType != ttnpb.MType_UNCONFIRMED_UP && parsedMHDR.MType != ttnpb.MType_CONFIRMED_UP {
		return nil, errUplinkDataFrame
	}

	micBytes, err := getInt32AsByteSlice(updf.MIC)
	if err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	var fPort uint32
	if updf.FPort == -1 {
		fPort = 0
	} else {
		fPort = uint32(updf.FPort)
	}

	var devAddr types.DevAddr
	devAddr.UnmarshalNumber(uint32(updf.DevAddr))

	var fctrl ttnpb.FCtrl
	if err := lorawan.UnmarshalFCtrl([]byte{byte(updf.FCtrl)}, &fctrl, true); err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	decFRMPayload, err := hex.DecodeString(updf.FRMPayload)
	if err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	decFOpts, err := hex.DecodeString(updf.FOpts)
	if err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	up.Payload = &ttnpb.Message{
		MHDR: parsedMHDR,
		MIC:  micBytes,
		Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
			FPort:      fPort,
			FRMPayload: decFRMPayload,
			FHDR: ttnpb.FHDR{
				DevAddr: devAddr,
				FCtrl:   fctrl,
				FCnt:    uint32(updf.FCnt),
				FOpts:   decFOpts,
			},
		}},
	}

	up.RawPayload, err = lorawan.MarshalMessage(*up.Payload)
	if err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	timestamp := uint32(updf.RadioMetaData.UpInfo.XTime & 0xFFFFFFFF)

	ids.EUI, err = GetEUIfromUID(ids.GatewayID)
	if err != nil {
		return nil, errGatewayID.WithAttributes("id", ids.GatewayID).WithCause(err)
	}

	ulToken := ttnpb.UplinkToken{
		GatewayAntennaIdentifiers: ttnpb.GatewayAntennaIdentifiers{
			GatewayIdentifiers: ids,
			AntennaIndex:       uint32(updf.RadioMetaData.UpInfo.RCtx),
		},
		Timestamp: timestamp,
	}
	ulTokenBytes, err := ulToken.Marshal()
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	var rxTime *time.Time
	sec, nsec := math.Modf(updf.RadioMetaData.UpInfo.RxTime)
	if sec != 0 {
		val := time.Unix(int64(sec), int64(nsec*(1e9)))
		rxTime = &val
	}

	rxMetadata := &ttnpb.RxMetadata{
		GatewayIdentifiers: ids,
		Time:               rxTime,
		Timestamp:          timestamp,
		RSSI:               updf.RadioMetaData.UpInfo.RSSI,
		SNR:                updf.RadioMetaData.UpInfo.SNR,
		UplinkToken:        ulTokenBytes,
	}
	up.RxMetadata = append(up.RxMetadata, rxMetadata)

	dataRate, isLora, err := getDataRateFromIndex(bandID, updf.RadioMetaData.DataRate)
	if err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	var codingRate string
	if isLora {
		codingRate = "4/5"
	}

	up.Settings = ttnpb.TxSettings{
		Frequency:  updf.RadioMetaData.Frequency,
		DataRate:   dataRate,
		CodingRate: codingRate,
		Timestamp:  timestamp,
		Time:       rxTime,
	}
	return &up, nil
}

// FromUplinkMessage extracts fields from ttnpb.UplinkMessage and creates the Basic Station UplinkDataFrame.
func (updf *UplinkDataFrame) FromUplinkMessage(up *ttnpb.UplinkMessage, bandID string) error {
	var payload ttnpb.Message
	err := lorawan.UnmarshalMessage(up.RawPayload, &payload)
	if err != nil {
		return errUplinkMessage
	}
	updf.MHdr = (uint(payload.MHDR.GetMType()) << 5) | uint(payload.MHDR.GetMajor())

	macPayload := payload.GetMACPayload()
	if macPayload == nil {
		return errUplinkMessage
	}

	updf.FPort = int(macPayload.GetFPort())

	updf.DevAddr = int32(macPayload.DevAddr.MarshalNumber())
	updf.FOpts = hex.EncodeToString(macPayload.GetFOpts())

	updf.FCtrl = getFCtrlAsUint(macPayload.FCtrl)
	updf.FCnt = uint(macPayload.GetFCnt())
	updf.FRMPayload = hex.EncodeToString(macPayload.GetFRMPayload())
	updf.MIC = int32(binary.LittleEndian.Uint32(payload.MIC[:]))

	dr, err := getDataRateIndexFromDataRate(bandID, up.Settings.GetDataRate())
	if err != nil {
		return err
	}

	rxMetadata := up.RxMetadata[0]
	antennaIDs, _, err := io.ParseUplinkToken(rxMetadata.UplinkToken)
	if err != nil {
		return err
	}

	var rxTime float64
	if rxMetadata.Time != nil {
		rxTime = float64(rxMetadata.Time.Unix()) + float64(rxMetadata.Time.Nanosecond())/(1e9)
	}

	updf.RadioMetaData = RadioMetaData{
		DataRate:  dr,
		Frequency: up.Settings.GetFrequency(),
		UpInfo: UpInfo{
			RCtx:   int64(antennaIDs.AntennaIndex),
			XTime:  int64(rxMetadata.Timestamp),
			RSSI:   rxMetadata.RSSI,
			SNR:    rxMetadata.SNR,
			RxTime: rxTime,
		},
	}
	return nil
}

// ToTxAcknowledgment extracts fields from the basic station TxConfirmation "dntxed" message and converts them into a TxAcknowledgment for the network server.
func ToTxAcknowledgment(correlationIDs []string) ttnpb.TxAcknowledgment {
	return ttnpb.TxAcknowledgment{
		CorrelationIDs: correlationIDs,
		Result:         ttnpb.TxAcknowledgment_SUCCESS,
	}
}
