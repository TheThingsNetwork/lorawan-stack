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

package objects

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Implemented as per https://www.loracloud.com/documentation/device_management?url=v1.html#object-formats

// DeviceUplinkResponse contains the uplink response and the error if applicable.
type DeviceUplinkResponse struct {
	Result ExtendedUplinkResponse `json:"result"`
	Error  string                 `json:"error"`
}

// DeviceUplinkResponses maps the device EUIs to the DeviceUplinkResponse.
type DeviceUplinkResponses map[EUI]DeviceUplinkResponse

// UnmarshalJSON implements json.Unmarshaler.
func (d DeviceUplinkResponses) UnmarshalJSON(b []byte) error {
	m := make(map[string]DeviceUplinkResponse)
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	for k, v := range m {
		eui, err := toEUI(k)
		if err != nil {
			return err
		}
		d[eui] = v
	}
	return nil
}

// PositionSolutionType is the algorithm used by a position solution query.
type PositionSolutionType uint8

const (
	// GNSSPositionSolutionType is GNSS position solution type.
	GNSSPositionSolutionType PositionSolutionType = iota
	// GNSSNGPositionSolutionType is GNSSNG position solution type.
	GNSSNGPositionSolutionType
	// WiFiPositionSolutionType is WiFi position solution type.
	WiFiPositionSolutionType
	// UnknownPositionSolutionType is used when the position solution type is unknown.
	UnknownPositionSolutionType
)

// PositionSolution is the result of a position query.
type PositionSolution struct {
	Algorithm PositionSolutionType `json:"algorithm_type"`
	LLH       []float64            `json:"llh"`
	Accuracy  float64              `json:"accuracy"`
}

const (
	gnssPositionSolutionType    = "gnss"
	gnssngPositionSolutionType  = "gnssng"
	wifiPositionSolutionType    = "wifi"
	unknownPositionSolutionType = "unknown"
)

// MarshalJSON implements the json.Marshaler interface.
func (t PositionSolutionType) MarshalJSON() ([]byte, error) {
	var tp string
	switch t {
	case GNSSPositionSolutionType:
		tp = gnssPositionSolutionType
	case GNSSNGPositionSolutionType:
		tp = gnssngPositionSolutionType
	case WiFiPositionSolutionType:
		tp = wifiPositionSolutionType
	default:
		tp = unknownPositionSolutionType
	}
	return json.Marshal(tp)
}

// UnmarshalJSON implements the json.Unarmshaler.
func (t *PositionSolutionType) UnmarshalJSON(b []byte) error {
	var tp string
	err := json.Unmarshal(b, &tp)
	if err != nil {
		return err
	}
	switch tp {
	case gnssPositionSolutionType:
		*t = GNSSPositionSolutionType
	case gnssngPositionSolutionType:
		*t = GNSSNGPositionSolutionType
	case wifiPositionSolutionType:
		*t = WiFiPositionSolutionType
	default:
		*t = UnknownPositionSolutionType
	}
	return nil
}

// String implements fmt.Stringer.
func (t PositionSolutionType) String() string {
	switch t {
	case GNSSPositionSolutionType:
		return gnssPositionSolutionType
	case GNSSNGPositionSolutionType:
		return gnssngPositionSolutionType
	case WiFiPositionSolutionType:
		return wifiPositionSolutionType
	default:
		return unknownPositionSolutionType
	}
}

// StreamRecord contains the offset and the data of a fully reconstructed stream frame.
type StreamRecord struct {
	Offset uint32
	Data   Hex
}

// MarshalJSON implements json.Marshaler.
func (r StreamRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{r.Offset, r.Data})
}

var errInvalidStreamRecord = errors.DefineCorruption("invalid_stream_record", "invalid stream record")

// UnmarshalJSON implements json.Marshaler.
func (r *StreamRecord) UnmarshalJSON(b []byte) error {
	raw := make([]json.RawMessage, 0, 2)
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return errInvalidStreamRecord.New()
	}
	if err := json.Unmarshal(raw[0], &r.Offset); err != nil {
		return err
	}
	return json.Unmarshal(raw[1], &r.Data)
}

// UplinkResponse contains the state changes and completed items due to an uplink message.
type UplinkResponse struct {
	Downlink      *LoRaDnlink       `json:"dnlink"`
	StreamRecords []StreamRecord    `json:"stream_records"`
	Position      *PositionSolution `json:"position_solution"`
}

// ExtendedUplinkResponse extends UplinkResponse with the raw JSON payload.
type ExtendedUplinkResponse struct {
	UplinkResponse

	Raw *json.RawMessage
}

// MarshalJSON implements json.Marshaler.
// Note that the Raw representation takes precedence
// in the marshaling process, if it is available.
func (r ExtendedUplinkResponse) MarshalJSON() ([]byte, error) {
	if r.Raw != nil {
		return r.Raw.MarshalJSON()
	}
	return json.Marshal(r.UplinkResponse)
}

// UnmarshalJSON implements json.Marshaler.
func (r *ExtendedUplinkResponse) UnmarshalJSON(b []byte) error {
	r.Raw = &json.RawMessage{}
	if err := r.Raw.UnmarshalJSON(b); err != nil {
		return err
	}
	return json.Unmarshal(b, &r.UplinkResponse)
}

