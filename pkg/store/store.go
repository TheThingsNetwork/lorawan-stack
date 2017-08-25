// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"errors"

	"github.com/oklog/ulid"
)

var (
	ErrNotFound = errors.New("not found")
)

type Store interface {
	Create(obj map[string]interface{}) (ulid.ULID, error)
	Find(id ulid.ULID) (map[string]interface{}, error)
	FindBy(map[string]interface{}) (map[ulid.ULID]map[string]interface{}, error)
	Update(id ulid.ULID, new, old map[string]interface{}) error
	Delete(id ulid.ULID) error
}
