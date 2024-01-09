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
	"fmt"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v MType) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *MType) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("MType", MType_name, b)
	if err != nil {
		return err
	}
	*v = MType(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v Major) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *Major) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("Major", Major_name, b)
	if err != nil {
		return err
	}
	*v = Major(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v MACVersion) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// EncodeMsgpack implements msgpack.CustomEncoder interface.
func (v MACVersion) EncodeMsgpack(enc *msgpack.Encoder) error {
	if v > 255 {
		panic(fmt.Errorf("MACVersion enum exceeds 255"))
	}
	return enc.EncodeUint8(uint8(v))
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *MACVersion) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("MACVersion", MACVersion_name, b)
	if err != nil {
		return err
	}
	*v = MACVersion(i)
	return nil
}

// DecodeMsgpack implements msgpack.CustomDecoder interface.
func (v *MACVersion) DecodeMsgpack(dec *msgpack.Decoder) error {
	i, err := dec.DecodeInt32()
	if err != nil {
		return err
	}
	*v = MACVersion(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v PHYVersion) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *PHYVersion) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("PHYVersion", PHYVersion_name, b)
	if err != nil {
		return err
	}
	*v = PHYVersion(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DataRateIndex) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *DataRateIndex) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("DataRateIndex", DataRateIndex_name, b)
	if err != nil {
		return err
	}
	*v = DataRateIndex(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *DataRateIndexValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *DataRateIndexValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *DataRateIndexValue) UnmarshalBinary(b []byte) error {
	var vv DataRateIndex
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = DataRateIndexValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DataRateIndexValue) UnmarshalText(b []byte) error {
	var vv DataRateIndex
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = DataRateIndexValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *DataRateIndexValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DataRateOffset) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *DataRateOffset) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("DataRateOffset", DataRateOffset_name, b)
	if err != nil {
		return err
	}
	*v = DataRateOffset(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *DataRateOffsetValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *DataRateOffsetValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *DataRateOffsetValue) UnmarshalBinary(b []byte) error {
	var vv DataRateOffset
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = DataRateOffsetValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DataRateOffsetValue) UnmarshalText(b []byte) error {
	var vv DataRateOffset
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = DataRateOffsetValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *DataRateOffsetValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *FrequencyValue) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(v.GetValue(), 10)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *FrequencyValue) UnmarshalText(b []byte) error {
	var vv uint64
	vv, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	*v = FrequencyValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *FrequencyValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *ZeroableFrequencyValue) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(v.GetValue(), 10)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ZeroableFrequencyValue) UnmarshalText(b []byte) error {
	var vv uint64
	vv, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	*v = ZeroableFrequencyValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *ZeroableFrequencyValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

func (v JoinRequestType) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *JoinRequestType) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("JoinRequestType", JoinRequestType_name, b)
	if err != nil {
		return err
	}
	*v = JoinRequestType(i)
	return nil
}

