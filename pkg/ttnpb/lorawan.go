// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import (
	"encoding/gob"
	"strconv"

	"github.com/blang/semver"
	"github.com/gogo/protobuf/jsonpb"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

func init() {
	for _, v := range []interface{}{
		&Message_MACPayload{},
		&Message_JoinRequestPayload{},
		&Message_RejoinRequestPayload{},
		&Message_JoinAcceptPayload{},
		&MACCommand{},
		&MACCommand_RawPayload{},
		&MACCommand_ResetInd{},
		&MACCommand_ResetInd_{},
		&MACCommand_ResetConf{},
		&MACCommand_ResetConf_{},
		&MACCommand_LinkCheckAns{},
		&MACCommand_LinkCheckAns_{},
		&MACCommand_LinkADRReq{},
		&MACCommand_LinkADRReq_{},
		&MACCommand_LinkADRAns{},
		&MACCommand_LinkADRAns_{},
		&MACCommand_DutyCycleReq_{},
		&MACCommand_RxParamSetupReq_{},
		&MACCommand_RxParamSetupAns_{},
		&MACCommand_DevStatusAns_{},
		&MACCommand_NewChannelReq_{},
		&MACCommand_NewChannelAns_{},
		&MACCommand_DlChannelReq{},
		&MACCommand_DlChannelAns{},
		&MACCommand_RxTimingSetupReq_{},
		&MACCommand_TxParamSetupReq_{},
		&MACCommand_RekeyInd_{},
		&MACCommand_RekeyConf_{},
		&MACCommand_ADRParamSetupReq{},
		&MACCommand_ADRParamSetupReq_{},
		&MACCommand_DeviceTimeAns_{},
		&MACCommand_ForceRejoinReq_{},
		&MACCommand_RejoinParamSetupReq_{},
		&MACCommand_RejoinParamSetupAns_{},
		&MACCommand_PingSlotInfoReq_{},
		&MACCommand_PingSlotChannelReq_{},
		&MACCommand_PingSlotChannelAns_{},
		&MACCommand_BeaconTimingAns_{},
		&MACCommand_BeaconFreqReq_{},
		&MACCommand_BeaconFreqAns_{},
		&MACCommand_DeviceModeInd_{},
		&MACCommand_DeviceModeConf_{},
	} {
		gob.Register(v)
	}
}

var errParsingSemanticVersion = unexpectedValue(
	errors.DefineInvalidArgument("parsing_semantic_version", "could not parse semantic version", valueKey),
)

// Validate reports whether v represents a valid MACVersion.
func (v MACVersion) Validate() error {
	if v < 1 || v >= MACVersion(len(MACVersion_name)) {
		return errExpectedBetween("MACVersion", 1, len(MACVersion_name)-1)(v)
	}

	_, err := semver.Parse(v.String())
	if err != nil {
		return errParsingSemanticVersion(v.String()).WithCause(err)
	}
	return nil
}

// String implements fmt.Stringer.
func (v MACVersion) String() string {
	switch v {
	case MAC_V1_0:
		return "1.0.0"
	case MAC_V1_0_1:
		return "1.0.1"
	case MAC_V1_0_2:
		return "1.0.2"
	case MAC_V1_1:
		return "1.1.0"
	}
	return "unknown"
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MACVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v MACVersion) MarshalJSON() ([]byte, error) {
	txt, err := v.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte("\"" + string(txt) + "\""), nil
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v MACVersion) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if m.EnumsAsInts {
		return []byte("\"" + strconv.Itoa(int(v)) + "\""), nil
	}
	return v.MarshalJSON()
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MACVersion) UnmarshalText(b []byte) error {
	switch string(b) {
	case MAC_V1_0.String():
		*v = MAC_V1_0
	case MAC_V1_0_1.String():
		*v = MAC_V1_0_1
	case MAC_V1_0_2.String():
		*v = MAC_V1_0_2
	case MAC_V1_0_2.String():
		*v = MAC_V1_0_2
	case MAC_V1_1.String():
		*v = MAC_V1_1
	case MAC_UNKNOWN.String():
		*v = MAC_UNKNOWN
	default:
		return errCouldNotParse("MACVersion")(string(b))
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MACVersion) UnmarshalJSON(b []byte) error {
	return v.UnmarshalText(b[1 : len(b)-1])
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *MACVersion) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
}

// Compare compares MACVersions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
// Compare panics, if v.Validate() returns non-nil error.
func (v MACVersion) Compare(o MACVersion) int {
	return semver.MustParse(v.String()).Compare(
		semver.MustParse(o.String()),
	)
}

// EncryptFOpts reports whether v requires MAC commands in FOpts to be encrypted.
// EncryptFOpts panics, if v.Validate() returns non-nil error.
func (v MACVersion) EncryptFOpts() bool {
	switch v {
	case MAC_V1_0, MAC_V1_0_1, MAC_V1_0_2:
		return false
	case MAC_V1_1:
		return true
	}
	panic(v.Validate())
}

// HasMaxFCntGap reports whether v defines a MaxFCntGap.
// HasMaxFCntGap panics, if v.Validate() returns non-nil error.
func (v MACVersion) HasMaxFCntGap() bool {
	switch v {
	case MAC_V1_0, MAC_V1_0_1, MAC_V1_0_2:
		return true
	case MAC_V1_1:
		return false
	}
	panic(v.Validate())
}

// String implements fmt.Stringer.
func (v PHYVersion) String() string {
	switch v {
	case PHY_V1_0:
		return "1.0.0"
	case PHY_V1_0_1:
		return "1.0.1"
	case PHY_V1_0_2_REV_A:
		return "1.0.2-a"
	case PHY_V1_0_2_REV_B:
		return "1.0.2-b"
	case PHY_V1_1_REV_A:
		return "1.1.0-a"
	case PHY_V1_1_REV_B:
		return "1.1.0-b"
	}
	return "unknown"
}

// MarshalText implements encoding.TextMarshaler interface.
func (v PHYVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v PHYVersion) MarshalJSON() ([]byte, error) {
	txt, err := v.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte("\"" + string(txt) + "\""), nil
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v PHYVersion) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if m.EnumsAsInts {
		return []byte("\"" + strconv.Itoa(int(v)) + "\""), nil
	}
	return v.MarshalJSON()
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PHYVersion) UnmarshalText(b []byte) error {
	switch string(b) {
	case PHY_V1_0.String():
		*v = PHY_V1_0
	case PHY_V1_0_1.String():
		*v = PHY_V1_0_1
	case PHY_V1_0_2_REV_A.String():
		*v = PHY_V1_0_2_REV_A
	case PHY_V1_0_2_REV_B.String():
		*v = PHY_V1_0_2_REV_B
	case PHY_V1_1_REV_A.String():
		*v = PHY_V1_1_REV_A
	case PHY_V1_1_REV_B.String():
		*v = PHY_V1_1_REV_B
	case PHY_UNKNOWN.String():
		*v = PHY_UNKNOWN
	default:
		return errCouldNotParse("PHYVersion")(string(b))
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PHYVersion) UnmarshalJSON(b []byte) error {
	return v.UnmarshalText(b[1 : len(b)-1])
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *PHYVersion) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
}
