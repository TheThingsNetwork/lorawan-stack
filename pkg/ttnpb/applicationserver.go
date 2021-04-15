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

import "strconv"

// MarshalText implements encoding.TextMarshaler interface.
func (v AsConfiguration_PubSub_Providers_Status) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *AsConfiguration_PubSub_Providers_Status) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := AsConfiguration_PubSub_Providers_Status_value[s]; ok {
		*v = AsConfiguration_PubSub_Providers_Status(i)
		return nil
	}
	return errCouldNotParse("AsConfiguration_PubSub_Providers_Status")(string(b))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (v *AsConfiguration_PubSub_Providers_Status) UnmarshalJSON(b []byte) error {
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return v.UnmarshalText(b[1 : len(b)-1])
	}
	i, err := strconv.Atoi(string(b))
	if err != nil {
		return errCouldNotParse("AsConfiguration_PubSub_Providers_Status")(string(b)).WithCause(err)
	}
	*v = AsConfiguration_PubSub_Providers_Status(i)
	return nil
}
