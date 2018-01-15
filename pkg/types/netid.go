// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// NetID is issued by the LoRa Alliance.
type NetID [3]byte

// IsZero returns true iff the type is zero.
func (id NetID) IsZero() bool { return id == [3]byte{} }

// String implements the Stringer interface.
func (id NetID) String() string { return strings.ToUpper(hex.EncodeToString(id[:])) }

// GoString implements the GoStringer interface.
func (id NetID) GoString() string { return id.String() }

// Size implements the Sizer interface.
func (id NetID) Size() int { return 3 }

// Equal returns true iff IDs are equal.
func (id NetID) Equal(other NetID) bool { return id == other }

// Marshal implements the proto.Marshaler interface.
func (id NetID) Marshal() ([]byte, error) { return id.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (id NetID) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, id[:]) }

// Unmarshal implements the proto.Unmarshaler interface.
func (id *NetID) Unmarshal(data []byte) error { return id.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (id NetID) MarshalJSON() ([]byte, error) { return marshalJSONBytes(id[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (id *NetID) UnmarshalJSON(data []byte) error {
	*id = [3]byte{}
	return unmarshalJSONBytes(id[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (id NetID) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(id[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (id *NetID) UnmarshalBinary(data []byte) error {
	*id = [3]byte{}
	return unmarshalBinaryBytes(id[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (id NetID) MarshalText() ([]byte, error) { return marshalTextBytes(id[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (id *NetID) UnmarshalText(data []byte) error {
	*id = [3]byte{}
	return unmarshalTextBytes(id[:], data)
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

// Type returns NetID type.
func (id NetID) Type() byte {
	return id[0] >> 5
}

// ID returns ID contained in the NetID.
func (id NetID) ID() []byte {
	switch id.Type() {
	case 0, 1:
		// 6 LSB
		return []byte{id[2] & 0x3f}
	case 2:
		// 9 LSB
		return []byte{id[1] & 0x01, id[2]}
	case 3, 4, 5, 6, 7:
		// 21 LSB
		return []byte{id[0] & 0x1f, id[1], id[2]}
	default:
		panic(fmt.Errorf("Unmatched NetID type: %d", id.Type()))
	}
}

// IDBits returns the bit-length of ID represented by the NetID.
func (id NetID) IDBits() uint {
	switch id.Type() {
	case 0, 1:
		return 6
	case 2:
		return 9
	case 3, 4, 5, 6, 7:
		return 21
	}
	panic(fmt.Errorf("Unmatched NetID type: %d", id.Type()))
}

// NewNetID returns new NetID.
func NewNetID(typ byte, id []byte) (netID NetID, err error) {
	if typ > 7 {
		return NetID{}, fmt.Errorf("NetID type must be lower or equal to 7, got: %d", typ)
	}

	if len(id) < 3 {
		id = append(make([]byte, 3-len(id)), id...)
	}
	if id[0]&0xe0 > 0 {
		return NetID{}, errors.New("Too many bits set in id")
	}
	copy(netID[:], id)
	netID[0] = netID[0]&0x1f | typ<<5
	return netID, nil
}
