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
	"encoding/json"
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// MessageType is the message type.
type MessageType string

const (
	MessageTypeJoinReq      MessageType = "JoinReq"
	MessageTypeJoinAns      MessageType = "JoinAns"
	MessageTypeRejoinReq    MessageType = "RejoinReq"
	MessageTypeRejoinAns    MessageType = "RejoinAns"
	MessageTypeAppSKeyReq   MessageType = "AppSKeyReq"
	MessageTypeAppSKeyAns   MessageType = "AppSKeyAns"
	MessageTypePRStartReq   MessageType = "PRStartReq"
	MessageTypePRStartAns   MessageType = "PRStartAns"
	MessageTypePRStartNotif MessageType = "PRStartNotif"
	MessageTypePRStopReq    MessageType = "PRStopReq"
	MessageTypePRStopAns    MessageType = "PRStopAns"
	MessageTypeHRStartReq   MessageType = "HRStartReq"
	MessageTypeHRStartAns   MessageType = "HRStartAns"
	MessageTypeHRStopReq    MessageType = "HRStopReq"
	MessageTypeHRStopAns    MessageType = "HRStopAns"
	MessageTypeHomeNSReq    MessageType = "HomeNSReq"
	MessageTypeHomeNSAns    MessageType = "HomeNSAns"
	MessageTypeProfileReq   MessageType = "ProfileReq"
	MessageTypeProfileAns   MessageType = "ProfileAns"
	MessageTypeXmitDataReq  MessageType = "XmitDataReq"
	MessageTypeXmitDataAns  MessageType = "XmitDataAns"
	MessageTypeXmitLocReq   MessageType = "XmitLocReq"
	MessageTypeXmitLocAns   MessageType = "XmitLocAns"
	MessageTypeErrorNotif   MessageType = "ErrorNotif"
)

// Result is the result of an answer message.
type Result string

const (
	ResultSuccess                Result = "Success"
	ResultNoAction               Result = "NoAction"
	ResultMICFailed              Result = "MICFailed"
	ResultFrameReplayed          Result = "FrameReplayed"
	ResultJoinReqFailed          Result = "JoinReqFailed"
	ResultNoRoamingAgreement     Result = "NoRoamingAgreement"
	ResultDevRoamingDisallowed   Result = "DevRoamingDisallowed"
	ResultRoamingActDisallowed   Result = "RoamingActDisallowed"
	ResultActivationDisallowed   Result = "ActivationDisallowed"
	ResultUnknownDevEUI          Result = "UnknownDevEUI"
	ResultUnknownDevAddr         Result = "UnknownDevAddr"
	ResultUnknownSender          Result = "UnknownSender"
	ResultUnknownReceiver        Result = "UnknownReceiver"
	ResultDeferred               Result = "Deferred"
	ResultXmitFailed             Result = "XmitFailed"
	ResultInvalidFPort           Result = "InvalidFPort"
	ResultInvalidProtocolVersion Result = "InvalidProtocolVersion"
	ResultStaleDeviceProfile     Result = "StaleDeviceProfile"
	ResultMalformedMessage       Result = "MalformedMessage"
	ResultFrameSizeError         Result = "FrameSizeError"
	ResultOther                  Result = "Other"
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

// KeyEnvelope contains a (encrypted) key.
type KeyEnvelope ttnpb.KeyEnvelope

// MarshalJSON marshals the key envelope to JSON.
func (k KeyEnvelope) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		KEKLabel string
		AESKey   Buffer
	}{
		KEKLabel: k.KEKLabel,
		AESKey:   Buffer(k.EncryptedKey),
	})
}

// UnmarshalJSON unmarshals the key envelope from JSON.
func (k *KeyEnvelope) UnmarshalJSON(data []byte) error {
	aux := struct {
		KEKLabel string
		AESKey   Buffer
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*k = KeyEnvelope{
		KEKLabel:     aux.KEKLabel,
		EncryptedKey: aux.AESKey,
	}
	return nil
}
