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

// Equal returns true iff keys are equal
func (key AES128Key) Equal(other AES128Key) bool { return key == other }

// Marshal implements the proto.Marshaler interface
func (key AES128Key) Marshal() ([]byte, error) { return key.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf
func (key AES128Key) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, key[:]) }

// Unmarshal implements the proto.Unmarshaler interface
func (key *AES128Key) Unmarshal(data []byte) error { return key.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface
func (key AES128Key) MarshalJSON() ([]byte, error) { return marshalJSONBytes(key[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface
func (key *AES128Key) UnmarshalJSON(data []byte) error {
	*key = [16]byte{}
	return unmarshalJSONBytes(key[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (key AES128Key) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(key[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (key *AES128Key) UnmarshalBinary(data []byte) error {
	*key = [16]byte{}
	return unmarshalBinaryBytes(key[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface
func (key AES128Key) MarshalText() ([]byte, error) { return marshalTextBytes(key[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (key *AES128Key) UnmarshalText(data []byte) error {
	*key = [16]byte{}
	return unmarshalTextBytes(key[:], data)
}
