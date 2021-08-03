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
	"encoding/base64"
	"encoding/json"
	"math"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"gopkg.in/square/go-jose.v2"
)

var (
	fromPBRegion = map[packetbroker.Region]string{
		packetbroker.Region_EU_863_870: band.EU_863_870,
		packetbroker.Region_US_902_928: band.US_902_928,
		packetbroker.Region_CN_779_787: band.CN_779_787,
		packetbroker.Region_EU_433:     band.EU_433,
		packetbroker.Region_AU_915_928: band.AU_915_928,
		packetbroker.Region_CN_470_510: band.CN_470_510,
		packetbroker.Region_AS_923:     band.AS_923,
		packetbroker.Region_AS_923_1:   band.AS_923,
		packetbroker.Region_KR_920_923: band.KR_920_923,
		packetbroker.Region_IN_865_867: band.IN_865_867,
		packetbroker.Region_RU_864_870: band.RU_864_870,
		packetbroker.Region_WW_2G4:     band.ISM_2400,
	}
	toPBRegion = map[string]packetbroker.Region{
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
		band.ISM_2400:   packetbroker.Region_WW_2G4,
	}
)

var (
	fromPBRegionalParameters = map[packetbroker.RegionalParametersVersion]ttnpb.PHYVersion{
		packetbroker.RegionalParametersVersion_TS001_V1_0:     ttnpb.TS001_V1_0,
		packetbroker.RegionalParametersVersion_TS001_V1_0_1:   ttnpb.TS001_V1_0_1,
		packetbroker.RegionalParametersVersion_RP001_V1_0_2_A: ttnpb.RP001_V1_0_2,
		packetbroker.RegionalParametersVersion_RP001_V1_0_2_B: ttnpb.RP001_V1_0_2_REV_B,
		packetbroker.RegionalParametersVersion_RP001_V1_0_3_A: ttnpb.RP001_V1_0_3_REV_A,
		packetbroker.RegionalParametersVersion_RP001_V1_1_A:   ttnpb.RP001_V1_1_REV_A,
		packetbroker.RegionalParametersVersion_RP001_V1_1_B:   ttnpb.RP001_V1_1_REV_B,
		packetbroker.RegionalParametersVersion_RP002_V1_0_0:   ttnpb.RP002_V1_0_0,
		packetbroker.RegionalParametersVersion_RP002_V1_0_1:   ttnpb.RP002_V1_0_1,
		packetbroker.RegionalParametersVersion_RP002_V1_0_2:   ttnpb.RP002_V1_0_2,
		packetbroker.RegionalParametersVersion_RP002_V1_0_3:   ttnpb.RP002_V1_0_3,
	}
	toPBRegionalParameters = map[ttnpb.PHYVersion]packetbroker.RegionalParametersVersion{
		ttnpb.TS001_V1_0:         packetbroker.RegionalParametersVersion_TS001_V1_0,
		ttnpb.TS001_V1_0_1:       packetbroker.RegionalParametersVersion_TS001_V1_0_1,
		ttnpb.RP001_V1_0_2:       packetbroker.RegionalParametersVersion_RP001_V1_0_2_A,
		ttnpb.RP001_V1_0_2_REV_B: packetbroker.RegionalParametersVersion_RP001_V1_0_2_B,
		ttnpb.RP001_V1_0_3_REV_A: packetbroker.RegionalParametersVersion_RP001_V1_0_3_A,
		ttnpb.RP001_V1_1_REV_A:   packetbroker.RegionalParametersVersion_RP001_V1_1_A,
		ttnpb.RP001_V1_1_REV_B:   packetbroker.RegionalParametersVersion_RP001_V1_1_B,
		ttnpb.RP002_V1_0_0:       packetbroker.RegionalParametersVersion_RP002_V1_0_0,
		ttnpb.RP002_V1_0_1:       packetbroker.RegionalParametersVersion_RP002_V1_0_1,
		ttnpb.RP002_V1_0_2:       packetbroker.RegionalParametersVersion_RP002_V1_0_2,
		ttnpb.RP002_V1_0_3:       packetbroker.RegionalParametersVersion_RP002_V1_0_3,
	}
)

func fromPBDataRate(dataRate *packetbroker.DataRate) (dr ttnpb.DataRate, codingRate string, ok bool) {
	switch mod := dataRate.GetModulation().(type) {
	case *packetbroker.DataRate_Lora:
		// TODO: Set coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466)
		return ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Lora{
				Lora: &ttnpb.LoRaDataRate{
					SpreadingFactor: mod.Lora.SpreadingFactor,
					Bandwidth:       mod.Lora.Bandwidth,
				},
			},
		}, mod.Lora.CodingRate, true
	case *packetbroker.DataRate_Fsk:
		return ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Fsk{
				Fsk: &ttnpb.FSKDataRate{
					BitRate: mod.Fsk.BitsPerSecond,
				},
			},
		}, "", true
	// TODO: Support LR-FHSS (https://github.com/TheThingsNetwork/lorawan-stack/issues/3806)
	// TODO: Set coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466)
	// case *packetbroker.DataRate_Lrfhss:
	// 	return ttnpb.DataRate{
	// 		Modulation: &ttnpb.DataRate_Lrfhss{
	// 			Lrfhss: &ttnpb.LRFHSSDataRate{
	// 				ModulationType:        mod.Lrfhss.ModulationType,
	// 				OperatingChannelWidth: mod.Lrfhss.OperatingChannelWidth,
	//  			CodingRate:            mod.Lrfhss.CodingRate,
	// 			},
	// 		},
	// 	}, mod.Lrfhss.CodingRate, true
	default:
		return ttnpb.DataRate{}, "", false
	}
}

