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

package store

import (
	"strings"

	"go.thethings.network/lorawan-stack/pkg/marshaling"
)

// NewResultFunc represents a constructor of some arbitrary type.
type NewResultFunc func() interface{}

// Client represents a generic interface to interact with different store implementations in CRUD manner.
//
// Create creates a new PrimaryKey, stores v under that key and returns it.
//
// Find searches for the value associated with PrimaryKey specified and stores it in v. v must be a pointer type.
//
// Range calls f sequentially for each key and value present in the store.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the underlying store's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.

// If batchSize argument is non-zero, Range will retrieve elements
// from the underlying store in chunks of (approximately) batchSize elements.
//
// Filter represents an AND relation, meaning that only entries matching all the fields in filter should be returned.
// newResultFunc is the constructor of a single value expected to be returned.
// Optional fields parameter is a list of fieldpaths separated by '.',
// which specifies the subset of fiters's fields, that should be used for searching.
// The fieldpaths can be grouped in most cases by specifying the parent identifier, i.e. to use fields "FOO.A", "FOO.B" and "FOO.C", it is possible to specify field path "FOO".
//
// Update overwrites stored fields under PrimaryKey with fields of v.
// Optional fields parameter is a list of fieldpaths separated by '.',
// which specifies the subset of v's fields, that should be updated.
// The fieldpaths can be grouped in most cases by specifying the parent identifier,
// i.e. to overwrite fields "FOO.A", "FOO.B" and "FOO.C", it is possible to specify field path "FOO".
//
// Delete deletes the value stored under PrimaryKey specified.
type Client interface {
	Create(v interface{}, fields ...string) (PrimaryKey, error)
	Find(id PrimaryKey, v interface{}) error
	Range(filter interface{}, newResult NewResultFunc, orderBy string, count, offset uint64, f func(PrimaryKey, interface{}) bool, fields ...string) (uint64, error)
	Update(id PrimaryKey, v interface{}, fields ...string) error
	Delete(id PrimaryKey) error
}

type typedMapStoreClient struct {
	TypedMapStore
}

// NewTypedMapStoreClient returns a new instance of the Client, which uses TypedMapStore as the storing backend.
func NewTypedMapStoreClient(s TypedMapStore) Client {
	return &typedMapStoreClient{s}
}

// filterFields returns a map, containing only values in m, which correspond to fieldpaths specified
// or m, if none are specified. Note that filterFields may modify the input map m.
func filterFields(m map[string]interface{}, fields ...string) map[string]interface{} {
	if len(fields) == 0 {
		return m
	}

	out := make(map[string]interface{}, len(m))
	for _, f := range fields {
		p := f + "."
		for k, v := range m {
			if k == f || strings.HasPrefix(k, p) {
				out[k] = v
				delete(m, k)
			}
		}
	}
	return out
}

func (cl *typedMapStoreClient) Create(v interface{}, fields ...string) (PrimaryKey, error) {
	m, err := marshaling.MarshalMap(v)
	if err != nil {
		return nil, err
	}
	return cl.TypedMapStore.Create(filterFields(m, fields...))
}

func (cl *typedMapStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.TypedMapStore.Find(id)
	if err != nil {
		return err
	}
	return marshaling.UnmarshalMap(m, v)
}

func (cl *typedMapStoreClient) Range(filter interface{}, newResult NewResultFunc, orderBy string, count, offset uint64, f func(PrimaryKey, interface{}) bool, fields ...string) (uint64, error) {
	fm, err := marshaling.MarshalMap(filter)
	if err != nil {
		return 0, err
	}

	var ierr error
	total, err := cl.TypedMapStore.Range(filterFields(fm, fields...), orderBy, count, offset, func(k PrimaryKey, v map[string]interface{}) bool {
		iface := newResult()
		if ierr = marshaling.UnmarshalMap(v, iface); ierr != nil {
			return false
		}
		return f(k, iface)
	})

	if err != nil {
		return 0, err
	}
	return total, ierr
}

