// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"strings"

	"github.com/TheThingsIndustries/protoc-gen-go-flags/flagsplugin"
	"github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/customflags"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errInvalidNetID = errors.DefineInvalidArgument("invalid_net_id", "invalid NetID")

// NetID is issued by the LoRa Alliance.
type NetID [3]byte

// IsZero returns true iff the type is zero.
func (id NetID) IsZero() bool { return id == [3]byte{} }

func (id NetID) String() string { return strings.ToUpper(hex.EncodeToString(id[:])) }

func (id NetID) GoString() string { return id.String() }

func (id NetID) Bytes() []byte {
	b := make([]byte, 3)
	copy(b, id[:])
	return b
}

// GetNetID gets a typed NetID from the bytes.
// It returns nil, nil if b is nil.
// It returns an error if unmarshaling fails.
func GetNetID(b []byte) (*NetID, error) {
	if b == nil {
		return nil, nil
	}
	var t NetID
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return &t, nil
}

// MustNetID returns a typed NetID from the bytes.
// It returns nil if b is nil.
// It panics if unmarshaling results in an error.
func MustNetID(b []byte) *NetID {
	t, err := GetNetID(b)
	if err != nil {
		panic(err)
	}
	return t
}

// OrZero returns the NetID value, or a zero value if the NetID was nil.
func (id *NetID) OrZero() NetID {
	if id != nil {
		return *id
	}
	return NetID{}
}

// Equal returns true iff IDs are equal.
func (id NetID) Equal(other NetID) bool { return id == other }

// Size implements the Sizer interface.
func (id NetID) Size() int { return 3 }

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
	if err := unmarshalJSONHexBytes(id[:], data); err != nil {
		return errInvalidNetID.WithCause(err)
	}
	return nil
}

// MarshalProtoJSON implements the jsonplugin.Marshaler interface.
func (id *NetID) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if id == nil {
		s.WriteNil()
		return
	}
	s.WriteString(fmt.Sprintf("%X", id[:]))
}

// UnmarshalProtoJSON implements the jsonplugin.Unmarshaler interface.
func (id *NetID) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	*id = [3]byte{}
	b, err := hex.DecodeString(s.ReadString())
	if err != nil {
		s.SetError(err)
		return
	}
	if len(b) != 3 {
		s.SetError(errInvalidNetID.WithCause(errInvalidLength.WithAttributes(
			"want", 3, "got", len(b), "got_type", "bytes",
		)))
		return
	}
	copy(id[:], b)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (id NetID) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(id[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (id *NetID) UnmarshalBinary(data []byte) error {
	*id = [3]byte{}
	if err := unmarshalBinaryBytes(id[:], data); err != nil {
		return errInvalidNetID.WithCause(err)
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (id NetID) MarshalText() ([]byte, error) { return marshalTextBytes(id[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (id *NetID) UnmarshalText(data []byte) error {
	*id = [3]byte{}
	if err := unmarshalTextBytes(id[:], data); err != nil {
		return errInvalidNetID.WithCause(err)
	}
	return nil
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
		return errNetIDOverflow.New()
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
		panic("unreachable")
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
	panic("unreachable")
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
		return NetID{}, errNetIDType.New()
	}

	if len(id) < 3 {
		id = append(make([]byte, 3-len(id)), id...)
	}
	if id[0]&0xe0 > 0 {
		return NetID{}, errNetIDBits.New()
	}
	copy(netID[:], id)
	netID[0] = netID[0]&0x1f | typ<<5
	return netID, nil
}

// GetNetIDFromFlag gets a NetID from a named flag in the flag set.
func GetNetIDFromFlag(fs *pflag.FlagSet, name string) (value NetID, set bool, err error) {
	flag := fs.Lookup(name)
	if flag == nil {
		return NetID{}, false, &flagsplugin.ErrFlagNotFound{FlagName: name}
	}
	var netID NetID
	if !flag.Changed {
		return netID, flag.Changed, nil
	}
	if err := netID.Unmarshal(flag.Value.(*customflags.ExactBytesValue).Value); err != nil {
		return netID, false, err
	}
	return netID, flag.Changed, nil
}