func toPBDataRate(dataRate ttnpb.DataRate, codingRate string) (*packetbroker.DataRate, bool) {
	switch mod := dataRate.GetModulation().(type) {
	case *ttnpb.DataRate_Lora:
		return &packetbroker.DataRate{
			Modulation: &packetbroker.DataRate_Lora{
				Lora: &packetbroker.LoRaDataRate{
					SpreadingFactor: mod.Lora.SpreadingFactor,
					Bandwidth:       mod.Lora.Bandwidth,
					// TODO: Consider getting coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466)
					CodingRate: codingRate,
				},
			},
		}, true
	case *ttnpb.DataRate_Fsk:
		return &packetbroker.DataRate{
			Modulation: &packetbroker.DataRate_Fsk{
				Fsk: &packetbroker.FSKDataRate{
					BitsPerSecond: mod.Fsk.BitRate,
				},
			},
		}, true
	// TODO: Support LR-FHSS (https://github.com/TheThingsNetwork/lorawan-stack/issues/3806)
	// TODO: Get coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466)
	// case *ttnpb.DataRate_Lrfhss:
	// 	return &packetbroker.DataRate{
	// 		Modulation: &packetbroker.DataRate_Lrfhss{
	// 			Lrfhss: &packetbroker.LRFHSSDataRate{
	// 				ModulationType:        mod.Lrfhss.ModulationType,
	// 				OperatingChannelWidth: mod.Lrfhss.OperatingChannelWidth,
	// 				CodingRate:            mod.Lrfhss.CodingRate,
	// 			},
	// 		},
	// 	}, true
	default:
		return nil, false
	}
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
		Altitude:  float64(loc.Altitude),
		Accuracy:  float32(loc.Accuracy),
	}
}

func toPBTerrestrialAntennaPlacement(p ttnpb.GatewayAntennaPlacement) packetbroker.TerrestrialAntennaPlacement {
	return packetbroker.TerrestrialAntennaPlacement(p)
}

type agentUplinkToken struct {
	ForwarderNetID     types.NetID `json:"fnid"`
	ForwarderTenantID  string      `json:"ftid,omitempty"`
	ForwarderClusterID string      `json:"fcid,omitempty"`
}

type compoundUplinkToken struct {
	Gateway   []byte            `json:"g,omitempty"`
	Forwarder []byte            `json:"f,omitempty"`
	Agent     *agentUplinkToken `json:"a,omitempty"`
}

func wrapUplinkTokens(gateway, forwarder []byte, agent *agentUplinkToken) ([]byte, error) {
	return json.Marshal(compoundUplinkToken{gateway, forwarder, agent})
}

func unwrapUplinkTokens(token []byte) (gateway, forwarder []byte, agent *agentUplinkToken, err error) {
	var t compoundUplinkToken
	if err := json.Unmarshal(token, &t); err != nil {
		return nil, nil, nil, err
	}
	return t.Gateway, t.Forwarder, t.Agent, nil
}

type gatewayUplinkToken struct {
	GatewayUID string `json:"uid"`
	Token      []byte `json:"t"`
}

