// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"strings"
)

// NewResultFunc represents a constructor of some arbitrary type.
type NewResultFunc func() interface{}

// Client represents a generic interface to interact with different store implementations in CRUD manner.
//
// Create creates a new PrimaryKey, stores v under that key and returns it.
// Find searches for the value associated with PrimaryKey specified and stores it in v. v must be a pointer type.
// FindBy returns mapping of PrimaryKey -> value, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned.
// newResultFunc is the constructor of a single value expected to be returned.
// Optional fields parameter is a list of fieldpaths separated by '.', which specifies the subset of fiters's fields, that should be used for searching. The fieldpaths can be grouped in most cases by specifying the parent identifier, i.e. to use fields "FOO.A", "FOO.B" and "FOO.C", it is possible to specify field path "FOO".
// Update overwrites stored fields under PrimaryKey with fields of v. Optional fields parameter is a list of fieldpaths separated by '.', which specifies the subset of v's fields, that should be updated. The fieldpaths can be grouped in most cases by specifying the parent identifier, i.e. to overwrite fields "FOO.A", "FOO.B" and "FOO.C", it is possible to specify field path "FOO".
// Delete deletes the value stored under PrimaryKey specified.
type Client interface {
	Create(v interface{}, fields ...string) (PrimaryKey, error)
	Find(id PrimaryKey, v interface{}) error
	FindBy(filter interface{}, newResult NewResultFunc, fields ...string) (map[PrimaryKey]interface{}, error)
	Update(id PrimaryKey, v interface{}, fields ...string) error
	Delete(id PrimaryKey) error
}

type typedStoreClient struct {
	TypedStore
}

// NewTypedStoreClient returns a new instance of the Client, which uses TypedStore as the storing backend.
func NewTypedStoreClient(s TypedStore) Client {
	return &typedStoreClient{s}
}

// filterFields returns a map, containing only values in m, which correspond to fieldpaths specified or m if none are specified. Note that filterFields may modify the input map m.
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

func (cl *typedStoreClient) Create(v interface{}, fields ...string) (PrimaryKey, error) {
	m, err := MarshalMap(v)
	if err != nil {
		return nil, err
	}
	return cl.TypedStore.Create(filterFields(m, fields...))
}

func (cl *typedStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.TypedStore.Find(id)
	if err != nil {
		return err
	}
	return UnmarshalMap(m, v)
}

func (cl *typedStoreClient) FindBy(filter interface{}, newResult NewResultFunc, fields ...string) (map[PrimaryKey]interface{}, error) {
	fm, err := MarshalMap(filter)
	if err != nil {
		return nil, err
	}
	m, err := cl.TypedStore.FindBy(filterFields(fm, fields...))
	if err != nil {
		return nil, err
	}

	filtered := make(map[PrimaryKey]interface{}, len(m))
	for k, v := range m {
		iface := newResult()
		if err = UnmarshalMap(v, iface); err != nil {
			return nil, err
		}
		filtered[k] = iface
	}
	return filtered, nil
}

func (cl *typedStoreClient) Update(id PrimaryKey, v interface{}, fields ...string) error {
	m, err := MarshalMap(v)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return cl.TypedStore.Update(id, m)
	}

	// Some values, i.e. empty slices/maps do not end up in the map returned by MarshalMap, so we assume that fields specified, but not in the map are those cases, hence, those fields should be cleared.
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
	return cl.TypedStore.Update(id, fm)
}

type byteStoreClient struct {
	ByteStore
}

// NewByteStoreClient returns a new instance of the Client, which uses ByteStore as the storing backend.
func NewByteStoreClient(s ByteStore) Client {
	return &byteStoreClient{s}
}

// filterByteFields returns a map, containing only values in m, which correspond to fieldpaths specified or m if none are specified. Note that filterByteFields may modify the input map m.
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

func (cl *byteStoreClient) Create(v interface{}, fields ...string) (PrimaryKey, error) {
	m, err := MarshalByteMap(v)
	if err != nil {
		return nil, err
	}
	return cl.ByteStore.Create(filterByteFields(m, fields...))
}

func (cl *byteStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.ByteStore.Find(id)
	if err != nil {
		return err
	}
	return UnmarshalByteMap(m, v)
}

func (cl *byteStoreClient) FindBy(filter interface{}, newResult NewResultFunc, fields ...string) (map[PrimaryKey]interface{}, error) {
	fm, err := MarshalByteMap(filter)
	if err != nil {
		return nil, err
	}

	m, err := cl.ByteStore.FindBy(filterByteFields(fm, fields...))
	if err != nil {
		return nil, err
	}
	filtered := make(map[PrimaryKey]interface{}, len(m))
	for k, v := range m {
		iface := newResult()
		if err = UnmarshalByteMap(v, iface); err != nil {
			return nil, err
		}
		filtered[k] = iface
	}
	return filtered, nil
}

func (cl *byteStoreClient) Update(id PrimaryKey, v interface{}, fields ...string) error {
	m, err := MarshalByteMap(v)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return cl.ByteStore.Update(id, m)
	}

	// Some values, i.e. empty slices/maps do not end up in the map returned by MarshalMap, so we assume that fields specified, but not in the map are those cases, hence, those fields should be cleared.
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
	return cl.ByteStore.Update(id, fm)
}
