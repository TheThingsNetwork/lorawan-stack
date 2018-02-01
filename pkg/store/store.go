// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"errors"
	"fmt"
)

// Encoding represents the encoding used to encode value into []byte representation.
// This is used as the first byte in the encoded []byte representation and allows for consistent decoding.
type Encoding byte

const (
	// Separator used to separate the flattened struct fields
	Separator = "."

	// NOTE: The following list MUST NOT be reordered

	// RawEncoding represents case when value is encoded into "raw" byte value.
	RawEncoding Encoding = 1
	// JSONEncoding represents case when MarshalJSON() method was used to encode value.
	JSONEncoding Encoding = 2
	// ProtoEncoding represents case when Proto() method was used to encode value.
	ProtoEncoding Encoding = 3
	// GobEncoding represents case when Gob was used to encode value.
	GobEncoding Encoding = 4
)

var (
	// ErrNotFound represents an error returned, when entity is not found.
	ErrNotFound = errors.New("Not found")

	// ErrInvalidData represents an error returned, when value stored is not valid.
	ErrInvalidData = errors.New("Invalid data")
)

// PrimaryKey represents the value used by store implementations to uniquely identify stored objects.
type PrimaryKey interface {
	fmt.Stringer
}

type Deleter interface {
	Delete(id PrimaryKey) error
}

type Trimmer interface {
	Trim(id PrimaryKey, n int) error
}

// TypedStore represents a store, modeled after CRUD, which stores typed data.
//
// Create creates a new PrimaryKey, stores fields under that key and returns it.
// Find returns the fields stored under PrimaryKey specified.
// FindBy returns mapping of PrimaryKey -> fields, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned.
// Update overwrites field values stored under PrimaryKey specified with values in diff.
// Delete deletes the fields stored under PrimaryKey specified.
type TypedStore interface {
	Create(fields map[string]interface{}) (PrimaryKey, error)
	Find(id PrimaryKey) (map[string]interface{}, error)
	FindBy(filter map[string]interface{}) (map[PrimaryKey]map[string]interface{}, error)
	Update(id PrimaryKey, diff map[string]interface{}) error
	Deleter
}

// ByteStore represents a store modeled after CRUD, which stores data as bytes.
//
// Create creates a new PrimaryKey, stores fields under that key and returns it.
// Find returns the fields stored under PrimaryKey specified.
// FindBy returns mapping of PrimaryKey -> fields, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned.
// Update overwrites field values stored under PrimaryKey specified with values in diff.
// Delete deletes the fields stored under PrimaryKey specified.
type ByteStore interface {
	Create(fields map[string][]byte) (PrimaryKey, error)
	Find(id PrimaryKey) (map[string][]byte, error)
	FindBy(filter map[string][]byte) (map[PrimaryKey]map[string][]byte, error)
	Update(id PrimaryKey, diff map[string][]byte) error
	Deleter
}

// ByteListStore represents a store, which stores lists of []byte values.
type ByteListStore interface {
	CreateList(bs ...[]byte) (PrimaryKey, error)
	FindList(id PrimaryKey) ([][]byte, error)
	Append(id PrimaryKey, bs ...[]byte) error
	Deleter
}

// ByteSetStore represents a store, which stores sets of []byte values.
type ByteSetStore interface {
	CreateSet(bs ...[]byte) (PrimaryKey, error)
	FindSet(id PrimaryKey) ([][]byte, error)
	Put(id PrimaryKey, bs ...[]byte) error
	Contains(id PrimaryKey, bs []byte) (bool, error)
	Remove(id PrimaryKey, bs ...[]byte) error
	Deleter
}
