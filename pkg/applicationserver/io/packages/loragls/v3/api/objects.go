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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// Gateway contains the description of a LoRaWAN gateway.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#gateway
type Gateway struct {
	GatewayID string  `json:"gatewayId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// Uplink contains the metadata of a LoRaWAN uplink.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#uplink
type Uplink struct {
	GatewayID string
	AntennaID *uint32
	TDOA      *uint64
	RSSI      float64
	SNR       float64
}

// MarshalJSON implements json.Marshaler.
func (u *Uplink) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		u.GatewayID,
		u.AntennaID,
		u.TDOA,
		u.RSSI,
		u.SNR,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *Uplink) UnmarshalJSON(b []byte) error {
	components := make([]json.RawMessage, 0, 5)
	if err := json.Unmarshal(b, &components); err != nil {
		return err
	}
	if n := len(components); n != 5 {
		return fmt.Errorf("invalid field count %d", n)
	}
	for i, c := range []any{
		&u.GatewayID,
		&u.AntennaID,
		&u.TDOA,
		&u.RSSI,
		&u.SNR,
	} {
		if err := json.Unmarshal(components[i], c); err != nil {
			return err
		}
	}
	return nil
}

// Frame contains the uplink metadata for each reception.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#frame
type Frame []Uplink

// SingleFrameRequest contains the location query request for a single LoRaWAN frame.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#singleframe-http-request
type SingleFrameRequest struct {
	Gateways []Gateway `json:"gateways"`
	Frame    Frame     `json:"frame"`
}

func parseRxMetadata(ctx context.Context, m *RxMetadata) (Gateway, Uplink) {
	ids := m.GatewayIDs.ToProto()
	gtwUID := unique.ID(ctx, ids)
	hashed := sha256.Sum256([]byte(gtwUID))
	hashedUID := hex.EncodeToString(hashed[:])
	var tdoa *uint64
	if m.FineTimestamp != 0 {
		tdoa = &m.FineTimestamp
	}
	return Gateway{
			GatewayID: hashedUID,
			Latitude:  m.Location.Latitude,
			Longitude: m.Location.Longitude,
			Altitude:  float64(m.Location.Altitude),
		}, Uplink{
			GatewayID: hashedUID,
			AntennaID: &m.AntennaIndex,
			TDOA:      tdoa,
			RSSI:      float64(m.RSSI),
			SNR:       float64(m.SNR),
		}
}

// BuildSingleFrameRequest builds a SingleFrameRequest from the provided metadata.
func BuildSingleFrameRequest(ctx context.Context, metadata []*RxMetadata) *SingleFrameRequest {
	r := &SingleFrameRequest{
		Gateways: []Gateway{},
		Frame:    Frame{},
	}
	for _, m := range metadata {
		if m.Location == nil || m.GatewayIDs == nil {
			continue
		}
		gtw, up := parseRxMetadata(ctx, m)
		r.Gateways = append(r.Gateways, gtw)
		r.Frame = append(r.Frame, up)
	}
	return r
}

// MultiFrameRequest contains the location query request for multiple LoRaWAN frames.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#multiframe-http-request
type MultiFrameRequest struct {
	Gateways []Gateway `json:"gateways"`
	Frames   []Frame   `json:"frames"`
}

// GatewayIDs contains the fields of the gateway identifiers used by the package.
type GatewayIDs struct {
	GatewayID string `json:"gateway_id"`
}

// ToProto converts the GatewayIDs to a protobuf representation.
func (g *GatewayIDs) ToProto() *ttnpb.GatewayIdentifiers {
	return &ttnpb.GatewayIdentifiers{
		GatewayId: g.GatewayID,
	}
}

// FromProto converts the GatewayIDs from a protobuf representation.
func (g *GatewayIDs) FromProto(pb *ttnpb.GatewayIdentifiers) error {
	g.GatewayID = pb.GatewayId
	return nil
}

// RxMDLocation contains the metadata location fields used by the package.
type RxMDLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  int32   `json:"altitude"`
	Accuracy  int32   `json:"accuracy"`
	Source    int32   `json:"source"`
}

// FromProto converts the Location from a protobuf representation.
func (l *RxMDLocation) FromProto(pb *ttnpb.Location) error {
	l.Latitude = pb.Latitude
	l.Longitude = pb.Longitude
	l.Altitude = pb.Altitude
	l.Accuracy = pb.Accuracy
	l.Source = int32(pb.Source)
	return nil
}

// RxMetadata contains the fields of the RxMetadata used by the package.
type RxMetadata struct {
	GatewayIDs    *GatewayIDs   `json:"gateway_ids"`
	AntennaIndex  uint32        `json:"antenna_index"`
	FineTimestamp uint64        `json:"fine_timestamp"`
	Location      *RxMDLocation `json:"location"`
	RSSI          float32       `json:"rssi"`
	SNR           float32       `json:"snr"`
}

// FromProto converts the RxMetadata from a protobuf representation.
func (r *RxMetadata) FromProto(pb *ttnpb.RxMetadata) error {
	r.GatewayIDs = &GatewayIDs{}
	if err := r.GatewayIDs.FromProto(pb.GatewayIds); err != nil {
		return err
	}

	r.Location = &RxMDLocation{}
	if err := r.Location.FromProto(pb.Location); err != nil {
		return err
	}

	r.AntennaIndex = pb.AntennaIndex
	r.FineTimestamp = pb.FineTimestamp
	r.RSSI = pb.Rssi
	r.SNR = pb.Snr

	return nil
}

// BuildMultiFrameRequest builds a MultiFrameRequest from the provided metadata.
func BuildMultiFrameRequest(ctx context.Context, mds [][]*RxMetadata) *MultiFrameRequest {
	r := &MultiFrameRequest{
		Gateways: []Gateway{},
		Frames:   []Frame{},
	}
	gateways := map[string]struct{}{}
	for _, metadata := range mds {
		frame := Frame{}
		for _, m := range metadata {
			if m.Location == nil || m.GatewayIDs == nil {
				continue
			}
			gtw, up := parseRxMetadata(ctx, m)
			if _, seen := gateways[gtw.GatewayID]; !seen {
				r.Gateways = append(r.Gateways, gtw)
				gateways[gtw.GatewayID] = struct{}{}
			}
			frame = append(frame, up)
		}
		r.Frames = append(r.Frames, frame)
	}
	return r
}

// Algorithms supported by the location solver.
const (
	Algorithm_TDOA     = "Tdoa"
	Algorithm_RSSI     = "Rssi"
	Algorithm_RSSITDOA = "RssiTdoaCombined"
)

// Location contains the coordinates contained in a location query result.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Tolerance uint64  `json:"toleranceHoriz"`
}

// LocationSolverResult contains the result of a location query.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#locationsolverresult
type LocationSolverResult struct {
	UsedGateways uint8    `json:"numUsedGateways"`
	HDOP         *float64 `json:"HDOP"`
	Algorithm    string   `json:"algorithmType"`
	Location     Location `json:"locationEst"`
}

// LocationSolverResponse contains the location query response.
// https://www.loracloud.com/documentation/geolocation?url=v3.html#singleframe-http-request
// https://www.loracloud.com/documentation/geolocation?url=v3.html#multiframe-http-request
type LocationSolverResponse struct {
	Result   *LocationSolverResult `json:"result"`
	Errors   []string              `json:"errors"`
	Warnings []string              `json:"warnings"`
}

// ExtendedLocationSolverResponse extends LocationSolverResponse with the raw JSON representation.
type ExtendedLocationSolverResponse struct {
	LocationSolverResponse

	Raw *json.RawMessage
}

// MarshalJSON implements json.Marshaler.
// Note that the Raw representation takes precedence
// in the marshaling process, if it is available.
func (r ExtendedLocationSolverResponse) MarshalJSON() ([]byte, error) {
	if r.Raw != nil {
		return r.Raw.MarshalJSON()
	}
	return json.Marshal(r.LocationSolverResponse)
}

// UnmarshalJSON implements json.Marshaler.
func (r *ExtendedLocationSolverResponse) UnmarshalJSON(b []byte) error {
	r.Raw = &json.RawMessage{}
	if err := r.Raw.UnmarshalJSON(b); err != nil {
		return err
	}
	return json.Unmarshal(b, &r.LocationSolverResponse)
}

// Hex represents hex encoded bytes.
type Hex []byte

// MarshalJSON implements json.Marshaler.
func (h Hex) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", hex.EncodeToString(h))), nil
}

// String implements fmt.Stringer.
func (h Hex) String() string {
	return hex.EncodeToString(h)
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *Hex) UnmarshalJSON(b []byte) (err error) {
	s := strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\"")
	*h, err = hex.DecodeString(s)
	return err
}

// GNSSRequest contains the location query request based on a GNSS payload.
// https://www.loracloud.com/documentation/geolocation?url=gnss.html#single-capture-http-request
type GNSSRequest struct {
	Payload Hex `json:"payload"`
}

// GNSSLocationSolverResult contains the result of a GNSS location query.
// https://www.loracloud.com/documentation/geolocation?url=gnss.html#locationsolverresult
type GNSSLocationSolverResult struct {
	LLH      []float64 `json:"llh"`
	Accuracy float64   `json:"accuracy"`
}

// GNSSLocationSolverResponse contains the GNSS location query response.
// https://www.loracloud.com/documentation/geolocation?url=gnss.html#single-capture-http-request
type GNSSLocationSolverResponse struct {
	Result   *GNSSLocationSolverResult `json:"result"`
	Errors   []string                  `json:"errors"`
	Warnings []string                  `json:"warnings"`
}

// ExtendedGNSSLocationSolverResponse extends GNSSLocationSolverResponse with the raw JSON representation.
type ExtendedGNSSLocationSolverResponse struct {
	GNSSLocationSolverResponse

	Raw *json.RawMessage
}

// MarshalJSON implements json.Marshaler.
// Note that the Raw representation takes precedence
// in the marshaling process, if it is available.
func (r ExtendedGNSSLocationSolverResponse) MarshalJSON() ([]byte, error) {
	if r.Raw != nil {
		return r.Raw.MarshalJSON()
	}
	return json.Marshal(r.GNSSLocationSolverResponse)
}

// UnmarshalJSON implements json.Marshaler.
func (r *ExtendedGNSSLocationSolverResponse) UnmarshalJSON(b []byte) error {
	r.Raw = &json.RawMessage{}
	if err := r.Raw.UnmarshalJSON(b); err != nil {
		return err
	}
	return json.Unmarshal(b, &r.GNSSLocationSolverResponse)
}
