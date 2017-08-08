// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"encoding/hex"
	"strings"
)

// AES128Key is an 128-bit AES key
type AES128Key [16]byte

// IsZero returns true iff the type is zero
func (key AES128Key) IsZero() bool { return key == [16]byte{} }

// String implements the Stringer interface
func (key AES128Key) String() string { return strings.ToUpper(hex.EncodeToString(key[:])) }

// GoString implements the GoStringer interface
func (key AES128Key) GoString() string { return key.String() }

// Size implements the Sizer interface
func (key AES128Key) Size() int { return 16 }

// MarshalTo implements the proto.MarshalerTo interface
func (key AES128Key) MarshalTo(data []byte) (int, error) { return copy(data, key[:]), nil }

// Unmarshal implements the proto.Unmarshaler interface
func (key *AES128Key) Unmarshal(data []byte) error {
	*key = [16]byte{}
	if len(data) != 16 || copy(key[:], data) != 16 {
		return ErrInvalidLength
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (key AES128Key) MarshalBinary() ([]byte, error) { return key[:], nil }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (key *AES128Key) UnmarshalBinary(data []byte) error { return key.Unmarshal(data) }

// MarshalText implements the encoding.TextMarshaler interface
func (key AES128Key) MarshalText() ([]byte, error) { return []byte(key.String()), nil }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (key *AES128Key) UnmarshalText(data []byte) error {
	if len(data) != 32 {
		return ErrInvalidLength
	}
	_, err := hex.Decode(key[:], data)
	if err != nil {
		return err
	}
	return nil
}
