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

package interop

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// ProtocolVersion is the LoRaWAN Backend Interfaces protocol version.
type ProtocolVersion string

const (
	// ProtocolV1_0 is LoRaWAN Backend Interfaces (TS002) 1.0.
	ProtocolV1_0 = "1.0"
	// ProtocolV1_1 is LoRaWAN Backend Interfaces (TS002) 1.1.x.
	ProtocolV1_1 = "1.1"
)

var errUnknownProtocol = errors.DefineInvalidArgument("unknown_protocol", "unknown protocol")

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *ProtocolVersion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	switch s {
	case "BI1.0":
		*p = ProtocolV1_0
		return nil
	case "BI1.1":
		*p = ProtocolV1_1
		return nil
	default:
		return errUnknownProtocol.New()
	}
}

// SupportsNSID returns true if the protocol version supports Network Server IDs (NSID).
func (p ProtocolVersion) SupportsNSID() bool {
	switch p {
	case ProtocolV1_1:
		return true
	default:
		return false
	}
}

// MessageType is the message type.
type MessageType string

// LoRaWAN Backend Interfaces message types.
const (
	MessageTypeJoinReq     MessageType = "JoinReq"
	MessageTypeJoinAns     MessageType = "JoinAns"
	MessageTypeRejoinReq   MessageType = "RejoinReq"
	MessageTypeRejoinAns   MessageType = "RejoinAns"
	MessageTypeAppSKeyReq  MessageType = "AppSKeyReq"
	MessageTypeAppSKeyAns  MessageType = "AppSKeyAns"
	MessageTypePRStartReq  MessageType = "PRStartReq"
	MessageTypePRStartAns  MessageType = "PRStartAns"
	MessageTypePRStopReq   MessageType = "PRStopReq"
	MessageTypePRStopAns   MessageType = "PRStopAns"
	MessageTypeHRStartReq  MessageType = "HRStartReq"
	MessageTypeHRStartAns  MessageType = "HRStartAns"
	MessageTypeHRStopReq   MessageType = "HRStopReq"
	MessageTypeHRStopAns   MessageType = "HRStopAns"
	MessageTypeHomeNSReq   MessageType = "HomeNSReq"
	MessageTypeHomeNSAns   MessageType = "HomeNSAns"
	MessageTypeProfileReq  MessageType = "ProfileReq"
	MessageTypeProfileAns  MessageType = "ProfileAns"
	MessageTypeXmitDataReq MessageType = "XmitDataReq"
	MessageTypeXmitDataAns MessageType = "XmitDataAns"
)

var requestAnswers = map[MessageType]MessageType{
	MessageTypeJoinReq:     MessageTypeJoinAns,
	MessageTypeRejoinReq:   MessageTypeRejoinAns,
	MessageTypeAppSKeyReq:  MessageTypeAppSKeyAns,
	MessageTypePRStartReq:  MessageTypePRStartAns,
	MessageTypePRStopReq:   MessageTypePRStopAns,
	MessageTypeHRStartReq:  MessageTypeHRStartAns,
	MessageTypeHRStopReq:   MessageTypeHRStopAns,
	MessageTypeHomeNSReq:   MessageTypeHomeNSAns,
	MessageTypeProfileReq:  MessageTypeProfileAns,
	MessageTypeXmitDataReq: MessageTypeXmitDataAns,
}

var protocolMessageTypes = map[ProtocolVersion]map[MessageType]bool{
	ProtocolV1_0: {
		MessageTypeJoinReq:     true,
		MessageTypeJoinAns:     true,
		MessageTypeAppSKeyReq:  true,
		MessageTypeAppSKeyAns:  true,
		MessageTypePRStartReq:  true,
		MessageTypePRStartAns:  true,
		MessageTypePRStopReq:   true,
		MessageTypePRStopAns:   true,
		MessageTypeHRStartReq:  true,
		MessageTypeHRStartAns:  true,
		MessageTypeHRStopReq:   true,
		MessageTypeHRStopAns:   true,
		MessageTypeHomeNSReq:   true,
		MessageTypeHomeNSAns:   true,
		MessageTypeProfileReq:  true,
		MessageTypeProfileAns:  true,
		MessageTypeXmitDataReq: true,
		MessageTypeXmitDataAns: true,
	},
	ProtocolV1_1: {
		MessageTypeJoinReq:     true,
		MessageTypeJoinAns:     true,
		MessageTypeRejoinReq:   true,
		MessageTypeRejoinAns:   true,
		MessageTypeAppSKeyReq:  true,
		MessageTypeAppSKeyAns:  true,
		MessageTypePRStartReq:  true,
		MessageTypePRStartAns:  true,
		MessageTypePRStopReq:   true,
		MessageTypePRStopAns:   true,
		MessageTypeHRStartReq:  true,
		MessageTypeHRStartAns:  true,
		MessageTypeHRStopReq:   true,
		MessageTypeHRStopAns:   true,
		MessageTypeHomeNSReq:   true,
		MessageTypeHomeNSAns:   true,
		MessageTypeProfileReq:  true,
		MessageTypeProfileAns:  true,
		MessageTypeXmitDataReq: true,
		MessageTypeXmitDataAns: true,
	},
}

