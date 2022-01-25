// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package api

import (
	"context"
	"encoding/json"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// BuildWiFiRequest builds a WiFiRequest from the provided metadata and payload.
func BuildWiFiRequest(ctx context.Context, metadata []*ttnpb.RxMetadata, accessPoints []AccessPoint) *WiFiRequest {
	removeNil := func(v *uint64) uint64 {
		if v == nil {
			return 0
		}
		return *v
	}
	r := &WiFiRequest{
		LoRaWAN:          []TDOAUplink{},
		WiFiAccessPoints: accessPoints,
	}
	for _, m := range metadata {
		if m.Location == nil || m.GatewayIds == nil {
			continue
		}
		gtw, up := parseRxMetadata(ctx, m)
		r.LoRaWAN = append(r.LoRaWAN, TDOAUplink{
			GatewayID: gtw.GatewayID,
			RSSI:      up.RSSI,
			SNR:       up.SNR,
			TDOA:      removeNil(up.TDOA),
			AntennaID: up.AntennaID,
			AntennaLocation: AntennaLocation{
				Latitude:  gtw.Latitude,
				Longitude: gtw.Longitude,
				Altitude:  gtw.Altitude,
			},
		})
	}
	return r
}

// WiFiRequest contains a WiFi / TDOA location query.
// https://www.loracloud.com/documentation/geolocation?url=v2.html#singleframe-wi-fi-tdoa-request
type WiFiRequest struct {
	LoRaWAN          []TDOAUplink  `json:"lorawan"`
	WiFiAccessPoints []AccessPoint `json:"wifiAccessPoints"`
}

// AntennaLocation contains the location information of a gateway.
// https://www.loracloud.com/documentation/geolocation?url=v2.html#antennalocation
type AntennaLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// TDOAUplink contains the metadata of an uplink.
// https://www.loracloud.com/documentation/geolocation?url=v2.html#uplinktdoa
type TDOAUplink struct {
	GatewayID       string          `json:"gatewayId"`
	RSSI            float64         `json:"rssi"`
	SNR             float64         `json:"snr"`
	TDOA            uint64          `json:"toa"`
	AntennaID       *uint32         `json:"antennaId"`
	AntennaLocation AntennaLocation `json:"antennaLocation"`
}

// AccessPoint contains the metadata of a WiFi access point.
// https://www.loracloud.com/documentation/geolocation?url=v2.html#wifiaccesspoint
type AccessPoint struct {
	MACAddress     string `json:"macAddress"`
	SignalStrength int64  `json:"signalStrength"`
}

// WiFiLocationSolverResponse contains the result of a WiFi location query.
// https://www.loracloud.com/documentation/geolocation?url=v2.html#singleframe-wi-fi-tdoa-request
type WiFiLocationSolverResponse struct {
	Result   *WiFiLocationSolverResult `json:"result"`
	Errors   []string                  `json:"errors"`
	Warnings []string                  `json:"warnings"`
}

// WiFiLocationSolverResult contains the result of a WiFi location query.
// https://www.loracloud.com/documentation/geolocation?url=v2.html#locationresult
type WiFiLocationSolverResult struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Altitude         float64 `json:"altitude"`
	Accuracy         float64 `json:"accuracy"`
	Algorithm        string  `json:"algorithmType"`
	GatewaysReceived float64 `json:"numberOfGatewaysReceived"`
	GatewaysUsed     float64 `json:"numberOfGatewaysUsed"`
}

// ExtendedWiFiLocationSolverResponse extends WiFiLocationQueryResponse with the raw JSON representation.
type ExtendedWiFiLocationSolverResponse struct {
	WiFiLocationSolverResponse

	Raw *json.RawMessage
}

// MarshalJSON implements json.Marshaler.
// Note that the Raw representation takes precedence
// in the marshaling process, if it is available.
func (r ExtendedWiFiLocationSolverResponse) MarshalJSON() ([]byte, error) {
	if r.Raw != nil {
		return r.Raw.MarshalJSON()
	}
	return json.Marshal(r.WiFiLocationSolverResponse)
}

// UnmarshalJSON implements json.Marshaler.
func (r *ExtendedWiFiLocationSolverResponse) UnmarshalJSON(b []byte) error {
	r.Raw = &json.RawMessage{}
	if err := r.Raw.UnmarshalJSON(b); err != nil {
		return err
	}
	return json.Unmarshal(b, &r.WiFiLocationSolverResponse)
}
