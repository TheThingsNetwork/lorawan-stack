// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"fmt"
	"strings"

	"github.com/kr/pretty"
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
// Update overwrites stored fields under PrimaryKey with fields of v. Optional fields parameter is a list of fieldpaths separated by a '.', which specifies the subset of v's fields, that should be updated.
// Delete deletes the value stored under PrimaryKey specified.
type Client interface {
	Create(v interface{}) (PrimaryKey, error)
	Find(id PrimaryKey, v interface{}) error
	FindBy(filter interface{}, newResult NewResultFunc) (map[PrimaryKey]interface{}, error)
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

func (cl *typedStoreClient) Create(v interface{}) (PrimaryKey, error) {
	return cl.TypedStore.Create(MarshalMap(v))
}

func (cl *typedStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.TypedStore.Find(id)
	if err != nil {
		return err
	}
	defer func() {
		if err := recover(); err != nil {
			pretty.Println(m)
			fmt.Println(err)
		}
	}()
	return UnmarshalMap(m, v)
}

func (cl *typedStoreClient) FindBy(filter interface{}, newResult NewResultFunc) (map[PrimaryKey]interface{}, error) {
	m, err := cl.TypedStore.FindBy(MarshalMap(filter))
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
	m := MarshalMap(v)
	pretty.Println(m)
	if len(fields) == 0 {
		return cl.TypedStore.Update(id, m)
	}

	fm := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		for k, mv := range m {
			if strings.HasPrefix(k, f) {
				fm[k] = mv
				delete(m, k)
			}
		}
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

func (cl *byteStoreClient) Create(v interface{}) (PrimaryKey, error) {
	m, err := MarshalByteMap(v)
	if err != nil {
		return nil, err
	}
	return cl.ByteStore.Create(m)
}

func (cl *byteStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.ByteStore.Find(id)
	if err != nil {
		return err
	}
	return UnmarshalByteMap(m, v)
}

func (cl *byteStoreClient) FindBy(filter interface{}, newResult NewResultFunc) (map[PrimaryKey]interface{}, error) {
	fm, err := MarshalByteMap(filter)
	if err != nil {
		return nil, err
	}

	m, err := cl.ByteStore.FindBy(fm)
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

	fm := make(map[string][]byte, len(fields))
	for _, f := range fields {
		for k, mv := range m {
			if strings.HasPrefix(k, f) {
				fm[k] = mv
				delete(m, k)
			}
		}
	}
	return cl.ByteStore.Update(id, fm)
}
