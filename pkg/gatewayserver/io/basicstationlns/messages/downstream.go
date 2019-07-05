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
	"encoding/hex"
	"encoding/json"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errDownlinkMessage = errors.Define("downlink_message", "could not translate downlink message")

// DownlinkMessage is the LoRaWAN downlink message sent to the basic station.
type DownlinkMessage struct {
	DevEUI      string  `json:"DevEui"`
	DeviceClass uint    `json:"dC"`
	Diid        int64   `json:"diid"`
	Pdu         string  `json:"pdu"`
	RxDelay     int     `json:"RxDelay"`
	Rx1DR       int     `json:"RX1DR"`
	Rx1Freq     int     `json:"RX1Freq"`
	Priority    int     `json:"priority"`
	XTime       int64   `json:"xtime"`
	GpsTime     int64   `json:"gpstime"`
	RCtx        int64   `json:"rctx"`
	MuxTime     float64 `json:"MuxTime"`
}

// MarshalJSON implements json.Marshaler.
func (dnmsg DownlinkMessage) MarshalJSON() ([]byte, error) {
	type Alias DownlinkMessage
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeDownstreamDownlinkMessage,
		Alias: Alias(dnmsg),
	})
}

// FromDownlinkMessage translates the ttnpb.DownlinkMessage to LNS DownlinkMessage "dnmsg".
func FromDownlinkMessage(ids ttnpb.GatewayIdentifiers, rawPayload []byte, scheduledMsg *ttnpb.TxSettings, dlToken int64, dlTime time.Time, xTime int64) DownlinkMessage {
	var dnmsg DownlinkMessage
	dnmsg.Pdu = hex.EncodeToString(rawPayload)
	dnmsg.RCtx = int64(scheduledMsg.Downlink.AntennaIndex)
	dnmsg.Diid = dlToken

	// This field is not used but needs to be defined for the station to parse the json.
	dnmsg.DevEUI = "00-00-00-00-00-00-00-00"

	// Chosen fixed values.
	dnmsg.Priority = 25
	dnmsg.RxDelay = 1

	// Fix the Tx Parameters since we don't use the gateway scheduler.
	dnmsg.Rx1DR = int(scheduledMsg.DataRateIndex)
	dnmsg.Rx1Freq = int(scheduledMsg.Frequency)

	// Add the MuxTime for RTT measurement
	dnmsg.MuxTime = float64(dlTime.UnixNano()) / float64(time.Second)

	// The GS controls the scheduling and hence for the gateway, its always Class A.
	dnmsg.DeviceClass = uint(ttnpb.CLASS_A)

	// Estimate the xtime based on the timestamp; xtime = timestamp - (rxdelay+1). The calculated offset is in microseconds.
	dnmsg.XTime = xTime - int64(dnmsg.RxDelay*(1e6))
	return dnmsg
}

// ToDownlinkMessage translates the LNS DownlinkMessage "dnmsg" to ttnpb.DownlinkMessage.
func (dnmsg *DownlinkMessage) ToDownlinkMessage() ttnpb.DownlinkMessage {
	// var dlMesg ttnpb.DownlinkMessage
	var absTime *time.Time
	if dnmsg.GpsTime != 0 {
		val := time.Unix(dnmsg.GpsTime, 0)
		absTime = &val
	}
	return ttnpb.DownlinkMessage{
		RawPayload: []byte(dnmsg.Pdu),
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				DataRateIndex: ttnpb.DataRateIndex(dnmsg.Rx1DR),
				Frequency:     uint64(dnmsg.Rx1Freq),
				Downlink: &ttnpb.TxSettings_Downlink{
					AntennaIndex: uint32(dnmsg.RCtx),
				},
				Timestamp: uint32(dnmsg.XTime),
				Time:      absTime,
			},
		},
	}
}
