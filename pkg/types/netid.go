// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"encoding/hex"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

const unmatchedNetID = "unmatched NetID type"

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
func (id NetID) MarshalJSON() ([]byte, error) { return marshalJSONHexBytes(id[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (id *NetID) UnmarshalJSON(data []byte) error {
	*id = [3]byte{}
	return unmarshalJSONHexBytes(id[:], data)
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

// MarshalNumber returns the numeric value.
func (id NetID) MarshalNumber() uint32 {
	return uint32(id[0])<<16 | uint32(id[1])<<8 | uint32(id[2])
}

var errNetIDOverflow = errors.DefineInvalidArgument("net_id_overflow", "NetID overflow")

// UnmarshalNumber unmarshals the NetID from a numeric value.
func (id *NetID) UnmarshalNumber(netID uint32) error {
	*id = [3]byte{}
	if netID > 0xFFFFFF {
		return errNetIDOverflow
	}
	id[0] = byte(netID >> 16)
	id[1] = byte(netID >> 8)
	id[2] = byte(netID)
	return nil
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
		panic(unmatchedNetID)
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
	panic(unmatchedNetID)
}

// Copy stores a copy of id in x and returns it.
func (id NetID) Copy(x *NetID) *NetID {
	copy(x[:], id[:])
	return x
}

var (
	errNetIDType = errors.DefineInvalidArgument("net_id_type", "NetID type must be lower or equal to 7")
	errNetIDBits = errors.DefineInvalidArgument("net_id_bits", "too many bits set in NetID")
)

// NewNetID returns new NetID.
func NewNetID(typ byte, id []byte) (netID NetID, err error) {
	if typ > 7 {
		return NetID{}, errNetIDType
	}

	if len(id) < 3 {
		id = append(make([]byte, 3-len(id)), id...)
	}
	if id[0]&0xe0 > 0 {
		return NetID{}, errNetIDBits
	}
	copy(netID[:], id)
	netID[0] = netID[0]&0x1f | typ<<5
	return netID, nil
}
