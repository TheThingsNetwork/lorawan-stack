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
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/basicstation"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	errJoinRequestMessage = errors.Define("join_request_message", "invalid join-request message received")
	errUplinkDataFrame    = errors.Define("uplink_data_frame", "invalid uplink data frame received")
	errUplinkMessage      = errors.Define("uplink_message", "invalid uplink message received")
	errMDHR               = errors.Define("mhdr", "invalid MHDR `{mhdr}` received")
	errDataRate           = errors.Define("data_rate", "invalid data rate")
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
	GPSTime int64            `json:"gpstime"`
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

func getInt32AsByteSlice(value int32) ([]byte, error) {
	b := &bytes.Buffer{}
	err := binary.Write(b, binary.LittleEndian, value)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// TimeSyncRequest is the time synchronization request from the BasicStation.
type TimeSyncRequest struct {
	TxTime float64 `json:"txtime"`
}

// MarshalJSON implements json.Marshaler.
func (tsr TimeSyncRequest) MarshalJSON() ([]byte, error) {
	type Alias TimeSyncRequest
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamTimeSync,
		Alias: Alias(tsr),
	})
}

// Response generates a TimeSyncResponse for this request.
func (tsr TimeSyncRequest) Response(t time.Time) TimeSyncResponse {
	return TimeSyncResponse{
		TxTime:  tsr.TxTime,
		GPSTime: TimeToGPSTime(t),
		MuxTime: TimeToUnixSeconds(t),
	}
}

