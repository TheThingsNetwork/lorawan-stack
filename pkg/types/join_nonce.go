// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"encoding/hex"
	"strings"
)

// JoinNonce is randomly generated in the join procedure
type JoinNonce [3]byte

// IsZero returns true iff the type is zero
func (jn JoinNonce) IsZero() bool { return jn == [3]byte{} }

// String implements the Stringer interface
func (jn JoinNonce) String() string { return strings.ToUpper(hex.EncodeToString(jn[:])) }

// GoString implements the GoStringer interface
func (jn JoinNonce) GoString() string { return jn.String() }

// Size implements the Sizer interface
func (jn JoinNonce) Size() int { return 3 }

// Equal returns true iff nonces are equal
func (jn JoinNonce) Equal(other JoinNonce) bool { return jn == other }

// MarshalTo implements the proto.MarshalerTo interface
func (jn JoinNonce) MarshalTo(data []byte) (int, error) { return copy(data, jn[:]), nil }

// Unmarshal implements the proto.Unmarshaler interface
func (jn *JoinNonce) Unmarshal(data []byte) error {
	*jn = [3]byte{}
	if len(data) != 3 || copy(jn[:], data) != 3 {
		return ErrInvalidLength
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (jn JoinNonce) MarshalBinary() ([]byte, error) { return jn[:], nil }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (jn *JoinNonce) UnmarshalBinary(data []byte) error { return jn.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (jn JoinNonce) MarshalText() ([]byte, error) { return []byte(jn.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (jn *JoinNonce) UnmarshalText(data []byte) error {
	if len(data) != 6 {
		return ErrInvalidLength
	}
	_, err := hex.Decode(jn[:], data)
	if err != nil {
		return err
	}
	return nil
}
