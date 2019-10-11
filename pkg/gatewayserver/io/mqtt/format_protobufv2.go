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
	"fmt"
	"strconv"
	"time"

	ttnpbv2 "go.thethings.network/lorawan-stack-legacy/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/datarate"
)

// eirpDelta is the delta between EIRP and ERP.
const eirpDelta = 2.15

var (
	sourceToV3 = map[ttnpbv2.LocationMetadata_LocationSource]ttnpb.LocationSource{
		ttnpbv2.LocationMetadata_GPS:            ttnpb.SOURCE_GPS,
		ttnpbv2.LocationMetadata_CONFIG:         ttnpb.SOURCE_REGISTRY,
		ttnpbv2.LocationMetadata_REGISTRY:       ttnpb.SOURCE_REGISTRY,
		ttnpbv2.LocationMetadata_IP_GEOLOCATION: ttnpb.SOURCE_IP_GEOLOCATION,
	}

	frequencyPlanToBand = map[uint32]string{
		0:  "EU_863_870",
		1:  "US_902_928",
		2:  "CN_779_787",
		3:  "EU_433",
		4:  "AU_915_928",
		5:  "CN_470_510",
		6:  "AS_923",
		7:  "KR_920_923",
		8:  "IN_865_867",
		9:  "RU_864_870",
		61: "AS_923",
		62: "AS_923",
	}

	errNotScheduled    = errors.DefineInvalidArgument("not_scheduled", "not scheduled")
	errLoRaWANPayload  = errors.DefineInvalidArgument("lorawan_payload", "invalid LoRaWAN payload")
	errLoRaWANMetadata = errors.DefineInvalidArgument("lorawan_metadata", "missing LoRaWAN metadata")
	errDataRate        = errors.DefineInvalidArgument("data_rate", "unknown data rate `{data_rate}`")
	errModulation      = errors.DefineInvalidArgument("modulation", "unknown modulation `{modulation}`")
	errFrequencyPlan   = errors.DefineNotFound("frequency_plan", "unknown frequency plan `{frequency_plan}`")
)

type protobufv2 struct {
	topics.Layout
}

