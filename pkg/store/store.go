// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound represents an error returned, when entity is not found.
	ErrNotFound = errors.New("not found")
)

// PrimaryKey represents the value used by store.Interface implementations to uniquely identify stored objects.
type PrimaryKey interface {
	fmt.Stringer
}

// Interface represents a store, modeled after CRUD.
//
// Create creates a new PrimaryKey, stores fields under that key and returns it.
// Find returns the fields stored under PrimaryKey specified.
// FindBy returns mapping of PrimaryKey -> fields, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned.
// Update overwrites field values stored under PrimaryKey specified with values in diff.
// Delete deletes the fields stored under PrimaryKey specified.
type Interface interface {
	Create(fields map[string]interface{}) (PrimaryKey, error)
	Find(id PrimaryKey) (map[string]interface{}, error)
	FindBy(filter map[string]interface{}) (map[PrimaryKey]map[string]interface{}, error)
	Update(id PrimaryKey, diff map[string]interface{}) error
	Delete(id PrimaryKey) error
}
