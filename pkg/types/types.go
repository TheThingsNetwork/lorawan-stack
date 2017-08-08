// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package types contains basic types that are used in The Things Network's protobuf messages.
package types

import (
	"errors"
)

// Interface all types in pkg/types must implement
type Interface interface {
	IsZero() bool
	String() string
	GoString() string
	Size() int
	MarshalTo(data []byte) (int, error)
	Unmarshal(data []byte) error
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	MarshalText() ([]byte, error)
	UnmarshalText(data []byte) error
}

// ErrInvalidLength can be returned when unmarshaling a slice of invalid length
var ErrInvalidLength = errors.New("invalid length")