func (v RejoinRequestType) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RejoinRequestType) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("RejoinRequestType", RejoinRequestType_name, b)
	if err != nil {
		return err
	}
	*v = RejoinRequestType(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v CFListType) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *CFListType) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("CFListType", CFListType_name, b)
	if err != nil {
		return err
	}
	*v = CFListType(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v Class) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *Class) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("Class", Class_name, b)
	if err != nil {
		return err
	}
	*v = Class(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v TxSchedulePriority) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *TxSchedulePriority) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("TxSchedulePriority", TxSchedulePriority_name, b)
	if err != nil {
		return err
	}
	*v = TxSchedulePriority(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v MACCommandIdentifier) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *MACCommandIdentifier) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("MACCommandIdentifier", MACCommandIdentifier_name, b)
	if err != nil {
		return err
	}
	*v = MACCommandIdentifier(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v AggregatedDutyCycle) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *AggregatedDutyCycle) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("AggregatedDutyCycle", AggregatedDutyCycle_name, b)
	if err != nil {
		return err
	}
	*v = AggregatedDutyCycle(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *AggregatedDutyCycleValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *AggregatedDutyCycleValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *AggregatedDutyCycleValue) UnmarshalBinary(b []byte) error {
	var vv AggregatedDutyCycle
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = AggregatedDutyCycleValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *AggregatedDutyCycleValue) UnmarshalText(b []byte) error {
	var vv AggregatedDutyCycle
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = AggregatedDutyCycleValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *AggregatedDutyCycleValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v PingSlotPeriod) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *PingSlotPeriod) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("PingSlotPeriod", PingSlotPeriod_name, b)
	if err != nil {
		return err
	}
	*v = PingSlotPeriod(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *PingSlotPeriodValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *PingSlotPeriodValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *PingSlotPeriodValue) UnmarshalBinary(b []byte) error {
	var vv PingSlotPeriod
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = PingSlotPeriodValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PingSlotPeriodValue) UnmarshalText(b []byte) error {
	var vv PingSlotPeriod
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = PingSlotPeriodValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *PingSlotPeriodValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RejoinCountExponent) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RejoinCountExponent) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("RejoinCountExponent", RejoinCountExponent_name, b)
	if err != nil {
		return err
	}
	*v = RejoinCountExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RejoinTimeExponent) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RejoinTimeExponent) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("RejoinTimeExponent", RejoinTimeExponent_name, b)
	if err != nil {
		return err
	}
	*v = RejoinTimeExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RejoinPeriodExponent) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RejoinPeriodExponent) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("RejoinPeriodExponent", RejoinPeriodExponent_name, b)
	if err != nil {
		return err
	}
	*v = RejoinPeriodExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DeviceEIRP) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *DeviceEIRP) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("DeviceEIRP", DeviceEIRP_name, b)
	if err != nil {
		return err
	}
	*v = DeviceEIRP(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v ADRAckLimitExponent) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *ADRAckLimitExponent) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("ADRAckLimitExponent", ADRAckLimitExponent_name, b)
	if err != nil {
		return err
	}
	*v = ADRAckLimitExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *ADRAckLimitExponentValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *ADRAckLimitExponentValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *ADRAckLimitExponentValue) UnmarshalBinary(b []byte) error {
	var vv ADRAckLimitExponent
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = ADRAckLimitExponentValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ADRAckLimitExponentValue) UnmarshalText(b []byte) error {
	var vv ADRAckLimitExponent
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = ADRAckLimitExponentValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *ADRAckLimitExponentValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v ADRAckDelayExponent) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *ADRAckDelayExponent) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("ADRAckDelayExponent", ADRAckDelayExponent_name, b)
	if err != nil {
		return err
	}
	*v = ADRAckDelayExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *ADRAckDelayExponentValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *ADRAckDelayExponentValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *ADRAckDelayExponentValue) UnmarshalBinary(b []byte) error {
	var vv ADRAckDelayExponent
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = ADRAckDelayExponentValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ADRAckDelayExponentValue) UnmarshalText(b []byte) error {
	var vv ADRAckDelayExponent
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = ADRAckDelayExponentValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *ADRAckDelayExponentValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RxDelay) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RxDelay) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("RxDelay", RxDelay_name, b)
	if err != nil {
		return err
	}
	*v = RxDelay(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *RxDelayValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *RxDelayValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RxDelayValue) UnmarshalBinary(b []byte) error {
	var vv RxDelay
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = RxDelayValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RxDelayValue) UnmarshalText(b []byte) error {
	var vv RxDelay
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = RxDelayValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *RxDelayValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v Minor) MarshalBinary() ([]byte, error) {
	return marshalBinaryEnum(int32(v)), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *Minor) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("Minor", Minor_name, b)
	if err != nil {
		return err
	}
	*v = Minor(i)
	return nil
}

// Validate reports whether v represents a valid MACVersion.
func (v MACVersion) Validate() error {
	if v < 1 || v >= MACVersion(len(MACVersion_name)) {
		return errExpectedBetween("MACVersion", 1, len(MACVersion_name)-1)(v)
	}
	return nil
}

// Validate reports whether v represents a valid PHYVersion.
func (v PHYVersion) Validate() error {
	if v < 1 || v >= PHYVersion(len(PHYVersion_name)) {
		return errExpectedBetween("PHYVersion", 1, len(PHYVersion_name)-1)(v)
	}
	return nil
}

// Duration returns v as time.Duration.
func (v RxDelay) Duration() time.Duration {
	switch v {
	case RxDelay_RX_DELAY_0, RxDelay_RX_DELAY_1:
		return time.Second
	default:
		return time.Duration(v) * time.Second
	}
}

// Validate reports whether v represents a valid RxDelay.
func (v RxDelay) Validate() error {
	if v < 0 || v >= RxDelay(len(RxDelay_name)) {
		return errExpectedBetween("RxDelay", 0, len(RxDelay_name)-1)(v)
	}
	return nil
}

