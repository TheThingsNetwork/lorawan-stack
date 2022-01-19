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
	"encoding/hex"
	"encoding/json"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// DownlinkMessage is the LoRaWAN downlink message sent to the LoRa Basics Station.
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
	RCtx        int64   `json:"rctx"`
	MuxTime     float64 `json:"MuxTime"`
}

// marshalJSON marshals dnmsg to a JSON byte array.
func (dnmsg DownlinkMessage) marshalJSON() ([]byte, error) {
	type Alias DownlinkMessage
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeDownstreamDownlinkMessage,
		Alias: Alias(dnmsg),
	})
}

// unmarshalJSON unmarshals dnmsg from a JSON byte array.
func (dnmsg *DownlinkMessage) unmarshalJSON(data []byte) error {
	return json.Unmarshal(data, dnmsg)
}

// FromDownlink implements Formatter.
func (f *lbsLNS) FromDownlink(ctx context.Context, down ttnpb.DownlinkMessage, bandID string, concentratorTime scheduling.ConcentratorTime, dlTime time.Time) ([]byte, error) {
	var dnmsg DownlinkMessage
	settings := down.GetScheduled()
	dnmsg.Pdu = hex.EncodeToString(down.GetRawPayload())
	dnmsg.RCtx = int64(settings.Downlink.AntennaIndex)
	dnmsg.Diid = int64(f.tokens.Next(&down, dlTime))

	// Chosen fixed values.
	dnmsg.Priority = 25
	dnmsg.RxDelay = 1

	// The first 16 bits of XTime gets the session ID from the upstream latestXTime and the other 48 bits are concentrator timestamp accounted for rollover.
	sessionID, found := getSessionID(ctx)
	if !found {
		return nil, errSessionStateNotFound.New()
	}
	xTime := ConcentratorTimeToXTime(sessionID, concentratorTime)

	// Estimate the xtime based on the timestamp; xtime = timestamp - (rxdelay). The calculated offset is in microseconds.
	dnmsg.XTime = xTime - int64(dnmsg.RxDelay*int(time.Second/time.Microsecond))

	log.FromContext(ctx).WithFields(log.Fields(
		"xtime", dnmsg.XTime,
		"mux_time", dlTime,
	)).Info("Prepare downlink message")

	// This field is not used but needs to be defined for the station to parse the json and should be non-zero for the station to return tx confirmations.
	dnmsg.DevEUI = "00-00-00-00-00-00-00-01"

	// Fix the Tx Parameters since we don't use the gateway scheduler.
	phy, err := band.GetLatest(bandID)
	if err != nil {
		return nil, err
	}
	drIdx, _, ok := phy.FindDownlinkDataRate(settings.DataRate)
	if !ok {
		return nil, errDataRate.New()
	}
	dnmsg.Rx1DR = int(drIdx)
	dnmsg.Rx1Freq = int(settings.Frequency)

	// Add the MuxTime for RTT measurement
	dnmsg.MuxTime = TimeToUnixSeconds(dlTime)

	// The GS controls the scheduling and hence for the gateway, its always Class A.
	dnmsg.DeviceClass = uint(ttnpb.Class_CLASS_A)

	return dnmsg.marshalJSON()
}

// ToDownlinkMessage translates the LNS DownlinkMessage "dnmsg" to ttnpb.DownlinkMessage.
func (dnmsg *DownlinkMessage) ToDownlinkMessage(bandID string) (*ttnpb.DownlinkMessage, error) {
	phy, err := band.GetLatest(bandID)
	if err != nil {
		return nil, err
	}
	bandDR, ok := phy.DataRates[ttnpb.DataRateIndex(dnmsg.Rx1DR)]
	if !ok {
		return nil, errDataRate.New()
	}
	return &ttnpb.DownlinkMessage{
		RawPayload: []byte(dnmsg.Pdu),
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				DataRate:  bandDR.Rate,
				Frequency: uint64(dnmsg.Rx1Freq),
				Downlink: &ttnpb.TxSettings_Downlink{
					AntennaIndex: uint32(dnmsg.RCtx),
				},
				Timestamp: uint32(dnmsg.XTime),
			},
		},
	}, nil
}

const (
	// transferTimeMinRTTCount is the minimum number of observed round-trip times that are taken into account before using
	// their statistics to determine the gateway GPS time.
	transferTimeMinRTTCount = 5
)

// TransferTime implements Formatter.
func (*lbsLNS) TransferTime(ctx context.Context, serverTime time.Time, gpsTime *time.Time, concentratorTime *scheduling.ConcentratorTime) ([]byte, error) {
	if enabled, ok := getSessionTimeSync(ctx); !ok || !enabled {
		return nil, nil
	}

	response := TimeSyncResponse{
		MuxTime: TimeToUnixSeconds(serverTime),
	}

	sessionID, found := getSessionID(ctx)
	if !found || gpsTime == nil || concentratorTime == nil {
		// Update only the MuxTime.
		// https://github.com/lorabasics/basicstation/blob/bd17e53ab1137de6abb5ae48d6f3d52f6c268299/src/s2e.c#L1616-L1619
		return response.MarshalJSON()
	}

	response.XTime = ConcentratorTimeToXTime(sessionID, *concentratorTime)
	response.GPSTime = TimeToGPSTime(*gpsTime)

	return response.MarshalJSON()
}
