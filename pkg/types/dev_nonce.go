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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errInvalidDevNonce = errors.DefineInvalidArgument("invalid_dev_nonce", "invalid DevNonce")

// DevNonce is device nonce used in the join procedure.
// - If LoRaWAN version <1.1 - it is randomly generated.
// - If LoRaWAN version >=1.1 - it is a scrictly increasing counter.
type DevNonce [2]byte

// IsZero returns true iff the type is zero.
func (dn DevNonce) IsZero() bool { return dn == [2]byte{} }

func (dn DevNonce) String() string { return strings.ToUpper(hex.EncodeToString(dn[:])) }

func (dn DevNonce) GoString() string { return dn.String() }

func (dn DevNonce) Bytes() []byte {
	b := make([]byte, 2)
	copy(b, dn[:])
	return b
}

// GetDevNonce gets a typed DevNonce from the bytes.
// It returns nil, nil if b is nil.
// It returns an error if unmarshaling fails.
func GetDevNonce(b []byte) (*DevNonce, error) {
	if b == nil {
		return nil, nil
	}
	var t DevNonce
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return &t, nil
}

// MustDevNonce returns a typed DevNonce from the bytes.
// It returns nil, nil if b is nil.
// It panics if unmarshaling results in an error.
func MustDevNonce(b []byte) *DevNonce {
	t, err := GetDevNonce(b)
	if err != nil {
		panic(err)
	}
	return t
}

// OrZero returns the nonce value, or a zero value if the nonce was nil.
func (dn *DevNonce) OrZero() DevNonce {
	if dn != nil {
		return *dn
	}
	return DevNonce{}
}

// Equal returns true iff nonces are equal.
func (dn DevNonce) Equal(other DevNonce) bool { return dn == other }

// Size implements the Sizer interface.
func (dn DevNonce) Size() int { return 2 }

// Marshal implements the proto.Marshaler interface.
func (dn DevNonce) Marshal() ([]byte, error) { return dn.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (dn DevNonce) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, dn[:]) }

// Unmarshal implements the proto.Unmarshaler interface.
func (dn *DevNonce) Unmarshal(data []byte) error { return dn.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (dn DevNonce) MarshalJSON() ([]byte, error) { return marshalJSONHexBytes(dn[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (dn *DevNonce) UnmarshalJSON(data []byte) error {
	*dn = [2]byte{}
	if err := unmarshalJSONHexBytes(dn[:], data); err != nil {
		return errInvalidDevNonce.WithCause(err)
	}
	return nil
}

// MarshalProtoJSON implements the jsonplugin.Marshaler interface.
func (dn *DevNonce) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if dn == nil {
		s.WriteNil()
		return
	}
	s.WriteString(fmt.Sprintf("%X", dn[:]))
}

// UnmarshalProtoJSON implements the jsonplugin.Unmarshaler interface.
func (dn *DevNonce) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	*dn = [2]byte{}
	b, err := hex.DecodeString(s.ReadString())
	if err != nil {
		s.SetError(err)
		return
	}
	if len(b) != 2 {
		s.SetError(errInvalidDevAddr.WithCause(errInvalidLength.WithAttributes(
			"want", 2, "got", len(b), "got_type", "bytes",
		)))
		return
	}
	copy(dn[:], b)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (dn DevNonce) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(dn[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (dn *DevNonce) UnmarshalBinary(data []byte) error {
	*dn = [2]byte{}
	if err := unmarshalBinaryBytes(dn[:], data); err != nil {
		return errInvalidDevNonce.WithCause(err)
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (dn DevNonce) MarshalText() ([]byte, error) { return marshalTextBytes(dn[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (dn *DevNonce) UnmarshalText(data []byte) error {
	*dn = [2]byte{}
	if err := unmarshalTextBytes(dn[:], data); err != nil {
		return errInvalidDevNonce.WithCause(err)
	}
	return nil
}

// MarshalNumber returns the DevNonce in decimal form.
func (dn DevNonce) MarshalNumber() uint16 {
	return binary.BigEndian.Uint16(dn[:])
}

// UnmarshalNumber retrieves the DevNonce from decimal form.
func (dn *DevNonce) UnmarshalNumber(n uint16) {
	*dn = [2]byte{}
	binary.BigEndian.PutUint16(dn[:], n)
}