// TimeSyncResponse is the time synchronization response to the BasicStation.
type TimeSyncResponse struct {
	TxTime  float64 `json:"txtime,omitempty"`
	XTime   int64   `json:"xtime,omitempty"`
	GPSTime int64   `json:"gpstime"`
	MuxTime float64 `json:"MuxTime,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (tsr TimeSyncResponse) MarshalJSON() ([]byte, error) {
	type Alias TimeSyncResponse
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamTimeSync,
		Alias: Alias(tsr),
	})
}

// toUplinkMessage extracts fields from the Basics Station Join Request "jreq" message and converts them into an UplinkMessage for the network server.
func (req *JoinRequest) toUplinkMessage(ids *ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	var up ttnpb.UplinkMessage
	up.ReceivedAt = ttnpb.ProtoTimePtr(receivedAt)

	parsedMHDR := &ttnpb.MHDR{}
	if err := lorawan.UnmarshalMHDR([]byte{byte(req.MHdr)}, parsedMHDR); err != nil {
		return nil, errMDHR.WithAttributes(`mhdr`, parsedMHDR)
	}

	micBytes, err := getInt32AsByteSlice(req.MIC)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}
	up.Payload = &ttnpb.Message{
		MHdr: parsedMHDR,
		Mic:  micBytes,
		Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
			JoinEui:  req.JoinEUI.EUI64.Bytes(),
			DevEui:   req.DevEUI.EUI64.Bytes(),
			DevNonce: []byte{byte(req.DevNonce >> 8), byte(req.DevNonce)},
		}},
	}

	up.RawPayload, err = lorawan.MarshalMessage(up.Payload)
	if err != nil {
		return nil, errJoinRequestMessage.WithCause(err)
	}

	timestamp := TimestampFromXTime(req.RadioMetaData.UpInfo.XTime)
	tm := TimePtrFromUpInfo(req.UpInfo.GPSTime, req.UpInfo.RxTime)
	gpsTime := TimePtrFromGPSTime(req.UpInfo.GPSTime)
	up.RxMetadata = []*ttnpb.RxMetadata{
		{
			GatewayIds:   ids,
			Time:         ttnpb.ProtoTime(tm),
			GpsTime:      ttnpb.ProtoTime(gpsTime),
			Timestamp:    timestamp,
			Rssi:         req.RadioMetaData.UpInfo.RSSI,
			ChannelRssi:  req.RadioMetaData.UpInfo.RSSI,
			Snr:          req.RadioMetaData.UpInfo.SNR,
			AntennaIndex: uint32(req.RadioMetaData.UpInfo.RCtx),
		},
	}

	phy, err := band.GetLatest(bandID)
	if err != nil {
		return nil, err
	}
	bandDR, ok := phy.DataRates[ttnpb.DataRateIndex(req.RadioMetaData.DataRate)]
	if !ok {
		return nil, errDataRate.New()
	}

	up.Settings = &ttnpb.TxSettings{
		Frequency: req.RadioMetaData.Frequency,
		DataRate:  bandDR.Rate,
		Timestamp: timestamp,
		Time:      ttnpb.ProtoTime(tm),
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
	req.MHdr = (uint(payload.MHdr.MType) << 5) | uint(payload.MHdr.Major)
	req.MIC = int32(binary.LittleEndian.Uint32(payload.Mic[:]))
	jreqPayload := payload.GetJoinRequestPayload()
	if jreqPayload == nil {
		return errUplinkMessage.New()
	}

	req.DevEUI = basicstation.EUI{
		EUI64: types.MustEUI64(jreqPayload.DevEui).OrZero(),
	}

	req.JoinEUI = basicstation.EUI{
		EUI64: types.MustEUI64(jreqPayload.JoinEui).OrZero(),
	}

	devNonce, err := types.MustDevNonce(jreqPayload.DevNonce).OrZero().Marshal()
	if err != nil {
		return err
	}
	req.DevNonce = uint(binary.BigEndian.Uint16(devNonce[:]))

	phy, err := band.GetLatest(bandID)
	if err != nil {
		return err
	}
	drIdx, _, ok := phy.FindUplinkDataRate(up.Settings.GetDataRate())
	if !ok {
		return errDataRate.New()
	}

	rxMetadata := up.RxMetadata[0]
	rxTime, gpsTime := TimePtrToUpInfoTime(ttnpb.StdTime(rxMetadata.Time))
	req.RadioMetaData = RadioMetaData{
		DataRate:  int(drIdx),
		Frequency: up.Settings.GetFrequency(),
		UpInfo: UpInfo{
			RCtx:    int64(rxMetadata.AntennaIndex),
			XTime:   int64(rxMetadata.Timestamp),
			RSSI:    rxMetadata.Rssi,
			SNR:     rxMetadata.Snr,
			RxTime:  rxTime,
			GPSTime: gpsTime,
		},
	}
	return nil
}

// toUplinkMessage extracts fields from the LoRa Basics Station Uplink Data Frame "updf" message and converts them into an UplinkMessage for the network server.
func (updf *UplinkDataFrame) toUplinkMessage(ids *ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	var up ttnpb.UplinkMessage
	up.ReceivedAt = ttnpb.ProtoTimePtr(receivedAt)

	parsedMHDR := &ttnpb.MHDR{}
	if err := lorawan.UnmarshalMHDR([]byte{byte(updf.MHdr)}, parsedMHDR); err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}
	if parsedMHDR.MType != ttnpb.MType_UNCONFIRMED_UP && parsedMHDR.MType != ttnpb.MType_CONFIRMED_UP {
		return nil, errMDHR.WithAttributes(`mhdr`, parsedMHDR)
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

	fctrl := &ttnpb.FCtrl{}
	if err := lorawan.UnmarshalFCtrl([]byte{byte(updf.FCtrl)}, fctrl, true); err != nil {
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
		MHdr: parsedMHDR,
		Mic:  micBytes,
		Payload: &ttnpb.Message_MacPayload{MacPayload: &ttnpb.MACPayload{
			FPort:      fPort,
			FrmPayload: decFRMPayload,
			FHdr: &ttnpb.FHDR{
				DevAddr: devAddr.Bytes(),
				FCtrl:   fctrl,
				FCnt:    uint32(updf.FCnt),
				FOpts:   decFOpts,
			},
		}},
	}

	up.RawPayload, err = lorawan.MarshalMessage(up.Payload)
	if err != nil {
		return nil, errUplinkDataFrame.WithCause(err)
	}

	timestamp := TimestampFromXTime(updf.RadioMetaData.UpInfo.XTime)
	gpsTime := TimePtrFromGPSTime(updf.UpInfo.GPSTime)
	tm := TimePtrFromUpInfo(updf.UpInfo.GPSTime, updf.UpInfo.RxTime)
	up.RxMetadata = []*ttnpb.RxMetadata{
		{
			GatewayIds:   ids,
			Time:         ttnpb.ProtoTime(tm),
			GpsTime:      ttnpb.ProtoTime(gpsTime),
			Timestamp:    timestamp,
			Rssi:         updf.RadioMetaData.UpInfo.RSSI,
			ChannelRssi:  updf.RadioMetaData.UpInfo.RSSI,
			Snr:          updf.RadioMetaData.UpInfo.SNR,
			AntennaIndex: uint32(updf.RadioMetaData.UpInfo.RCtx),
		},
	}

	phy, err := band.GetLatest(bandID)
	if err != nil {
		return nil, err
	}
	bandDR, ok := phy.DataRates[ttnpb.DataRateIndex(updf.RadioMetaData.DataRate)]
	if !ok {
		return nil, errDataRate.New()
	}

	up.Settings = &ttnpb.TxSettings{
		Frequency: updf.RadioMetaData.Frequency,
		DataRate:  bandDR.Rate,
		Timestamp: timestamp,
		Time:      ttnpb.ProtoTime(tm),
	}
	return &up, nil
}

func getFCtrlAsUint(fCtrl *ttnpb.FCtrl) uint {
	var ret uint
	if fCtrl.GetAdr() {
		ret = ret | 0x80
	}
	if fCtrl.GetAdrAckReq() {
		ret = ret | 0x40
	}
	if fCtrl.GetAck() {
		ret = ret | 0x20
	}
	if fCtrl.GetFPending() || fCtrl.GetClassB() {
		ret = ret | 0x10
	}
	return ret
}

// FromUplinkMessage extracts fields from ttnpb.UplinkMessage and creates the LoRa Basics Station UplinkDataFrame.
func (updf *UplinkDataFrame) FromUplinkMessage(up *ttnpb.UplinkMessage, bandID string) error {
	var payload ttnpb.Message
	err := lorawan.UnmarshalMessage(up.RawPayload, &payload)
	if err != nil {
		return errUplinkMessage.New()
	}
	updf.MHdr = (uint(payload.MHdr.MType) << 5) | uint(payload.MHdr.Major)

	macPayload := payload.GetMacPayload()
	if macPayload == nil {
		return errUplinkMessage.New()
	}

	updf.FPort = int(macPayload.GetFPort())

	updf.DevAddr = int32(types.MustDevAddr(macPayload.FHdr.DevAddr).OrZero().MarshalNumber())
	updf.FOpts = hex.EncodeToString(macPayload.FHdr.FOpts)

	updf.FCtrl = getFCtrlAsUint(macPayload.FHdr.FCtrl)
	updf.FCnt = uint(macPayload.FHdr.FCnt)
	updf.FRMPayload = hex.EncodeToString(macPayload.GetFrmPayload())
	updf.MIC = int32(binary.LittleEndian.Uint32(payload.Mic[:]))

	phy, err := band.GetLatest(bandID)
	if err != nil {
		return err
	}
	drIdx, _, ok := phy.FindUplinkDataRate(up.Settings.GetDataRate())
	if !ok {
		return errDataRate.New()
	}

	rxMetadata := up.RxMetadata[0]
	rxTime, gpsTime := TimePtrToUpInfoTime(ttnpb.StdTime(rxMetadata.Time))

	updf.RadioMetaData = RadioMetaData{
		DataRate:  int(drIdx),
		Frequency: up.Settings.GetFrequency(),
		UpInfo: UpInfo{
			RCtx:    int64(rxMetadata.AntennaIndex),
			XTime:   int64(rxMetadata.Timestamp),
			RSSI:    rxMetadata.Rssi,
			SNR:     rxMetadata.Snr,
			RxTime:  rxTime,
			GPSTime: gpsTime,
		},
	}
	return nil
}

// ToTxAck converts the LoRa Basics Station TxConfirmation message to ttnpb.TxAcknowledgment
func (conf TxConfirmation) ToTxAck(ctx context.Context, tokens io.DownlinkTokens, receivedAt time.Time) *ttnpb.TxAcknowledgment {
	var txAck ttnpb.TxAcknowledgment
	if msg, _, ok := tokens.Get(uint16(conf.Diid), receivedAt); ok && msg != nil {
		txAck.DownlinkMessage = msg
		txAck.CorrelationIds = msg.CorrelationIds
		txAck.Result = ttnpb.TxAcknowledgment_SUCCESS
	} else {
		logger := log.FromContext(ctx)
		logger.WithField("diid", conf.Diid).Debug("Tx acknowledgment either does not correspond to a downlink message or arrived too late")
	}
	return &txAck
}

// HandleUp implements Formatter.
func (f *lbsLNS) HandleUp(ctx context.Context, raw []byte, ids *ttnpb.GatewayIdentifiers, conn *io.Connection, receivedAt time.Time) ([]byte, error) {
	logger := log.FromContext(ctx)
	typ, err := Type(raw)
	if err != nil {
		logger.WithError(err).Debug("Failed to parse message type")
		return nil, err
	}
	logger = logger.WithFields(log.Fields(
		"upstream_type", typ,
	))

	recordRTT := func(refTimeUnix float64) {
		if refTimeUnix == 0.0 {
			return
		}
		refTime := TimeFromUnixSeconds(refTimeUnix)
		if delta := receivedAt.Sub(refTime); delta > f.maxRoundTripDelay {
			logger.WithFields(log.Fields(
				"delta", delta,
				"ref_time_unix", refTimeUnix,
				"ref_time", refTime,
				"received_at", receivedAt,
			)).Warn("Gateway reported RefTime greater than the valid maximum. Skip RTT measurement")
		} else {
			conn.RecordRTT(delta, receivedAt)
		}
	}
	syncClock := func(xTime int64, gpsTime int64, rxTime float64, onlyWithGPS bool) *io.FrontendClockSynchronization {
		if onlyWithGPS && gpsTime == 0 {
			return nil
		}
		return &io.FrontendClockSynchronization{
			Timestamp:  TimestampFromXTime(xTime),
			ServerTime: receivedAt,
			// RxTime is an undocumented field with unspecified precision.
			// Using 0.0 for RxTime here means that the GatewayTime is nil and is not used for syncing the gateway clock.
			GatewayTime:      TimePtrFromUpInfo(gpsTime, 0.0),
			ConcentratorTime: ConcentratorTimeFromXTime(xTime),
		}
	}
	recordTime := func(refTimeUnix float64, xTime int64, gpsTime int64, rxTime float64) *io.FrontendClockSynchronization {
		recordRTT(refTimeUnix)
		return syncClock(xTime, gpsTime, rxTime, false)
	}

	switch typ {
	case TypeUpstreamVersion:
		var antennaGain int
		antennas := conn.Gateway().Antennas
		if len(antennas) != 0 && antennas[0] != nil {
			// TODO: Support downlink path to multiple antennas (https://github.com/TheThingsNetwork/lorawan-stack/issues/48).
			// Need to set different gain value to each `SX1301_conf` object as per https://doc.sm.tc/station/gw_v1.5.html#multi-board-sample-configuration.
			// Currently we support downlink to only one antenna so the gain of the first antenna is applied to all (though the latter ones are not used).
			// FPs and Antennas need to be synchronized. See https://github.com/TheThingsNetwork/lorawan-stack/issues/48#issuecomment-983412639.
			antennaGain = int(antennas[0].Gain)
		}
		ctx, msg, stat, err := f.GetRouterConfig(ctx, raw, conn.BandID(), conn.FrequencyPlans(), antennaGain, receivedAt)
		if err != nil {
			logger.WithError(err).Warn("Failed to generate router configuration")
			return nil, err
		}
		logger = log.FromContext(ctx)
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
		updateSessionID(ctx, SessionIDFromXTime(jreq.UpInfo.XTime))
		ct := recordTime(jreq.RefTime, jreq.UpInfo.XTime, jreq.UpInfo.GPSTime, jreq.UpInfo.RxTime)
		if err := conn.HandleUp(up, ct); err != nil {
			logger.WithError(err).Warn("Failed to handle upstream message")
		}

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
		updateSessionID(ctx, SessionIDFromXTime(updf.UpInfo.XTime))
		ct := recordTime(updf.RefTime, updf.UpInfo.XTime, updf.UpInfo.GPSTime, updf.UpInfo.RxTime)
		if err := conn.HandleUp(up, ct); err != nil {
			logger.WithError(err).Warn("Failed to handle upstream message")
		}

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
		updateSessionID(ctx, SessionIDFromXTime(txConf.XTime))
		// Transmission confirmation messages do not contain a RefTime, and cannot be used for
		// RTT computations. The GPS timestamp is present only if the downlink is a class
		// B downlink. We allow clock synchronization to occur only if GPSTime is present.
		// References https://github.com/lorabasics/basicstation/issues/134.
		syncClock(txConf.XTime, txConf.GPSTime, 0.0, true)

	case TypeUpstreamTimeSync:
		// If the gateway sends a `timesync` request, it means that it has access to a PPS
		// source. As such, there is no point in doing time transfers with this particular
		// gateway.
		updateSessionTimeSync(ctx, false)
		var req TimeSyncRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, err
		}
		return req.Response(receivedAt).MarshalJSON()

	case TypeUpstreamProprietaryDataFrame, TypeUpstreamRemoteShell:
		logger.WithField("message_type", typ).Debug("Message type not implemented")

	default:
		logger.WithField("message_type", typ).Debug("Unknown message type")

	}
	return nil, nil
}
