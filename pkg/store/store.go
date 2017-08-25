// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
)

type Store interface {
	Create(obj map[string]interface{}) (string, error)
	Find(id string) (map[string]interface{}, error)
	FindBy(map[string]interface{}) (map[string]map[string]interface{}, error)
	Update(id string, new, old map[string]interface{}) error
	Delete(id string) error
}
