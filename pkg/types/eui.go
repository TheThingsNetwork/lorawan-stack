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

package types

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

// EUI64 is a 64-bit Extended Unique Identifier.
type EUI64 [8]byte

// IsZero returns true iff the type is zero.
func (eui EUI64) IsZero() bool { return eui == [8]byte{} }

// String implements the Stringer interface.
func (eui EUI64) String() string { return strings.ToUpper(hex.EncodeToString(eui[:])) }

// GoString implements the GoStringer interface.
func (eui EUI64) GoString() string { return eui.String() }

// Size implements the Sizer interface.
func (eui EUI64) Size() int { return 8 }

// Equal returns true iff EUIs are equal.
func (eui EUI64) Equal(other EUI64) bool { return eui == other }

// Marshal implements the proto.Marshaler interface.
func (eui EUI64) Marshal() ([]byte, error) { return eui.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (eui EUI64) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, eui[:]) }

// Unmarshal implements the proto.Unmarshaler interface.
func (eui *EUI64) Unmarshal(data []byte) error { return eui.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (eui EUI64) MarshalJSON() ([]byte, error) { return marshalJSONHexBytes(eui[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (eui *EUI64) UnmarshalJSON(data []byte) error {
	*eui = [8]byte{}
	return unmarshalJSONHexBytes(eui[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (eui EUI64) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(eui[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (eui *EUI64) UnmarshalBinary(data []byte) error {
	*eui = [8]byte{}
	return unmarshalBinaryBytes(eui[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (eui EUI64) MarshalText() ([]byte, error) { return marshalTextBytes(eui[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (eui *EUI64) UnmarshalText(data []byte) error {
	*eui = [8]byte{}
	return unmarshalTextBytes(eui[:], data)
}

// MarshalNumber returns the EUI64 in a decimal form.
func (eui EUI64) MarshalNumber() uint64 {
	return binary.BigEndian.Uint64(eui[:])
}

// UnmarshalNumber retrieves a EUI64 from a decimal form.
func (eui *EUI64) UnmarshalNumber(n uint64) {
	*eui = [8]byte{}
	binary.BigEndian.PutUint64(eui[:], n)
}

// EUI64Prefix is an EUI64 with a prefix length.
type EUI64Prefix struct {
	EUI64  EUI64
	Length uint8
}

// IsZero returns true iff the type is zero.
func (prefix EUI64Prefix) IsZero() bool { return prefix.Length == 0 }

// String implements the Stringer interface.
func (prefix EUI64Prefix) String() string {
	return fmt.Sprintf("%s/%d", prefix.EUI64, prefix.Length)
}

// GoString implements the GoStringer interface.
func (prefix EUI64Prefix) GoString() string { return prefix.String() }

// Size implements the Sizer interface.
func (prefix EUI64Prefix) Size() int { return 9 }

// Equal returns true iff prefixes are equal.
func (prefix EUI64Prefix) Equal(other EUI64Prefix) bool {
	return prefix.Length == other.Length && prefix.EUI64.Equal(other.EUI64)
}

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (prefix EUI64Prefix) MarshalTo(data []byte) (int, error) {
	return marshalBinaryBytesTo(data, append(prefix.EUI64[:], prefix.Length))
}

// Marshal implements the proto.Marshaler interface.
func (prefix EUI64Prefix) Marshal() ([]byte, error) { return prefix.MarshalBinary() }

// Unmarshal implements the proto.Unmarshaler interface.
func (prefix *EUI64Prefix) Unmarshal(data []byte) error { return prefix.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (prefix EUI64Prefix) MarshalJSON() ([]byte, error) {
	result := append([]byte(`"`+hex.EncodeToString(prefix.EUI64[:])), '/')
	length := strconv.Itoa(int(prefix.Length))
	result = append(result, []byte(length)...)
	return append(result, '"'), nil
}

var errInvalidEUIPrefix = errors.DefineInvalidArgument("eui_prefix", "invalid EUI prefix")

// UnmarshalJSON implements the json.Unmarshaler interface.
func (prefix *EUI64Prefix) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		*prefix = EUI64Prefix{}
		return nil
	}
	if len(data) != 20 && len(data) != 21 {
		return errInvalidEUIPrefix
	}
	if data[0] != '"' || data[len(data)-1] != '"' {
		return errInvalidJSON.WithAttributes("json", string(data))
	}
	if data[17] != '/' {
		return errInvalidEUIPrefix
	}
	b := make([]byte, hex.DecodedLen(16))
	n, err := hex.Decode(b, data[1:17])
	if err != nil {
		return err
	}
	if n != 8 || copy(prefix.EUI64[:], b) != 8 {
		return errInvalidEUIPrefix
	}
	length, err := strconv.Atoi(string(data[18 : len(data)-1]))
	if err != nil {
		return err
	}
	prefix.Length = uint8(length)
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (prefix EUI64Prefix) MarshalBinary() ([]byte, error) {
	return marshalBinaryBytes(append(prefix.EUI64[:], prefix.Length))
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (prefix *EUI64Prefix) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		*prefix = EUI64Prefix{}
		return nil
	}
	if len(data) != 9 {
		return errInvalidEUIPrefix
	}
	if err := prefix.EUI64.Unmarshal(data[:8]); err != nil {
		return err
	}
	prefix.Length = data[8]
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (prefix EUI64Prefix) MarshalText() ([]byte, error) {
	b, err := prefix.EUI64.MarshalText()
	if err != nil {
		return nil, err
	}
	// transform length into digit character range
	return append(append(b, '/'), []byte(strconv.Itoa(int(prefix.Length)))...), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (prefix *EUI64Prefix) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*prefix = EUI64Prefix{}
		return nil
	}
	if len(data) != 18 && len(data) != 19 {
		return errInvalidEUIPrefix
	}
	if data[16] != '/' {
		return errInvalidEUIPrefix
	}
	if err := prefix.EUI64.UnmarshalText(data[:16]); err != nil {
		return err
	}
	// transform length from number character range
	if len(data) == 18 {
		prefix.Length = data[17] - '0'
	} else {
		prefix.Length = (data[17]-'0')*10 + (data[18] - '0')
	}
	return nil
}

// FromConfigString implements the config.Configurable interface
func (prefix EUI64Prefix) FromConfigString(in string) (interface{}, error) {
	p := new(EUI64Prefix)
	if err := p.UnmarshalText([]byte(in)); err != nil {
		return nil, err
	}
	return p, nil
}

// ConfigString implements the config.Stringer interface
func (prefix EUI64Prefix) ConfigString() string {
	return prefix.String()
}

// WithPrefix returns the EUI64, but with the first length bits replaced by the Prefix.
func (eui EUI64) WithPrefix(prefix EUI64Prefix) (prefixed EUI64) {
	k := uint(prefix.Length)
	for i := 0; i < 8; i++ {
		if k < 8 {
			prefixed[i] = (prefix.EUI64[i] & ^byte(0xff>>k)) | (eui[i] & byte(0xff>>k))
			return
		}
		prefixed[i] = prefix.EUI64[i] & 0xff
		k -= 8
	}
	return
}

// Mask returns a copy of the EUI64 with only the first "bits" bits.
func (eui EUI64) Mask(bits uint8) (masked EUI64) {
	return (EUI64{}).WithPrefix(EUI64Prefix{eui, bits})
}

// HasPrefix returns true iff the EUI64 has a prefix of given length.
func (eui EUI64) HasPrefix(prefix EUI64Prefix) bool { return prefix.Matches(eui) }

// Matches returns true iff the EUI64 matches the prefix.
func (prefix EUI64Prefix) Matches(eui EUI64) bool {
	return eui.Mask(prefix.Length) == prefix.EUI64.Mask(prefix.Length)
}

// Copy stores a copy of eui in x and returns it.
func (eui EUI64) Copy(x *EUI64) *EUI64 {
	copy(x[:], eui[:])
	return x
}
