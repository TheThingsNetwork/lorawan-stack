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

	"github.com/mohae/deepcopy"
	"github.com/oklog/ulid"
	"go.thethings.network/lorawan-stack/pkg/store"
)

var _ store.TypedMapStore = &MapStore{}

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
	defer s.mu.RUnlock()

	fields := s.data[id]
	return deepcopy.Copy(fields).(map[string]interface{}), nil
}

func (s *MapStore) Range(filter map[string]interface{}, _ uint64, f func(store.PrimaryKey, map[string]interface{}) bool) error {
	if len(filter) == 0 {
		return store.ErrEmptyFilter.New(nil)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

outer:
	for id, fields := range s.data {
		for k, fv := range filter {
			if !reflect.DeepEqual(fields[k], fv) {
				continue outer
			}
		}
		if !f(id, fields) {
			return nil
		}
	}
	return nil
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
	defer s.mu.Unlock()

	delete(s.data, id)
	return nil
}
