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

	"github.com/blang/semver"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func init() {
	for i := range PHYVersion_name {
		PHYVersion_value[PHYVersion(i).String()] = i
	}
	PHYVersion_value["1.0"] = int32(PHY_V1_0)           // 1.0 is the official version number
	PHYVersion_value["1.0.2"] = int32(PHY_V1_0_2_REV_A) // Revisions were added from 1.0.2-b
	PHYVersion_value["1.1-a"] = int32(PHY_V1_1_REV_A)   // 1.1 is the official version number
	PHYVersion_value["1.1-b"] = int32(PHY_V1_1_REV_B)   // 1.1 is the official version number
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v MType) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("MType enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MType) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromText("MType", MType_value, b)
	if err != nil {
		return err
	}
	*v = MType(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MType) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("MType", MType_name, b)
	if err != nil {
		return err
	}
	*v = MType(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v Major) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("Major enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v Major) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Major) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("Major", Major_value, "LORAWAN_", b)
	if err != nil {
		return err
	}
	*v = Major(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Major) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("Major", Major_name, b)
	if err != nil {
		return err
	}
	*v = Major(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v MACVersion) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("MACVersion enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MACVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MACVersion) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("MACVersion", MACVersion_value, "MAC_", b)
	if err != nil {
		return err
	}
	*v = MACVersion(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MACVersion) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("MACVersion", MACVersion_name, b)
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
	if v > 255 {
		panic(fmt.Errorf("PHYVersion enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v PHYVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PHYVersion) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("PHYVersion", PHYVersion_value, "PHY_", b)
	if err != nil {
		return err
	}
	*v = PHYVersion(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PHYVersion) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("PHYVersion", PHYVersion_name, b)
	if err != nil {
		return err
	}
	*v = PHYVersion(i)
	return nil
}

// String implements fmt.Stringer.
func (v DataRateIndex) String() string {
	return strconv.Itoa(int(v))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DataRateIndex) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("DataRateIndex enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DataRateIndex) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v DataRateIndex) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DataRateIndex) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("DataRateIndex", DataRateIndex_value, "DATA_RATE_", b)
	if err != nil {
		return err
	}
	*v = DataRateIndex(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DataRateIndex) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("DataRateIndex", DataRateIndex_name, b)
	if err != nil {
		return err
	}
	*v = DataRateIndex(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DataRateIndexValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DataRateIndexValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v DataRateIndexValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v DataRateIndexValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (DataRateIndexValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DataRateIndexValue) UnmarshalJSON(b []byte) error {
	var vv DataRateIndex
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = DataRateIndexValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *DataRateIndexValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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

// String implements fmt.Stringer.
func (v DataRateOffset) String() string {
	return strconv.Itoa(int(v))
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DataRateOffset) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("DataRateOffset enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DataRateOffset) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v DataRateOffset) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DataRateOffset) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("DataRateOffset", DataRateOffset_value, "DATA_RATE_OFFSET_", b)
	if err != nil {
		return err
	}
	*v = DataRateOffset(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DataRateOffset) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("DataRateOffset", DataRateOffset_name, b)
	if err != nil {
		return err
	}
	*v = DataRateOffset(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DataRateOffsetValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DataRateOffsetValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v DataRateOffsetValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v DataRateOffsetValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (DataRateOffsetValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DataRateOffsetValue) UnmarshalJSON(b []byte) error {
	var vv DataRateOffset
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = DataRateOffsetValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *DataRateOffsetValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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
func (v FrequencyValue) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(v.Value, 10)), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v FrequencyValue) MarshalJSON() ([]byte, error) {
	b, err := v.MarshalText()
	if err != nil {
		return nil, err
	}
	// Ensure number is marshaled as a string according to the spec.
	return append(append([]byte{'"'}, b...), '"'), nil
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v FrequencyValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (FrequencyValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *FrequencyValue) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	return v.UnmarshalText(b)
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *FrequencyValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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

func (v RejoinType) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("RejoinType enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RejoinType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (v *RejoinType) UnmarshalBinary(b []byte) error {
	i, err := unmarshalEnumFromBinary("RejoinType", RejoinType_name, b)
	if err != nil {
		return err
	}
	*v = RejoinType(i)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinType) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromText("RejoinType", RejoinType_value, b)
	if err != nil {
		return err
	}
	*v = RejoinType(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinType) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("RejoinType", RejoinType_name, b)
	if err != nil {
		return err
	}
	*v = RejoinType(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v CFListType) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("CFListType enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v CFListType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *CFListType) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromText("CFListType", CFListType_value, b)
	if err != nil {
		return err
	}
	*v = CFListType(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *CFListType) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("CFListType", CFListType_name, b)
	if err != nil {
		return err
	}
	*v = CFListType(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v Class) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("Class enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v Class) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Class) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("Class", Class_value, "CLASS_", b)
	if err != nil {
		return err
	}
	*v = Class(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Class) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("Class", Class_name, b)
	if err != nil {
		return err
	}
	*v = Class(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v TxSchedulePriority) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("TxSchedulePriority enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v TxSchedulePriority) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *TxSchedulePriority) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromText("TxSchedulePriority", TxSchedulePriority_value, b)
	if err != nil {
		return err
	}
	*v = TxSchedulePriority(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *TxSchedulePriority) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("TxSchedulePriority", TxSchedulePriority_name, b)
	if err != nil {
		return err
	}
	*v = TxSchedulePriority(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v MACCommandIdentifier) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("MACCommandIdentifier enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v MACCommandIdentifier) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *MACCommandIdentifier) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("MACCommandIdentifier", MACCommandIdentifier_value, "CID_", b)
	if err != nil {
		return err
	}
	*v = MACCommandIdentifier(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *MACCommandIdentifier) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("MACCommandIdentifier", MACCommandIdentifier_name, b)
	if err != nil {
		return err
	}
	*v = MACCommandIdentifier(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v AggregatedDutyCycle) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("AggregatedDutyCycle enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v AggregatedDutyCycle) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v AggregatedDutyCycle) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *AggregatedDutyCycle) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("AggregatedDutyCycle", AggregatedDutyCycle_value, "DUTY_CYCLE_", b)
	if err != nil {
		return err
	}
	*v = AggregatedDutyCycle(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *AggregatedDutyCycle) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("AggregatedDutyCycle", AggregatedDutyCycle_name, b)
	if err != nil {
		return err
	}
	*v = AggregatedDutyCycle(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v AggregatedDutyCycleValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v AggregatedDutyCycleValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v AggregatedDutyCycleValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v AggregatedDutyCycleValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (AggregatedDutyCycleValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *AggregatedDutyCycleValue) UnmarshalJSON(b []byte) error {
	var vv AggregatedDutyCycle
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = AggregatedDutyCycleValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *AggregatedDutyCycleValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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
	if v > 255 {
		panic(fmt.Errorf("PingSlotPeriod enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v PingSlotPeriod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v PingSlotPeriod) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PingSlotPeriod) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("PingSlotPeriod", PingSlotPeriod_value, "PING_EVERY_", b)
	if err != nil {
		return err
	}
	*v = PingSlotPeriod(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PingSlotPeriod) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("PingSlotPeriod", PingSlotPeriod_name, b)
	if err != nil {
		return err
	}
	*v = PingSlotPeriod(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v PingSlotPeriodValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v PingSlotPeriodValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v PingSlotPeriodValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v PingSlotPeriodValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (PingSlotPeriodValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PingSlotPeriodValue) UnmarshalJSON(b []byte) error {
	var vv PingSlotPeriod
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = PingSlotPeriodValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *PingSlotPeriodValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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
	if v > 255 {
		panic(fmt.Errorf("RejoinCountExponent enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RejoinCountExponent) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinCountExponent) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("RejoinCountExponent", RejoinCountExponent_value, "REJOIN_COUNT_", b)
	if err != nil {
		return err
	}
	*v = RejoinCountExponent(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinCountExponent) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("RejoinCountExponent", RejoinCountExponent_name, b)
	if err != nil {
		return err
	}
	*v = RejoinCountExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RejoinTimeExponent) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("RejoinTimeExponent enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RejoinTimeExponent) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinTimeExponent) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("RejoinTimeExponent", RejoinTimeExponent_value, "REJOIN_TIME_", b)
	if err != nil {
		return err
	}
	*v = RejoinTimeExponent(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinTimeExponent) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("RejoinTimeExponent", RejoinTimeExponent_name, b)
	if err != nil {
		return err
	}
	*v = RejoinTimeExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RejoinPeriodExponent) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("RejoinPeriodExponent enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RejoinPeriodExponent) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RejoinPeriodExponent) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("RejoinPeriodExponent", RejoinPeriodExponent_value, "REJOIN_PERIOD_", b)
	if err != nil {
		return err
	}
	*v = RejoinPeriodExponent(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RejoinPeriodExponent) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("RejoinPeriodExponent", RejoinPeriodExponent_name, b)
	if err != nil {
		return err
	}
	*v = RejoinPeriodExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v DeviceEIRP) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("DeviceEIRP enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v DeviceEIRP) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DeviceEIRP) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("DeviceEIRP", DeviceEIRP_value, "DEVICE_EIRP_", b)
	if err != nil {
		return err
	}
	*v = DeviceEIRP(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DeviceEIRP) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("DeviceEIRP", DeviceEIRP_name, b)
	if err != nil {
		return err
	}
	*v = DeviceEIRP(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v ADRAckLimitExponent) MarshalBinary() ([]byte, error) {
	if v > 255 {
		panic(fmt.Errorf("ADRAckLimitExponent enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v ADRAckLimitExponent) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v ADRAckLimitExponent) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ADRAckLimitExponent) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("ADRAckLimitExponent", ADRAckLimitExponent_value, "ADR_ACK_LIMIT_", b)
	if err != nil {
		return err
	}
	*v = ADRAckLimitExponent(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ADRAckLimitExponent) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("ADRAckLimitExponent", ADRAckLimitExponent_name, b)
	if err != nil {
		return err
	}
	*v = ADRAckLimitExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v ADRAckLimitExponentValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v ADRAckLimitExponentValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v ADRAckLimitExponentValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v ADRAckLimitExponentValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (ADRAckLimitExponentValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ADRAckLimitExponentValue) UnmarshalJSON(b []byte) error {
	var vv ADRAckLimitExponent
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = ADRAckLimitExponentValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *ADRAckLimitExponentValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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
	if v > 255 {
		panic(fmt.Errorf("ADRAckDelayExponent enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v ADRAckDelayExponent) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v ADRAckDelayExponent) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ADRAckDelayExponent) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("ADRAckDelayExponent", ADRAckDelayExponent_value, "ADR_ACK_DELAY_", b)
	if err != nil {
		return err
	}
	*v = ADRAckDelayExponent(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ADRAckDelayExponent) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("ADRAckDelayExponent", ADRAckDelayExponent_name, b)
	if err != nil {
		return err
	}
	*v = ADRAckDelayExponent(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v ADRAckDelayExponentValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v ADRAckDelayExponentValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v ADRAckDelayExponentValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v ADRAckDelayExponentValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (ADRAckDelayExponentValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ADRAckDelayExponentValue) UnmarshalJSON(b []byte) error {
	var vv ADRAckDelayExponent
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = ADRAckDelayExponentValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *ADRAckDelayExponentValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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
	if v > 255 {
		panic(fmt.Errorf("RxDelay enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RxDelay) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// MarshalJSON implements json.Marshaler interface.
func (v RxDelay) MarshalJSON() ([]byte, error) {
	return marshalJSONEnum(int32(v)), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *RxDelay) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("RxDelay", RxDelay_value, "RX_DELAY_", b)
	if err != nil {
		return err
	}
	*v = RxDelay(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RxDelay) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("RxDelay", RxDelay_name, b)
	if err != nil {
		return err
	}
	*v = RxDelay(i)
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface.
func (v RxDelayValue) MarshalBinary() ([]byte, error) {
	return v.Value.MarshalBinary()
}

// MarshalText implements encoding.TextMarshaler interface.
func (v RxDelayValue) MarshalText() ([]byte, error) {
	return v.Value.MarshalText()
}

// MarshalJSON implements json.Marshaler interface.
func (v RxDelayValue) MarshalJSON() ([]byte, error) {
	return v.Value.MarshalJSON()
}

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface.
func (v RxDelayValue) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	if v == (RxDelayValue{}) && !m.EmitDefaults {
		return nil, nil
	}
	if !m.EnumsAsInts {
		return nil, errEnumAsJSONString.New()
	}
	return v.MarshalJSON()
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

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *RxDelayValue) UnmarshalJSON(b []byte) error {
	var vv RxDelay
	if err := vv.UnmarshalJSON(b); err != nil {
		return err
	}
	*v = RxDelayValue{
		Value: vv,
	}
	return nil
}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface.
func (v *RxDelayValue) UnmarshalJSONPB(m *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
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
	if v > 255 {
		panic(fmt.Errorf("Minor enum exceeds 255"))
	}
	return []byte{byte(v)}, nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v Minor) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
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

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *Minor) UnmarshalText(b []byte) error {
	i, err := unmarshalEnumFromTextPrefix("Minor", Minor_value, "MINOR_", b)
	if err != nil {
		return err
	}
	*v = Minor(i)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *Minor) UnmarshalJSON(b []byte) error {
	if bt, ok := unmarshalJSONString(b); ok {
		return v.UnmarshalText(bt)
	}
	i, err := unmarshalEnumFromNumber("Minor", Minor_name, b)
	if err != nil {
		return err
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
	case MAC_V1_0_4:
		return "1.0.4"
	case MAC_V1_1:
		return "1.1.0"
	}
	return "unknown"
}

func init() {
	for i := range MACVersion_name {
		MACVersion_value[MACVersion(i).String()] = i
	}
	MACVersion_value["1.0"] = int32(MAC_V1_0) // 1.0 is the official version number
	MACVersion_value["1.1"] = int32(MAC_V1_1) // 1.1 is the official version number
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
	return v.Compare(MAC_V1_0_4) < 0
}

// HasNoChangeTXPowerIndex reports whether v defines a no-change TxPowerIndex value.
// HasNoChangeTXPowerIndex panics, if v.Validate() returns non-nil error.
func (v MACVersion) HasNoChangeTXPowerIndex() bool {
	return v.Compare(MAC_V1_0_4) >= 0
}

// HasNoChangeDataRateIndex reports whether v defines a no-change DataRateIndex value.
// HasNoChangeDataRateIndex panics, if v.Validate() returns non-nil error.
func (v MACVersion) HasNoChangeDataRateIndex() bool {
	return v.Compare(MAC_V1_0_4) >= 0
}

// IgnoreUplinksExceedingLengthLimit reports whether v requires Network Server to
// silently drop uplinks exceeding selected data rate payload length limits.
// IgnoreUplinksExceedingLengthLimit panics, if v.Validate() returns non-nil error.
func (v MACVersion) IgnoreUplinksExceedingLengthLimit() bool {
	return v.Compare(MAC_V1_0_4) >= 0 && v.Compare(MAC_V1_1) < 0
}

// IncrementDevNonce reports whether v defines DevNonce as an incrementing counter.
// IncrementDevNonce panics, if v.Validate() returns non-nil error.
func (v MACVersion) IncrementDevNonce() bool {
	return v.Compare(MAC_V1_0_4) >= 0
}

// UseNwkKey reports whether v uses a root NwkKey.
// UseNwkKey panics, if v.Validate() returns non-nil error.
func (v MACVersion) UseNwkKey() bool {
	return v.Compare(MAC_V1_1) >= 0
}

// UseLegacyMIC reports whether v uses legacy MIC computation algorithm.
// UseLegacyMIC panics, if v.Validate() returns non-nil error.
func (v MACVersion) UseLegacyMIC() bool {
	return v.Compare(MAC_V1_1) < 0
}

// RequireDevEUIForABP reports whether v requires ABP devices to have a DevEUI associated.
// RequireDevEUIForABP panics, if v.Validate() returns non-nil error.
func (v MACVersion) RequireDevEUIForABP() bool {
	return v.Compare(MAC_V1_0_4) >= 0 && v.Compare(MAC_V1_1) < 0
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

// Duration returns v as time.Duration.
func (v RxDelay) Duration() time.Duration {
	switch v {
	case RX_DELAY_0, RX_DELAY_1:
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

// String implements fmt.Stringer.
func (v RxDelay) String() string {
	return strconv.Itoa(int(v))
}

func (v LoRaDataRate) DataRate() DataRate {
	return DataRate{
		Modulation: &DataRate_LoRa{
			LoRa: &v,
		},
	}
}

func (v FSKDataRate) DataRate() DataRate {
	return DataRate{
		Modulation: &DataRate_FSK{
			FSK: &v,
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
		return v.Rx1DROffset == 0
	case "rx2_dr":
		return v.Rx2DR == 0
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
		return v.CFList == nil
	case "cf_list.ch_masks":
		return v.CFList.FieldIsZero("ch_masks")
	case "cf_list.freq":
		return v.CFList.FieldIsZero("freq")
	case "cf_list.type":
		return v.CFList.FieldIsZero("type")
	case "dev_addr":
		return v.DevAddr == types.DevAddr{}
	case "dl_settings":
		return v.DLSettings == DLSettings{}
	case "dl_settings.opt_neg":
		return v.DLSettings.FieldIsZero("opt_neg")
	case "dl_settings.rx1_dr_offset":
		return v.DLSettings.FieldIsZero("rx1_dr_offset")
	case "dl_settings.rx2_dr":
		return v.DLSettings.FieldIsZero("rx2_dr")
	case "encrypted":
		return v.Encrypted == nil
	case "join_nonce":
		return v.JoinNonce == types.JoinNonce{}
	case "net_id":
		return v.NetID == types.NetID{}
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
		return v.DevEUI == types.EUI64{}
	case "dev_nonce":
		return v.DevNonce == types.DevNonce{}
	case "join_eui":
		return v.JoinEUI == types.EUI64{}
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
		return !v.ADR
	case "adr_ack_req":
		return !v.ADRAckReq
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
		return v.DevAddr == types.DevAddr{}
	case "f_cnt":
		return v.FCnt == 0
	case "f_ctrl":
		return v.FCtrl == FCtrl{}
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
		return fieldsAreZero(&v.FHDR, FHDRFieldPathsTopLevel...)
	case "f_hdr.dev_addr":
		return v.FHDR.FieldIsZero("dev_addr")
	case "f_hdr.f_cnt":
		return v.FHDR.FieldIsZero("f_cnt")
	case "f_hdr.f_ctrl":
		return v.FHDR.FieldIsZero("f_ctrl")
	case "f_hdr.f_ctrl.ack":
		return v.FHDR.FieldIsZero("f_ctrl.ack")
	case "f_hdr.f_ctrl.adr":
		return v.FHDR.FieldIsZero("f_ctrl.adr")
	case "f_hdr.f_ctrl.adr_ack_req":
		return v.FHDR.FieldIsZero("f_ctrl.adr_ack_req")
	case "f_hdr.f_ctrl.class_b":
		return v.FHDR.FieldIsZero("f_ctrl.class_b")
	case "f_hdr.f_ctrl.f_pending":
		return v.FHDR.FieldIsZero("f_ctrl.f_pending")
	case "f_hdr.f_opts":
		return v.FHDR.FieldIsZero("f_opts")
	case "f_port":
		return v.FPort == 0
	case "frm_payload":
		return v.FRMPayload == nil
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
		return v.DevEUI == types.EUI64{}
	case "join_eui":
		return v.JoinEUI == types.EUI64{}
	case "net_id":
		return v.NetID == types.NetID{}
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
		return v.GetMACPayload() == nil
	case "Payload.mac_payload.decoded_payload":
		return v.GetMACPayload().FieldIsZero("decoded_payload")
	case "Payload.mac_payload.f_hdr":
		return v.GetMACPayload().FieldIsZero("f_hdr")
	case "Payload.mac_payload.f_hdr.dev_addr":
		return v.GetMACPayload().FieldIsZero("f_hdr.dev_addr")
	case "Payload.mac_payload.f_hdr.f_cnt":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_cnt")
	case "Payload.mac_payload.f_hdr.f_ctrl":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_ctrl")
	case "Payload.mac_payload.f_hdr.f_ctrl.ack":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_ctrl.ack")
	case "Payload.mac_payload.f_hdr.f_ctrl.adr":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_ctrl.adr")
	case "Payload.mac_payload.f_hdr.f_ctrl.adr_ack_req":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_ctrl.adr_ack_req")
	case "Payload.mac_payload.f_hdr.f_ctrl.class_b":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_ctrl.class_b")
	case "Payload.mac_payload.f_hdr.f_ctrl.f_pending":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_ctrl.f_pending")
	case "Payload.mac_payload.f_hdr.f_opts":
		return v.GetMACPayload().FieldIsZero("f_hdr.f_opts")
	case "Payload.mac_payload.f_port":
		return v.GetMACPayload().FieldIsZero("f_port")
	case "Payload.mac_payload.frm_payload":
		return v.GetMACPayload().FieldIsZero("frm_payload")
	case "Payload.mac_payload.full_f_cnt":
		return v.GetMACPayload().FieldIsZero("full_f_cnt")
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
		return v.MHDR == MHDR{}
	case "m_hdr.m_type":
		return v.MHDR.FieldIsZero("m_type")
	case "m_hdr.major":
		return v.MHDR.FieldIsZero("major")
	case "mic":
		return v.MIC == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}