// IsRequest returns whether the message type is a request that has an answer message type defined.
func (m MessageType) IsRequest() bool {
	_, ok := requestAnswers[m]
	return ok
}

// Answer returns the answer message type.
// If the message type is not a request, this method returns false.
func (m MessageType) Answer() (MessageType, bool) {
	ans, ok := requestAnswers[m]
	return ans, ok
}

// Validate returns an error if the message type is not valid for the given protocol version.
func (m MessageType) Validate(version ProtocolVersion) error {
	protocolMessageTypes, ok := protocolMessageTypes[version]
	if !ok {
		return ErrProtocolVersion.New()
	}
	for mt := range protocolMessageTypes {
		if mt == m {
			return nil
		}
	}
	return ErrMalformedMessage.New()
}

// ToJoinServer indicates whether the message goes to a Join Server.
// If this method returns true, the message's ReceiverID must be a JoinEUI.
func (m MessageType) ToJoinServer() bool {
	switch m {
	case MessageTypeJoinReq,
		MessageTypeRejoinReq,
		MessageTypeAppSKeyReq,
		MessageTypeHomeNSReq:
		return true
	default:
		return false
	}
}

// ResultCode is the result of an answer message.
type ResultCode string

const (
	ResultSuccess              ResultCode = "Success"
	ResultNoAction             ResultCode = "NoAction"
	ResultMICFailed            ResultCode = "MICFailed"
	ResultFrameReplayed        ResultCode = "FrameReplayed"
	ResultJoinReqFailed        ResultCode = "JoinReqFailed"
	ResultNoRoamingAgreement   ResultCode = "NoRoamingAgreement"
	ResultDevRoamingDisallowed ResultCode = "DevRoamingDisallowed"
	ResultRoamingActDisallowed ResultCode = "RoamingActDisallowed"
	ResultActivationDisallowed ResultCode = "ActivationDisallowed"
	ResultUnknownDevEUI        ResultCode = "UnknownDevEUI"
	ResultUnknownDevAddr       ResultCode = "UnknownDevAddr"
	ResultUnknownSender        ResultCode = "UnknownSender"
	ResultUnkownReceiver       ResultCode = "UnkownReceiver" // sic
	// ResultUnknownReceiver is not specified in LoRaWAN Backend Interfaces 1.0 but is processed as ResultUnkownReceiver.
	ResultUnknownReceiver        ResultCode = "UnknownReceiver"
	ResultDeferred               ResultCode = "Deferred"
	ResultXmitFailed             ResultCode = "XmitFailed"
	ResultInvalidFPort           ResultCode = "InvalidFPort"
	ResultInvalidProtocolVersion ResultCode = "InvalidProtocolVersion"
	ResultStaleDeviceProfile     ResultCode = "StaleDeviceProfile"
	ResultMalformedRequest       ResultCode = "MalformedRequest"
	ResultMalformedMessage       ResultCode = "MalformedMessage"
	ResultFrameSizeError         ResultCode = "FrameSizeError"
	ResultOther                  ResultCode = "Other"
)

// MACVersion is the MAC version.
type MACVersion ttnpb.MACVersion

