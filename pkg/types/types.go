// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package types contains basic types that are used in The Things Network's protobuf messages.
package types

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"

	"bytes"
	"database/sql/driver"
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
	Value() (driver.Value, error)
	Scan(src interface{}) error
}

// ErrInvalidLength can be returned when unmarshaling a slice of invalid length.
var ErrInvalidLength = errors.New("invalid length")

// ErrTypeAssertion can be returned when trying to assert one variable.
var ErrTypeAssertion = errors.New("invalid type assertion")

// ErrInvalidJSONString is returned when an invalid JSON string is passed as an argument.
var ErrInvalidJSONString = &errors.ErrDescriptor{
	MessageFormat:  "Invalid JSON string: `{json_string}`",
	Code:           1,
	Type:           errors.InvalidArgument,
	SafeAttributes: []string{"json_string"},
}

func init() {
	ErrInvalidJSONString.Register()
}

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
		return ErrInvalidJSONString.New(errors.Attributes{
			"json_string": string(data),
		})
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
		return ErrInvalidLength
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
		return 0, ErrInvalidLength
	}
	return copy(dst, src), nil
}

func unmarshalBinaryBytes(dst, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) != len(dst) || copy(dst[:], data) != len(dst) {
		return ErrInvalidLength
	}
	return nil
}
