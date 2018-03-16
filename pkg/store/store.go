// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package store defines generic storage interfaces and provides deterministic data encoding and decoding mechanisms, as well as utilities associated with this process.
package store

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

var protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()

// Encoding represents the encoding used to encode value into []byte representation.
// This is used as the first byte in the encoded []byte representation and allows for consistent decoding.
type Encoding byte

const (
	// SeparatorByte is character used to separate the flattened struct fields.
	SeparatorByte = '.'

	// Separator is SeparatorByte converted to a string.
	Separator = string(SeparatorByte)

	// NOTE: The following list MUST NOT be reordered

	// RawEncoding represents case when value is encoded into "raw" byte value.
	RawEncoding Encoding = 1
	// JSONEncoding represents case when MarshalJSON() method was used to encode value.
	JSONEncoding Encoding = 2
	// ProtoEncoding represents case when Proto() method was used to encode value.
	ProtoEncoding Encoding = 3
	// GobEncoding represents case when Gob was used to encode value.
	GobEncoding Encoding = 4
	// MsgPackEncoding represents case when MsgPack was used to encode value.
	MsgPackEncoding Encoding = 5
)

// PrimaryKey represents the value used by store implementations to uniquely identify stored objects.
type PrimaryKey interface {
	fmt.Stringer
}

// Deleter is an interface, which allows deleting of values stored under specified PrimaryKey.
type Deleter interface {
	Delete(id PrimaryKey) error
}

// Trimmer is an interface, which allows trimming size of
// the data structure stored under PrimaryKey id to a size of n elements.
type Trimmer interface {
	Trim(id PrimaryKey, n int) error
}

// TypedStore represents a store, modeled after CRUD, which stores typed data.
//
// Create creates a new PrimaryKey, stores fields under that key and returns it.
// Find returns the fields stored under PrimaryKey specified. It returns a nil map, if key is not found.
// FindBy returns mapping of PrimaryKey -> fields, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned. It returns a nil map, if no value matching filter is found.
// Update overwrites field values stored under PrimaryKey specified with values in diff.
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
// Find returns the fields stored under PrimaryKey specified. It returns a nil map, if key is not found.
// FindBy returns mapping of PrimaryKey -> fields, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned. It returns a nil map, if key is not found.
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
// CreateList creates a new list, containing bs.
// FindList returns list identified by id.
// Append appends bs to list identified by id.
type ByteListStore interface {
	CreateList(bs ...[]byte) (PrimaryKey, error)
	FindList(id PrimaryKey) ([][]byte, error)
	Append(id PrimaryKey, bs ...[]byte) error
	Deleter
}

// ByteSetStore represents a store, which stores sets of []byte values.
// CreateSet creates a new set, containing bs.
// FindSet returns set identified by id.
// Put adds bs to set identified by id.
// Contains reports whether b is contained in set identified by id.
// Remove removes bs from set identified by id.
type ByteSetStore interface {
	CreateSet(bs ...[]byte) (PrimaryKey, error)
	FindSet(id PrimaryKey) ([][]byte, error)
	Put(id PrimaryKey, bs ...[]byte) error
	Contains(id PrimaryKey, bs []byte) (bool, error)
	Remove(id PrimaryKey, bs ...[]byte) error
	Deleter
}