func wrapGatewayUplinkToken(ctx context.Context, ids ttnpb.GatewayIdentifiers, ulToken []byte, encrypter jose.Encrypter) ([]byte, error) {
	plaintext, err := json.Marshal(gatewayUplinkToken{
		GatewayUID: unique.ID(ctx, ids),
		Token:      ulToken,
	})
	if err != nil {
		return nil, err
	}
	obj, err := encrypter.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	s, err := obj.CompactSerialize()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func unwrapGatewayUplinkToken(token, key []byte) (string, []byte, error) {
	obj, err := jose.ParseEncrypted(string(token))
	if err != nil {
		return "", nil, err
	}
	plaintext, err := obj.Decrypt(key)
	if err != nil {
		return "", nil, err
	}
	var t gatewayUplinkToken
	if err := json.Unmarshal(plaintext, &t); err != nil {
		return "", nil, err
	}
	return t.GatewayUID, t.Token, nil
}

type gatewayIdentifier interface {
	GetGatewayId() string
	GetEui() *types.EUI64
}

func toPBGatewayIdentifier(ids gatewayIdentifier, config ForwarderConfig) (res *packetbroker.GatewayIdentifier) {
	if config.IncludeGatewayEUI && ids.GetEui() != nil {
		res = &packetbroker.GatewayIdentifier{
			Eui: &pbtypes.UInt64Value{
				Value: ids.GetEui().MarshalNumber(),
			},
		}
	}
	if config.IncludeGatewayID {
		if res == nil {
			res = &packetbroker.GatewayIdentifier{}
		}
		if config.HashGatewayID {
			hash := sha256.Sum256([]byte(ids.GetGatewayId()))
			res.Id = &packetbroker.GatewayIdentifier_Hash{
				Hash: hash[:],
			}
		} else {
			res.Id = &packetbroker.GatewayIdentifier_Plain{
				Plain: ids.GetGatewayId(),
			}
		}
	}
	return
}

var (
	errDecodePayload             = errors.DefineInvalidArgument("decode_payload", "decode LoRaWAN payload")
	errUnsupportedLoRaWANVersion = errors.DefineAborted("unsupported_lorawan_version", "unsupported LoRaWAN version `{version}`")
	errUnknownBand               = errors.DefineFailedPrecondition("unknown_band", "unknown band `{band_id}`")
	errUnknownDataRate           = errors.DefineFailedPrecondition("unknown_data_rate", "unknown data rate")
	errUnknownDataRateIndex      = errors.DefineFailedPrecondition("unknown_data_rate_index", "unknown data rate `{index}` in region `{region}`")
	errUnsupportedMType          = errors.DefineAborted("unsupported_m_type", "unsupported LoRaWAN MType `{m_type}`")
	errWrapGatewayUplinkToken    = errors.DefineAborted("wrap_gateway_uplink_token", "wrap gateway uplink token")
)

func toPBUplink(ctx context.Context, msg *ttnpb.GatewayUplinkMessage, config ForwarderConfig) (*packetbroker.UplinkMessage, error) {
	msg.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(msg.RawPayload, msg.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}
	if msg.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"version", msg.Payload.Major,
		)
	}

	hash := sha256.Sum256(msg.RawPayload[:len(msg.RawPayload)-4]) // The hash is without MIC to detect retransmissions.
	up := &packetbroker.UplinkMessage{
		PhyPayload: &packetbroker.UplinkMessage_PHYPayload{
			Teaser: &packetbroker.PHYPayloadTeaser{
				Hash:   hash[:],
				Length: uint32(len(msg.RawPayload)),
			},
			Value: &packetbroker.UplinkMessage_PHYPayload_Plain{
				Plain: msg.RawPayload,
			},
		},
		Frequency:  msg.Settings.Frequency,
		CodingRate: msg.Settings.CodingRate,
	}

	var ok bool
	if up.GatewayRegion, ok = toPBRegion[msg.BandId]; !ok {
		return nil, errUnknownBand.WithAttributes("band_id", msg.BandId)
	}
	if up.DataRate, ok = toPBDataRate(msg.Settings.DataRate, msg.Settings.CodingRate); !ok {
		return nil, errUnknownDataRate.New()
	}

	// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478).
	phy, err := band.GetByID(msg.BandId)
	if err != nil {
		return nil, errUnknownBand.WithAttributes("band_id", msg.BandId)
	}
	if dataRateIndex, _, ok := phy.FindUplinkDataRate(msg.Settings.DataRate); ok {
		up.DataRateIndex = uint32(dataRateIndex)
	} else {
		return nil, errUnknownDataRate.New()
	}

	switch pld := msg.Payload.Payload.(type) {
	case *ttnpb.Message_JoinRequestPayload:
		up.PhyPayload.Teaser.Payload = &packetbroker.PHYPayloadTeaser_JoinRequest{
			JoinRequest: &packetbroker.PHYPayloadTeaser_JoinRequestTeaser{
				JoinEui:  pld.JoinRequestPayload.JoinEui.MarshalNumber(),
				DevEui:   pld.JoinRequestPayload.DevEui.MarshalNumber(),
				DevNonce: uint32(pld.JoinRequestPayload.DevNonce.MarshalNumber()),
			},
		}
	case *ttnpb.Message_MacPayload:
		up.PhyPayload.Teaser.Payload = &packetbroker.PHYPayloadTeaser_Mac{
			Mac: &packetbroker.PHYPayloadTeaser_MACPayloadTeaser{
				Confirmed:        pld.MacPayload.Ack,
				DevAddr:          pld.MacPayload.DevAddr.MarshalNumber(),
				FOpts:            len(pld.MacPayload.FOpts) > 0,
				FCnt:             pld.MacPayload.FCnt,
				FPort:            pld.MacPayload.FPort,
				FrmPayloadLength: uint32(len(pld.MacPayload.FrmPayload)),
			},
		}
	default:
		return nil, errUnsupportedMType.WithAttributes("m_type", msg.Payload.MType)
	}

	var gatewayReceiveTime *time.Time
	var gatewayUplinkToken []byte
	if len(msg.RxMetadata) > 0 {
		md := msg.RxMetadata[0]
		up.GatewayId = toPBGatewayIdentifier(&md.GatewayIdentifiers, config)

		var teaser packetbroker.GatewayMetadataTeaser_Terrestrial
		var signalQuality packetbroker.GatewayMetadataSignalQuality_Terrestrial
		var localization *packetbroker.GatewayMetadataLocalization_Terrestrial
		for _, md := range msg.RxMetadata {
			var rssiStandardDeviation *pbtypes.FloatValue
			if md.RssiStandardDeviation > 0 {
				rssiStandardDeviation = &pbtypes.FloatValue{
					Value: md.RssiStandardDeviation,
				}
			}

			sqAnt := &packetbroker.GatewayMetadataSignalQuality_Terrestrial_Antenna{
				Index: md.AntennaIndex,
				Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
					ChannelRssi:           md.ChannelRssi,
					SignalRssi:            md.SignalRssi,
					RssiStandardDeviation: rssiStandardDeviation,
					Snr:                   md.Snr,
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
			if len(gatewayUplinkToken) == 0 {
				var err error
				gatewayUplinkToken, err = wrapGatewayUplinkToken(ctx, md.GatewayIdentifiers, md.UplinkToken, config.TokenEncrypter)
				if err != nil {
					return nil, errWrapGatewayUplinkToken.WithCause(err)
				}
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
	up.GatewayUplinkToken = gatewayUplinkToken

	return up, nil
}

var errWrapUplinkTokens = errors.DefineAborted("wrap_uplink_tokens", "wrap uplink tokens")

func fromPBUplink(ctx context.Context, msg *packetbroker.RoutedUplinkMessage, receivedAt time.Time, includeHops bool) (*ttnpb.UplinkMessage, error) {
	// NOTE: The data rate index is set for informative purposes only. The band cannot be versioned because there is no
	// Regional Parameters version available in uplink messages.
	// TODO: Remove usage of data rate index (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478).
	bandID, ok := fromPBRegion[msg.Message.GatewayRegion]
	if !ok {
		return nil, errUnknownRegion.WithAttributes("region", msg.Message.GatewayRegion)
	}
	phy, err := band.GetByID(bandID)
	if err != nil {
		return nil, errUnknownBand.WithAttributes("band_id", bandID)
	}
	var (
		dataRate      ttnpb.DataRate
		dataRateIndex ttnpb.DataRateIndex
		codingRate    = msg.Message.CodingRate
	)
	if msg.Message.DataRate != nil {
		dataRate, codingRate, ok = fromPBDataRate(msg.Message.DataRate)
		if !ok {
			return nil, errUnknownDataRate.New()
		}
		dataRateIndex, _, ok = phy.FindUplinkDataRate(dataRate)
		if !ok {
			return nil, errUnknownDataRate.New()
		}
	} else {
		if msg.Message.DataRateIndex > math.MaxInt32 {
			return nil, errUnknownDataRate.New()
		}
		phyDR, ok := phy.DataRates[ttnpb.DataRateIndex(msg.Message.DataRateIndex)]
		if !ok {
			return nil, errUnknownDataRateIndex.WithAttributes(
				"index", msg.Message.DataRateIndex,
				"region", msg.Message.GatewayRegion,
			)
		}
		// TODO: Set coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466)
		dataRate, dataRateIndex = phyDR.Rate, ttnpb.DataRateIndex(msg.Message.DataRateIndex)
	}

	var forwarderNetID, homeNetworkNetID types.NetID
	if err := forwarderNetID.UnmarshalNumber(msg.ForwarderNetId); err != nil {
		return nil, errNetID.WithCause(err).WithAttributes("net_id", msg.ForwarderNetId)
	}
	if err := homeNetworkNetID.UnmarshalNumber(msg.HomeNetworkNetId); err != nil {
		return nil, errNetID.WithCause(err).WithAttributes("net_id", msg.HomeNetworkNetId)
	}
	var (
		downlinkPathConstraint = ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER
		uplinkToken            []byte
	)
	if len(msg.Message.GatewayUplinkToken) > 0 || len(msg.Message.ForwarderUplinkToken) > 0 {
		downlinkPathConstraint = ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE
		token := &agentUplinkToken{
			ForwarderNetID:     forwarderNetID,
			ForwarderTenantID:  msg.ForwarderTenantId,
			ForwarderClusterID: msg.ForwarderClusterId,
		}
		var err error
		uplinkToken, err = wrapUplinkTokens(msg.Message.GatewayUplinkToken, msg.Message.ForwarderUplinkToken, token)
		if err != nil {
			return nil, errWrapUplinkTokens.WithCause(err)
		}
	}

	up := &ttnpb.UplinkMessage{
		RawPayload: msg.Message.PhyPayload.GetPlain(),
		Settings: ttnpb.TxSettings{
			DataRate:      dataRate,
			DataRateIndex: dataRateIndex,
			Frequency:     msg.Message.Frequency,
			CodingRate:    codingRate,
		},
		ReceivedAt:     receivedAt,
		CorrelationIds: events.CorrelationIDsFromContext(ctx),
	}

	var receiveTime *time.Time
	if t, err := pbtypes.TimestampFromProto(msg.Message.GatewayReceiveTime); err == nil {
		receiveTime = &t
	}
	if gtwMd := msg.Message.GatewayMetadata; gtwMd != nil {
		pbMD := &ttnpb.PacketBrokerMetadata{
			MessageId:            msg.Id,
			ForwarderNetId:       forwarderNetID,
			ForwarderTenantId:    msg.ForwarderTenantId,
			ForwarderClusterId:   msg.ForwarderClusterId,
			HomeNetworkNetId:     homeNetworkNetID,
			HomeNetworkTenantId:  msg.HomeNetworkTenantId,
			HomeNetworkClusterId: msg.HomeNetworkClusterId,
		}
		if id := msg.GetMessage().GetGatewayId(); id != nil {
			if eui := id.Eui; eui != nil {
				pbMD.ForwarderGatewayEui = &types.EUI64{}
				pbMD.ForwarderGatewayEui.UnmarshalNumber(eui.Value)
			}
			switch s := id.Id.(type) {
			case *packetbroker.GatewayIdentifier_Hash:
				pbMD.ForwarderGatewayId = &pbtypes.StringValue{
					Value: base64.StdEncoding.EncodeToString(s.Hash),
				}
			case *packetbroker.GatewayIdentifier_Plain:
				pbMD.ForwarderGatewayId = &pbtypes.StringValue{
					Value: s.Plain,
				}
			}
		}
		if includeHops {
			pbMD.Hops = make([]*ttnpb.PacketBrokerRouteHop, 0, len(msg.Hops))
			for _, h := range msg.Hops {
				receivedAt, err := pbtypes.TimestampFromProto(h.ReceivedAt)
				if err != nil {
					continue
				}
				pbMD.Hops = append(pbMD.Hops, &ttnpb.PacketBrokerRouteHop{
					ReceivedAt:    receivedAt,
					SenderName:    h.SenderName,
					SenderAddress: h.SenderAddress,
					ReceiverName:  h.ReceiverName,
					ReceiverAgent: h.ReceiverAgent,
				})
			}
		}
		if md := gtwMd.GetPlainLocalization().GetTerrestrial(); md != nil {
			for _, ant := range md.Antennas {
				up.RxMetadata = append(up.RxMetadata, &ttnpb.RxMetadata{
					GatewayIdentifiers:     cluster.PacketBrokerGatewayID,
					PacketBroker:           pbMD,
					AntennaIndex:           ant.Index,
					Time:                   receiveTime,
					FineTimestamp:          ant.FineTimestamp.GetValue(),
					Rssi:                   ant.SignalQuality.GetChannelRssi(),
					ChannelRssi:            ant.SignalQuality.GetChannelRssi(),
					SignalRssi:             ant.SignalQuality.GetSignalRssi(),
					RssiStandardDeviation:  ant.SignalQuality.GetRssiStandardDeviation().GetValue(),
					Snr:                    ant.SignalQuality.GetSnr(),
					FrequencyOffset:        ant.SignalQuality.GetFrequencyOffset(),
					Location:               fromPBLocation(ant.Location),
					DownlinkPathConstraint: downlinkPathConstraint,
					UplinkToken:            uplinkToken,
				})
			}
		} else if md := gtwMd.GetPlainSignalQuality().GetTerrestrial(); md != nil {
			for _, ant := range md.Antennas {
				up.RxMetadata = append(up.RxMetadata, &ttnpb.RxMetadata{
					GatewayIdentifiers:     cluster.PacketBrokerGatewayID,
					PacketBroker:           pbMD,
					AntennaIndex:           ant.Index,
					Time:                   receiveTime,
					Rssi:                   ant.Value.GetChannelRssi(),
					ChannelRssi:            ant.Value.GetChannelRssi(),
					SignalRssi:             ant.Value.GetSignalRssi(),
					RssiStandardDeviation:  ant.Value.GetRssiStandardDeviation().GetValue(),
					Snr:                    ant.Value.GetSnr(),
					FrequencyOffset:        ant.Value.GetFrequencyOffset(),
					DownlinkPathConstraint: downlinkPathConstraint,
					UplinkToken:            uplinkToken,
				})
			}
		}
	}

	return up, nil
}

var (
	fromPBClass = map[packetbroker.DownlinkMessageClass]ttnpb.Class{
		packetbroker.DownlinkMessageClass_CLASS_A: ttnpb.CLASS_A,
		packetbroker.DownlinkMessageClass_CLASS_B: ttnpb.CLASS_B,
		packetbroker.DownlinkMessageClass_CLASS_C: ttnpb.CLASS_C,
	}
	toPBClass = map[ttnpb.Class]packetbroker.DownlinkMessageClass{
		ttnpb.CLASS_A: packetbroker.DownlinkMessageClass_CLASS_A,
		ttnpb.CLASS_B: packetbroker.DownlinkMessageClass_CLASS_B,
		ttnpb.CLASS_C: packetbroker.DownlinkMessageClass_CLASS_C,
	}
	fromPBPriority = map[packetbroker.DownlinkMessagePriority]ttnpb.TxSchedulePriority{
		packetbroker.DownlinkMessagePriority_LOWEST:  ttnpb.TxSchedulePriority_LOWEST,
		packetbroker.DownlinkMessagePriority_LOW:     ttnpb.TxSchedulePriority_LOW,
		packetbroker.DownlinkMessagePriority_NORMAL:  ttnpb.TxSchedulePriority_NORMAL,
		packetbroker.DownlinkMessagePriority_HIGH:    ttnpb.TxSchedulePriority_HIGH,
		packetbroker.DownlinkMessagePriority_HIGHEST: ttnpb.TxSchedulePriority_HIGHEST,
	}
	toPBPriority = map[ttnpb.TxSchedulePriority]packetbroker.DownlinkMessagePriority{
		ttnpb.TxSchedulePriority_LOWEST:       packetbroker.DownlinkMessagePriority_LOWEST,
		ttnpb.TxSchedulePriority_LOW:          packetbroker.DownlinkMessagePriority_LOW,
		ttnpb.TxSchedulePriority_BELOW_NORMAL: packetbroker.DownlinkMessagePriority_LOW,
		ttnpb.TxSchedulePriority_NORMAL:       packetbroker.DownlinkMessagePriority_NORMAL,
		ttnpb.TxSchedulePriority_ABOVE_NORMAL: packetbroker.DownlinkMessagePriority_HIGH,
		ttnpb.TxSchedulePriority_HIGH:         packetbroker.DownlinkMessagePriority_HIGH,
		ttnpb.TxSchedulePriority_HIGHEST:      packetbroker.DownlinkMessagePriority_HIGHEST,
	}
)

var (
	errNoRequest                  = errors.DefineFailedPrecondition("no_request", "downlink message is not a transmission request")
	errUnknownPHYVersion          = errors.DefineInvalidArgument("unknown_phy_version", "unknown LoRaWAN Regional Parameters version `{version}`")
	errUnknownClass               = errors.DefineInvalidArgument("unknown_class", "unknown class `{class}`")
	errUnknownPriority            = errors.DefineInvalidArgument("unknown_priority", "unknown priority `{priority}`")
	errNoDownlinkPaths            = errors.DefineFailedPrecondition("no_downlink_paths", "no downlink paths")
	errInvalidDownlinkPath        = errors.DefineFailedPrecondition("downlink_path", "invalid uplink token downlink path")
	errFrequencyPlanNotConfigured = errors.DefineInvalidArgument("frequency_plan_not_configured", "frequency plan `{id}` is not configured")
	errDataRateNotFound           = errors.DefineInvalidArgument("data_rate_not_found", "no data rate with index `{index}`")
	errIncompatibleDataRate       = errors.DefineInvalidArgument("incompatible_data_rate", "incompatible data rate in Rx{rx_window}")
)

func toPBDownlink(ctx context.Context, msg *ttnpb.DownlinkMessage, fps frequencyPlansStore) (*packetbroker.DownlinkMessage, *agentUplinkToken, error) {
	req := msg.GetRequest()
	if req == nil {
		return nil, nil, errNoRequest.New()
	}

	// TODO: Convert data rate and fill out new rx{1,2}_data_rate fields (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478).
	// That will lead to removing getting the frequency plan and the band and versioning the band.
	fp, err := fps.GetByID(req.FrequencyPlanId)
	if err != nil {
		return nil, nil, errFrequencyPlanNotConfigured.WithAttributes("id", req.FrequencyPlanId)
	}
	phy, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, nil, errUnknownBand.WithAttributes("band_id", fp.BandID)
	}
	phy, err = phy.Version(req.LorawanPhyVersion)
	if err != nil {
		return nil, nil, err
	}

	down := &packetbroker.DownlinkMessage{
		PhyPayload: msg.RawPayload,
		Rx1Delay:   pbtypes.DurationProto(req.Rx1Delay.Duration()),
	}
	var ok bool
	if down.Region, ok = toPBRegion[phy.ID]; !ok {
		return nil, nil, errUnknownBand.WithAttributes("band_id", phy.ID)
	}
	rpVersion, ok := toPBRegionalParameters[req.LorawanPhyVersion]
	if !ok {
		return nil, nil, errUnknownPHYVersion.WithAttributes("version", req.LorawanPhyVersion)
	}
	down.RegionalParametersVersion = &packetbroker.RegionalParametersVersionValue{
		Value: rpVersion,
	}
	for i, rx := range []struct {
		dataRateIndex ttnpb.DataRateIndex
		frequency     uint64
		dst           **packetbroker.DownlinkMessage_RXSettings
	}{
		{req.Rx1DataRateIndex, req.Rx1Frequency, &down.Rx1},
		{req.Rx2DataRateIndex, req.Rx2Frequency, &down.Rx2},
	} {
		if rx.frequency == 0 {
			continue
		}
		dr, ok := phy.DataRates[rx.dataRateIndex]
		if !ok {
			return nil, nil, errDataRateNotFound.WithAttributes("index", rx.dataRateIndex)
		}
		// TODO: Get coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466)
		var codingRate string
		switch dr.Rate.Modulation.(type) {
		case *ttnpb.DataRate_Lora:
			codingRate = phy.LoRaCodingRate
		}
		pbDR, ok := toPBDataRate(dr.Rate, codingRate)
		if !ok {
			return nil, nil, errIncompatibleDataRate.WithAttributes("rx_window", i+1)
		}
		*rx.dst = &packetbroker.DownlinkMessage_RXSettings{
			DataRate:      pbDR,
			DataRateIndex: uint32(rx.dataRateIndex),
			Frequency:     rx.frequency,
		}
	}
	if down.Class, ok = toPBClass[req.Class]; !ok {
		return nil, nil, errUnknownClass.WithAttributes("class", req.Class)
	}
	if down.Priority, ok = toPBPriority[req.Priority]; !ok {
		return nil, nil, errUnknownPriority.WithAttributes("priority", req.Priority)
	}
	if len(req.DownlinkPaths) == 0 {
		return nil, nil, errNoDownlinkPaths.New()
	}
	uplinkToken := req.DownlinkPaths[0].GetUplinkToken()
	if len(uplinkToken) == 0 {
		return nil, nil, errInvalidDownlinkPath.New()
	}
	var token *agentUplinkToken
	down.GatewayUplinkToken, down.ForwarderUplinkToken, token, err = unwrapUplinkTokens(uplinkToken)
	if err != nil {
		return nil, nil, errInvalidDownlinkPath.WithCause(err)
	}

	return down, token, nil
}

var (
	errUnwrapGatewayUplinkToken = errors.DefineAborted("unwrap_gateway_uplink_token", "unwrap gateway uplink token")
	errUnknownRegion            = errors.DefineFailedPrecondition("unknown_region", "unknown region `{region}`")
	errInvalidRx1Delay          = errors.DefineInvalidArgument("invalid_rx1_delay", "invalid Rx1 delay")
)

func fromPBDownlink(ctx context.Context, msg *packetbroker.DownlinkMessage, receivedAt time.Time, conf ForwarderConfig) (uid string, res *ttnpb.DownlinkMessage, err error) {
	uid, token, err := unwrapGatewayUplinkToken(msg.GatewayUplinkToken, conf.TokenKey)
	if err != nil {
		return "", nil, errUnwrapGatewayUplinkToken.WithCause(err)
	}

	req := &ttnpb.TxRequest{
		DownlinkPaths: []*ttnpb.DownlinkPath{
			{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: token,
				},
			},
		},
	}
	var (
		ok  bool
		phy *band.Band
	)
	// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478)
	if msg.Region != packetbroker.Region_UNKNOWN_REGION {
		var bandID string
		if bandID, ok = fromPBRegion[msg.Region]; !ok {
			return "", nil, errUnknownRegion.WithAttributes("region", msg.Region)
		}
		p, err := band.GetByID(bandID)
		if err != nil {
			return "", nil, errUnknownRegion.WithCause(err).WithAttributes("region", msg.Region)
		}
		phy = &p
		// NOTE: The Things Stack expects the frequency plan ID; not the band ID. Since the frequency plan ID cannot be
		// inferred from the downlink message from Packet Broker, it is intentionally left blank. This makes the Gateway
		// Server fallback to a single frequency plan configured for the gateway. This does not work if there are multiple
		// frequency plans. (https://github.com/TheThingsNetwork/lorawan-stack/issues/1394)
	}
	if msg.RegionalParametersVersion != nil {
		var phyVersion ttnpb.PHYVersion
		if phyVersion, ok = fromPBRegionalParameters[msg.RegionalParametersVersion.Value]; !ok {
			return "", nil, errUnknownPHYVersion.WithAttributes("version", msg.RegionalParametersVersion.Value)
		}
		req.LorawanPhyVersion = phyVersion
		// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478)
		if phy != nil {
			p, err := phy.Version(phyVersion)
			if err != nil {
				return "", nil, err
			}
			phy = &p
		}
	}
	if req.Class, ok = fromPBClass[msg.Class]; !ok {
		return "", nil, errUnknownClass.WithAttributes("class", msg.Class)
	}
	if req.Priority, ok = fromPBPriority[msg.Priority]; !ok {
		return "", nil, errUnknownPriority.WithAttributes("priority", msg.Priority)
	}
	rx1Delay, err := pbtypes.DurationFromProto(msg.Rx1Delay)
	if err != nil {
		return "", nil, errInvalidRx1Delay.WithCause(err)
	}
	req.Rx1Delay = ttnpb.RxDelay(rx1Delay / time.Second)
	for i, rx := range []struct {
		settings      *packetbroker.DownlinkMessage_RXSettings
		dataRateIndex *ttnpb.DataRateIndex
		frequency     *uint64
	}{
		{msg.Rx1, &req.Rx1DataRateIndex, &req.Rx1Frequency},
		{msg.Rx2, &req.Rx2DataRateIndex, &req.Rx2Frequency},
	} {
		if rx.settings == nil {
			continue
		}
		if rx.settings.DataRate != nil && phy != nil {
			dr, _, ok := fromPBDataRate(rx.settings.DataRate)
			if !ok {
				return "", nil, errIncompatibleDataRate.WithAttributes("rx_window", i+1)
			}
			// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478)
			if *rx.dataRateIndex, _, ok = phy.FindDownlinkDataRate(dr); !ok {
				return "", nil, errIncompatibleDataRate.WithAttributes("rx_window", i+1)
			}
		} else {
			// NOTE: This is fallback behavior: downlink messages sent by newer clients of Packet Broker, including
			// The Things Stack from 3.14.1, will send the fully defined data rate in the downlink message.
			// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/4478)
			*rx.dataRateIndex = ttnpb.DataRateIndex(rx.settings.DataRateIndex)
		}
		*rx.frequency = rx.settings.Frequency
	}

	down := &ttnpb.DownlinkMessage{
		RawPayload:     msg.PhyPayload,
		CorrelationIds: events.CorrelationIDsFromContext(ctx),
		Settings: &ttnpb.DownlinkMessage_Request{
			Request: req,
		},
	}
	return uid, down, nil
}

func fromPBDevAddrBlocks(blocks []*packetbroker.DevAddrBlock) []*ttnpb.PacketBrokerDevAddrBlock {
	res := make([]*ttnpb.PacketBrokerDevAddrBlock, len(blocks))
	for i, b := range blocks {
		res[i] = &ttnpb.PacketBrokerDevAddrBlock{
			DevAddrPrefix: &ttnpb.DevAddrPrefix{
				DevAddr: &types.DevAddr{},
				Length:  b.GetPrefix().GetLength(),
			},
			HomeNetworkClusterId: b.GetHomeNetworkClusterId(),
		}
		res[i].DevAddrPrefix.DevAddr.UnmarshalNumber(b.GetPrefix().GetValue())
	}
	return res
}

func toPBDevAddrBlocks(blocks []*ttnpb.PacketBrokerDevAddrBlock) []*packetbroker.DevAddrBlock {
	res := make([]*packetbroker.DevAddrBlock, len(blocks))
	for i, b := range blocks {
		res[i] = &packetbroker.DevAddrBlock{
			Prefix: &packetbroker.DevAddrPrefix{
				Value:  b.GetDevAddrPrefix().DevAddr.MarshalNumber(),
				Length: b.GetDevAddrPrefix().GetLength(),
			},
			HomeNetworkClusterId: b.GetHomeNetworkClusterId(),
		}
	}
	return res
}

func fromPBContactInfo(admin, technical *packetbroker.ContactInfo) []*ttnpb.ContactInfo {
	res := make([]*ttnpb.ContactInfo, 0, 2)
	if email := admin.GetEmail(); email != "" {
		res = append(res, &ttnpb.ContactInfo{
			ContactType:   ttnpb.CONTACT_TYPE_OTHER,
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         email,
		})
	}
	if email := technical.GetEmail(); email != "" {
		res = append(res, &ttnpb.ContactInfo{
			ContactType:   ttnpb.CONTACT_TYPE_TECHNICAL,
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         email,
		})
	}
	return res
}

func toPBContactInfo(info []*ttnpb.ContactInfo) (admin, technical *packetbroker.ContactInfo) {
	for _, c := range info {
		if c.GetContactMethod() != ttnpb.CONTACT_METHOD_EMAIL || c.GetValue() == "" {
			continue
		}
		switch c.GetContactType() {
		case ttnpb.CONTACT_TYPE_OTHER:
			admin = &packetbroker.ContactInfo{
				Email: c.GetValue(),
			}
		case ttnpb.CONTACT_TYPE_TECHNICAL:
			technical = &packetbroker.ContactInfo{
				Email: c.GetValue(),
			}
		}
	}
	return
}

func fromPBUplinkRoutingPolicy(policy *packetbroker.RoutingPolicy_Uplink) *ttnpb.PacketBrokerRoutingPolicyUplink {
	return &ttnpb.PacketBrokerRoutingPolicyUplink{
		JoinRequest:     policy.GetJoinRequest(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
		SignalQuality:   policy.GetSignalQuality(),
		Localization:    policy.GetLocalization(),
	}
}

func fromPBDownlinkRoutingPolicy(policy *packetbroker.RoutingPolicy_Downlink) *ttnpb.PacketBrokerRoutingPolicyDownlink {
	return &ttnpb.PacketBrokerRoutingPolicyDownlink{
		JoinAccept:      policy.GetJoinAccept(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
	}
}

func fromPBDefaultRoutingPolicy(policy *packetbroker.RoutingPolicy) *ttnpb.PacketBrokerDefaultRoutingPolicy {
	return &ttnpb.PacketBrokerDefaultRoutingPolicy{
		UpdatedAt: policy.GetUpdatedAt(),
		Uplink:    fromPBUplinkRoutingPolicy(policy.GetUplink()),
		Downlink:  fromPBDownlinkRoutingPolicy(policy.GetDownlink()),
	}
}

func fromPBRoutingPolicy(policy *packetbroker.RoutingPolicy) *ttnpb.PacketBrokerRoutingPolicy {
	var homeNetworkID *ttnpb.PacketBrokerNetworkIdentifier
	if policy.HomeNetworkNetId != 0 || policy.HomeNetworkTenantId != "" {
		homeNetworkID = &ttnpb.PacketBrokerNetworkIdentifier{
			NetId:    policy.GetHomeNetworkNetId(),
			TenantId: policy.GetHomeNetworkTenantId(),
		}
	}
	return &ttnpb.PacketBrokerRoutingPolicy{
		ForwarderId: &ttnpb.PacketBrokerNetworkIdentifier{
			NetId:    policy.GetForwarderNetId(),
			TenantId: policy.GetForwarderTenantId(),
		},
		HomeNetworkId: homeNetworkID,
		UpdatedAt:     policy.GetUpdatedAt(),
		Uplink:        fromPBUplinkRoutingPolicy(policy.GetUplink()),
		Downlink:      fromPBDownlinkRoutingPolicy(policy.GetDownlink()),
	}
}

func toPBUplinkRoutingPolicy(policy *ttnpb.PacketBrokerRoutingPolicyUplink) *packetbroker.RoutingPolicy_Uplink {
	return &packetbroker.RoutingPolicy_Uplink{
		JoinRequest:     policy.GetJoinRequest(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
		SignalQuality:   policy.GetSignalQuality(),
		Localization:    policy.GetLocalization(),
	}
}

func toPBDownlinkRoutingPolicy(policy *ttnpb.PacketBrokerRoutingPolicyDownlink) *packetbroker.RoutingPolicy_Downlink {
	return &packetbroker.RoutingPolicy_Downlink{
		JoinAccept:      policy.GetJoinAccept(),
		MacData:         policy.GetMacData(),
		ApplicationData: policy.GetApplicationData(),
	}
}

var errInconsistentBands = errors.DefineInvalidArgument("inconsistent_bands", "inconsistent bands")

func toPBFrequencyPlan(fps ...*frequencyplans.FrequencyPlan) (*packetbroker.GatewayFrequencyPlan, error) {
	if len(fps) == 0 {
		return nil, nil
	}
	phy, err := band.GetByID(fps[0].BandID)
	if err != nil {
		return nil, err
	}
	res := &packetbroker.GatewayFrequencyPlan{
		Region: toPBRegion[phy.ID],
	}

	type singleSFChannel struct {
		frequency uint64
		sf, bw    uint32
	}
	singleSFChs := make(map[singleSFChannel]struct{})
	multiSFChs := make(map[uint64]struct{})

	for _, fp := range fps {
		if fp.BandID != phy.ID {
			return nil, errInconsistentBands.New()
		}
		for _, ch := range fp.UplinkChannels {
			if idx := ch.MinDataRate; idx == ch.MaxDataRate {
				dr, ok := phy.DataRates[ttnpb.DataRateIndex(idx)]
				if !ok {
					continue
				}
				switch mod := dr.Rate.Modulation.(type) {
				case *ttnpb.DataRate_Fsk:
					res.FskChannel = &packetbroker.GatewayFrequencyPlan_FSKChannel{
						Frequency: ch.Frequency,
					}
				case *ttnpb.DataRate_Lora:
					chKey := singleSFChannel{ch.Frequency, mod.Lora.SpreadingFactor, mod.Lora.Bandwidth}
					if _, ok := singleSFChs[chKey]; ok {
						continue
					}
					res.LoraSingleSfChannels = append(res.LoraSingleSfChannels, &packetbroker.GatewayFrequencyPlan_LoRaSingleSFChannel{
						Frequency:       ch.Frequency,
						SpreadingFactor: mod.Lora.SpreadingFactor,
						Bandwidth:       mod.Lora.Bandwidth,
					})
					singleSFChs[chKey] = struct{}{}
				}
			} else {
				if _, ok := multiSFChs[ch.Frequency]; ok {
					continue
				}
				res.LoraMultiSfChannels = append(res.LoraMultiSfChannels, &packetbroker.GatewayFrequencyPlan_LoRaMultiSFChannel{
					Frequency: ch.Frequency,
				})
				multiSFChs[ch.Frequency] = struct{}{}
			}
		}
	}
	return res, nil
}
