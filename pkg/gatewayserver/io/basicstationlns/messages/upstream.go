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
	"encoding/json"
	"time"

	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errJoinRequestMessage = errors.Define("join_request_message", "invalid join-request message received")

// DataRates encodes the available datarates of the channel plan for the Station in the format below
// [0] -> SF (Spreading Factor; Range: 7...12 for LoRa, 0 for FSK)
// [1] -> BW (Bandwidth; 125/250/500 for LoRa, ignored for FSK)
// [2] -> DNONLY (Downlink Only; 1 = true, 0 = false)
type DataRates [16][3]int

// UpInfo provides additional metadata on each upstream message.
type UpInfo struct {
	RxTime  int64   `json:"rxtime"`
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

// JoinRequest is the LoRaWAN Join Request message
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
// TODO: Make MarshalJSON() messages generic.
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

// ToUplinkMessage extracts fields from the basic station Join Request "jreq" message and converts them into an UplinkMessage for the network server.
func (req *JoinRequest) ToUplinkMessage(ids ttnpb.GatewayIdentifiers, bandID string) (ttnpb.UplinkMessage, error) {
	up := ttnpb.UplinkMessage{}
	up.ReceivedAt = time.Now()

	parsedMHDR := ttnpb.MHDR{}
	err := lorawan.UnmarshalMHDR([]byte{byte(req.MHdr)}, &parsedMHDR)
	if err != nil {
		return ttnpb.UplinkMessage{}, errJoinRequestMessage.WithCause(err)
	}

	micBytes, err := getInt32AsByteSlice(req.MIC)
	if err != nil {
		return ttnpb.UplinkMessage{}, errJoinRequestMessage.WithCause(err)
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
		return ttnpb.UplinkMessage{}, errJoinRequestMessage.WithCause(err)
	}

	rxTime := time.Unix(req.RadioMetaData.UpInfo.RxTime, 0)
	rxMetadata := &ttnpb.RxMetadata{
		GatewayIdentifiers: ids,
		Time:               &rxTime,
		Timestamp:          (uint32)(req.RadioMetaData.UpInfo.XTime & 0xFFFFFFFF),
		RSSI:               req.RadioMetaData.UpInfo.RSSI,
		SNR:                req.RadioMetaData.UpInfo.SNR,
	}
	up.RxMetadata = append(up.RxMetadata, rxMetadata)

	loraDR, err := getDataRateFromIndex(bandID, req.RadioMetaData.DataRate)
	if err != nil {
		return ttnpb.UplinkMessage{}, errJoinRequestMessage.WithCause(err)
	}
	up.Settings = ttnpb.TxSettings{
		Frequency:     req.RadioMetaData.Frequency,
		DataRateIndex: (ttnpb.DataRateIndex)(req.RadioMetaData.DataRate),
		DataRate:      loraDR,
	}
	return up, nil
}
