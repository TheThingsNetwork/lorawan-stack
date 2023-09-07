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

var errInvalidAESKey = errors.DefineInvalidArgument("invalid_aes_key", "invalid AES key")

// AES128Key is an 128-bit AES key.
type AES128Key [16]byte

// IsZero returns true iff the type is zero.
func (key AES128Key) IsZero() bool { return key == [16]byte{} }

func (key AES128Key) String() string { return strings.ToUpper(hex.EncodeToString(key[:])) }

func (key AES128Key) GoString() string { return key.String() }

func (key AES128Key) Bytes() []byte {
	b := make([]byte, 16)
	copy(b, key[:])
	return b
}

// GetAES128Key gets a typed AES128Key from the bytes.
// It returns nil, nil if b is nil.
// It returns an error if unmarshaling fails.
func GetAES128Key(b []byte) (*AES128Key, error) {
	if b == nil {
		return nil, nil
	}
	var t AES128Key
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return &t, nil
}

// MustAES128Key returns a typed AES128Key from the bytes.
// It returns nil if b is nil.
// It panics if unmarshaling results in an error.
func MustAES128Key(b []byte) *AES128Key {
	t, err := GetAES128Key(b)
	if err != nil {
		panic(err)
	}
	return t
}

// OrZero returns the key value, or a zero value if the key was nil.
func (key *AES128Key) OrZero() AES128Key {
	if key != nil {
		return *key
	}
	return AES128Key{}
}

// Equal returns true iff keys are equal.
func (key AES128Key) Equal(other AES128Key) bool { return key == other }

// Size implements the Sizer interface.
func (key AES128Key) Size() int { return 16 }

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

// MarshalProtoJSON implements the jsonplugin.Marshaler interface.
func (key *AES128Key) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if key == nil {
		s.WriteNil()
		return
	}
	s.WriteString(fmt.Sprintf("%X", key[:]))
}

// UnmarshalProtoJSON implements the jsonplugin.Unmarshaler interface.
func (key *AES128Key) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	*key = [16]byte{}
	b, err := hex.DecodeString(s.ReadString())
	if err != nil {
		s.SetError(err)
		return
	}
	if len(b) != 16 {
		s.SetError(errInvalidDevAddr.WithCause(errInvalidLength.WithAttributes(
			"want", 16, "got", len(b), "got_type", "bytes",
		)))
		return
	}
	copy(key[:], b)
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
