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

package lbslns

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/basicstation"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/util"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	errJoinRequestMessage = errors.Define("join_request_message", "invalid join-request message received")
	errUplinkDataFrame    = errors.Define("uplink_data_Frame", "invalid uplink data frame received")
	errUplinkMessage      = errors.Define("uplink_message", "invalid uplink message received")
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

// JoinRequest is the LoRaWAN Join Request message from LoRa Basics Station protocol.
type JoinRequest struct {
	MHdr     uint             `json:"MHdr"`
	JoinEUI  basicstation.EUI `json:"JoinEui"`
	DevEUI   basicstation.EUI `json:"DevEui"`
	DevNonce uint             `json:"DevNonce"`
	MIC      int32            `json:"MIC"`
	RefTime  float64          `json:"RefTime"`
	RadioMetaData
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

// UplinkDataFrame is the LoRaWAN Uplink message of the LoRa Basics Station protocol.
type UplinkDataFrame struct {
	MHdr       uint    `json:"MHdr"`
	DevAddr    int32   `json:"DevAddr"`
	FCtrl      uint    `json:"FCtrl"`
	FCnt       uint    `json:"Fcnt"`
	FOpts      string  `json:"FOpts"`
	FPort      int     `json:"FPort"`
	FRMPayload string  `json:"FRMPayload"`
	MIC        int32   `json:"MIC"`
	RefTime    float64 `json:"RefTime"`
	RadioMetaData
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
	RefTime float64          `json:"RefTime,omitempty"`
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

// toUplinkMessage extracts fields from the Basics Station Join Request "jreq" message and converts them into an UplinkMessage for the network server.
func (req *JoinRequest) toUplinkMessage(ids ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	var up ttnpb.UplinkMessage
	up.ReceivedAt = receivedAt

	var parsedMHDR ttnpb.MHDR
	if err := lorawan.UnmarshalMHDR([]byte{byte(req.MHdr)}, &parsedMHDR); err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	micBytes, err := util.GetInt32AsByteSlice(req.MIC)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}
	up.Payload = &ttnpb.Message{
		MHDR: parsedMHDR,
		MIC:  micBytes,
		Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
			JoinEUI:  req.JoinEUI.EUI64,
			DevEUI:   req.DevEUI.EUI64,
			DevNonce: [2]byte{byte(req.DevNonce >> 8), byte(req.DevNonce)},
		}},
	}

	up.RawPayload, err = lorawan.MarshalMessage(*up.Payload)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	timestamp := uint32(req.RadioMetaData.UpInfo.XTime & 0xFFFFFFFF)

	var rxTime *time.Time
	sec, nsec := math.Modf(req.RadioMetaData.UpInfo.RxTime)
	if sec != 0 {
		val := time.Unix(int64(sec), int64(nsec*(1e9)))
		rxTime = &val
	}

	rxMetadata := &ttnpb.RxMetadata{
		GatewayIdentifiers: ids,
		Time:               rxTime,
		Timestamp:          timestamp,
		RSSI:               req.RadioMetaData.UpInfo.RSSI,
		ChannelRSSI:        req.RadioMetaData.UpInfo.RSSI,
		SNR:                req.RadioMetaData.UpInfo.SNR,
		AntennaIndex:       uint32(req.RadioMetaData.UpInfo.RCtx),
	}
	up.RxMetadata = append(up.RxMetadata, rxMetadata)

	dataRate, isLora, err := util.GetDataRateFromIndex(bandID, req.RadioMetaData.DataRate)
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

