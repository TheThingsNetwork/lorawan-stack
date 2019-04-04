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

package ttnpb

import (
	"strconv"

	"github.com/blang/semver"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

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
	case MAC_V1_0_3:
		return "1.0.3"
	case MAC_V1_1:
		return "1.1.0"
	}
	return "unknown"
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MACVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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
	case MAC_V1_0_3.String():
		*v = MAC_V1_0_3
	case MAC_V1_1.String():
		*v = MAC_V1_1
	case MAC_UNKNOWN.String():
		*v = MAC_UNKNOWN
	default:
		return errCouldNotParse("MACVersion")(string(b))
	}
	return nil
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
	case MAC_V1_0, MAC_V1_0_1, MAC_V1_0_2, MAC_V1_0_3:
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
	case MAC_V1_0, MAC_V1_0_1, MAC_V1_0_2, MAC_V1_0_3:
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
	case PHY_V1_0_3_REV_A:
		return "1.0.3-a"
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
	case PHY_V1_0_3_REV_A.String():
		*v = PHY_V1_0_3_REV_A
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

// String implements fmt.Stringer.
func (v DataRateIndex) String() string {
	return strconv.Itoa(int(v))
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DataRateIndex) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DataRateIndex) UnmarshalText(b []byte) error {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("DataRateIndex")(string(b)).WithCause(err)
	}
	if i > int(DATA_RATE_15) {
		return errFieldHasMax.WithAttributes(
			"lorawan_field", "DataRateIndex",
			"max", DATA_RATE_15,
		)
	}
	*v = DataRateIndex(i)
	return nil
}

// String implements fmt.Stringer.
func (v RxDelay) String() string {
	return strconv.Itoa(int(v))
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RxDelay) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RxDelay) UnmarshalText(b []byte) error {
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("RxDelay")(string(b)).WithCause(err)
	}
	if i > int(RX_DELAY_15) {
		return errFieldHasMax.WithAttributes(
			"lorawan_field", "RxDelay",
			"max", RX_DELAY_15,
		)
	}
	*v = RxDelay(i)
	return nil
}
