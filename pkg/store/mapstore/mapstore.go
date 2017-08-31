// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"io"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/mohae/deepcopy"
	"github.com/oklog/ulid"
)

// New returns a new MapStore
func New() store.Interface {
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

func (s *mapStore) ulid() store.PrimaryKey {
	return ulid.MustNew(ulid.Now(), s.entropy)
}

func (s *mapStore) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	id := s.ulid()
	return id, s.Update(id, fields)
}

func (s *mapStore) Find(id store.PrimaryKey) (map[string]interface{}, error) {
	s.mu.RLock()
	fields, ok := s.data[id]
	s.mu.RUnlock()
	if ok {
		return deepcopy.Copy(fields).(map[string]interface{}), nil
	}
	return nil, store.ErrNotFound
}

func (s *mapStore) FindBy(filters map[string]interface{}) (map[store.PrimaryKey]map[string]interface{}, error) {
	matches := make(map[store.PrimaryKey]map[string]interface{})
	s.mu.RLock()
outer:
	for id, fields := range s.data {
		for k, fv := range filters {
			dv, ok := fields[k]
			if !ok || !reflect.DeepEqual(dv, fv) {
				continue outer
			}
		}
		matches[id] = fields
	}
	s.mu.RUnlock()
	return matches, nil
}

func (s *mapStore) Update(id store.PrimaryKey, diff map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fields, ok := s.data[id]
	if !ok {
		s.data[id] = diff
		return nil
	}
	for k, v := range diff {
		if v == nil {
			delete(fields, k)
			continue
		}
		fields[k] = v
	}
	s.data[id] = fields
	return nil
}

func (s *mapStore) Delete(id store.PrimaryKey) error {
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
	return nil
}