// FromUplinkMessage extracts fields from ttnpb.UplinkMessage and creates the LoRa Basics Station Join Request Frame.
func (req *JoinRequest) FromUplinkMessage(up *ttnpb.UplinkMessage, bandID string) error {
	var payload ttnpb.Message
	err := lorawan.UnmarshalMessage(up.RawPayload, &payload)
	if err != nil {
		return errUplinkMessage.New()
	}
	req.MHdr = (uint(payload.MHDR.GetMType()) << 5) | uint(payload.MHDR.GetMajor())
	req.MIC = int32(binary.LittleEndian.Uint32(payload.MIC[:]))
	jreqPayload := payload.GetJoinRequestPayload()
	if jreqPayload == nil {
		return errUplinkMessage.New()
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
	req.DevNonce = uint(binary.BigEndian.Uint16(devNonce[:]))

	dr, err := util.GetDataRateIndexFromDataRate(bandID, up.Settings.GetDataRate())
	if err != nil {
		return err
	}

	rxMetadata := up.RxMetadata[0]

	var rxTime float64
	if rxMetadata.Time != nil {
		rxTime = float64(rxMetadata.Time.Unix()) + float64(rxMetadata.Time.Nanosecond())/(1e9)
	}

	req.RadioMetaData = RadioMetaData{
		DataRate:  dr,
		Frequency: up.Settings.GetFrequency(),
		UpInfo: UpInfo{
			RCtx:   int64(rxMetadata.AntennaIndex),
			XTime:  int64(rxMetadata.Timestamp),
			RSSI:   rxMetadata.RSSI,
			SNR:    rxMetadata.SNR,
			RxTime: rxTime,
		},
	}
	return nil
}

// toUplinkMessage extracts fields from the LoRa Basics Station Uplink Data Frame "updf" message and converts them into an UplinkMessage for the network server.
func (updf *UplinkDataFrame) toUplinkMessage(ids ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	var up ttnpb.UplinkMessage
	up.ReceivedAt = receivedAt

	var parsedMHDR ttnpb.MHDR
	if err := lorawan.UnmarshalMHDR([]byte{byte(updf.MHdr)}, &parsedMHDR); err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}
	if parsedMHDR.MType != ttnpb.MType_UNCONFIRMED_UP && parsedMHDR.MType != ttnpb.MType_CONFIRMED_UP {
		return nil, errUplinkDataFrame.New()
	}

	micBytes, err := util.GetInt32AsByteSlice(updf.MIC)
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
		ChannelRSSI:        updf.RadioMetaData.UpInfo.RSSI,
		SNR:                updf.RadioMetaData.UpInfo.SNR,
		AntennaIndex:       uint32(updf.RadioMetaData.UpInfo.RCtx),
	}
	up.RxMetadata = append(up.RxMetadata, rxMetadata)

	dataRate, isLora, err := util.GetDataRateFromIndex(bandID, updf.RadioMetaData.DataRate)
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

// FromUplinkMessage extracts fields from ttnpb.UplinkMessage and creates the LoRa Basics Station UplinkDataFrame.
func (updf *UplinkDataFrame) FromUplinkMessage(up *ttnpb.UplinkMessage, bandID string) error {
	var payload ttnpb.Message
	err := lorawan.UnmarshalMessage(up.RawPayload, &payload)
	if err != nil {
		return errUplinkMessage.New()
	}
	updf.MHdr = (uint(payload.MHDR.GetMType()) << 5) | uint(payload.MHDR.GetMajor())

	macPayload := payload.GetMACPayload()
	if macPayload == nil {
		return errUplinkMessage.New()
	}

	updf.FPort = int(macPayload.GetFPort())

	updf.DevAddr = int32(macPayload.DevAddr.MarshalNumber())
	updf.FOpts = hex.EncodeToString(macPayload.GetFOpts())

	updf.FCtrl = util.GetFCtrlAsUint(macPayload.FCtrl)
	updf.FCnt = uint(macPayload.GetFCnt())
	updf.FRMPayload = hex.EncodeToString(macPayload.GetFRMPayload())
	updf.MIC = int32(binary.LittleEndian.Uint32(payload.MIC[:]))

	dr, err := util.GetDataRateIndexFromDataRate(bandID, up.Settings.GetDataRate())
	if err != nil {
		return err
	}

	rxMetadata := up.RxMetadata[0]

	var rxTime float64
	if rxMetadata.Time != nil {
		rxTime = float64(rxMetadata.Time.Unix()) + float64(rxMetadata.Time.Nanosecond())/(1e9)
	}

	updf.RadioMetaData = RadioMetaData{
		DataRate:  dr,
		Frequency: up.Settings.GetFrequency(),
		UpInfo: UpInfo{
			RCtx:   int64(rxMetadata.AntennaIndex),
			XTime:  int64(rxMetadata.Timestamp),
			RSSI:   rxMetadata.RSSI,
			SNR:    rxMetadata.SNR,
			RxTime: rxTime,
		},
	}
	return nil
}

// ToTxAck converts the LoRa Basics Station TxConfirmation message to ttnpb.TxAcknowledgment
func (conf TxConfirmation) ToTxAck(ctx context.Context, tokens io.DownlinkTokens, receivedAt time.Time) *ttnpb.TxAcknowledgment {
	var txAck ttnpb.TxAcknowledgment
	if cids, _, ok := tokens.Get(uint16(conf.Diid), receivedAt); ok {
		txAck.CorrelationIDs = cids
		txAck.Result = ttnpb.TxAcknowledgment_SUCCESS
	} else {
		logger := log.FromContext(ctx)
		logger.WithField("diid", conf.Diid).Debug("Tx acknowledgement either does not correspond to a downlink message or arrived too late")
	}
	return &txAck
}

