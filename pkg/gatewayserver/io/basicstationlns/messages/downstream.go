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

	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
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

// GetFromNSDownlinkMessage ...
func (dnmsg *DownlinkMessage) GetFromNSDownlinkMessage(ids ttnpb.GatewayIdentifiers, down ttnpb.DownlinkMessage) error {
	txReq := down.GetRequest()
	dnmsg.DevEUI = basicstation.EUI{Prefix: "DevEui", EUI64: *down.EndDeviceIDs.DevEUI}
	dnmsg.Pdu = string(down.GetRawPayload())
	dnmsg.DeviceClass = uint(txReq.Class)
	dnmsg.RxDelay = int(txReq.Rx1Delay)

	//TODO: Check if this is from Band or FP
	dnmsg.Rx1DR = int(txReq.Rx1DataRateIndex)
	dnmsg.Rx1Freq = int(txReq.Rx1Frequency)
	dnmsg.Rx2DR = int(txReq.Rx2DataRateIndex)
	dnmsg.Rx2Freq = int(txReq.Rx2Frequency)

	dnmsg.Priority = int(txReq.Priority)

	//TODO: Use GS Scheduler similar to the other frontends:
	for _, path := range txReq.DownlinkPaths {
		fixedPath := path.GetFixed()
		if fixedPath != nil {
			if fixedPath.GatewayID == ids.GatewayID {
				dnmsg.RCtx = int64(fixedPath.AntennaIndex)
				break
			}
			continue
		}
		if token := path.GetUplinkToken(); len(token) == 0 {
			antennaIDs, timestamp, err := io.ParseUplinkToken(token)
			if err != nil {
				return errDownlinkMessage
			}
			if antennaIDs.GatewayID == ids.GatewayID {
				dnmsg.RCtx = int64(antennaIDs.AntennaIndex)
				dnmsg.XTime = int64(timestamp)
				break
			}
		}
	}
	return nil
}
