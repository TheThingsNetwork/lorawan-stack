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
	"crypto/sha256"
	"encoding/json"
	"math"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	packetbroker "go.packetbroker.org/api/v1"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var fromPBRegion = map[packetbroker.Region]string{
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

var toPBRegion = map[string]packetbroker.Region{
	band.EU_863_870: packetbroker.Region_EU_863_870,
	band.US_902_928: packetbroker.Region_US_902_928,
	band.CN_779_787: packetbroker.Region_CN_779_787,
	band.EU_433:     packetbroker.Region_EU_433,
	band.AU_915_928: packetbroker.Region_AU_915_928,
	band.CN_470_510: packetbroker.Region_CN_470_510,
	band.AS_923:     packetbroker.Region_AS_923,
	band.KR_920_923: packetbroker.Region_KR_920_923,
	band.IN_865_867: packetbroker.Region_IN_865_867,
	band.RU_864_870: packetbroker.Region_RU_864_870,
}

func fromPBDataRate(region packetbroker.Region, index int) (ttnpb.DataRate, bool) {
	bandID, ok := fromPBRegion[region]
	if !ok {
		return ttnpb.DataRate{}, false
	}
	phy, err := band.GetByID(bandID)
	if err != nil {
		return ttnpb.DataRate{}, false
	}
	if index < 0 || index > math.MaxInt32 {
		// All protobuf enums are int32-typed, so ensure it does not overflow.
		return ttnpb.DataRate{}, false
	}
	dr, ok := phy.DataRates[ttnpb.DataRateIndex(index)]
	if !ok {
		return ttnpb.DataRate{}, false
	}
	return dr.Rate, true
}

func toPBDataRateIndex(region packetbroker.Region, dr ttnpb.DataRate) (uint32, bool) {
	bandID, ok := fromPBRegion[region]
	if !ok {
		return 0, false
	}
	phy, err := band.GetByID(bandID)
	if err != nil {
		return 0, false
	}
	for i, phyDR := range phy.DataRates {
		if phyDR.Rate.Equal(dr) {
			return uint32(i), true
		}
	}
	return 0, false
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

func toPBLocation(loc *ttnpb.Location) *packetbroker.Location {
	if loc == nil {
		return nil
	}
	return &packetbroker.Location{
		Longitude: loc.Longitude,
		Latitude:  loc.Latitude,
		Altitude:  float32(loc.Altitude),
		Accuracy:  float32(loc.Accuracy),
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
	if err := json.Unmarshal(token, &t); err != nil {
		return nil, nil, err
	}
	return t.Forwarder, t.Gateway, nil
}

var (
	errDecodePayload             = errors.DefineInvalidArgument("decode_payload", "decode LoRaWAN payload")
	errUnsupportedLoRaWANVersion = errors.DefineAborted("unsupported_lorawan_version", "unsupported LoRaWAN version `{version}`")
	errUnknownBand               = errors.DefineFailedPrecondition("unknown_band", "unknown band `{band_id}`")
	errUnknownDataRate           = errors.DefineFailedPrecondition("unknown_data_rate", "unknown data rate in region `{region}`")
	errUnsupportedMType          = errors.DefineAborted("unsupported_m_type", "unsupported LoRaWAN MType `{m_type}`")
)

func toPBUplink(ctx context.Context, msg *ttnpb.GatewayUplinkMessage) (*packetbroker.UplinkMessage, error) {
	msg.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(msg.RawPayload, msg.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}
	if msg.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"version", msg.Payload.Major,
		)
	}

	hash := sha256.Sum256(msg.RawPayload)
	up := &packetbroker.UplinkMessage{
		PhyPayload: &packetbroker.UplinkMessage_PHYPayload{
			Teaser: &packetbroker.PHYPayloadTeaser{
				Hash: hash[:],
			},
			Value: &packetbroker.UplinkMessage_PHYPayload_Plain{
				Plain: msg.RawPayload,
			},
		},
		Frequency: msg.Settings.Frequency,
	}

	var ok bool
	if up.GatewayRegion, ok = toPBRegion[msg.BandID]; !ok {
		return nil, errUnknownBand.WithAttributes("band_id", msg.BandID)
	}
	if up.DataRateIndex, ok = toPBDataRateIndex(up.GatewayRegion, msg.Settings.DataRate); !ok {
		return nil, errUnknownDataRate.WithAttributes("region", up.GatewayRegion)
	}

	switch pld := msg.Payload.Payload.(type) {
	case *ttnpb.Message_JoinRequestPayload:
		up.PhyPayload.Teaser.Payload = &packetbroker.PHYPayloadTeaser_JoinRequest{
			JoinRequest: &packetbroker.PHYPayloadTeaser_JoinRequestTeaser{
				JoinEui:  pld.JoinRequestPayload.JoinEUI.MarshalNumber(),
				DevEui:   pld.JoinRequestPayload.DevEUI.MarshalNumber(),
				DevNonce: uint32(pld.JoinRequestPayload.DevNonce.MarshalNumber()),
			},
		}
	case *ttnpb.Message_MACPayload:
		up.PhyPayload.Teaser.Payload = &packetbroker.PHYPayloadTeaser_Mac{
			Mac: &packetbroker.PHYPayloadTeaser_MACPayloadTeaser{
				Confirmed:        pld.MACPayload.Ack,
				DevAddr:          pld.MACPayload.DevAddr.MarshalNumber(),
				FOpts:            len(pld.MACPayload.FOpts) > 0,
				FCnt:             pld.MACPayload.FCnt,
				FPort:            pld.MACPayload.FPort,
				FrmPayloadLength: uint32(len(pld.MACPayload.FRMPayload)),
			},
		}
	default:
		return nil, errUnsupportedMType.WithAttributes("m_type", msg.Payload.MType)
	}

	var gatewayReceiveTime *time.Time
	var uplinkToken []byte
	if len(msg.RxMetadata) > 0 {
		var teaser packetbroker.GatewayMetadataTeaser_Terrestrial
		var signalQuality packetbroker.GatewayMetadataSignalQuality_Terrestrial
		var localization *packetbroker.GatewayMetadataLocalization_Terrestrial
		for _, md := range msg.RxMetadata {
			var rssiStandardDeviation *pbtypes.FloatValue
			if md.RSSIStandardDeviation > 0 {
				rssiStandardDeviation = &pbtypes.FloatValue{
					Value: md.RSSIStandardDeviation,
				}
			}

			sqAnt := &packetbroker.GatewayMetadataSignalQuality_Terrestrial_Antenna{
				Index: md.AntennaIndex,
				Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
					ChannelRssi:           md.ChannelRSSI,
					SignalRssi:            md.SignalRSSI,
					RssiStandardDeviation: rssiStandardDeviation,
					Snr:                   md.SNR,
					FrequencyOffset:       md.FrequencyOffset,
				},
			}
			signalQuality.Antennas = append(signalQuality.Antennas, sqAnt)

			if md.Location != nil {
				if localization == nil {
					localization = &packetbroker.GatewayMetadataLocalization_Terrestrial{}
				}
				locAnt := &packetbroker.GatewayMetadataLocalization_Terrestrial_Antenna{
					Index:         md.AntennaIndex,
					Location:      toPBLocation(md.Location),
					SignalQuality: sqAnt.Value,
				}
				if md.FineTimestamp > 0 {
					teaser.FineTimestamp = true
					locAnt.FineTimestamp = &pbtypes.UInt64Value{
						Value: md.FineTimestamp,
					}
				}
				localization.Antennas = append(localization.Antennas, locAnt)
			}

			if md.Time != nil {
				t := *md.Time
				if gatewayReceiveTime == nil || t.Before(*gatewayReceiveTime) {
					gatewayReceiveTime = &t
				}
			}
			if len(uplinkToken) == 0 {
				uplinkToken = md.UplinkToken
			}
		}

		up.GatewayMetadata = &packetbroker.UplinkMessage_GatewayMetadata{
			Teaser: &packetbroker.GatewayMetadataTeaser{
				Value: &packetbroker.GatewayMetadataTeaser_Terrestrial_{
					Terrestrial: &teaser,
				},
			},
			SignalQuality: &packetbroker.UplinkMessage_GatewayMetadata_PlainSignalQuality{
				PlainSignalQuality: &packetbroker.GatewayMetadataSignalQuality{
					Value: &packetbroker.GatewayMetadataSignalQuality_Terrestrial_{
						Terrestrial: &signalQuality,
					},
				},
			},
		}
		if localization != nil {
			up.GatewayMetadata.Localization = &packetbroker.UplinkMessage_GatewayMetadata_PlainLocalization{
				PlainLocalization: &packetbroker.GatewayMetadataLocalization{
					Value: &packetbroker.GatewayMetadataLocalization_Terrestrial_{
						Terrestrial: localization,
					},
				},
			}
		}
	}

	if t, err := pbtypes.TimestampProto(msg.ReceivedAt); err == nil {
		up.ForwarderReceiveTime = t
	}
	if gatewayReceiveTime != nil {
		if t, err := pbtypes.TimestampProto(*gatewayReceiveTime); err == nil {
			up.GatewayReceiveTime = t
		}
	}
	// TODO: Set uplink token when NsPba is implemented.
	// up.ForwarderUplinkToken = uplinkToken

	return up, nil
}

var errWrapUplinkTokens = errors.DefineAborted("wrap_uplink_tokens", "wrap uplink tokens")

func fromPBUplink(ctx context.Context, msg *packetbroker.UplinkMessage, receivedAt time.Time) (*ttnpb.UplinkMessage, error) {
	dataRate, ok := fromPBDataRate(msg.GatewayRegion, int(msg.DataRateIndex))
	if !ok {
		return nil, errUnknownDataRate.WithAttributes(
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
