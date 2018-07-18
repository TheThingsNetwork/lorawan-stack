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
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/oklog/ulid"
	"go.thethings.network/lorawan-stack/pkg/store"
)

var _ store.TypedMapStore = &MapStore{}

// MapStore is a store.TypedMapStore implementation to use for testing.
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

// Create implements store.TypedMapStore.
func (s *MapStore) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	id := s.newULID()
	if len(fields) == 0 {
		return id, nil
	}
	return id, s.Update(id, fields)
}

// Find implements store.TypedMapStore.
func (s *MapStore) Find(id store.PrimaryKey) (map[string]interface{}, error) {
	if id == nil {
		return nil, store.ErrNilKey.New(nil)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	fields := s.data[id]
	return deepcopy.Copy(fields).(map[string]interface{}), nil
}

// Range implements store.TypedMapStore.
func (s *MapStore) Range(filter map[string]interface{}, orderBy string, count, offset uint64, f func(store.PrimaryKey, map[string]interface{}) bool) (uint64, error) {
	if len(filter) == 0 {
		return 0, store.ErrEmptyFilter.New(nil)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	matches := make(map[store.PrimaryKey]map[string]interface{}, count)
outer:
	for id, fields := range s.data {
		for k, fv := range filter {
			if !reflect.DeepEqual(fields[k], fv) {
				continue outer
			}
		}
		matches[id] = fields
	}

	type value struct {
		id     store.PrimaryKey
		fields map[string]interface{}
	}

	sl := make([]value, 0, len(matches))
	for k, v := range matches {
		sl = append(sl, value{
			id:     k,
			fields: v,
		})
	}

	if orderBy != "" {
		sort.Slice(sl, func(i, j int) bool {
			is, ok := sl[i].fields[orderBy].(string)
			if !ok {
				panic(fmt.Errorf("mapstore only supports sorting by string values, %s is %T", orderBy, sl[i].fields[orderBy]))
			}
			js, ok := sl[j].fields[orderBy].(string)
			if !ok {
				panic(fmt.Errorf("mapstore only supports sorting by string values, %s is %T", orderBy, sl[j].fields[orderBy]))
			}
			return is < js
		})
	}

	for i, v := range sl {
		switch {
		case uint64(i) < offset:

		case count > 0 && uint64(i) >= offset+count,
			!f(v.id, deepcopy.Copy(v.fields).(map[string]interface{})):
			return uint64(len(sl)), nil
		}
	}
	return uint64(len(sl)), nil
}

// Update implements store.TypedMapStore.
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

// Delete implements store.TypedMapStore.
func (s *MapStore) Delete(id store.PrimaryKey) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, id)
	return nil
}
