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

	"github.com/gogo/protobuf/jsonpb"
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

// MarshalText implements encoding.TextMarshaler interface.
func (v KeyProvisioning) MarshalText() ([]byte, error) {
	switch v {
	case KEY_PROVISIONING_UNKNOWN:
		return []byte("unknown"), nil
	case KEY_PROVISIONING_CUSTOM:
		return []byte("custom"), nil
	case KEY_PROVISIONING_JOIN_SERVER:
		return []byte("join server"), nil
	case KEY_PROVISIONING_MANIFEST:
		return []byte("manifest"), nil
	}

	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaller interface.
func (v *KeyProvisioning) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := KeyProvisioning_value[s]; ok {
		*v = KeyProvisioning(i)
		return nil
	}
	switch s {
	case "unknown", "KEY_PROVISIONING_UNKNOWN":
		*v = KEY_PROVISIONING_UNKNOWN
	case "custom", "KEY_PROVISIONING_CUSTOM":
		*v = KEY_PROVISIONING_CUSTOM
	case "join server", "KEY_PROVISIONING_JOIN_SERVER":
		*v = KEY_PROVISIONING_JOIN_SERVER
	case "manifest", "KEY_PROVISIONING_MANIFEST":
		*v = KEY_PROVISIONING_MANIFEST
	default:
		errCouldNotParse("KeyProvisioning")(string(b))
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *KeyProvisioning) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}

	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("KeyProvisioning")(string(b)).WithCause(err)
	}
	*v = KeyProvisioning(i)
	return nil
}

// UnmarshalJSONPB implements json.Unmarshaler interface.
func (v *KeyProvisioning) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
}

// MarshalText implements encoding.TextMarshaler interface.
func (v KeySecurity) MarshalText() ([]byte, error) {
	switch v {
	case KEY_SECURITY_UNKNOWN:
		return []byte("unknown"), nil
	case KEY_SECURITY_NONE:
		return []byte("none"), nil
	case KEY_SECURITY_READ_PROTECTED:
		return []byte("read protected"), nil
	case KEY_SECURITY_SECURE_ELEMENT:
		return []byte("secure element"), nil
	}

	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaller interface.
func (v *KeySecurity) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := KeySecurity_value[s]; ok {
		*v = KeySecurity(i)
		return nil
	}
	switch s {
	case "unknown", "KEY_SECURITY_UNKNOWN":
		*v = KEY_SECURITY_UNKNOWN
	case "none", "KEY_SECURITY_NONE":
		*v = KEY_SECURITY_NONE
	case "read protected", "KEY_SECURITY_READ_PROTECTED":
		*v = KEY_SECURITY_READ_PROTECTED
	case "secure element", "KEY_SECURITY_SECURE_ELEMENT":
		*v = KEY_SECURITY_SECURE_ELEMENT
	default:
		errCouldNotParse("KeySecurity")(string(b))
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *KeySecurity) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("KeySecurity")(string(b)).WithCause(err)
	}
	*v = KeySecurity(i)
	return nil
}

// UnmarshalJSONPB implements json.Unmarshaler interface.
func (v *KeySecurity) UnmarshalJSONPB(u *jsonpb.Unmarshaler, b []byte) error {
	return v.UnmarshalJSON(b)
}
