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

package messages

import (
	"fmt"
	"regexp"
	"strconv"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// EUI is an EUI that can be marshaled to an ID6 string and unmarshaled from a ID6 or hex string.
type EUI types.EUI64

// MarshalJSON implements json.Marshaler.
func (eui EUI) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%x:%x:%x:%x"`,
		uint16(eui[0])<<8|uint16(eui[1]),
		uint16(eui[2])<<8|uint16(eui[3]),
		uint16(eui[4])<<8|uint16(eui[5]),
		uint16(eui[6])<<8|uint16(eui[7]),
	)), nil
}

var (
	hexPattern = regexp.MustCompile(`^"([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})"$`)
	id6Pattern = regexp.MustCompile(`^"([a-fA-F0-9]{0,4}):([a-fA-F0-9]{0,4}):([a-fA-F0-9]{0,4}):([a-fA-F0-9]{0,4})"$`)
)

var errFormat = errors.DefineInvalidArgument("format", "invalid format")

// UnmarshalJSON implements json.Unmarshaler.
func (eui *EUI) UnmarshalJSON(data []byte) error {
	if bytes := hexPattern.FindStringSubmatch(string(data)); bytes != nil {
		for i, b := range bytes[1:] {
			v, err := strconv.ParseUint(b, 16, 8)
			if err != nil {
				return errFormat.WithCause(err)
			}
			eui[i] = uint8(v)
		}
		return nil
	}
	if bytes := id6Pattern.FindStringSubmatch(string(data)); bytes != nil {
		for i, b := range bytes[1:] {
			if b == "" {
				b = "0"
			}
			v, err := strconv.ParseUint(b, 16, 16)
			if err != nil {
				return errFormat.WithCause(err)
			}
			eui[2*i] = uint8(v >> 8)
			eui[2*i+1] = uint8(v)
		}
		return nil
	}
	return errFormat
}
