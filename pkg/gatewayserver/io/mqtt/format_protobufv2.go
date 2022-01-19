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

package mqtt

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	ttnpbv2 "go.thethings.network/lorawan-stack-legacy/v2/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/datarate"
)

// eirpDelta is the delta between EIRP and ERP.
const eirpDelta = 2.15

var (
	sourceToV3 = map[ttnpbv2.LocationMetadata_LocationSource]ttnpb.LocationSource{
		ttnpbv2.LocationMetadata_GPS:            ttnpb.LocationSource_SOURCE_GPS,
		ttnpbv2.LocationMetadata_CONFIG:         ttnpb.LocationSource_SOURCE_REGISTRY,
		ttnpbv2.LocationMetadata_REGISTRY:       ttnpb.LocationSource_SOURCE_REGISTRY,
		ttnpbv2.LocationMetadata_IP_GEOLOCATION: ttnpb.LocationSource_SOURCE_IP_GEOLOCATION,
	}

	errNotScheduled    = errors.DefineInvalidArgument("not_scheduled", "not scheduled")
	errLoRaWANMetadata = errors.DefineInvalidArgument("lorawan_metadata", "missing LoRaWAN metadata")
	errModulation      = errors.DefineInvalidArgument("modulation", "unknown modulation `{modulation}`")
)

type protobufv2 struct {
	topics.Layout
}

func (protobufv2) FromDownlink(down *ttnpb.DownlinkMessage, _ ttnpb.GatewayIdentifiers) ([]byte, error) {
	settings := down.GetScheduled()
	if settings == nil {
		return nil, errNotScheduled.New()
	}
	lorawan := &ttnpbv2.LoRaWANTxConfiguration{}
	if pld, ok := down.GetPayload().GetPayload().(*ttnpb.Message_MacPayload); ok && pld != nil {
		lorawan.FCnt = pld.MacPayload.FHdr.GetFCnt()
	}
	switch dr := settings.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_Lora:
		lorawan.Modulation = ttnpbv2.Modulation_LORA
		lorawan.CodingRate = settings.CodingRate
		lorawan.DataRate = fmt.Sprintf("SF%dBW%d", dr.Lora.SpreadingFactor, dr.Lora.Bandwidth/1000)
	case *ttnpb.DataRate_Fsk:
		lorawan.Modulation = ttnpbv2.Modulation_FSK
		lorawan.BitRate = dr.Fsk.BitRate
	default:
		return nil, errModulation.New()
	}

	v2downlink := &ttnpbv2.DownlinkMessage{
		Payload: down.RawPayload,
		GatewayConfiguration: &ttnpbv2.GatewayTxConfiguration{
			Frequency:             settings.Frequency,
			Power:                 int32(settings.Downlink.TxPower - eirpDelta),
			PolarizationInversion: true,
			RfChain:               0,
			Timestamp:             settings.Timestamp,
		},
		ProtocolConfiguration: &ttnpbv2.ProtocolTxConfiguration{
			Lorawan: lorawan,
		},
	}
	return v2downlink.Marshal()
}

func (protobufv2) ToUplink(message []byte, ids ttnpb.GatewayIdentifiers) (*ttnpb.UplinkMessage, error) {
	v2uplink := &ttnpbv2.UplinkMessage{}
	err := v2uplink.Unmarshal(message)
	if err != nil {
		return nil, err
	}

	lorawanMetadata := v2uplink.GetProtocolMetadata().GetLorawan()
	if lorawanMetadata == nil {
		return nil, errLoRaWANMetadata.New()
	}
	gwMetadata := v2uplink.GatewayMetadata
	uplink := &ttnpb.UplinkMessage{
		RawPayload: v2uplink.Payload,
	}

	settings := ttnpb.TxSettings{
		Frequency: gwMetadata.Frequency,
		Timestamp: gwMetadata.Timestamp,
	}
	switch lorawanMetadata.Modulation {
	case ttnpbv2.Modulation_LORA:
		loraDr, err := datarate.ParseLoRa(lorawanMetadata.DataRate)
		if err != nil {
			return nil, err
		}
		settings.DataRate = loraDr.DataRate
		settings.CodingRate = lorawanMetadata.CodingRate
	case ttnpbv2.Modulation_FSK:
		settings.DataRate = &ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Fsk{
				Fsk: &ttnpb.FSKDataRate{
					BitRate: lorawanMetadata.BitRate,
				},
			},
		}
	default:
		return nil, errModulation.WithAttributes("modulation", lorawanMetadata.Modulation)
	}

	mdTime := ttnpb.ProtoTimePtr(time.Unix(0, gwMetadata.Time))
	if antennas := gwMetadata.Antennas; len(antennas) > 0 {
		for _, antenna := range antennas {
			rssi := antenna.ChannelRssi
			if rssi == 0 {
				rssi = antenna.Rssi
			}
			uplink.RxMetadata = append(uplink.RxMetadata, &ttnpb.RxMetadata{
				GatewayIds:            &ids,
				AntennaIndex:          antenna.Antenna,
				ChannelRssi:           rssi,
				FrequencyOffset:       antenna.FrequencyOffset,
				Rssi:                  rssi,
				RssiStandardDeviation: antenna.RssiStandardDeviation,
				Snr:                   antenna.Snr,
				Time:                  mdTime,
				Timestamp:             gwMetadata.Timestamp,
			})
		}
	} else {
		uplink.RxMetadata = append(uplink.RxMetadata, &ttnpb.RxMetadata{
			GatewayIds:   &ids,
			AntennaIndex: 0,
			ChannelRssi:  gwMetadata.Rssi,
			Rssi:         gwMetadata.Rssi,
			Snr:          gwMetadata.Snr,
			Time:         mdTime,
			Timestamp:    gwMetadata.Timestamp,
		})
	}
	uplink.Settings = &settings
	return uplink, nil
}