// DeviceUplinks maps device EUIs to LoRaUplink
type DeviceUplinks map[EUI]*LoRaUplink

// MarshalJSON implements json.Marshaler.
func (u DeviceUplinks) MarshalJSON() ([]byte, error) {
	m := make(map[string]*LoRaUplink)
	for k, v := range u {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

// LoRaUplink encapsulates the information of a LoRa message.
type LoRaUplink struct {
	Type LoRaUplinkType `json:"msgtype"`

	FCnt        *uint32  `json:"fcnt,omitempty"`
	Port        *uint8   `json:"port,omitempty"`
	Payload     *Hex     `json:"payload,omitempty"`
	DR          *uint8   `json:"dr,omitempty"`
	Freq        *uint32  `json:"freq,omitempty"`
	Timestamp   *float64 `json:"timestamp,omitempty"`
	DownlinkMTU *uint32  `json:"dn_mtu,omitempty"`

	GNSSCaptureTime         *float64  `json:"gnss_capture_time,omitempty"`
	GNSSCaptureTimeAccuracy *float64  `json:"gnss_capture_time_accuracy,omitempty"`
	GNSSAssistPosition      []float64 `json:"gnss_assist_position,omitempty"`
	GNSSAssistAltitude      *float64  `json:"gnss_assist_altitude,omitempty"`
	GNSSUse2DSolver         *bool     `json:"gnss_use_2D_solver,omitempty"`
}

// LoRaUplinkType is the type of a LoRaCloud DMS uplink.
type LoRaUplinkType uint8

const (
	// UplinkUplinkType is LoRaWAN Message Type.
	UplinkUplinkType LoRaUplinkType = iota
	// ModemUplinkType is DMS Protocol Message Type.
	ModemUplinkType
	// JoiningUplinkType is Session Reset Message Type.
	JoiningUplinkType
	// GNSSUplinkType is DMS GNSS Message Type.
	GNSSUplinkType
	// WiFiUplinkType is DMS WiFi Message Type.
	WiFiUplinkType
	// UnknownUplinkType is used when the uplink type is unknown.
	UnknownUplinkType
)

const (
	uplinkUplinkType  = "updf"
	modemUplinkType   = "modem"
	joiningUplinkType = "joining"
	gnssUplinkType    = "gnss"
	wifiUplinkType    = "wifi"
	unknownUplinkType = "unknown"
)

// MarshalJSON implements the json.Marshaler interface.
func (t LoRaUplinkType) MarshalJSON() ([]byte, error) {
	var tp string
	switch t {
	case UplinkUplinkType:
		tp = uplinkUplinkType
	case ModemUplinkType:
		tp = modemUplinkType
	case JoiningUplinkType:
		tp = joiningUplinkType
	case GNSSUplinkType:
		tp = gnssUplinkType
	case WiFiUplinkType:
		tp = wifiUplinkType
	default:
		tp = unknownUplinkType
	}
	return json.Marshal(tp)
}

// UnmarshalJSON implements the json.Unarmshaler.
func (t *LoRaUplinkType) UnmarshalJSON(b []byte) error {
	var tp string
	err := json.Unmarshal(b, &tp)
	if err != nil {
		return err
	}
	switch tp {
	case uplinkUplinkType:
		*t = UplinkUplinkType
	case modemUplinkType:
		*t = ModemUplinkType
	case joiningUplinkType:
		*t = JoiningUplinkType
	case gnssUplinkType:
		*t = GNSSUplinkType
	case wifiUplinkType:
		*t = WiFiUplinkType
	default:
		*t = UnknownUplinkType
	}
	return nil
}

// LoRaDnlink is a specification for a modem device.
type LoRaDnlink struct {
	Port    uint8 `json:"port"`
	Payload Hex   `json:"payload"`
}

// Fields implements log.Fielder.
func (u LoRaDnlink) Fields() map[string]any {
	return map[string]any{
		"port":    u.Port,
		"payload": u.Payload,
	}
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

// EUI represents a dash-separated EUI64.
type EUI types.EUI64

const (
	hyphenatedEUIPattern = "\"%02X-%02X-%02X-%02X-%02X-%02X-%02X-%02X\""
	euiPattern           = "%02X-%02X-%02X-%02X-%02X-%02X-%02X-%02X"
)

// MarshalJSON implements json.Marshaler.
func (e EUI) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(hyphenatedEUIPattern, e[0], e[1], e[2], e[3], e[4], e[5], e[6], e[7])), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *EUI) UnmarshalJSON(b []byte) error {
	_, err := fmt.Sscanf(string(b), hyphenatedEUIPattern, &e[0], &e[1], &e[2], &e[3], &e[4], &e[5], &e[6], &e[7])
	return err
}

func (e EUI) String() string {
	return fmt.Sprintf(euiPattern, e[0], e[1], e[2], e[3], e[4], e[5], e[6], e[7])
}

func toEUI(s string) (e EUI, err error) {
	_, err = fmt.Sscanf(s, euiPattern, &e[0], &e[1], &e[2], &e[3], &e[4], &e[5], &e[6], &e[7])
	return e, err
}
