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
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TimestampDownlinkMessage encapsulates the information used for downlinks
// which are meant to be sent at fixed concentrator timestamps.
type TimestampDownlinkMessage struct {
	RxDelay int   `json:"RxDelay"`
	Rx1DR   int   `json:"RX1DR"`
	Rx1Freq int   `json:"RX1Freq"`
	XTime   int64 `json:"xtime"`
}

// AbsoluteTimeDownlinkMessage encapsulates the information used for downlinks
// which are meant to be sent at fixed absolute GPS times.
type AbsoluteTimeDownlinkMessage struct {
	DR      int   `json:"DR"`
	Freq    int   `json:"Freq"`
	GPSTime int64 `json:"gpstime"`
}

// DownlinkMessage is the LoRaWAN downlink message sent to the LoRa Basics Station.
type DownlinkMessage struct {
	DevEUI      string  `json:"DevEui"`
	DeviceClass uint    `json:"dC"`
	Diid        int64   `json:"diid"`
	Pdu         string  `json:"pdu"`
	Priority    int     `json:"priority"`
	RCtx        int64   `json:"rctx"`
	MuxTime     float64 `json:"MuxTime"`

	*TimestampDownlinkMessage    `json:",omitempty"`
	*AbsoluteTimeDownlinkMessage `json:",omitempty"`
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
func (f *lbsLNS) FromDownlink(
	ctx context.Context, down *ttnpb.DownlinkMessage, bandID string, dlTime time.Time,
) ([]byte, error) {
	settings := down.GetScheduled()
	dnmsg := DownlinkMessage{
		DevEUI:   "00-00-00-00-00-00-00-01", // The DevEUI is required for transmission acknowledgements.
		Diid:     int64(f.tokens.Next(down, dlTime)),
		Pdu:      hex.EncodeToString(down.GetRawPayload()),
		Priority: 25,
		RCtx:     int64(settings.Downlink.AntennaIndex),
		MuxTime:  ws.TimeToUnixSeconds(dlTime),
	}

	phy, err := band.GetLatest(bandID)
	if err != nil {
		return nil, err
	}
	drIdx, _, ok := phy.FindDownlinkDataRate(settings.DataRate)
	if !ok {
		return nil, errDataRate.New()
	}

	if transmitAt := ttnpb.StdTime(settings.Time); transmitAt != nil {
		// Absolute time downlinks are scheduled as class B downlinks.
		dnmsg.DeviceClass = uint(ttnpb.Class_CLASS_B)
		dnmsg.AbsoluteTimeDownlinkMessage = &AbsoluteTimeDownlinkMessage{
			DR:      int(drIdx),
			Freq:    int(settings.Frequency),
			GPSTime: ws.TimeToGPSTime(*transmitAt),
		}
	} else {
		// The first 16 bits of XTime gets the session ID from the upstream
		// latest XTime and the other 48 bits are concentrator timestamp accounted for rollover.
		sessionID, found := ws.GetSessionID(ctx)
		if !found {
			return nil, errSessionStateNotFound.New()
		}

		xTime := ws.ConcentratorTimeToXTime(sessionID, scheduling.ConcentratorTime(settings.ConcentratorTimestamp))
		xTime = xTime - int64(time.Second/time.Microsecond) // Subtract a second, since the RX delay is 1.

		// Timestamp based downlinks are scheduled as class A downlinks.
		dnmsg.DeviceClass = uint(ttnpb.Class_CLASS_A)
		dnmsg.TimestampDownlinkMessage = &TimestampDownlinkMessage{
			RxDelay: 1,
			Rx1DR:   int(drIdx),
			Rx1Freq: int(settings.Frequency),
			XTime:   xTime,
		}
	}

	return dnmsg.marshalJSON()
}

// ToDownlinkMessage translates the LNS DownlinkMessage "dnmsg" to ttnpb.DownlinkMessage.
func (dnmsg *DownlinkMessage) ToDownlinkMessage(bandID string) (*ttnpb.DownlinkMessage, error) {
	phy, err := band.GetLatest(bandID)
	if err != nil {
		return nil, err
	}
	down := &ttnpb.DownlinkMessage{
		RawPayload: []byte(dnmsg.Pdu),
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				Downlink: &ttnpb.TxSettings_Downlink{
					AntennaIndex: uint32(dnmsg.RCtx),
				},
			},
		},
	}
	switch dnmsg.DeviceClass {
	case uint(ttnpb.Class_CLASS_A):
		bandDR, ok := phy.DataRates[ttnpb.DataRateIndex(dnmsg.Rx1DR)]
		if !ok {
			return nil, errDataRate.New()
		}
		down.GetScheduled().DataRate = bandDR.Rate
		down.GetScheduled().Frequency = uint64(dnmsg.Rx1Freq)
		down.GetScheduled().Timestamp = ws.TimestampFromXTime(dnmsg.XTime)
	case uint(ttnpb.Class_CLASS_B):
		bandDR, ok := phy.DataRates[ttnpb.DataRateIndex(dnmsg.DR)]
		if !ok {
			return nil, errDataRate.New()
		}
		down.GetScheduled().DataRate = bandDR.Rate
		down.GetScheduled().Frequency = uint64(dnmsg.Freq)
		down.GetScheduled().Time = timestamppb.New(ws.TimeFromGPSTime(dnmsg.GPSTime))
	default:
		panic("unreachable")
	}
	return down, nil
}

// TransferTime implements Formatter.
func (*lbsLNS) TransferTime(
	ctx context.Context, serverTime time.Time, gpsTime *time.Time, concentratorTime *scheduling.ConcentratorTime,
) ([]byte, error) {
	if enabled, ok := ws.GetSessionTimeSync(ctx); !ok || !enabled {
		return nil, nil
	}

	response := TimeSyncResponse{
		MuxTime: ws.TimeToUnixSeconds(serverTime),
	}

	sessionID, found := ws.GetSessionID(ctx)
	if !found || gpsTime == nil || concentratorTime == nil {
		// Update only the MuxTime.
		// https://github.com/lorabasics/basicstation/blob/bd17e53ab1137de6abb5ae48d6f3d52f6c268299/src/s2e.c#L1616-L1619
		return response.MarshalJSON()
	}

	response.XTime = ws.ConcentratorTimeToXTime(sessionID, *concentratorTime)
	response.GPSTime = ws.TimeToGPSTime(*gpsTime)

	return response.MarshalJSON()
}
