// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"io"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/oklog/ulid"
)

// New returns a new MapStore
func New() store.Store {
	return &mapStore{
		entropy: rand.New(rand.NewSource(time.Now().UnixNano())),
		data:    make(map[ulid.ULID]map[string]interface{}),
	}
}

type mapStore struct {
	entropy io.Reader
	mu      sync.RWMutex
	data    map[ulid.ULID]map[string]interface{}
}

func (s *mapStore) ulid() ulid.ULID {
	return ulid.MustNew(uint64(time.Now().UnixNano()/1000000), s.entropy)
}

func (s *mapStore) Create(obj map[string]interface{}) (ulid.ULID, error) {
	id := s.ulid()
	return id, s.Update(id, obj, nil)
}

func (s *mapStore) Find(id ulid.ULID) (map[string]interface{}, error) {
	s.mu.RLock()
	obj, ok := s.data[id]
	s.mu.RUnlock()
	if ok {
		return obj, nil
	}
	return nil, store.ErrNotFound
}

func (s *mapStore) FindBy(filters map[string]interface{}) (map[ulid.ULID]map[string]interface{}, error) {
	matches := make(map[ulid.ULID]map[string]interface{})
	s.mu.RLock()
	for id, obj := range s.data {
		for filterK, filterV := range filters {
			v, ok := obj[filterK]
			if !ok || !reflect.DeepEqual(v, filterV) {
				continue
			}
			matches[id] = obj
			break
		}
	}
	s.mu.RUnlock()
	return matches, nil
}

func (s *mapStore) Update(id ulid.ULID, new, old map[string]interface{}) error {
	diff := store.Diff(new, old)
	s.mu.Lock()
	defer s.mu.Unlock()
	obj, ok := s.data[id]
	if !ok {
		s.data[id] = diff
		return nil
	}
	for k, v := range diff {
		if v == nil {
			delete(obj, k)
			continue
		}
		obj[k] = v
	}
	s.data[id] = obj
	return nil
}

func (s *mapStore) Delete(id ulid.ULID) error {
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
	return nil
}
