// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
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

// MarshalTo implements the proto.MarshalerTo interface
func (id NetID) MarshalTo(data []byte) (int, error) { return copy(data, id[:]), nil }

// Unmarshal implements the proto.Unmarshaler interface
func (id *NetID) Unmarshal(data []byte) error {
	*id = [3]byte{}
	if len(data) != 3 || copy(id[:], data) != 3 {
		return ErrInvalidLength
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (id NetID) MarshalBinary() ([]byte, error) { return id[:], nil }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (id *NetID) UnmarshalBinary(data []byte) error { return id.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (id NetID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (id *NetID) UnmarshalText(data []byte) error {
	if len(data) != 6 {
		return ErrInvalidLength
	}
	_, err := hex.Decode(id[:], data)
	if err != nil {
		return err
	}
	return nil
}

// NwkID contained in the NetID
func (id NetID) NwkID() byte {
	return id[2] & 127
}