func (protobufv2) FromDownlink(down *ttnpb.DownlinkMessage, _ ttnpb.GatewayIdentifiers) ([]byte, error) {
	settings := down.GetScheduled()
	if settings == nil {
		return nil, errNotScheduled
	}
	lorawan := &ttnpbv2.LoRaWANTxConfiguration{}
	if pld, ok := down.GetPayload().GetPayload().(*ttnpb.Message_MACPayload); ok && pld != nil {
		lorawan.FCnt = pld.MACPayload.FHDR.FCnt
	}
	switch dr := settings.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_LoRa:
		lorawan.Modulation = ttnpbv2.Modulation_LORA
		lorawan.CodingRate = settings.CodingRate
		lorawan.DataRate = fmt.Sprintf("SF%dBW%d", dr.LoRa.SpreadingFactor, dr.LoRa.Bandwidth/1000)
	case *ttnpb.DataRate_FSK:
		lorawan.Modulation = ttnpbv2.Modulation_FSK
		lorawan.BitRate = dr.FSK.BitRate
	default:
		return nil, errModulation
	}

	v2downlink := &ttnpbv2.DownlinkMessage{
		Payload: down.RawPayload,
		GatewayConfiguration: ttnpbv2.GatewayTxConfiguration{
			Frequency:             settings.Frequency,
			Power:                 int32(settings.Downlink.TxPower - eirpDelta),
			PolarizationInversion: true,
			RfChain:               0,
			Timestamp:             settings.Timestamp,
		},
		ProtocolConfiguration: ttnpbv2.ProtocolTxConfiguration{
			LoRaWAN: lorawan,
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

	if v2uplink.ProtocolMetadata.LoRaWAN == nil {
		return nil, errLoRaWANMetadata
	}
	lorawanMetadata := v2uplink.ProtocolMetadata.LoRaWAN
	gwMetadata := v2uplink.GatewayMetadata
	uplink := &ttnpb.UplinkMessage{
		RawPayload: v2uplink.Payload,
	}

	settings := ttnpb.TxSettings{
		Frequency: gwMetadata.Frequency,
	}
	switch lorawanMetadata.Modulation {
	case ttnpbv2.Modulation_LORA:
		bandID, ok := frequencyPlanToBand[lorawanMetadata.FrequencyPlan]
		if !ok {
			return nil, errFrequencyPlan.WithAttributes("frequency_plan", lorawanMetadata.FrequencyPlan)
		}
		band, err := band.GetByID(bandID)
		if err != nil {
			return nil, err
		}
		var drIndex ttnpb.DataRateIndex
		var found bool
		loraDr, err := datarate.ParseLoRa(lorawanMetadata.DataRate)
		if err != nil {
			return nil, err
		}
		for bandDRIndex, bandDR := range band.DataRates {
			if bandDR.Rate.Equal(loraDr.DataRate) {
				found = true
				drIndex = ttnpb.DataRateIndex(bandDRIndex)
				break
			}
		}
		if !found {
			return nil, errDataRate.WithAttributes("data_rate", lorawanMetadata.DataRate)
		}
		settings.DataRate = loraDr.DataRate
		settings.CodingRate = lorawanMetadata.CodingRate
		settings.DataRateIndex = drIndex
	case ttnpbv2.Modulation_FSK:
		settings.DataRate = ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_FSK{
				FSK: &ttnpb.FSKDataRate{
					BitRate: lorawanMetadata.BitRate,
				},
			},
		}
	default:
		return nil, errModulation.WithAttributes("modulation", lorawanMetadata.Modulation)
	}

	mdTime := time.Unix(0, gwMetadata.Time)
	if antennas := gwMetadata.Antennas; len(antennas) > 0 {
		for _, antenna := range antennas {
			uplink.RxMetadata = append(uplink.RxMetadata, &ttnpb.RxMetadata{
				AntennaIndex:          antenna.Antenna,
				ChannelRSSI:           antenna.ChannelRSSI,
				FrequencyOffset:       antenna.FrequencyOffset,
				GatewayIdentifiers:    ids,
				RSSI:                  antenna.RSSI,
				RSSIStandardDeviation: antenna.RSSIStandardDeviation,
				SNR:                   antenna.SNR,
				Time:                  &mdTime,
				Timestamp:             gwMetadata.Timestamp,
			})
		}
	} else {
		uplink.RxMetadata = append(uplink.RxMetadata, &ttnpb.RxMetadata{
			AntennaIndex:       0,
			GatewayIdentifiers: ids,
			RSSI:               gwMetadata.RSSI,
			SNR:                gwMetadata.SNR,
			Time:               &mdTime,
			Timestamp:          gwMetadata.Timestamp,
		})
	}
	uplink.Settings = settings

	return uplink, nil
}

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
		"lpps": float32(v2status.LPPS),
		"rxin": float32(v2status.RxIn),
		"rxok": float32(v2status.RxOk),
		"txin": float32(v2status.TxIn),
		"txok": float32(v2status.TxOk),
	}
	if os := v2status.OS; os != nil {
		metrics["cpu_percentage"] = os.CPUPercentage
		metrics["load_1"] = os.Load_1
		metrics["load_5"] = os.Load_5
		metrics["load_15"] = os.Load_15
		metrics["memory_percentage"] = os.MemoryPercentage
		metrics["temp"] = os.Temperature
	}
	if v2status.RTT != 0 {
		metrics["rtt_ms"] = float32(v2status.RTT)
	}
	versions := make(map[string]string)
	if v2status.DSP > 0 {
		versions["dsp"] = strconv.Itoa(int(v2status.DSP))
	}
	if v2status.FPGA > 0 {
		versions["fpga"] = strconv.Itoa(int(v2status.FPGA))
	}
	if v2status.HAL != "" {
		versions["hal"] = v2status.HAL
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
		BootTime:         time.Unix(0, v2status.BootTime),
		IP:               v2status.IP,
		Metrics:          metrics,
		Time:             time.Unix(0, v2status.Time),
		Versions:         versions,
	}, nil
}

func (protobufv2) ToTxAck(message []byte, _ ttnpb.GatewayIdentifiers) (*ttnpb.TxAcknowledgment, error) {
	return nil, errNotSupported
}

// ProtobufV2 is a format that uses the legacy The Things Stack V2 Protocol Buffers marshaling and unmarshaling.
var ProtobufV2 Format = &protobufv2{
	Layout: topics.V2,
}
