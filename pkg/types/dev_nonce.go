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
)

// DevNonce is device nonce used in the join procedure.
// - If LoRaWAN version <1.1 - it is randomly generated.
// - If LoRaWAN version >=1.1 - it is a scrictly increasing counter.
type DevNonce [2]byte

// IsZero returns true iff the type is zero.
func (dn DevNonce) IsZero() bool { return dn == [2]byte{} }

// String implements the Stringer interface.
func (dn DevNonce) String() string { return strings.ToUpper(hex.EncodeToString(dn[:])) }

// GoString implements the GoStringer interface.
func (dn DevNonce) GoString() string { return dn.String() }

// Size implements the Sizer interface.
func (dn DevNonce) Size() int { return 2 }

// Equal returns true iff nonces are equal.
func (dn DevNonce) Equal(other DevNonce) bool { return dn == other }

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
	return unmarshalJSONHexBytes(dn[:], data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (dn DevNonce) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(dn[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (dn *DevNonce) UnmarshalBinary(data []byte) error {
	*dn = [2]byte{}
	return unmarshalBinaryBytes(dn[:], data)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (dn DevNonce) MarshalText() ([]byte, error) { return marshalTextBytes(dn[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (dn *DevNonce) UnmarshalText(data []byte) error {
	*dn = [2]byte{}
	return unmarshalTextBytes(dn[:], data)
}
