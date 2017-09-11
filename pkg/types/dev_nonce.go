// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"encoding/hex"
	"strings"
)

// DevNonce is randomly generated in the join procedure
type DevNonce [2]byte

// IsZero returns true iff the type is zero
func (dn DevNonce) IsZero() bool { return dn == [2]byte{} }

// String implements the Stringer interface
func (dn DevNonce) String() string { return strings.ToUpper(hex.EncodeToString(dn[:])) }

// GoString implements the GoStringer interface
func (dn DevNonce) GoString() string { return dn.String() }

// Size implements the Sizer interface
func (dn DevNonce) Size() int { return 2 }

// Equal returns true iff nonces are equal
func (dn DevNonce) Equal(other DevNonce) bool { return dn == other }

// MarshalTo implements the proto.MarshalerTo interface
func (dn DevNonce) MarshalTo(data []byte) (int, error) { return copy(data, dn[:]), nil }

// Unmarshal implements the proto.Unmarshaler interface
func (dn *DevNonce) Unmarshal(data []byte) error {
	*dn = [2]byte{}
	if len(data) != 2 || copy(dn[:], data) != 2 {
		return ErrInvalidLength
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (dn DevNonce) MarshalBinary() ([]byte, error) { return dn[:], nil }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (dn *DevNonce) UnmarshalBinary(data []byte) error { return dn.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (dn DevNonce) MarshalText() ([]byte, error) { return []byte(dn.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (dn *DevNonce) UnmarshalText(data []byte) error {
	if len(data) != 4 {
		return ErrInvalidLength
	}
	_, err := hex.Decode(dn[:], data)
	if err != nil {
		return err
	}
	return nil
}