// HandleUp implements Formatter.
func (f *lbsLNS) HandleUp(ctx context.Context, raw []byte, ids ttnpb.GatewayIdentifiers, conn *io.Connection, receivedAt time.Time) ([]byte, error) {
	logger := log.FromContext(ctx)
	typ, err := Type(raw)
	if err != nil {
		logger.WithError(err).Debug("Failed to parse message type")
		return nil, err
	}
	logger = logger.WithFields(log.Fields(
		"upstream_type", typ,
	))

	recordTime := func(refTime float64, xTime int64, server time.Time) {
		sec, nsec := math.Modf(refTime)
		if sec != 0 {
			ref := time.Unix(int64(sec), int64(nsec*1e9))
			conn.RecordRTT(server.Sub(ref), server)
		}
		conn.SyncWithGatewayConcentrator(
			// The concentrator timestamp is the 32 LSB.
			uint32(xTime&0xFFFFFFFF),
			server,
			// The Basic Station epoch is the 48 LSB.
			scheduling.ConcentratorTime(time.Duration(xTime&0xFFFFFFFFFF)*time.Microsecond),
		)
	}

	switch typ {
	case TypeUpstreamVersion:
		ctx, msg, stat, err := f.GetRouterConfig(ctx, raw, conn.BandID(), conn.FrequencyPlans(), receivedAt)
		logger = log.FromContext(ctx)
		if err != nil {
			logger.WithError(err).Warn("Failed to generate router configuration")
			return nil, err
		}
		if err := conn.HandleStatus(stat); err != nil {
			logger.WithError(err).Warn("Failed to handle status message")
			return nil, err
		}
		return msg, nil

	case TypeUpstreamJoinRequest:
		var jreq JoinRequest
		if err := json.Unmarshal(raw, &jreq); err != nil {
			return nil, err
		}
		// TODO: Remove (https://github.com/lorabasics/basicstation/issues/74)
		if jreq.UpInfo.XTime == 0 {
			logger.Warn("Received join-request without xtime, drop message")
			break
		}
		up, err := jreq.toUplinkMessage(ids, conn.BandID(), receivedAt)
		if err != nil {
			logger.WithError(err).Warn("Failed to parse join request")
			return nil, err
		}
		if err := conn.HandleUp(up); err != nil {
			logger.WithError(err).Warn("Failed to handle upstream message")
			return nil, err
		}
		session := ws.SessionFromContext(ctx)
		session.DataMu.Lock()
		session.Data = State{
			ID: int32(jreq.UpInfo.XTime >> 48),
		}
		session.DataMu.Unlock()
		recordTime(jreq.RefTime, jreq.UpInfo.XTime, receivedAt)

	case TypeUpstreamUplinkDataFrame:
		var updf UplinkDataFrame
		if err := json.Unmarshal(raw, &updf); err != nil {
			return nil, err
		}
		// TODO: Remove (https://github.com/lorabasics/basicstation/issues/74)
		if updf.UpInfo.XTime == 0 {
			logger.Warn("Received uplink without xtime, drop message")
			return nil, nil
		}
		up, err := updf.toUplinkMessage(ids, conn.BandID(), receivedAt)
		if err != nil {
			logger.WithError(err).Warn("Failed to parse uplink message")
			return nil, err
		}
		if err := conn.HandleUp(up); err != nil {
			logger.WithError(err).Warn("Failed to handle upstream message")
			return nil, err
		}
		session := ws.SessionFromContext(ctx)
		session.DataMu.Lock()
		session.Data = State{
			ID: int32(updf.UpInfo.XTime >> 48),
		}
		session.DataMu.Unlock()
		recordTime(updf.RefTime, updf.UpInfo.XTime, receivedAt)

	case TypeUpstreamTxConfirmation:
		var txConf TxConfirmation
		if err := json.Unmarshal(raw, &txConf); err != nil {
			return nil, err
		}
		txAck := txConf.ToTxAck(ctx, f.tokens, receivedAt)
		if txAck == nil {
			break
		}
		if err := conn.HandleTxAck(txAck); err != nil {
			logger.WithError(err).Warn("Failed to handle tx ack message")
			return nil, err
		}
		session := ws.SessionFromContext(ctx)
		session.DataMu.Lock()
		session.Data = State{
			ID: int32(txConf.XTime >> 48),
		}
		session.DataMu.Unlock()
		recordTime(txConf.RefTime, txConf.XTime, receivedAt)
		return nil, err

	case TypeUpstreamProprietaryDataFrame, TypeUpstreamRemoteShell, TypeUpstreamTimeSync:
		logger.WithField("message_type", typ).Debug("Message type not implemented")
		break
	default:
		logger.WithField("message_type", typ).Debug("Unknown message type")
		break
	}
	return nil, nil
}