// MarshalText implements encoding.TextMarshaler.
func (v MACVersion) MarshalText() ([]byte, error) {
	var res string
	switch ttnpb.MACVersion(v) {
	case ttnpb.MACVersion_MAC_V1_0:
		res = "1.0"
	case ttnpb.MACVersion_MAC_V1_0_1:
		res = "1.0.1"
	case ttnpb.MACVersion_MAC_V1_0_2:
		res = "1.0.2"
	case ttnpb.MACVersion_MAC_V1_0_3:
		res = "1.0.3"
	case ttnpb.MACVersion_MAC_V1_0_4:
		res = "1.0.4"
	case ttnpb.MACVersion_MAC_V1_1:
		res = "1.1"
	default:
		return nil, errUnknownMACVersion.New()
	}
	return []byte(res), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (v *MACVersion) UnmarshalText(data []byte) error {
	var res ttnpb.MACVersion
	switch string(data) {
	case "1.0":
		res = ttnpb.MACVersion_MAC_V1_0
	case "1.0.1":
		res = ttnpb.MACVersion_MAC_V1_0_1
	case "1.0.2":
		res = ttnpb.MACVersion_MAC_V1_0_2
	case "1.0.3":
		res = ttnpb.MACVersion_MAC_V1_0_3
	case "1.0.4":
		res = ttnpb.MACVersion_MAC_V1_0_4
	case "1.1":
		res = ttnpb.MACVersion_MAC_V1_1
	default:
		return errUnknownMACVersion.New()
	}
	*v = MACVersion(res)
	return nil
}

// Buffer is a binary buffer that is represented as hexadecimal in text.
type Buffer []byte

// MarshalText implements encoding.TextMarshaler.
func (b Buffer) MarshalText() ([]byte, error) {
	return []byte(strings.ToUpper(hex.EncodeToString(b))), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Buffer) UnmarshalText(data []byte) error {
	buf, err := hex.DecodeString(strings.TrimPrefix(string(data), "0x"))
	if err != nil {
		return err
	}
	*b = Buffer(buf)
	return nil
}

// VendorID is an IEEE MAC-L (OUI) assignment to indicate the vendor.
type VendorID [3]byte

// MarshalText implements encoding.TextMarshaler.
func (v VendorID) MarshalText() ([]byte, error) {
	return Buffer(v[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (v *VendorID) UnmarshalText(data []byte) error {
	var buf Buffer
	if err := buf.UnmarshalText(data); err != nil {
		return errInvalidVendorID.WithCause(err)
	}
	if len(buf) != 3 {
		return errInvalidVendorID.WithCause(errInvalidLength.New())
	}
	copy(v[:], buf)
	return nil
}

// KeyEnvelope contains a (encrypted) key.
type KeyEnvelope ttnpb.KeyEnvelope

// MarshalJSON marshals the key envelope to JSON.
func (k KeyEnvelope) MarshalJSON() ([]byte, error) {
	var key []byte
	if k.KekLabel != "" {
		key = k.EncryptedKey
	} else if k.Key != nil {
		key = k.Key[:]
	}
	return json.Marshal(struct {
		KEKLabel string
		AESKey   Buffer
	}{
		KEKLabel: k.KekLabel,
		AESKey:   Buffer(key),
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
	var key *types.AES128Key
	var encryptedKey []byte
	if aux.KEKLabel != "" {
		encryptedKey = aux.AESKey
	} else {
		key = new(types.AES128Key)
		copy(key[:], aux.AESKey)
	}
	*k = KeyEnvelope{
		KekLabel:     aux.KEKLabel,
		Key:          key,
		EncryptedKey: encryptedKey,
	}
	return nil
}

// NetID is a LoRaWAN NetID.
type NetID types.NetID

// MarshalJSON marshals the NetID to text.
func (n NetID) MarshalText() ([]byte, error) {
	return Buffer(n[:]).MarshalText()
}

// UnmarshalText unmarshals the NetID from text.
func (n *NetID) UnmarshalText(data []byte) error {
	var buf Buffer
	if err := buf.UnmarshalText(data); err != nil {
		return err
	}
	if len(buf) != 3 {
		return errInvalidLength.New()
	}
	copy(n[:], buf)
	return nil
}

// EUI64 is an 64-bit EUI, e.g. a DevEUI or JoinEUI.
type EUI64 types.EUI64

// MarshalText implements encoding.TextMarshaler.
func (n EUI64) MarshalText() ([]byte, error) {
	return Buffer(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *EUI64) UnmarshalText(data []byte) error {
	var buf Buffer
	if err := buf.UnmarshalText(data); err != nil {
		return err
	}
	if len(buf) != 8 {
		return errInvalidLength.New()
	}
	copy(n[:], buf)
	return nil
}

// DevAddr is a LoRaWAN DevAddr.
type DevAddr types.DevAddr

// MarshalText implements encoding.TextMarshaler.
func (n DevAddr) MarshalText() ([]byte, error) {
	return Buffer(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *DevAddr) UnmarshalText(data []byte) error {
	var buf Buffer
	if err := buf.UnmarshalText(data); err != nil {
		return err
	}
	if len(buf) != 4 {
		return errInvalidLength.New()
	}
	copy(n[:], buf)
	return nil
}
