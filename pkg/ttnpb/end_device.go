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
	"context"
	"strconv"
	"strings"
)

// MarshalText implements encoding.TextMarshaler interface.
func (v PowerState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *PowerState) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := PowerState_value[s]; ok {
		*v = PowerState(i)
		return nil
	}
	if !strings.HasPrefix(s, "POWER_") {
		if i, ok := PowerState_value["POWER_"+s]; ok {
			*v = PowerState(i)
			return nil
		}
	}
	return errCouldNotParse("PowerState")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *PowerState) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("PowerState")(string(b)).WithCause(err)
	}
	*v = PowerState(i)
	return nil
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (m *UpdateEndDeviceRequest) ValidateContext(context.Context) error {
	if len(m.FieldMask.Paths) == 0 {
		return m.ValidateFields()
	}
	return m.ValidateFields(append(fieldsWithPrefix("end_device", m.FieldMask.Paths...),
		"end_device.ids.application_ids",
		"end_device.ids.device_id",
	)...)
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (m *SetEndDeviceRequest) ValidateContext(context.Context) error {
	if len(m.FieldMask.Paths) == 0 {
		return m.ValidateFields()
	}
	return m.ValidateFields(append(fieldsWithPrefix("end_device", m.FieldMask.Paths...),
		"end_device.ids.application_ids",
		"end_device.ids.device_id",
	)...)
}
