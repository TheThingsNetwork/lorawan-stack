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

	"github.com/TheThingsIndustries/protoc-gen-go-flags/flagsplugin"
	"github.com/spf13/pflag"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/customflags"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errInvalidAESKey = errors.DefineInvalidArgument("invalid_aes_key", "invalid AES key")

// AES128Key is an 128-bit AES key.
type AES128Key [16]byte

// IsZero returns true iff the type is zero.
func (key *AES128Key) IsZero() bool { return key == nil || *key == [16]byte{} }

// String implements the Stringer interface.
func (key AES128Key) String() string { return strings.ToUpper(hex.EncodeToString(key[:])) }

// GoString implements the GoStringer interface.
func (key AES128Key) GoString() string { return key.String() }

// Size implements the Sizer interface.
func (key AES128Key) Size() int { return 16 }

// Equal returns true iff keys are equal.
func (key AES128Key) Equal(other AES128Key) bool { return key == other }

// Marshal implements the proto.Marshaler interface.
func (key AES128Key) Marshal() ([]byte, error) { return key.MarshalBinary() }

// MarshalTo implements the MarshalerTo function required by generated protobuf.
func (key AES128Key) MarshalTo(data []byte) (int, error) { return marshalBinaryBytesTo(data, key[:]) }

// Unmarshal implements the proto.Unmarshaler interface.
func (key *AES128Key) Unmarshal(data []byte) error { return key.UnmarshalBinary(data) }

// MarshalJSON implements the json.Marshaler interface.
func (key AES128Key) MarshalJSON() ([]byte, error) { return marshalJSONHexBytes(key[:]) }

// UnmarshalJSON implements the json.Unmarshaler interface.
func (key *AES128Key) UnmarshalJSON(data []byte) error {
	*key = [16]byte{}
	if err := unmarshalJSONHexBytes(key[:], data); err != nil {
		return errInvalidAESKey.WithCause(err)
	}
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (key AES128Key) MarshalBinary() ([]byte, error) { return marshalBinaryBytes(key[:]) }

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (key *AES128Key) UnmarshalBinary(data []byte) error {
	*key = [16]byte{}
	if err := unmarshalBinaryBytes(key[:], data); err != nil {
		return errInvalidAESKey.WithCause(err)
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (key AES128Key) MarshalText() ([]byte, error) { return marshalTextBytes(key[:]) }

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (key *AES128Key) UnmarshalText(data []byte) error {
	*key = [16]byte{}
	if err := unmarshalTextBytes(key[:], data); err != nil {
		return errInvalidAESKey.WithCause(err)
	}
	return nil
}

// EncodeMsgpack implements msgpack.CustomEncoder interface.
func (key AES128Key) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeString(hex.EncodeToString(key[:]))
}

// DecodeMsgpack implements msgpack.CustomDecoder interface.
func (key *AES128Key) DecodeMsgpack(dec *msgpack.Decoder) error {
	s, err := dec.DecodeString()
	if err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(b) != 16 {
		return errInvalidLength.WithAttributes("want", 16, "got", len(b))
	}
	copy(key[:], b[:])
	return nil
}

// GetAES128KeyFromFlag gets an AES128Key from a named flag in the flag set.
func GetAES128KeyFromFlag(fs *pflag.FlagSet, name string) (value AES128Key, set bool, err error) {
	flag := fs.Lookup(name)
	if flag == nil {
		return AES128Key{}, false, &flagsplugin.ErrFlagNotFound{FlagName: name}
	}
	var aes AES128Key
	if !flag.Changed {
		return aes, flag.Changed, nil
	}
	if err := aes.Unmarshal(flag.Value.(*customflags.ExactBytesValue).Value); err != nil {
		return aes, false, err
	}
	return aes, flag.Changed, nil
}