func (v *LoRaDataRate) DataRate() *DataRate {
	if v == nil {
		return nil
	}
	return &DataRate{
		Modulation: &DataRate_Lora{
			Lora: v,
		},
	}
}

func (v *FSKDataRate) DataRate() *DataRate {
	if v == nil {
		return nil
	}
	return &DataRate{
		Modulation: &DataRate_Fsk{
			Fsk: v,
		},
	}
}

func (v *LRFHSSDataRate) DataRate() *DataRate {
	if v == nil {
		return nil
	}
	return &DataRate{
		Modulation: &DataRate_Lrfhss{
			Lrfhss: v,
		},
	}
}

// FieldIsZero returns whether path p is zero.
func (v *CFList) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "ch_masks":
		return v.ChMasks == nil
	case "freq":
		return v.Freq == nil
	case "type":
		return v.Type == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *DLSettings) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "opt_neg":
		return !v.OptNeg
	case "rx1_dr_offset":
		return v.Rx1DrOffset == 0
	case "rx2_dr":
		return v.Rx2Dr == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MHDR) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "m_type":
		return v.MType == 0
	case "major":
		return v.Major == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *JoinAcceptPayload) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "cf_list":
		return v.CfList == nil
	case "cf_list.ch_masks":
		return v.CfList.FieldIsZero("ch_masks")
	case "cf_list.freq":
		return v.CfList.FieldIsZero("freq")
	case "cf_list.type":
		return v.CfList.FieldIsZero("type")
	case "dev_addr":
		return types.MustDevAddr(v.DevAddr).OrZero().IsZero()
	case "dl_settings":
		return v.DlSettings == nil
	case "dl_settings.opt_neg":
		return v.DlSettings.FieldIsZero("opt_neg")
	case "dl_settings.rx1_dr_offset":
		return v.DlSettings.FieldIsZero("rx1_dr_offset")
	case "dl_settings.rx2_dr":
		return v.DlSettings.FieldIsZero("rx2_dr")
	case "encrypted":
		return v.Encrypted == nil
	case "join_nonce":
		return types.MustJoinNonce(v.JoinNonce).OrZero().IsZero()
	case "net_id":
		return types.MustNetID(v.NetId).OrZero().IsZero()
	case "rx_delay":
		return v.RxDelay == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *JoinRequestPayload) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "dev_eui":
		return types.MustEUI64(v.DevEui).OrZero().IsZero()
	case "dev_nonce":
		return types.MustDevNonce(v.DevNonce).OrZero().IsZero()
	case "join_eui":
		return types.MustEUI64(v.JoinEui).OrZero().IsZero()
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *FCtrl) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "ack":
		return !v.Ack
	case "adr":
		return !v.Adr
	case "adr_ack_req":
		return !v.AdrAckReq
	case "class_b":
		return !v.ClassB
	case "f_pending":
		return !v.FPending
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *FHDR) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "dev_addr":
		return types.MustDevAddr(v.DevAddr).OrZero().IsZero()
	case "f_cnt":
		return v.FCnt == 0
	case "f_ctrl":
		return v.FCtrl == nil
	case "f_ctrl.ack":
		return v.FCtrl.FieldIsZero("ack")
	case "f_ctrl.adr":
		return v.FCtrl.FieldIsZero("adr")
	case "f_ctrl.adr_ack_req":
		return v.FCtrl.FieldIsZero("adr_ack_req")
	case "f_ctrl.class_b":
		return v.FCtrl.FieldIsZero("class_b")
	case "f_ctrl.f_pending":
		return v.FCtrl.FieldIsZero("f_pending")
	case "f_opts":
		return v.FOpts == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *MACPayload) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "decoded_payload":
		return v.DecodedPayload == nil
	case "f_hdr":
		return fieldsAreZero(v.FHdr, FHDRFieldPathsTopLevel...)
	case "f_hdr.dev_addr":
		return v.FHdr.FieldIsZero("dev_addr")
	case "f_hdr.f_cnt":
		return v.FHdr.FieldIsZero("f_cnt")
	case "f_hdr.f_ctrl":
		return v.FHdr.FieldIsZero("f_ctrl")
	case "f_hdr.f_ctrl.ack":
		return v.FHdr.FieldIsZero("f_ctrl.ack")
	case "f_hdr.f_ctrl.adr":
		return v.FHdr.FieldIsZero("f_ctrl.adr")
	case "f_hdr.f_ctrl.adr_ack_req":
		return v.FHdr.FieldIsZero("f_ctrl.adr_ack_req")
	case "f_hdr.f_ctrl.class_b":
		return v.FHdr.FieldIsZero("f_ctrl.class_b")
	case "f_hdr.f_ctrl.f_pending":
		return v.FHdr.FieldIsZero("f_ctrl.f_pending")
	case "f_hdr.f_opts":
		return v.FHdr.FieldIsZero("f_opts")
	case "f_port":
		return v.FPort == 0
	case "frm_payload":
		return v.FrmPayload == nil
	case "full_f_cnt":
		return v.FullFCnt == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *RejoinRequestPayload) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "dev_eui":
		return types.MustEUI64(v.DevEui).OrZero().IsZero()
	case "join_eui":
		return types.MustEUI64(v.JoinEui).OrZero().IsZero()
	case "net_id":
		return types.MustNetID(v.NetId).OrZero().IsZero()
	case "rejoin_cnt":
		return v.RejoinCnt == 0
	case "rejoin_type":
		return v.RejoinType == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *Message) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "Payload":
		return v.Payload == nil
	case "Payload.join_accept_payload":
		return v.GetJoinAcceptPayload() == nil
	case "Payload.join_accept_payload.cf_list":
		return v.GetJoinAcceptPayload().FieldIsZero("cf_list")
	case "Payload.join_accept_payload.cf_list.ch_masks":
		return v.GetJoinAcceptPayload().FieldIsZero("cf_list.ch_masks")
	case "Payload.join_accept_payload.cf_list.freq":
		return v.GetJoinAcceptPayload().FieldIsZero("cf_list.freq")
	case "Payload.join_accept_payload.cf_list.type":
		return v.GetJoinAcceptPayload().FieldIsZero("cf_list.type")
	case "Payload.join_accept_payload.dev_addr":
		return v.GetJoinAcceptPayload().FieldIsZero("dev_addr")
	case "Payload.join_accept_payload.dl_settings":
		return v.GetJoinAcceptPayload().FieldIsZero("dl_settings")
	case "Payload.join_accept_payload.dl_settings.opt_neg":
		return v.GetJoinAcceptPayload().FieldIsZero("dl_settings.opt_neg")
	case "Payload.join_accept_payload.dl_settings.rx1_dr_offset":
		return v.GetJoinAcceptPayload().FieldIsZero("dl_settings.rx1_dr_offset")
	case "Payload.join_accept_payload.dl_settings.rx2_dr":
		return v.GetJoinAcceptPayload().FieldIsZero("dl_settings.rx2_dr")
	case "Payload.join_accept_payload.encrypted":
		return v.GetJoinAcceptPayload().FieldIsZero("encrypted")
	case "Payload.join_accept_payload.join_nonce":
		return v.GetJoinAcceptPayload().FieldIsZero("join_nonce")
	case "Payload.join_accept_payload.net_id":
		return v.GetJoinAcceptPayload().FieldIsZero("net_id")
	case "Payload.join_accept_payload.rx_delay":
		return v.GetJoinAcceptPayload().FieldIsZero("rx_delay")
	case "Payload.join_request_payload":
		return v.GetJoinRequestPayload() == nil
	case "Payload.join_request_payload.dev_eui":
		return v.GetJoinRequestPayload().FieldIsZero("dev_eui")
	case "Payload.join_request_payload.dev_nonce":
		return v.GetJoinRequestPayload().FieldIsZero("dev_nonce")
	case "Payload.join_request_payload.join_eui":
		return v.GetJoinRequestPayload().FieldIsZero("join_eui")
	case "Payload.mac_payload":
		return v.GetMacPayload() == nil
	case "Payload.mac_payload.decoded_payload":
		return v.GetMacPayload().FieldIsZero("decoded_payload")
	case "Payload.mac_payload.f_hdr":
		return v.GetMacPayload().FieldIsZero("f_hdr")
	case "Payload.mac_payload.f_hdr.dev_addr":
		return v.GetMacPayload().FieldIsZero("f_hdr.dev_addr")
	case "Payload.mac_payload.f_hdr.f_cnt":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_cnt")
	case "Payload.mac_payload.f_hdr.f_ctrl":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_ctrl")
	case "Payload.mac_payload.f_hdr.f_ctrl.ack":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_ctrl.ack")
	case "Payload.mac_payload.f_hdr.f_ctrl.adr":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_ctrl.adr")
	case "Payload.mac_payload.f_hdr.f_ctrl.adr_ack_req":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_ctrl.adr_ack_req")
	case "Payload.mac_payload.f_hdr.f_ctrl.class_b":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_ctrl.class_b")
	case "Payload.mac_payload.f_hdr.f_ctrl.f_pending":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_ctrl.f_pending")
	case "Payload.mac_payload.f_hdr.f_opts":
		return v.GetMacPayload().FieldIsZero("f_hdr.f_opts")
	case "Payload.mac_payload.f_port":
		return v.GetMacPayload().FieldIsZero("f_port")
	case "Payload.mac_payload.frm_payload":
		return v.GetMacPayload().FieldIsZero("frm_payload")
	case "Payload.mac_payload.full_f_cnt":
		return v.GetMacPayload().FieldIsZero("full_f_cnt")
	case "Payload.rejoin_request_payload":
		return v.GetRejoinRequestPayload() == nil
	case "Payload.rejoin_request_payload.dev_eui":
		return v.GetRejoinRequestPayload().FieldIsZero("dev_eui")
	case "Payload.rejoin_request_payload.join_eui":
		return v.GetRejoinRequestPayload().FieldIsZero("join_eui")
	case "Payload.rejoin_request_payload.net_id":
		return v.GetRejoinRequestPayload().FieldIsZero("net_id")
	case "Payload.rejoin_request_payload.rejoin_cnt":
		return v.GetRejoinRequestPayload().FieldIsZero("rejoin_cnt")
	case "Payload.rejoin_request_payload.rejoin_type":
		return v.GetRejoinRequestPayload().FieldIsZero("rejoin_type")
	case "m_hdr":
		return v.MHdr == nil
	case "m_hdr.m_type":
		return v.MHdr.FieldIsZero("m_type")
	case "m_hdr.major":
		return v.MHdr.FieldIsZero("major")
	case "mic":
		return v.Mic == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v *DeviceEIRPValue) MarshalBinary() ([]byte, error) {
	return v.GetValue().MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *DeviceEIRPValue) MarshalText() ([]byte, error) {
	return v.GetValue().MarshalText()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *DeviceEIRPValue) UnmarshalBinary(b []byte) error {
	var vv DeviceEIRP
	if err := vv.UnmarshalBinary(b); err != nil {
		return err
	}
	*v = DeviceEIRPValue{
		Value: vv,
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DeviceEIRPValue) UnmarshalText(b []byte) error {
	var vv DeviceEIRP
	if err := vv.UnmarshalText(b); err != nil {
		return err
	}
	*v = DeviceEIRPValue{
		Value: vv,
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *DeviceEIRPValue) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "value":
		return v.Value == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// EndDeviceIdentifiers returns the end device identifiers (DevEUI/JoinEUI/DevAddr) available for the message payload.
// Note that if the payload is nil, the end device identifiers will be nil.
func (m *Message) EndDeviceIdentifiers() *EndDeviceIdentifiers {
	if h := m.GetMacPayload().GetFHdr(); h != nil {
		devAddr := types.MustDevAddr(h.DevAddr).OrZero()
		return &EndDeviceIdentifiers{
			DevAddr: devAddr.Bytes(),
		}
	}
	if p := m.GetJoinRequestPayload(); p != nil {
		devEUI, joinEUI := types.MustEUI64(p.DevEui).OrZero(), types.MustEUI64(p.JoinEui).OrZero()
		return &EndDeviceIdentifiers{
			DevEui:  devEUI.Bytes(),
			JoinEui: joinEUI.Bytes(),
		}
	}
	if p := m.GetJoinAcceptPayload(); p != nil {
		devAddr := types.MustDevAddr(p.DevAddr).OrZero()
		return &EndDeviceIdentifiers{
			DevAddr: devAddr.Bytes(),
		}
	}
	if p := m.GetRejoinRequestPayload(); p != nil {
		devEUI, joinEUI := types.MustEUI64(p.DevEui).OrZero(), types.MustEUI64(p.JoinEui).OrZero()
		return &EndDeviceIdentifiers{
			DevEui:  devEUI.Bytes(),
			JoinEui: joinEUI.Bytes(),
		}
	}
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *RelaySecondChannel) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "ack_offset":
		return v.AckOffset == 0
	case "data_rate_index":
		return v.DataRateIndex == 0
	case "frequency":
		return v.Frequency == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *RelayEndDeviceDynamicMode) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "smart_enable_level":
		return v.SmartEnableLevel == 0
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *RelayForwardDownlinkReq) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "raw_payload":
		return v.RawPayload == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}