func (cl *typedMapStoreClient) Update(id PrimaryKey, v interface{}, fields ...string) error {
	m, err := marshaling.MarshalMap(v)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return cl.TypedMapStore.Update(id, m)
	}

	// Some values, i.e. empty slices/maps do not end up in the map returned by MarshalMap,
	// so we assume that fields specified, but not in the map are those cases, hence, those fields should be cleared.
	toDel := make(map[string]struct{}, len(fields))
	for _, k := range fields {
		toDel[k] = struct{}{}
	}

	fm := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		p := f + "."
		for k, mv := range m {
			if k == f || strings.HasPrefix(k, p) {
				fm[k] = mv
				delete(m, k)
				delete(toDel, f)
			}
		}
	}
	for k := range toDel {
		fm[k] = nil
	}
	return cl.TypedMapStore.Update(id, fm)
}

type byteMapStoreClient struct {
	ByteMapStore
}

// NewByteMapStoreClient returns a new instance of the Client, which uses ByteMapStore as the storing backend.
func NewByteMapStoreClient(s ByteMapStore) Client {
	return &byteMapStoreClient{s}
}

// filterByteFields returns a map, containing only values in m, which correspond to fieldpaths specified
// or m if none are specified. Note that filterByteFields may modify the input map m.
func filterByteFields(m map[string][]byte, fields ...string) map[string][]byte {
	if len(fields) == 0 {
		return m
	}

	out := make(map[string][]byte, len(m))
	for _, f := range fields {
		p := f + "."
		for k, v := range m {
			if k == f || strings.HasPrefix(k, p) {
				out[k] = v
				delete(m, k)
			}
		}
	}
	return out
}

func (cl *byteMapStoreClient) Create(v interface{}, fields ...string) (PrimaryKey, error) {
	m, err := marshaling.MarshalByteMap(v)
	if err != nil {
		return nil, err
	}
	return cl.ByteMapStore.Create(filterByteFields(m, fields...))
}

func (cl *byteMapStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.ByteMapStore.Find(id)
	if err != nil {
		return err
	}
	return marshaling.UnmarshalByteMap(m, v)
}

func (cl *byteMapStoreClient) Range(filter interface{}, newResult NewResultFunc, orderBy string, count, offset uint64, f func(PrimaryKey, interface{}) bool, fields ...string) (uint64, error) {
	fm, err := marshaling.MarshalByteMap(filter)
	if err != nil {
		return 0, err
	}

	var ierr error
	total, err := cl.ByteMapStore.Range(filterByteFields(fm, fields...), orderBy, count, offset, func(k PrimaryKey, v map[string][]byte) bool {
		iface := newResult()
		if ierr = marshaling.UnmarshalByteMap(v, iface); ierr != nil {
			return false
		}
		return f(k, iface)
	})

	if err != nil {
		return 0, err
	}
	return total, ierr
}

func (cl *byteMapStoreClient) Update(id PrimaryKey, v interface{}, fields ...string) error {
	m, err := marshaling.MarshalByteMap(v)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return cl.ByteMapStore.Update(id, m)
	}

	// Some values, i.e. empty slices/maps do not end up in the map returned by MarshalByteMap,
	// so we assume that fields specified, but not in the map are those cases, hence, those fields should be cleared.
	toDel := make(map[string]struct{}, len(fields))
	for _, k := range fields {
		toDel[k] = struct{}{}
	}

	fm := make(map[string][]byte, len(fields))
	for _, f := range fields {
		p := f + "."
		for k, mv := range m {
			if k == f || strings.HasPrefix(k, p) {
				fm[k] = mv
				delete(m, k)
				delete(toDel, f)
			}
		}
	}
	for k := range toDel {
		fm[k] = nil
	}
	return cl.ByteMapStore.Update(id, fm)
}
