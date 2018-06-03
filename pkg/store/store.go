// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package store defines generic storage interfaces and provides deterministic data encoding and decoding mechanisms, as well as utilities associated with this process.
package store

import (
	"fmt"
)

const (
	// SeparatorByte is character used to separate the flattened struct fields.
	SeparatorByte = '.'

	// Separator is SeparatorByte converted to a string.
	Separator = string(SeparatorByte)
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

// TypedMapStore represents a store, which stores typed data.
//
// Create creates a new PrimaryKey, stores fields under that key and returns it.
//
// Find returns the fields stored under PrimaryKey specified.
// It returns a nil map, if key is not found.
//
// Range calls f sequentially for each key and value present in the store.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the TypedMapStore's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.

// If batchSize argument is non-zero, Range will retrieve elements
// from the underlying store in chunks of (approximately) batchSize elements.
//
// Update overwrites field values stored under PrimaryKey specified with values in diff.
type TypedMapStore interface {
	Create(fields map[string]interface{}) (PrimaryKey, error)
	Find(id PrimaryKey) (map[string]interface{}, error)
	Range(filter map[string]interface{}, batchSize uint64, f func(PrimaryKey, map[string]interface{}) bool) error
	Update(id PrimaryKey, diff map[string]interface{}) error
	Deleter
}

// TypedMapStore represents a store, which stores data as []byte.
//
// Create creates a new PrimaryKey, stores fields under that key and returns it.
//
// Find returns the fields stored under PrimaryKey specified.
// It returns a nil map, if key is not found.
//
// Range calls f sequentially for each key and value present in the store.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the ByteMapStore's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.

// If batchSize argument is non-zero, Range will retrieve elements
// from the underlying store in chunks of (approximately) batchSize elements.
//
// Update overwrites field values stored under PrimaryKey specified with values in diff.
type ByteMapStore interface {
	Create(fields map[string][]byte) (PrimaryKey, error)
	Find(id PrimaryKey) (map[string][]byte, error)
	Range(filter map[string][]byte, batchSize uint64, f func(PrimaryKey, map[string][]byte) bool) error
	Update(id PrimaryKey, diff map[string][]byte) error
	Deleter
}

// ByteListStore represents a store, which stores lists of []byte values.
//
// CreateList creates a new list, containing bs.
//
// FindList returns list identified by id.
//
// Append appends bs to list identified by id.
//
// Pop returns the value stored at last index of list identified by id and removes it from the list.
//
// Len returns the length of the list identified by id.
type ByteListStore interface {
	CreateList(bs ...[]byte) (PrimaryKey, error)
	FindList(id PrimaryKey) ([][]byte, error)
	Append(id PrimaryKey, bs ...[]byte) error
	Pop(id PrimaryKey) ([]byte, error)
	Len(id PrimaryKey) (int64, error)
	Deleter
}

// ByteSetStore represents a store, which stores sets of []byte values.
//
// CreateSet creates a new set, containing bs.
//
// FindSet returns set identified by id.
//
// Put adds bs to set identified by id.
//
// Contains reports whether b is contained in set identified by id.
//
// Remove removes bs from set identified by id.
type ByteSetStore interface {
	CreateSet(bs ...[]byte) (PrimaryKey, error)
	FindSet(id PrimaryKey) ([][]byte, error)
	Put(id PrimaryKey, bs ...[]byte) error
	Contains(id PrimaryKey, bs []byte) (bool, error)
	Remove(id PrimaryKey, bs ...[]byte) error
	Deleter
}
