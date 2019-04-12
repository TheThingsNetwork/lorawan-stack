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
func (v DownlinkPathConstraint) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *DownlinkPathConstraint) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := DownlinkPathConstraint_value[s]; ok {
		*v = DownlinkPathConstraint(i)
		return nil
	}
	if !strings.HasPrefix(s, "DOWNLINK_PATH_CONSTRAINT_") {
		if i, ok := DownlinkPathConstraint_value["DOWNLINK_PATH_CONSTRAINT_"+s]; ok {
			*v = DownlinkPathConstraint(i)
			return nil
		}
	}
	return errCouldNotParse("DownlinkPathConstraint")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *DownlinkPathConstraint) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("DownlinkPathConstraint")(string(b)).WithCause(err)
	}
	*v = DownlinkPathConstraint(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v State) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *State) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := State_value[s]; ok {
		*v = State(i)
		return nil
	}
	if !strings.HasPrefix(s, "STATE_") {
		if i, ok := State_value["STATE_"+s]; ok {
			*v = State(i)
			return nil
		}
	}
	return errCouldNotParse("State")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *State) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("State")(string(b)).WithCause(err)
	}
	*v = State(i)
	return nil
}
