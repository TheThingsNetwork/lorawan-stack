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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errDownlinkMessage = errors.Define("downlink_message", "could not translate downlink message")

// DownlinkMessage is the LoRaWAN downlink message sent to the basic station.
type DownlinkMessage struct {
	DevEUI      basicstation.EUI `json:"DevEui"`
	DeviceClass uint             `json:"dC"`
	Diid        int64            `json:"diid"`
	Pdu         string           `json:"pdu"`
	RxDelay     int              `json:"RxDelay"`
	Rx1DR       int              `json:"Rx1DR"`
	Rx1Freq     int              `json:"Rx1Freq"`
	Rx2DR       int              `json:"Rx2DR"`
	Rx2Freq     int              `json:"Rx2Freq"`
	Priority    int              `json:"priority"`
	XTime       int64            `json:"xtime"`
	RCtx        int64            `json:"rctx"`
}

// MarshalJSON implements json.Marshaler.
func (dnmsg DownlinkMessage) MarshalJSON() ([]byte, error) {
	type Alias DownlinkMessage
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamJoinRequest,
		Alias: Alias(dnmsg),
	})
}

// FromNSDownlinkMessage translates the ttnpb.DownlinkMessage to LNS DownlinkMessage "dnmsg".
func (dnmsg *DownlinkMessage) FromNSDownlinkMessage(ids ttnpb.GatewayIdentifiers, down ttnpb.DownlinkMessage, dlToken int64) error {
	scheduledMsg := down.GetScheduled()
	dnmsg.DevEUI = basicstation.EUI{Prefix: "DevEui", EUI64: *down.EndDeviceIDs.DevEUI}
	dnmsg.Pdu = string(down.GetRawPayload())

	// TODO: Use Class_B for absolute timebound scheduling
	if (scheduledMsg.RequestInfo.Class == ttnpb.CLASS_A) || (scheduledMsg.RequestInfo.Class == ttnpb.CLASS_C) {
		dnmsg.DeviceClass = uint(ttnpb.CLASS_A)
	} else {
		dnmsg.DeviceClass = uint(ttnpb.CLASS_B)
	}

	// TODO: Choose a Sane value
	dnmsg.Priority = int(0)

	// The gateway is made to think that it's RX2 even if the network chooses RX1.
	// This way the network will enforce the chosen window instead of the gateway trying it's internal RX2.
	dnmsg.Rx2DR = int(scheduledMsg.DataRateIndex)
	dnmsg.Rx2Freq = int(scheduledMsg.Frequency)

	dnmsg.RxDelay = 1

	var xtime int64
	if scheduledMsg.RequestInfo.RxWindow == 0 {
		xtime = time.Now().Add(time.Duration(-1) * time.Minute).Unix()
	} else {
		xtime = time.Now().Add(time.Duration(-2) * time.Minute).Unix()
	}
	dnmsg.XTime = xtime

	dnmsg.RCtx = int64(scheduledMsg.RequestInfo.AntennaIndex)
	dnmsg.Diid = dlToken

	return nil
}
