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

// Package mapstore provides an example implementation of store.TypedMapStore, which is used for testing.
package mapstore

import (
	"crypto/rand"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/mohae/deepcopy"
	"github.com/oklog/ulid"
)

type MapStore struct {
	mu      sync.RWMutex
	data    map[store.PrimaryKey]map[string]interface{}
	entropy io.Reader
}

// New returns a new MapStore.
func New() *MapStore {
	return &MapStore{
		data:    make(map[store.PrimaryKey]map[string]interface{}),
		entropy: rand.Reader,
	}
}

func (s *MapStore) newULID() store.PrimaryKey {
	return ulid.MustNew(ulid.Now(), s.entropy)
}

func (s *MapStore) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	id := s.newULID()
	if len(fields) == 0 {
		return id, nil
	}
	return id, s.Update(id, fields)
}

func (s *MapStore) Find(id store.PrimaryKey) (map[string]interface{}, error) {
	if id == nil {
		return nil, store.ErrNilKey.New(nil)
	}

	s.mu.RLock()
	fields := s.data[id]
	s.mu.RUnlock()
	return deepcopy.Copy(fields).(map[string]interface{}), nil
}

func (s *MapStore) FindBy(filter map[string]interface{}) (map[store.PrimaryKey]map[string]interface{}, error) {
	if len(filter) == 0 {
		return nil, store.ErrEmptyFilter.New(nil)
	}

	matches := make(map[store.PrimaryKey]map[string]interface{})
	s.mu.RLock()
outer:
	for id, fields := range s.data {
		for k, fv := range filter {
			if !reflect.DeepEqual(fields[k], fv) {
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

func (s *MapStore) Update(id store.PrimaryKey, diff map[string]interface{}) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
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
		for sk := range fields {
			if strings.HasPrefix(sk, k+store.Separator) {
				delete(fields, sk)
			}
		}

		if i := strings.LastIndexByte(k, store.SeparatorByte); i != -1 {
			delete(fields, k[:i])
		}

		fields[k] = v
	}
	s.data[id] = fields
	return nil
}

func (s *MapStore) Delete(id store.PrimaryKey) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}

	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
	return nil
}
