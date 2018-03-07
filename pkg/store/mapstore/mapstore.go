// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"io"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/mohae/deepcopy"
	"github.com/oklog/ulid"
)

// New returns a new MapStore
func New() store.TypedStore {
	return &mapStore{
		entropy: rand.New(rand.NewSource(time.Now().UnixNano())),
		data:    make(map[store.PrimaryKey]map[string]interface{}),
	}
}

type mapStore struct {
	entropy io.Reader
	mu      sync.RWMutex
	data    map[store.PrimaryKey]map[string]interface{}
}

func (s *mapStore) newULID() store.PrimaryKey {
	return ulid.MustNew(ulid.Now(), s.entropy)
}

func (s *mapStore) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	id := s.newULID()
	if len(fields) == 0 {
		return id, nil
	}
	return id, s.Update(id, fields)
}

func (s *mapStore) Find(id store.PrimaryKey) (map[string]interface{}, error) {
	if id == nil {
		return nil, store.ErrNilKey
	}

	s.mu.RLock()
	fields := s.data[id]
	s.mu.RUnlock()
	return deepcopy.Copy(fields).(map[string]interface{}), nil
}

func (s *mapStore) FindBy(filter map[string]interface{}) (map[store.PrimaryKey]map[string]interface{}, error) {
	matches := make(map[store.PrimaryKey]map[string]interface{})
	s.mu.RLock()
outer:
	for id, fields := range s.data {
		for k, fv := range filter {
			dv, ok := fields[k]
			if !ok || !reflect.DeepEqual(dv, fv) {
				continue outer
			}
		}
		matches[id] = fields
	}
	s.mu.RUnlock()
	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

func (s *mapStore) Update(id store.PrimaryKey, diff map[string]interface{}) error {
	if id == nil {
		return store.ErrNilKey
	}
	if len(diff) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	fields, ok := s.data[id]
	if !ok {
		s.data[id] = diff
		return nil
	}
	for k, v := range diff {
		p := k + store.Separator
		for sk := range fields {
			if strings.HasPrefix(sk, p) {
				delete(fields, sk)
			}
		}

		if v == nil {
			delete(fields, k)
		} else {
			fields[k] = v
		}
	}
	s.data[id] = fields
	return nil
}

func (s *mapStore) Delete(id store.PrimaryKey) error {
	if id == nil {
		return store.ErrNilKey
	}

	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
	return nil
}
