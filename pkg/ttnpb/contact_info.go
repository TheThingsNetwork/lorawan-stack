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
func (v ContactType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ContactType) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := ContactType_value[s]; ok {
		*v = ContactType(i)
		return nil
	}
	if !strings.HasPrefix(s, "CONTACT_TYPE_") {
		if i, ok := ContactType_value["CONTACT_TYPE_"+s]; ok {
			*v = ContactType(i)
			return nil
		}
	}
	return errCouldNotParse("ContactType")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ContactType) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("ContactType")(string(b)).WithCause(err)
	}
	*v = ContactType(i)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v ContactMethod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ContactMethod) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := ContactMethod_value[s]; ok {
		*v = ContactMethod(i)
		return nil
	}
	if !strings.HasPrefix(s, "CONTACT_METHOD_") {
		if i, ok := ContactMethod_value["CONTACT_METHOD_"+s]; ok {
			*v = ContactMethod(i)
			return nil
		}
	}
	return errCouldNotParse("ContactMethod")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *ContactMethod) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("ContactMethod")(string(b)).WithCause(err)
	}
	*v = ContactMethod(i)
	return nil
}
