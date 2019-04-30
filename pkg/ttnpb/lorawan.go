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
	"strings"

	"github.com/blang/semver"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// MarshalText implements encoding.TextMarshaler interface.
func (v MType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MType) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := MType_value[s]; ok {
		*v = MType(i)
		return nil
	}
	return errCouldNotParse("MType")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MType) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("MType")(string(b)).WithCause(err)
	}
	*v = MType(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v Major) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Major) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := Major_value[s]; ok {
		*v = Major(i)
		return nil
	}
	return errCouldNotParse("Major")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Major) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("Major")(string(b)).WithCause(err)
	}
	*v = Major(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MACVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MACVersion) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := MACVersion_value[s]; ok {
		*v = MACVersion(i)
		return nil
	}
	if !strings.HasPrefix(s, "MAC_") {
		if i, ok := MACVersion_value["MAC_"+s]; ok {
			*v = MACVersion(i)
			return nil
		}
	}
	return errCouldNotParse("MACVersion")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MACVersion) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("MACVersion")(string(b)).WithCause(err)
	}
	*v = MACVersion(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v PHYVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PHYVersion) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := PHYVersion_value[s]; ok {
		*v = PHYVersion(i)
		return nil
	}
	if !strings.HasPrefix(s, "PHY_") {
		if i, ok := PHYVersion_value["PHY_"+s]; ok {
			*v = PHYVersion(i)
			return nil
		}
	}
	return errCouldNotParse("PHYVersion")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PHYVersion) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("PHYVersion")(string(b)).WithCause(err)
	}
	*v = PHYVersion(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DataRateIndex) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v DataRateIndex) MarshalJSON() ([]byte, error) {
	return v.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DataRateIndex) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := DataRateIndex_value[s]; ok {
		*v = DataRateIndex(i)
		return nil
	}
	if !strings.HasPrefix(s, "DATA_RATE_") {
		if i, ok := DataRateIndex_value["DATA_RATE_"+s]; ok {
			*v = DataRateIndex(i)
			return nil
		}
	}
	return errCouldNotParse("DataRateIndex")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DataRateIndex) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("DataRateIndex")(string(b)).WithCause(err)
	}
	*v = DataRateIndex(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinType) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := RejoinType_value[s]; ok {
		*v = RejoinType(i)
		return nil
	}
	return errCouldNotParse("RejoinType")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinType) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("RejoinType")(string(b)).WithCause(err)
	}
	*v = RejoinType(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *CFListType) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := CFListType_value[s]; ok {
		*v = CFListType(i)
		return nil
	}
	return errCouldNotParse("CFListType")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *CFListType) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("CFListType")(string(b)).WithCause(err)
	}
	*v = CFListType(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v Class) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Class) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := Class_value[s]; ok {
		*v = Class(i)
		return nil
	}
	if !strings.HasPrefix(s, "CLASS_") {
		if i, ok := Class_value["CLASS_"+s]; ok {
			*v = Class(i)
			return nil
		}
	}
	return errCouldNotParse("Class")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Class) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("Class")(string(b)).WithCause(err)
	}
	*v = Class(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v TxSchedulePriority) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *TxSchedulePriority) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := TxSchedulePriority_value[s]; ok {
		*v = TxSchedulePriority(i)
		return nil
	}
	return errCouldNotParse("TxSchedulePriority")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *TxSchedulePriority) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("TxSchedulePriority")(string(b)).WithCause(err)
	}
	*v = TxSchedulePriority(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MACCommandIdentifier) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MACCommandIdentifier) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := MACCommandIdentifier_value[s]; ok {
		*v = MACCommandIdentifier(i)
		return nil
	}
	if !strings.HasPrefix(s, "CID_") {
		if i, ok := MACCommandIdentifier_value["CID_"+s]; ok {
			*v = MACCommandIdentifier(i)
			return nil
		}
	}
	return errCouldNotParse("MACCommandIdentifier")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MACCommandIdentifier) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("MACCommandIdentifier")(string(b)).WithCause(err)
	}
	*v = MACCommandIdentifier(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *AggregatedDutyCycle) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := AggregatedDutyCycle_value[s]; ok {
		*v = AggregatedDutyCycle(i)
		return nil
	}
	if !strings.HasPrefix(s, "DUTY_CYCLE_") {
		if i, ok := AggregatedDutyCycle_value["DUTY_CYCLE_"+s]; ok {
			*v = AggregatedDutyCycle(i)
			return nil
		}
	}
	return errCouldNotParse("AggregatedDutyCycle")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *AggregatedDutyCycle) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("AggregatedDutyCycle")(string(b)).WithCause(err)
	}
	*v = AggregatedDutyCycle(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v PingSlotPeriod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PingSlotPeriod) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := PingSlotPeriod_value[s]; ok {
		*v = PingSlotPeriod(i)
		return nil
	}
	if !strings.HasPrefix(s, "PING_EVERY_") {
		if i, ok := PingSlotPeriod_value["PING_EVERY_"+s]; ok {
			*v = PingSlotPeriod(i)
			return nil
		}
	}
	return errCouldNotParse("PingSlotPeriod")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PingSlotPeriod) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("PingSlotPeriod")(string(b)).WithCause(err)
	}
	*v = PingSlotPeriod(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinCountExponent) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := RejoinCountExponent_value[s]; ok {
		*v = RejoinCountExponent(i)
		return nil
	}
	if !strings.HasPrefix(s, "REJOIN_COUNT_") {
		if i, ok := RejoinCountExponent_value["REJOIN_COUNT_"+s]; ok {
			*v = RejoinCountExponent(i)
			return nil
		}
	}
	return errCouldNotParse("RejoinCountExponent")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinCountExponent) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("RejoinCountExponent")(string(b)).WithCause(err)
	}
	*v = RejoinCountExponent(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinTimeExponent) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := RejoinTimeExponent_value[s]; ok {
		*v = RejoinTimeExponent(i)
		return nil
	}
	if !strings.HasPrefix(s, "REJOIN_TIME_") {
		if i, ok := RejoinTimeExponent_value["REJOIN_TIME_"+s]; ok {
			*v = RejoinTimeExponent(i)
			return nil
		}
	}
	return errCouldNotParse("RejoinTimeExponent")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinTimeExponent) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("RejoinTimeExponent")(string(b)).WithCause(err)
	}
	*v = RejoinTimeExponent(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinPeriodExponent) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := RejoinPeriodExponent_value[s]; ok {
		*v = RejoinPeriodExponent(i)
		return nil
	}
	if !strings.HasPrefix(s, "REJOIN_PERIOD_") {
		if i, ok := RejoinPeriodExponent_value["REJOIN_PERIOD_"+s]; ok {
			*v = RejoinPeriodExponent(i)
			return nil
		}
	}
	return errCouldNotParse("RejoinPeriodExponent")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinPeriodExponent) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("RejoinPeriodExponent")(string(b)).WithCause(err)
	}
	*v = RejoinPeriodExponent(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DeviceEIRP) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DeviceEIRP) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := DeviceEIRP_value[s]; ok {
		*v = DeviceEIRP(i)
		return nil
	}
	if !strings.HasPrefix(s, "DEVICE_EIRP_") {
		if i, ok := DeviceEIRP_value["DEVICE_EIRP_"+s]; ok {
			*v = DeviceEIRP(i)
			return nil
		}
	}
	return errCouldNotParse("DeviceEIRP")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DeviceEIRP) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("DeviceEIRP")(string(b)).WithCause(err)
	}
	*v = DeviceEIRP(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ADRAckLimitExponent) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := ADRAckLimitExponent_value[s]; ok {
		*v = ADRAckLimitExponent(i)
		return nil
	}
	if !strings.HasPrefix(s, "ADR_ACK_LIMIT_") {
		if i, ok := ADRAckLimitExponent_value["ADR_ACK_LIMIT_"+s]; ok {
			*v = ADRAckLimitExponent(i)
			return nil
		}
	}
	return errCouldNotParse("ADRAckLimitExponent")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ADRAckLimitExponent) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("ADRAckLimitExponent")(string(b)).WithCause(err)
	}
	*v = ADRAckLimitExponent(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ADRAckDelayExponent) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := ADRAckDelayExponent_value[s]; ok {
		*v = ADRAckDelayExponent(i)
		return nil
	}
	if !strings.HasPrefix(s, "ADR_ACK_DELAY_") {
		if i, ok := ADRAckDelayExponent_value["ADR_ACK_DELAY_"+s]; ok {
			*v = ADRAckDelayExponent(i)
			return nil
		}
	}
	return errCouldNotParse("ADRAckDelayExponent")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ADRAckDelayExponent) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("ADRAckDelayExponent")(string(b)).WithCause(err)
	}
	*v = ADRAckDelayExponent(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RxDelay) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := RxDelay_value[s]; ok {
		*v = RxDelay(i)
		return nil
	}
	if !strings.HasPrefix(s, "RX_DELAY_") {
		if i, ok := RxDelay_value["RX_DELAY_"+s]; ok {
			*v = RxDelay(i)
			return nil
		}
	}
	return errCouldNotParse("RxDelay")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RxDelay) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("RxDelay")(string(b)).WithCause(err)
	}
	*v = RxDelay(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Minor) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := Minor_value[s]; ok {
		*v = Minor(i)
		return nil
	}
	if !strings.HasPrefix(s, "MINOR_") {
		if i, ok := Minor_value["MINOR_"+s]; ok {
			*v = Minor(i)
			return nil
		}
	}
	return errCouldNotParse("Minor")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Minor) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("Minor")(string(b)).WithCause(err)
	}
	*v = Minor(i)
	return nil
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
	case MAC_V1_0_3:
		return "1.0.3"
	case MAC_V1_1:
		return "1.1.0"
	}
	return "unknown"
}

