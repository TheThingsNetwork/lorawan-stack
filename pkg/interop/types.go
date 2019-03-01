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

package interop

import (
	"encoding/hex"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// MessageType is the message type.
type MessageType string

const (
	MessageTypeJoinReq     MessageType = "JoinReq"
	MessageTypeJoinAns                 = "JoinAns"
	MessageTypeRejoinReq               = "RejoinReq"
	MessageTypeRejoinAns               = "RejoinAns"
	MessageTypePRStartReq              = "PRStartReq"
	MessageTypePRStartAns              = "PRStartAns"
	MessageTypePRStopReq               = "PRStopReq"
	MessageTypePRStopAns               = "PRStopAns"
	MessageTypeHRStartReq              = "HRStartReq"
	MessageTypeHRStartAns              = "HRStartAns"
	MessageTypeHRStopReq               = "HRStopReq"
	MessageTypeHRStopAns               = "HRStopAns"
	MessageTypeHomeNSReq               = "HomeNSReq"
	MessageTypeHomeNSAns               = "HomeNSAns"
	MessageTypeProfileReq              = "ProfileReq"
	MessageTypeProfileAns              = "ProfileAns"
	MessageTypeXmitDataReq             = "XmitDataReq"
	MessageTypeXmitDataAns             = "XmitDataAns"
)

// Result is the result of an answer message.
type Result string

const (
	ResultSuccess                = "Success"
	ResultNoAction               = "NoAction"
	ResultMICFailed              = "MICFailed"
	ResultFrameReplayed          = "FrameReplayed"
	ResultJoinReqFailed          = "JoinReqFailed"
	ResultNoRoamingAgreement     = "NoRoamingAgreement"
	ResultDevRoamingDisallowed   = "DevRoamingDisallowed"
	ResultRoamingActDisallowed   = "RoamingActDisallowed"
	ResultActivationDisallowed   = "ActivationDisallowed"
	ResultUnknownDevEUI          = "UnknownDevEUI"
	ResultUnknownDevAddr         = "UnknownDevAddr"
	ResultUnknownSender          = "UnknownSender"
	ResultUnknownReceiver        = "UnknownReceiver"
	ResultDeferred               = "Deferred"
	ResultXmitFailed             = "XmitFailed"
	ResultInvalidFPort           = "InvalidFPort"
	ResultInvalidProtocolVersion = "InvalidProtocolVersion"
	ResultStaleDeviceProfile     = "StaleDeviceProfile"
	ResultMalformedMessage       = "MalformedMessage"
	ResultFrameSizeError         = "FrameSizeError"
	ResultOther                  = "Other"
)

// MACVersion is the MAC version.
type MACVersion ttnpb.MACVersion

// MarshalJSON marshals the version to text format.
func (v MACVersion) MarshalJSON() ([]byte, error) {
	var res string
	switch ttnpb.MACVersion(v) {
	case ttnpb.MAC_V1_0:
		res = "1.0"
	case ttnpb.MAC_V1_0_1:
		res = "1.0.1"
	case ttnpb.MAC_V1_0_2:
		res = "1.0.2"
	case ttnpb.MAC_V1_0_3:
		res = "1.0.2"
	case ttnpb.MAC_V1_1:
		res = "1.1"
	default:
		return nil, errUnknownMACVersion
	}
	return []byte(fmt.Sprintf(`"%s"`, res)), nil
}

// UnmarshalJSON unmarshals a version in text format.
func (v *MACVersion) UnmarshalJSON(data []byte) error {
	var res ttnpb.MACVersion
	switch strings.Trim(string(data), `"`) {
	case "1.0":
		res = ttnpb.MAC_V1_0
	case "1.0.1":
		res = ttnpb.MAC_V1_0_1
	case "1.0.2":
		res = ttnpb.MAC_V1_0_2
	case "1.0.3":
		res = ttnpb.MAC_V1_0_3
	case "1.1":
		res = ttnpb.MAC_V1_1
	default:
		return errUnknownMACVersion
	}
	*v = MACVersion(res)
	return nil
}

// Buffer contains binary data.
type Buffer []byte

// MarshalJSON marshals the binary data to a hexadecimal string.
func (b Buffer) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, hex.EncodeToString(b))), nil
}

// UnmarshalJSON unmarshals a hexadecimal string to binary data.
func (b *Buffer) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errInvalidLength
	}
	buf, err := hex.DecodeString(strings.TrimPrefix(string(data[1:len(data)-1]), "0x"))
	if err != nil {
		return err
	}
	*b = Buffer(buf)
	return nil
}

// KeyEnvelope contains a wrapped session key.
type KeyEnvelope struct {
	KEKLabel string
	AESKey   Buffer
}
