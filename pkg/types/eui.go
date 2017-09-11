// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
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

// MarshalTo implements the proto.MarshalerTo interface
func (eui EUI64) MarshalTo(data []byte) (int, error) { return copy(data, eui[:]), nil }

// Unmarshal implements the proto.Unmarshaler interface
func (eui *EUI64) Unmarshal(data []byte) error {
	*eui = [8]byte{}
	if len(data) != 8 || copy(eui[:], data) != 8 {
		return ErrInvalidLength
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (eui EUI64) MarshalBinary() ([]byte, error) { return eui[:], nil }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (eui *EUI64) UnmarshalBinary(data []byte) error { return eui.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (eui EUI64) MarshalText() ([]byte, error) { return []byte(eui.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (eui *EUI64) UnmarshalText(data []byte) error {
	if len(data) != 16 {
		return ErrInvalidLength
	}
	_, err := hex.Decode(eui[:], data)
	if err != nil {
		return err
	}
	return nil
}
