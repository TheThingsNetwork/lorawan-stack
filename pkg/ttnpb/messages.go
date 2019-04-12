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
