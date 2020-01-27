// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"context"
	"encoding/json"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	packetbroker "go.packetbroker.org/api/v1"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var regionBands = map[packetbroker.Region]string{
	packetbroker.Region_EU_863_870: band.EU_863_870,
	packetbroker.Region_US_902_928: band.US_902_928,
	packetbroker.Region_CN_779_787: band.CN_779_787,
	packetbroker.Region_EU_433:     band.EU_433,
	packetbroker.Region_AU_915_928: band.AU_915_928,
	packetbroker.Region_CN_470_510: band.CN_470_510,
	packetbroker.Region_AS_923:     band.AS_923,
	packetbroker.Region_KR_920_923: band.KR_920_923,
	packetbroker.Region_IN_865_867: band.IN_865_867,
	packetbroker.Region_RU_864_870: band.RU_864_870,
}

func fromPBDataRate(region packetbroker.Region, index int) (ttnpb.DataRate, bool) {
	bandID, ok := regionBands[region]
	if !ok {
		return ttnpb.DataRate{}, false
	}
	phy, err := band.GetByID(bandID)
	if err != nil {
		return ttnpb.DataRate{}, false
	}
	if index >= len(phy.DataRates) {
		return ttnpb.DataRate{}, false
	}
	return phy.DataRates[index].Rate, true
}

func fromPBLocation(loc *packetbroker.Location) *ttnpb.Location {
	if loc == nil {
		return nil
	}
	return &ttnpb.Location{
		Longitude: loc.Longitude,
		Latitude:  loc.Latitude,
		Altitude:  int32(loc.Altitude),
		Accuracy:  int32(loc.Accuracy),
	}
}

type compoundUplinkToken struct {
	Forwarder []byte `json:"f,omitempty"`
	Gateway   []byte `json:"g,omitempty"`
}

func wrapUplinkTokens(forwarder, gateway []byte) ([]byte, error) {
	if forwarder == nil || gateway == nil {
		return nil, nil
	}
	return json.Marshal(compoundUplinkToken{forwarder, gateway})
}

func unwrapUplinkTokens(token []byte) (forwarder, gateway []byte, err error) {
	var t compoundUplinkToken
	if json.Unmarshal(token, &t); err != nil {
		return nil, nil, err
	}
	return t.Forwarder, t.Gateway, nil
}

var (
	errDataRate         = errors.DefineFailedPrecondition("data_rate", "invalid data rate index `{index}` in region `{region}`")
	errWrapUplinkTokens = errors.DefineAborted("wrap_uplink_tokens", "wrap uplink tokens")
)

func fromPBUplink(ctx context.Context, msg *packetbroker.UplinkMessage, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	dataRate, ok := fromPBDataRate(msg.GatewayRegion, int(msg.DataRateIndex))
	if !ok {
		return nil, errDataRate.WithAttributes(
			"index", msg.DataRateIndex,
			"region", msg.GatewayRegion,
		)
	}

	uplinkToken, err := wrapUplinkTokens(msg.ForwarderUplinkToken, msg.GatewayUplinkToken)
	if err != nil {
		return nil, errWrapUplinkTokens.WithCause(err)
	}

	up := &ttnpb.UplinkMessage{
		RawPayload: msg.PhyPayload.GetPlain(),
		Settings: ttnpb.TxSettings{
			DataRate:      dataRate,
			DataRateIndex: ttnpb.DataRateIndex(msg.DataRateIndex),
			Frequency:     msg.Frequency,
		},
		ReceivedAt:     receivedAt,
		CorrelationIDs: events.CorrelationIDsFromContext(ctx),
	}

	var receiveTime *time.Time
	if t, err := pbtypes.TimestampFromProto(msg.GatewayReceiveTime); err == nil {
		receiveTime = &t
	}
	if gtwMd := msg.GatewayMetadata; gtwMd != nil {
		if md := gtwMd.GetPlainLocalization().GetTerrestrial(); md != nil {
			for _, ant := range md.Antennas {
				up.RxMetadata = append(up.RxMetadata, &ttnpb.RxMetadata{
					GatewayIdentifiers:    cluster.PacketBrokerGatewayID,
					AntennaIndex:          ant.Index,
					Time:                  receiveTime,
					FineTimestamp:         ant.FineTimestamp.GetValue(),
					RSSI:                  ant.SignalQuality.GetChannelRssi(),
					SignalRSSI:            ant.SignalQuality.GetSignalRssi(),
					RSSIStandardDeviation: ant.SignalQuality.GetRssiStandardDeviation().GetValue(),
					SNR:                   ant.SignalQuality.GetSnr(),
					FrequencyOffset:       ant.SignalQuality.GetFrequencyOffset(),
					Location:              fromPBLocation(ant.Location),
					UplinkToken:           uplinkToken,
				})
			}
		} else if md := gtwMd.GetPlainSignalQuality().GetTerrestrial(); md != nil {
			for _, ant := range md.Antennas {
				up.RxMetadata = append(up.RxMetadata, &ttnpb.RxMetadata{
					GatewayIdentifiers:    cluster.PacketBrokerGatewayID,
					AntennaIndex:          ant.Index,
					Time:                  receiveTime,
					RSSI:                  ant.Value.GetChannelRssi(),
					SignalRSSI:            ant.Value.GetSignalRssi(),
					RSSIStandardDeviation: ant.Value.GetRssiStandardDeviation().GetValue(),
					SNR:                   ant.Value.GetSnr(),
					FrequencyOffset:       ant.Value.GetFrequencyOffset(),
					UplinkToken:           uplinkToken,
				})
			}
		}
	}

	return up, nil
}
