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

package basicstation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// EUI is an EUI that can be marshaled to an ID6 string and unmarshaled from a ID6 or hex string.
type EUI struct {
	types.EUI64
	Prefix string
}

// MarshalJSON implements json.Marshaler.
func (eui EUI) MarshalJSON() ([]byte, error) {
	var res string
	if eui.Prefix != "" {
		res += strings.ToLower(eui.Prefix) + "-"
	}
	if eui.EUI64[0] != 0 || eui.EUI64[1] != 0 {
		res += fmt.Sprintf("%x:", uint16(eui.EUI64[0])<<8|uint16(eui.EUI64[1]))
	}
	for _, g := range []uint16{
		uint16(eui.EUI64[2])<<8 | uint16(eui.EUI64[3]),
		uint16(eui.EUI64[4])<<8 | uint16(eui.EUI64[5]),
	} {
		if g != 0 {
			res += fmt.Sprintf("%x", g)
		}
		res += ":"
	}
	res += fmt.Sprintf("%x", uint16(eui.EUI64[6])<<8|uint16(eui.EUI64[7]))
	return []byte(`"` + res + `"`), nil
}

var (
	hexPatternWithDashes = regexp.MustCompile(`^"([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})-([a-fA-F0-9]{2})"$`)
	hexPatternWithColons = regexp.MustCompile(`^"([a-fA-F0-9]{2}):([a-fA-F0-9]{2}):([a-fA-F0-9]{2}):([a-fA-F0-9]{2}):([a-fA-F0-9]{2}):([a-fA-F0-9]{2}):([a-fA-F0-9]{2}):([a-fA-F0-9]{2})"$`)
	id6Pattern           = regexp.MustCompile(`^"(?:([a-z]+)-)?(?:([a-f0-9]{0,4}):)?([a-f0-9]{0,4}):([a-f0-9]{0,4}):([a-f0-9]{0,4})"$`)
)

var errFormat = errors.DefineInvalidArgument("format", "invalid format")

// UnmarshalJSON implements json.Unmarshaler.
func (eui *EUI) UnmarshalJSON(data []byte) error {
	if bytes := hexPatternWithDashes.FindStringSubmatch(string(data)); bytes != nil {
		for i, b := range bytes[1:] {
			v, err := strconv.ParseUint(b, 16, 8)
			if err != nil {
				return errFormat.WithCause(err)
			}
			eui.EUI64[i] = uint8(v)
		}
		return nil
	}
	if bytes := hexPatternWithColons.FindStringSubmatch(string(data)); bytes != nil {
		for i, b := range bytes[1:] {
			v, err := strconv.ParseUint(b, 16, 8)
			if err != nil {
				return errFormat.WithCause(err)
			}
			eui.EUI64[i] = uint8(v)
		}
		return nil
	}
	if bytes := id6Pattern.FindStringSubmatch(string(data)); bytes != nil {
		eui.Prefix = bytes[1]
		for i, b := range bytes[2:] {
			if b == "" {
				b = "0"
			}
			v, err := strconv.ParseUint(b, 16, 16)
			if err != nil {
				return errFormat.WithCause(err)
			}
			eui.EUI64[2*i] = uint8(v >> 8)
			eui.EUI64[2*i+1] = uint8(v)
		}
		return nil
	}
	return errFormat
}