func init() {
	for i := range MACVersion_name {
		MACVersion_value[MACVersion(i).String()] = i
	}
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
	return v.Compare(MAC_V1_1) >= 0
}

// HasMaxFCntGap reports whether v defines a MaxFCntGap.
// HasMaxFCntGap panics, if v.Validate() returns non-nil error.
func (v MACVersion) HasMaxFCntGap() bool {
	return v.Compare(MAC_V1_1) < 0
}

// Validate reports whether v represents a valid PHYVersion.
func (v PHYVersion) Validate() error {
	if v < 1 || v >= PHYVersion(len(PHYVersion_name)) {
		return errExpectedBetween("PHYVersion", 1, len(PHYVersion_name)-1)(v)
	}

	_, err := semver.Parse(v.String())
	if err != nil {
		return errParsingSemanticVersion(v.String()).WithCause(err)
	}
	return nil
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

// Compare compares PHYVersions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
// Compare panics, if v.Validate() returns non-nil error.
func (v PHYVersion) Compare(o PHYVersion) int {
	return semver.MustParse(v.String()).Compare(
		semver.MustParse(o.String()),
	)
}

func init() {
	for i := range PHYVersion_name {
		PHYVersion_value[PHYVersion(i).String()] = i
	}
}

// String implements fmt.Stringer.
func (v DataRateIndex) String() string {
	return strconv.Itoa(int(v))
}

// String implements fmt.Stringer.
func (v RxDelay) String() string {
	return strconv.Itoa(int(v))
}
