// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package types contains basic types that are used in The Things Network's protobuf messages.
package types

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

// Interface all types in pkg/types must implement
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

// ErrInvalidLength can be returned when unmarshaling a slice of invalid length
var ErrInvalidLength = errors.New("invalid length")

var base64Encoding = base64.StdEncoding

func marshalJSONBytes(data []byte) ([]byte, error) {
	n := 2 + base64.StdEncoding.EncodedLen(len(data))
	b := make([]byte, n-1, n)
	b[0] = '"'
	base64.StdEncoding.Encode(b[1:], data)
	return append(b, '"'), nil
}

func unmarshalJSONBytes(dst, data []byte) error {
	if string(data) == `""` {
		return nil
	}

	b := make([]byte, base64.StdEncoding.DecodedLen(len(data)-2))
	n, err := base64Encoding.Decode(b, data[1:len(data)-1])
	if err != nil {
		return err
	}
	if n != len(dst) || copy(dst, b) != len(dst) {
		return ErrInvalidLength
	}
	return nil
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
