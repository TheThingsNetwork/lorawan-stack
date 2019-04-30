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

// Package types contains basic types that are used in The Things Network's protobuf messages.
package types

import (
	"go.thethings.network/lorawan-stack/pkg/errors"

	"bytes"
	"encoding/hex"
)

// Interface all types in pkg/types must implement.
type Interface interface {
	IsZero() bool
	String() string
	GoString() string
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
	errScanArgumentType = errors.DefineInternal("src_type", "invalid type for src") // DB schema problem.
	errInvalidJSON      = errors.DefineInvalidArgument("invalid_json", "invalid JSON: `{json}`")
	errInvalidLength    = errors.DefineInvalidArgument("invalid_length", "invalid slice length")
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
		return errInvalidLength
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
		return 0, errInvalidLength
	}
	return copy(dst, src), nil
}

func unmarshalBinaryBytes(dst, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) != len(dst) || copy(dst[:], data) != len(dst) {
		return errInvalidLength
	}
	return nil
}
