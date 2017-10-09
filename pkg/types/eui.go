// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"encoding/hex"
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
