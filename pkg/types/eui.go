// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// EUI64 is a 64-bit Extended Unique Identifier
type EUI64 [8]byte

// IsZero returns true iff the type is zero
func (eui EUI64) IsZero() bool { return eui == [8]byte{} }

// String implements the Stringer interface
func (eui EUI64) String() string { return strings.ToUpper(hex.EncodeToString(eui[:])) }

// GoString implements the GoStringer interface
func (eui EUI64) GoString() string { return eui.String() }

// Size implements the Sizer interface
func (eui EUI64) Size() int { return 8 }

// Equal returns true iff EUIs are equal
func (eui EUI64) Equal(other EUI64) bool { return eui == other }

// Marshal implements the proto.Marshaler interface
func (eui EUI64) Marshal() ([]byte, error) { return eui.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf
func (eui EUI64) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, eui[:]) }

// Unmarshal implements the proto.Unmarshaler interface
func (eui *EUI64) Unmarshal(data []byte) error { return eui.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface
func (eui EUI64) MarshalJSON() ([]byte, error) { return marshalJSONBytes(eui[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface
func (eui *EUI64) UnmarshalJSON(data []byte) error {
	*eui = [8]byte{}
	return unmarshalJSONBytes(eui[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (eui EUI64) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(eui[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (eui *EUI64) UnmarshalBinary(data []byte) error {
	*eui = [8]byte{}
	return unmarshalBinaryBytes(eui[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface
func (eui EUI64) MarshalText() ([]byte, error) { return marshalTextBytes(eui[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (eui *EUI64) UnmarshalText(data []byte) error {
	*eui = [8]byte{}
	return unmarshalTextBytes(eui[:], data)
}

// Value implements driver.Valuer interface.
func (eui EUI64) Value() (driver.Value, error) {
	return eui.MarshalText()
}

// Scan implements sql.Scanner interface.
func (eui *EUI64) Scan(src interface{}) error {
	data, ok := src.([]byte)
	if !ok {
		return ErrTypeAssertion
	}
	return eui.UnmarshalText(data)
}

// MarshalNumber returns the EUI64 in a decimal form
func (eui EUI64) MarshalNumber() uint64 {
	return binary.BigEndian.Uint64(eui[:])
}

// UnmarshalNumber retrieves a EUI64 from a decimal form
func (eui *EUI64) UnmarshalNumber(n uint64) {
	*eui = [8]byte{}
	binary.BigEndian.PutUint64(eui[:], n)
}

// Before returns true if the EUI64 is strictly inferior to the EUI64 passed as an argument
func (eui EUI64) Before(a EUI64) bool {
	if eui.MarshalNumber() < a.MarshalNumber() {
		return true
	}
	return false
}

// After returns true if the EUI64 is strictly superior to the EUI64 passed as an argument
func (eui EUI64) After(a EUI64) bool {
	if eui.MarshalNumber() > a.MarshalNumber() {
		return true
	}
	return false
}

// BeforeOrEqual returns true if the EUI64 is inferior or equal to the EUI64 passed as an argument
func (eui EUI64) BeforeOrEqual(a EUI64) bool {
	return eui.Before(a) || eui == a
}

// AfterOrEqual returns true if the EUI64 is superior or equal to the EUI64 passed as an argument
func (eui EUI64) AfterOrEqual(a EUI64) bool {
	return eui.After(a) || eui == a
}

// ErrInvalidEUI64Prefix can be returned when unmarshaling an invalid slice into a prefix
var ErrInvalidEUI64Prefix = errors.New("invalid EUI prefix")

type EUI64Prefix struct {
	EUI64  EUI64
	Length uint8
}

// NbItems returns the number of items that this prefix encapsulates
func (prefix EUI64Prefix) NbItems() uint64 {
	return uint64(math.Pow(2, float64(64-prefix.Length)))
}

// FirstEUI64Covered returns the first EUI64 covered, in the numeric order
func (prefix EUI64Prefix) FirstEUI64Covered() EUI64 {
	return prefix.EUI64.Mask(prefix.Length)
}

func (prefix EUI64Prefix) firstNumericEUI64Covered() uint64 {
	return prefix.FirstEUI64Covered().MarshalNumber()
}

func (prefix EUI64Prefix) lastNumericEUI64Covered() uint64 {
	return prefix.firstNumericEUI64Covered() + prefix.NbItems() - 1
}

// LastEUI64Covered returns the last EUI64 covered, in the numeric order
func (prefix EUI64Prefix) LastEUI64Covered() EUI64 {
	result := EUI64{}

	lastEUI64Numeric := prefix.lastNumericEUI64Covered()
	hex := fmt.Sprintf("%08X", lastEUI64Numeric)
	result.UnmarshalText([]byte(hex))

	return result
}

// IsZero returns true iff the type is zero
func (prefix EUI64Prefix) IsZero() bool { return prefix.Length == 0 }

// String implements the Stringer interface
func (prefix EUI64Prefix) String() string {
	return fmt.Sprintf("%s/%d", prefix.EUI64, prefix.Length)
}

// GoString implements the GoStringer interface
func (prefix EUI64Prefix) GoString() string { return prefix.String() }

// Size implements the Sizer interface
func (prefix EUI64Prefix) Size() int { return 9 }

// Equal returns true iff prefixes are equal
func (prefix EUI64Prefix) Equal(other EUI64Prefix) bool {
	return prefix.Length == other.Length && prefix.EUI64.Equal(other.EUI64)
}

// MarshalTo implements the MarshalerTo function required by generated protobuf
func (prefix EUI64Prefix) MarshalTo(data []byte) (int, error) {
	return marshalBinaryBytesTo(data, append(prefix.EUI64[:], prefix.Length))
}

// Marshal implements the proto.Marshaler interface
func (prefix EUI64Prefix) Marshal() ([]byte, error) { return prefix.MarshalBinary() }

// Unmarshal implements the proto.Unmarshaler interface
func (prefix *EUI64Prefix) Unmarshal(data []byte) error { return prefix.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface
func (prefix EUI64Prefix) MarshalJSON() ([]byte, error) {
	return append([]byte(`"`+base64Encoding.EncodeToString(prefix.EUI64[:])), '/', prefix.Length, '"'), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (prefix *EUI64Prefix) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		*prefix = EUI64Prefix{}
		return nil
	}
	if len(data) != 16 {
		return ErrInvalidLength
	}
	if data[13] != '/' {
		return ErrInvalidEUI64Prefix
	}
	b := make([]byte, base64Encoding.DecodedLen(16))
	n, err := base64Encoding.Decode(b, data[1:13])
	if err != nil {
		return err
	}
	if n != 8 || copy(prefix.EUI64[:], b) != 8 {
		return ErrInvalidEUI64Prefix
	}
	prefix.Length = data[14]
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (prefix EUI64Prefix) MarshalBinary() ([]byte, error) {
	return marshalBinaryBytes(append(prefix.EUI64[:], prefix.Length))
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (prefix *EUI64Prefix) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		*prefix = EUI64Prefix{}
		return nil
	}
	if len(data) != 9 {
		return ErrInvalidLength
	}
	if err := prefix.EUI64.Unmarshal(data[:8]); err != nil {
		return err
	}
	prefix.Length = data[8]
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (prefix EUI64Prefix) MarshalText() ([]byte, error) {
	b, err := prefix.EUI64.MarshalText()
	if err != nil {
		return nil, err
	}
	// transform length into digit character range
	return append(append(b, '/'), []byte(strconv.Itoa(int(prefix.Length)))...), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (prefix *EUI64Prefix) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*prefix = EUI64Prefix{}
		return nil
	}
	if len(data) != 18 && len(data) != 19 {
		return ErrInvalidLength
	}
	if data[16] != '/' {
		return ErrInvalidEUI64Prefix
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

// Value implements driver.Valuer interface.
func (prefix EUI64Prefix) Value() (driver.Value, error) {
	return prefix.MarshalText()
}

// Scan implements sql.Scanner interface.
func (prefix *EUI64Prefix) Scan(src interface{}) error {
	data, ok := src.([]byte)
	if !ok {
		return ErrTypeAssertion
	}
	return prefix.UnmarshalText(data)
}

// WithPrefix returns the EUI64, but with the first length bits replaced by the Prefix
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

// Mask returns a copy of the EUI64 with only the first "bits" bits
func (eui EUI64) Mask(bits uint8) (masked EUI64) {
	return (EUI64{}).WithPrefix(EUI64Prefix{eui, bits})
}

// HasPrefix returns true iff the EUI64 has a prefix of given length
func (eui EUI64) HasPrefix(prefix EUI64Prefix) bool { return prefix.Matches(eui) }

// Matches returns true iff the EUI64 matches the prefix
func (prefix EUI64Prefix) Matches(eui EUI64) bool {
	return eui.Mask(prefix.Length) == prefix.EUI64.Mask(prefix.Length)
}
