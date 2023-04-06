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

// Package types contains basic types that are used in The Things Network's protobuf messages.
package types

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// Interface all types in pkg/types must implement.
type Interface interface {
	IsZero() bool
	String() string
	GoString() string
	Bytes() []byte
	Size() int
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (int, error)
	Unmarshal(data []byte) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	MarshalText() ([]byte, error)
	UnmarshalText(data []byte) error
}

var (
	errInvalidJSON   = errors.DefineInvalidArgument("invalid_json", "invalid JSON: `{json}`")
	errInvalidLength = errors.DefineInvalidArgument(
		"invalid_length", "invalid length: want {want} bytes, got {got} {got_type}",
	)
)

func marshalJSONHexBytes(data []byte) ([]byte, error) {
	hexData, err := marshalTextBytes(data)
	if err != nil {
		return nil, err
	}
	b := []byte{'"'}
	b = append(b, hexData...)
	return append(b, '"'), nil
}

func unmarshalJSONHexBytes(dst, data []byte) error {
	if string(data) == `""` {
		return nil
	}

	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errInvalidJSON.WithAttributes("json", string(data))
	}
	return unmarshalTextBytes(dst, data[1:len(data)-1])
}

func marshalTextBytes(data []byte) ([]byte, error) {
	b := make([]byte, hex.EncodedLen(len(data)))
	hex.Encode(b, data)
	return bytes.ToUpper(b), nil
}

func unmarshalTextBytes(dst, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	b := make([]byte, hex.DecodedLen(len(data)))
	n, err := hex.Decode(b, data)
	if err != nil {
		return err
	}
	if n != len(dst) || copy(dst, b) != len(dst) {
		return errInvalidLength.WithAttributes("want", len(dst), "got", n, "got_type", "bytes")
	}
	return nil
}

func marshalBinaryBytes(data []byte) ([]byte, error) {
	b := make([]byte, len(data))
	copy(b, data)
	return b, nil
}

func marshalBinaryBytesTo(dst, src []byte) (int, error) {
	if len(dst) < len(src) {
		return 0, errInvalidLength.WithAttributes("want", len(dst), "got", len(src), "got_type", "bytes")
	}
	return copy(dst, src), nil
}

func unmarshalBinaryBytes(dst, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) != len(dst) || copy(dst[:], data) != len(dst) {
		return errInvalidLength.WithAttributes("want", len(dst), "got", len(data), "got_type", "bytes")
	}
	return nil
}

// MarshalHEXBytes marshals bytes to JSON as HEX.
func MarshalHEXBytes(s *jsonplugin.MarshalState, b []byte) {
	s.WriteString(strings.ToUpper(hex.EncodeToString(b)))
}

var base64Replacer = strings.NewReplacer("_", "/", "-", "+")

// unmarshalNBytes unmarshals N bytes from JSON. For n > 1, it accepts both hex and base64 encoding.
func unmarshalNBytes(s *jsonplugin.UnmarshalState, n int) []byte {
	enc := s.ReadString()
	if s.Err() != nil {
		return nil
	}
	trimmed := strings.TrimRight(enc, "=")

	switch len(trimmed) {
	case 0:
		b := make([]byte, n)
		return b
	case hex.EncodedLen(n):
		b, err := hex.DecodeString(trimmed)
		if err != nil {
			s.SetError(err)
			return nil
		}
		return b
	case base64.RawStdEncoding.EncodedLen(n):
		b, err := base64.RawStdEncoding.DecodeString(base64Replacer.Replace(trimmed))
		if err != nil {
			s.SetError(err)
			return nil
		}
		return b
	default:
		s.SetError(errInvalidLength.WithAttributes(
			"want", n, "got", len(enc), "got_type", "runes",
		))
		return nil
	}
}

// Unmarshal2Bytes unmarshals 2 bytes from JSON. It accepts both hex and base64 encoding.
func Unmarshal2Bytes(s *jsonplugin.UnmarshalState) []byte { return unmarshalNBytes(s, 2) }

// Unmarshal3Bytes unmarshals 3 bytes from JSON. It accepts both hex and base64 encoding.
func Unmarshal3Bytes(s *jsonplugin.UnmarshalState) []byte { return unmarshalNBytes(s, 3) }

// Unmarshal4Bytes unmarshals 4 bytes from JSON. It accepts both hex and base64 encoding.
func Unmarshal4Bytes(s *jsonplugin.UnmarshalState) []byte { return unmarshalNBytes(s, 4) }

// Unmarshal8Bytes unmarshals 8 bytes from JSON. It accepts both hex and base64 encoding.
func Unmarshal8Bytes(s *jsonplugin.UnmarshalState) []byte { return unmarshalNBytes(s, 8) }

// Unmarshal16Bytes unmarshals 16 bytes from JSON. It accepts both hex and base64 encoding.
func Unmarshal16Bytes(s *jsonplugin.UnmarshalState) []byte { return unmarshalNBytes(s, 16) }
