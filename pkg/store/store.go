// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
)

type PrimaryKey interface {
	fmt.Stringer
}

type Store interface {
	Create(obj map[string]interface{}) (PrimaryKey, error)
	Find(id PrimaryKey) (map[string]interface{}, error)
	FindBy(map[string]interface{}) (map[PrimaryKey]map[string]interface{}, error)
	Update(id PrimaryKey, new, old map[string]interface{}) error
	Delete(id PrimaryKey) error
}
