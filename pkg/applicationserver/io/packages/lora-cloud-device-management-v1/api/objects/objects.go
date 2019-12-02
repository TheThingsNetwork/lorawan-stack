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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Implemented as per https://www.loracloud.com/documentation/device_management?url=v1.html#object-formats

// DeviceInfo encapsulates the current state of a modem as known to the server.
type DeviceInfo struct {
	// DMPorts contains the ports currently accepted as "dmport".
	DMPorts []uint8 `json:"dmports"`
	// InfoFields is the current view of modem fields by service.
	InfoFields InfoFields `json:"info_fields"`
	// UploadSessions contains the current upload sessions.
	UploadSessions []UploadSession `json:"upload_sessions"`
	// StreamSessions contains the current streaming sessions.
	StreamSessions  []StreamSession `json:"stream_sessions"`
	PendingRequests []struct {
		// Upcount is the "upcount" communicated to modem.
		Upcount uint8 `json:"upcount"`
		// Updelay is the "updelay" communicated to modem.
		Updelay  uint8            `json:"updelay"`
		Requests []PendingRequest `json:"requests"`
	} `json:"pending_requests"`
	// LogMessages contains the from service related to this device.
	LogMessages []LogMessage `json:"log_messages"`
	// UploadedFiles contains the history of uploaded files.
	UploadedFiles []File `json:"uploaded_files"`
	// UploadedStreamRecords contains the history of uploaded records.
	UploadedStreamRecords []Stream `json:"uploaded_stream_records"`
	// LastUplink is the last handled uplink.
	LastUplink LoRaUplink `json:"last_uplink"`
}

