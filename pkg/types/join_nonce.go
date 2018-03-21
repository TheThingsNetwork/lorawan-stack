// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"encoding/hex"
	"strings"
)

// JoinNonce is randomly generated in the join procedure.
type JoinNonce [3]byte

// IsZero returns true iff the type is zero.
func (jn JoinNonce) IsZero() bool { return jn == [3]byte{} }

// String implements the Stringer interface.
func (jn JoinNonce) String() string { return strings.ToUpper(hex.EncodeToString(jn[:])) }

// GoString implements the GoStringer interface.
func (jn JoinNonce) GoString() string { return jn.String() }

// Size implements the Sizer interface.
func (jn JoinNonce) Size() int { return 3 }

// Equal returns true iff nonces are equal.
func (jn JoinNonce) Equal(other JoinNonce) bool { return jn == other }

// Marshal implements the proto.Marshaler interface.
func (jn JoinNonce) Marshal() ([]byte, error) { return jn.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (jn JoinNonce) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, jn[:]) }

// Unmarshal implements the proto.Unmarshaler interface.
func (jn *JoinNonce) Unmarshal(data []byte) error { return jn.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (jn JoinNonce) MarshalJSON() ([]byte, error) { return marshalJSONHexBytes(jn[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (jn *JoinNonce) UnmarshalJSON(data []byte) error {
	*jn = [3]byte{}
	return unmarshalJSONHexBytes(jn[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (jn JoinNonce) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(jn[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (jn *JoinNonce) UnmarshalBinary(data []byte) error {
	*jn = [3]byte{}
	return unmarshalBinaryBytes(jn[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (jn JoinNonce) MarshalText() ([]byte, error) { return marshalTextBytes(jn[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (jn *JoinNonce) UnmarshalText(data []byte) error {
	*jn = [3]byte{}
	return unmarshalTextBytes(jn[:], data)
}

// Value implements driver.Valuer interface.
func (jn JoinNonce) Value() (driver.Value, error) {
	return jn.MarshalText()
}

// Scan implements sql.Scanner interface.
func (jn *JoinNonce) Scan(src interface{}) error {
	data, ok := src.([]byte)
	if !ok {
		return ErrTypeAssertion
	}
	return jn.UnmarshalText(data)
}