var ttkgPlatformRegex = regexp.MustCompile(`The Things Gateway v1 - BL (r[0-9]+\-[0-9a-f]+) \(([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z)\) - Firmware (v[0-9]+.[0-9]+.[0-9]+\-[0-9a-f]+) \(([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z)\)`)

func (protobufv2) ToStatus(message []byte, _ ttnpb.GatewayIdentifiers) (*ttnpb.GatewayStatus, error) {
	v2status := &ttnpbv2.StatusMessage{}
	err := v2status.Unmarshal(message)
	if err != nil {
		return nil, err
	}
	metrics := map[string]float32{
		"lmnw": float32(v2status.LmNw),
		"lmst": float32(v2status.LmSt),
		"lmok": float32(v2status.LmOk),
		"lpps": float32(v2status.LPps),
		"rxin": float32(v2status.RxIn),
		"rxok": float32(v2status.RxOk),
		"txin": float32(v2status.TxIn),
		"txok": float32(v2status.TxOk),
	}
	if os := v2status.Os; os != nil {
		metrics["cpu_percentage"] = os.CpuPercentage
		metrics["load_1"] = os.Load_1
		metrics["load_5"] = os.Load_5
		metrics["load_15"] = os.Load_15
		metrics["memory_percentage"] = os.MemoryPercentage
		metrics["temp"] = os.Temperature
	}
	if v2status.Rtt != 0 {
		metrics["rtt_ms"] = float32(v2status.Rtt)
	}
	versions := make(map[string]string)
	if v2status.Dsp > 0 {
		versions["dsp"] = strconv.Itoa(int(v2status.Dsp))
	}
	if v2status.Fpga > 0 {
		versions["fpga"] = strconv.Itoa(int(v2status.Fpga))
	}
	if v2status.Hal != "" {
		versions["hal"] = v2status.Hal
	}
	if v2status.Platform != "" {
		if matches := ttkgPlatformRegex.FindStringSubmatch(v2status.Platform); len(matches) == 5 {
			versions["model"] = "The Things Kickstarter Gateway v1"
			versions["firmware"] = matches[3]
		}
		versions["platform"] = v2status.Platform
	}
	var antennasLocation []*ttnpb.Location
	if loc := v2status.Location; loc.Validate() {
		antennasLocation = []*ttnpb.Location{
			{
				Accuracy:  loc.Accuracy,
				Altitude:  loc.Altitude,
				Latitude:  float64(loc.Latitude),
				Longitude: float64(loc.Longitude),
				Source:    sourceToV3[loc.Source],
			},
		}
	}
	return &ttnpb.GatewayStatus{
		AntennaLocations: antennasLocation,
		BootTime:         ttnpb.ProtoTimePtr(time.Unix(0, v2status.BootTime)),
		Ip:               v2status.Ip,
		Metrics:          metrics,
		Time:             ttnpb.ProtoTimePtr(time.Unix(0, v2status.Time)),
		Versions:         versions,
	}, nil
}

func (protobufv2) ToTxAck(message []byte, _ ttnpb.GatewayIdentifiers) (*ttnpb.TxAcknowledgment, error) {
	return nil, errNotSupported.New()
}

// NewProtobufV2 returns a format that uses the legacy The Things Stack V2 Protocol Buffers marshaling and unmarshaling.
func NewProtobufV2(ctx context.Context) Format {
	return &protobufv2{
		Layout: topics.NewV2(ctx),
	}
}