// DeviceUplinkResponse contains the uplink response and the error if applicable.
type DeviceUplinkResponse struct {
	Result UplinkResponse `json:"result"`
	Error  string         `json:"error"`
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

// UplinkResponse contains the state changes and completed items due to an uplink message.
type UplinkResponse struct {
	File              *File            `json:"file"`
	StreamRecords     []Stream         `json:"stream_records"`
	FullfiledRequests []PendingRequest `json:"fulfilled_requests"`
	Downlink          *LoRaDnlink      `json:"dnlink"`
	InfoFields        InfoFields       `json:"info_fields"`
	LogMessages       []LogMessage     `json:"log_messages"`
}

// InfoFields contains the value of the various information fields and the timestamp of their last update.
type InfoFields struct {
	Status *struct {
		Timestamp float64     `json:"timestamp"`
		Value     StatusField `json:"value"`
	} `json:"status"`
	Charge *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"charge"`
	Voltage *struct {
		Timestamp float64 `json:"timestamp"`
		Value     float64 `json:"value"`
	} `json:"voltage"`
	Temp *struct {
		Timestamp float64 `json:"timestamp"`
		Value     int32   `json:"value"`
	} `json:"temp"`
	Signal *struct {
		Timestamp float64 `json:"timestamp"`
		Value     struct {
			RSSI float64 `json:"rssi"`
			SNR  float64 `json:"snr"`
		} `json:"value"`
	} `json:"signal"`
	Uptime *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"uptime"`
	RxTime *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"rxtime"`
	Firmware *struct {
		Timestamp float64 `json:"timestamp"`
		Value     struct {
			FwCRC string `json:"fwcrc"`
			FwCnt uint16 `json:"fwcnt"`
		} `json:"value"`
	} `json:"firmware"`
	ADRMode *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint8   `json:"value"`
	} `json:"adrmode"`
	JoinEUI *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint8   `json:"value"`
	} `json:"joineui"`
	Interval *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint8   `json:"value"`
	} `json:"interval"`
	Region *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint8   `json:"value"`
	} `json:"region"`
	OpMode *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint32  `json:"value"`
	} `json:"opmode"`
	CrashLog *struct {
		Timestamp float64 `json:"timestamp"`
		Value     string  `json:"value"`
	} `json:"crashlog"`
	RstCount *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"rstcount"`
	DevEUI *struct {
		Timestamp float64 `json:"timestamp"`
		Value     string  `json:"value"`
	} `json:"deveui"`
	FactRst *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"factrst"`
	Session *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"session"`
	ChipEUI *struct {
		Timestamp float64 `json:"timestamp"`
		Value     uint16  `json:"value"`
	} `json:"chipeui"`
	StreamPar *struct {
		Timestamp float64   `json:"timestamp"`
		Value     StreamPar `json:"value"`
	} `json:"streampar"`
	AppStatus *struct {
		Timestamp float64 `json:"timestamp"`
		Value     string  `json:"value"`
	} `json:"appstatus"`
}

// StatusField contains the status flags of a device.
type StatusField struct {
	Brownout bool `json:"brownout"`
	Crash    bool `json:"crash"`
	Mute     bool `json:"mute"`
	Joined   bool `json:"joined"`
	Suspend  bool `json:"suspend"`
	Upload   bool `json:"upload"`
}

// StreamPar contains te
type StreamPar struct {
	Port    uint8 `json:"port"`
	EncMode bool  `json:"encmode"`
}

// PendingRequest encapsulates a pending request.
type PendingRequest struct {
	ID        uint32      `json:"id"`
	Cookie    interface{} `json:"cookie"`
	Timestamp float64     `json:"timestamp"`
	Request   Request     `json:"request"`
}

type requestParam interface {
	isRequestParam()
}

// RequestType is the type of a modem request.
type RequestType uint8

const (
	// ResetRequestType identifies a RESET request.
	ResetRequestType RequestType = iota
	// RejoinRequestType identifiers a REJOIN request.
	RejoinRequestType
	// MuteRequestType identifies a MUTE request.
	MuteRequestType
	// GetInfoRequestType identifies a GETINFO request.
	GetInfoRequestType
	// SetConfRequestType identifies a SETCONF request.
	SetConfRequestType
	// FileDoneRequestType identifies a FILEDONE request.
	FileDoneRequestType
	// FUOTARequestType identifies a FUOTA request.
	FUOTARequestType
)

// MarshalJSON implements the json.Marshaler interface.
func (t RequestType) MarshalJSON() ([]byte, error) {
	var tp string
	switch t {
	case ResetRequestType:
		tp = resetRequestType
	case RejoinRequestType:
		tp = rejoinRequestType
	case MuteRequestType:
		tp = muteRequestType
	case GetInfoRequestType:
		tp = getInfoRequestType
	case SetConfRequestType:
		tp = setConfRequestType
	case FileDoneRequestType:
		tp = fileDoneRequestType
	case FUOTARequestType:
		tp = fuotaRequestType
	default:
		panic(fmt.Sprintf("RequestType %v is unsupported", t))
	}
	return json.Marshal(tp)
}

// UnmarshalJSON implements the json.Unarmshaler.
func (t *RequestType) UnmarshalJSON(b []byte) error {
	var tp string
	err := json.Unmarshal(b, &tp)
	if err != nil {
		return err
	}
	switch tp {
	case resetRequestType:
		*t = ResetRequestType
	case rejoinRequestType:
		*t = RejoinRequestType
	case muteRequestType:
		*t = MuteRequestType
	case getInfoRequestType:
		*t = GetInfoRequestType
	case setConfRequestType:
		*t = SetConfRequestType
	case fileDoneRequestType:
		*t = FileDoneRequestType
	case fuotaRequestType:
		*t = FUOTARequestType
	default:
		panic(fmt.Sprintf("RequestType %v is unsupported", t))
	}
	return nil
}

// Request encapsulates a modem request.
type Request struct {
	Type  RequestType  `json:"type"`
	Param requestParam `json:"param,omitempty"`
}

const (
	resetRequestType    = "RESET"
	rejoinRequestType   = "REJOIN"
	muteRequestType     = "MUTE"
	getInfoRequestType  = "GETINFO"
	setConfRequestType  = "SETCONF"
	fileDoneRequestType = "FILEDONE"
	fuotaRequestType    = "FUOTA"
)

type baseRequest struct {
	Type  RequestType     `json:"type"`
	Param json.RawMessage `json:"param"`
}

var errInvalidRequestType = errors.DefineInvalidArgument("invalid_request_type", "request type `{type}` is invalid")

// UnmarshalJSON implements json.Unmarshaler.
func (r *Request) UnmarshalJSON(b []byte) error {
	var br baseRequest
	err := json.Unmarshal(b, &br)
	if err != nil {
		return err
	}
	r.Type = br.Type
	switch br.Type {
	case ResetRequestType:
		var p ResetRequestParam
		r.Param = &p
	case RejoinRequestType, MuteRequestType:
		return nil
	case GetInfoRequestType:
		var p GetInfoRequestParam
		r.Param = &p
	case SetConfRequestType:
		var p SetConfRequestParam
		r.Param = &p
	case FileDoneRequestType:
		var p FileDoneRequestParam
		r.Param = &p
	case FUOTARequestType:
		var p FUOTARequestParam
		r.Param = &p
	default:
		return errInvalidRequestType.WithAttributes("type", br.Type)
	}
	return json.Unmarshal(br.Param, r.Param)
}

// ResetRequestParam is the reset type of a Reset request.
type ResetRequestParam uint8

func (ResetRequestParam) isRequestParam() {}

// GetInfoRequestParam contains the requested fields of a GetInfo request.
type GetInfoRequestParam []string

func (GetInfoRequestParam) isRequestParam() {}

// SetConfRequestParam is the configuration request of a SetConf request.
type SetConfRequestParam struct {
	ADRMode  *uint8  `json:"adrmode,omitempty"`
	JoinEUI  EUI     `json:"joineui,omitempty"`
	Interval *uint8  `json:"interval,omitempty"`
	Region   *uint8  `json:"region,omitempty"`
	OpMode   *uint32 `json:"opmode,omitempty"`
}

func (SetConfRequestParam) isRequestParam() {}

// FileDoneRequestParam is the configuration of a file upload request of a FileDone request.
type FileDoneRequestParam struct {
	SID  int32 `json:"sid"`
	SCtr int32 `json:"sctr"`
}

func (FileDoneRequestParam) isRequestParam() {}

// FUOTARequestParam is the FUOTA binary chunk of a FUOTA request.
type FUOTARequestParam Hex

// MarshalJSON implements json.Marshaler.
func (f FUOTARequestParam) MarshalJSON() ([]byte, error) {
	return Hex(f).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *FUOTARequestParam) UnmarshalJSON(b []byte) error {
	h := Hex{}
	err := h.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	*f = FUOTARequestParam(h)
	return nil
}

func (FUOTARequestParam) isRequestParam() {}

// UploadSession contains information for an active, obsolete or finished file upload.
type UploadSession struct {
	State  string   `json:"string"`
	SID    uint8    `json:"sid"`
	SCtr   uint8    `json:"sctr"`
	CCt    uint16   `json:"cct"`
	CSz    uint8    `json:"csz"`
	Chunks []string `json:"chunks"`
}

// StreamSession contains information for a defined streaming session.
type StreamSession struct {
	Port    uint8   `json:"port"`
	Decoder *string `json:"decoder,omitempty"`
}

// DeviceUplinks maps device EUIs to LoRaUplink
type DeviceUplinks map[EUI]LoRaUplink

// MarshalJSON implements json.Marshaler.
func (u DeviceUplinks) MarshalJSON() ([]byte, error) {
	m := make(map[string]LoRaUplink)
	for k, v := range u {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

// LoRaUplink encapsulates the information of a LoRa message.
type LoRaUplink struct {
	FCnt      uint32  `json:"fcnt"`
	Port      uint8   `json:"port"`
	Payload   Hex     `json:"payload"`
	DR        uint8   `json:"dr"`
	Freq      uint32  `json:"freq"`
	Timestamp float64 `json:"timestamp"`
}

// LoRaDnlink is a specification for a modem device.
type LoRaDnlink struct {
	Port    uint8 `json:"port"`
	Payload Hex   `json:"payload"`
}

// File carries the contents of an uploaded, defragmented file by a modem to the service.
type File struct {
	SCtr      uint8   `json:"sctr"`
	Timestamp float64 `json:"timestamp"`
	Port      uint8   `json:"port"`
	Data      Hex     `json:"string"`
	Hash      Hex     `json:"hash"`
	EncMode   bool    `json:"encmode"`
	Message   *string `json:"message,omitempty"`
}

// Stream is a fully assembled record of the stream session stored
// in the stream record history with the device state.
type Stream struct {
	Timestamp float64 `json:"timestamp"`
	Port      uint8   `json:"port"`
	Data      Hex     `json:"data"`
	Off       uint16  `json:"off"`
}

// LogMessage is a log message associated with a device.
type LogMessage struct {
	LogMsg    string  `json:"logmsg"`
	Level     string  `json:"level"`
	Timestamp float64 `json:"timestamp"`
}

// DeviceSettings holds the initial settings for a device.
type DeviceSettings struct {
	DMPorts *int32 `json:"dmports,omitempty"`
	Streams []struct {
		Port  int32 `json:"port"`
		AppSz int32 `json:"appsz"`
	} `json:"streams,omitempty"`
}

// TokenInfo holds token information.
type TokenInfo struct {
	Name         string   `json:"name"`
	Token        string   `json:"token"`
	Capabilities []string `json:"capabilities"`
}

// Hex represents hex encoded bytes.
type Hex []byte

// MarshalJSON implements json.Marshaler.
func (h Hex) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", hex.EncodeToString(h))), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *Hex) UnmarshalJSON(b []byte) (err error) {
	s := strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\"")
	*h, err = hex.DecodeString(s)
	return
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
	return
}
