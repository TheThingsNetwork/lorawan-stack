// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"encoding/hex"
	"strings"
)

// NetID is issued by the LoRa Alliance
type NetID [3]byte

// IsZero returns true iff the type is zero
func (id NetID) IsZero() bool { return id == [3]byte{} }

// String implements the Stringer interface
func (id NetID) String() string { return strings.ToUpper(hex.EncodeToString(id[:])) }

// GoString implements the GoStringer interface
func (id NetID) GoString() string { return id.String() }

// Size implements the Sizer interface
func (id NetID) Size() int { return 3 }

// Equal returns true iff IDs are equal
func (id NetID) Equal(other NetID) bool { return id == other }

// Marshal implements the proto.Marshaler interface
func (id NetID) Marshal() ([]byte, error) { return id.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf
func (id NetID) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, id[:]) }

// Unmarshal implements the proto.Unmarshaler interface
func (id *NetID) Unmarshal(data []byte) error { return id.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface
func (id NetID) MarshalJSON() ([]byte, error) { return marshalJSONBytes(id[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface
func (id *NetID) UnmarshalJSON(data []byte) error {
	*id = [3]byte{}
	return unmarshalJSONBytes(id[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (id NetID) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(id[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (id *NetID) UnmarshalBinary(data []byte) error {
	*id = [3]byte{}
	return unmarshalBinaryBytes(id[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface
func (id NetID) MarshalText() ([]byte, error) { return marshalTextBytes(id[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (id *NetID) UnmarshalText(data []byte) error {
	*id = [3]byte{}
	return unmarshalTextBytes(id[:], data)
}

// NwkID contained in the NetID
func (id NetID) NwkID() byte {
	return id[2] & 127
}

// Value implements driver.Valuer interface.
func (id NetID) Value() (driver.Value, error) {
	return id.MarshalText()
}

// Scan implements sql.Scanner interface.
func (id *NetID) Scan(src interface{}) error {
	data, ok := src.([]byte)
	if !ok {
		return ErrTypeAssertion
	}
	return id.UnmarshalText(data)
}
