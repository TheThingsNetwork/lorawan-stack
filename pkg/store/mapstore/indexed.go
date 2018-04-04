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

package mapstore

import (
	"fmt"
	"sort"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/store"
)

// NewIndexed returns a new MapStore that keeps indexes for the given fields
func NewIndexed(indexed ...string) store.TypedStore {
	internal := New().(*mapStore)
	s := &indexedStore{
		mapStore: internal,
		indexes:  make(map[string]map[string]store.KeySet),
	}
	for _, field := range indexed {
		s.indexes[field] = make(map[string]store.KeySet)
	}
	return s
}

type indexedStore struct {
	*mapStore
	mu      sync.RWMutex
	indexes map[string]map[string]store.KeySet
}

func (s *indexedStore) transform(i interface{}) string {
	return fmt.Sprint(i)
}

func (s *indexedStore) index(field string, val interface{}, id store.PrimaryKey) {
	index := s.indexes[field]
	ik := s.transform(val)
	if _, ok := index[ik]; !ok {
		index[ik] = store.NewSet()
	}
	index[ik].Add(id)
}

func (s *indexedStore) deindex(field string, val interface{}, id store.PrimaryKey) {
	index := s.indexes[field]
	ik := s.transform(val)
	if idx, ok := index[ik]; ok {
		idx.Remove(id)
		if idx.IsEmpty() {
			delete(index, ik)
		}
	}
}

func (s *indexedStore) filterIndex(filter map[string]interface{}) ([]store.KeySet, error) {
	filtered := make([]store.KeySet, 0, len(filter))
	for k, v := range filter {
		index, ok := s.indexes[k]
		if !ok {
			return nil, fmt.Errorf(`no index "%s"`, k)
		}

		idxs, ok := index[s.transform(v)]
		if !ok {
			filtered = append(filtered, make(store.KeySet, 0))
		} else {
			filtered = append(filtered, idxs)
		}
	}
	return filtered, nil
}

func (s *indexedStore) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, err := s.mapStore.Create(fields)
	if err != nil {
		return id, err
	}
	if len(fields) == 0 {
		return id, nil
	}

	for field := range s.indexes {
		if val, ok := fields[field]; ok {
			s.index(field, val, id)
		}
	}
	return id, nil
}

func (s *indexedStore) Update(id store.PrimaryKey, diff map[string]interface{}) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}
	if len(diff) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.mapStore.Find(id)
	if err != nil {
		return err
	}

	err = s.mapStore.Update(id, diff)
	if err != nil {
		return err
	}
	for field := range s.indexes {
		newVal, newOK := diff[field]
		if !newOK {
			continue
		}
		oldVal, oldOK := old[field]
		if oldOK {
			s.deindex(field, oldVal, id)
		}
		s.index(field, newVal, id)
	}
	return nil
}

func (s *indexedStore) FindBy(filter map[string]interface{}) (matches map[store.PrimaryKey]map[string]interface{}, err error) {
	if len(filter) == 0 {
		return nil, store.ErrEmptyFilter.New(nil)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	idxs := make(map[string]interface{}, len(filter))
	fields := make(map[string]interface{}, len(filter))
	for k, v := range filter {
		_, ok := s.indexes[k]
		if ok {
			idxs[k] = v
		} else {
			fields[k] = v
		}
	}

	var byFields map[store.PrimaryKey]map[string]interface{}
	if len(fields) > 0 {
		byFields, err = s.mapStore.FindBy(fields)
		if err != nil {
			return nil, err
		}
	}

	idxKeys, err := s.filterIndex(idxs)
	if err != nil {
		return nil, err
	}
	sort.Slice(idxKeys, func(i, j int) bool { // Optimization: start with the smallest set
		return idxKeys[i].Size() < idxKeys[j].Size()
	})
	var filterSet store.KeySet
	for _, set := range idxKeys {
		if filterSet == nil {
			filterSet = set
			continue
		}
		filterSet.Intersect(set)
	}

	matches = make(map[store.PrimaryKey]map[string]interface{})
	switch {
	case len(idxs) != 0 && len(fields) != 0:
		for k, v := range byFields {
			if filterSet.Contains(k) {
				matches[k] = v
			}
		}
	case len(idxs) != 0:
		for k := range filterSet {
			if v, err := s.Find(k); err == nil {
				matches[k] = v
			}
		}
	default:
		matches = byFields
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

func (s *indexedStore) Delete(id store.PrimaryKey) error {
	if id == nil {
		return store.ErrNilKey.New(nil)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.mapStore.Find(id)
	if err != nil {
		return err
	}
	err = s.mapStore.Delete(id)
	if err != nil {
		return err
	}
	for field := range s.indexes {
		val, ok := old[field]
		if ok {
			s.deindex(field, val, id)
		}
	}
	return nil
}
