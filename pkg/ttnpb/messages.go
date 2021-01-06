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
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// MarshalText implements encoding.TextMarshaler interface.
func (v PayloadFormatter) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PayloadFormatter) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := PayloadFormatter_value[s]; ok {
		*v = PayloadFormatter(i)
		return nil
	}
	if !strings.HasPrefix(s, "FORMATTER_") {
		if i, ok := PayloadFormatter_value["FORMATTER_"+s]; ok {
			*v = PayloadFormatter(i)
			return nil
		}
	}
	return errCouldNotParse("PayloadFormatter")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PayloadFormatter) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("PayloadFormatter")(string(b)).WithCause(err)
	}
	*v = PayloadFormatter(i)
	return nil
}

// FieldIsZero returns whether path p is zero.
func (v *MessagePayloadFormatters) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "down_formatter":
		return v.DownFormatter == 0
	case "down_formatter_parameter":
		return v.DownFormatterParameter == ""
	case "up_formatter":
		return v.UpFormatter == 0
	case "up_formatter_parameter":
		return v.UpFormatterParameter == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ApplicationDownlink_ClassBC) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "absolute_time":
		return v.AbsoluteTime == nil
	case "gateways":
		return v.Gateways == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *ApplicationDownlink) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "class_b_c":
		return v.ClassBC == nil
	case "class_b_c.absolute_time":
		return v.ClassBC.FieldIsZero("absolute_time")
	case "class_b_c.gateways":
		return v.ClassBC.FieldIsZero("gateways")
	case "confirmed":
		return !v.Confirmed
	case "correlation_ids":
		return v.CorrelationIDs == nil
	case "decoded_payload":
		return v.DecodedPayload == nil
	case "decoded_payload_warnings":
		return v.DecodedPayloadWarnings == nil
	case "f_cnt":
		return v.FCnt == 0
	case "f_port":
		return v.FPort == 0
	case "frm_payload":
		return v.FRMPayload == nil
	case "priority":
		return v.Priority == 0
	case "session_key_id":
		return v.SessionKeyID == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// MarshalText implements encoding.TextMarshaler interface.
func (v TxAcknowledgment_Result) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *TxAcknowledgment_Result) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := TxAcknowledgment_Result_value[s]; ok {
		*v = TxAcknowledgment_Result(i)
		return nil
	}
	return errCouldNotParse("TxAcknowledgment_Result")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *TxAcknowledgment_Result) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("TxAcknowledgment_Result")(string(b)).WithCause(err)
	}
	*v = TxAcknowledgment_Result(i)
	return nil
}

// PartitionDownlinks partitions downlinks based on the general predicate p.
func PartitionDownlinks(p func(down *ApplicationDownlink) bool, downs ...*ApplicationDownlink) (t, f []*ApplicationDownlink) {
	t, f = downs[:0:0], downs[:0:0]
	for _, down := range downs {
		if p(down) {
			t = append(t, down)
		} else {
			f = append(f, down)
		}
	}
	return t, f
}

// PartitionDownlinksBySessionKeyID partitions the downlinks based on the session key ID predicate p.
func PartitionDownlinksBySessionKeyID(p func([]byte) bool, downs ...*ApplicationDownlink) (t, f []*ApplicationDownlink) {
	return PartitionDownlinks(func(down *ApplicationDownlink) bool { return p(down.SessionKeyID) }, downs...)
}

// PartitionDownlinksBySessionKeyIDEquality partitions the downlinks based on the equality to the given session key ID.
func PartitionDownlinksBySessionKeyIDEquality(id []byte, downs ...*ApplicationDownlink) (t, f []*ApplicationDownlink) {
	return PartitionDownlinksBySessionKeyID(func(downID []byte) bool { return bytes.Equal(downID, id) }, downs...)
}
