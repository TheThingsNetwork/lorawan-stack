// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"fmt"
	"sort"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/store"
)

// NewIndexed returns a new MapStore that keeps indexes for the given fields
func NewIndexed(indexed ...string) store.Interface {
	internal := New().(*mapStore)
	s := &indexedStore{
		mapStore: internal,
		indexes:  make(map[string]map[string]store.Set),
	}
	for _, field := range indexed {
		s.indexes[field] = make(map[string]store.Set)
	}
	return s
}

type indexedStore struct {
	*mapStore
	mu      sync.RWMutex
	indexes map[string]map[string]store.Set
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

func (s *indexedStore) filterIndex(filters map[string]interface{}) ([]store.Set, error) {
	filtered := make([]store.Set, 0, len(filters))
	for k, v := range filters {
		index, ok := s.indexes[k]
		if !ok {
			return nil, fmt.Errorf(`no index "%s"`, k)
		}

		idxs, ok := index[s.transform(v)]
		if !ok {
			filtered = append(filtered, make(store.Set, 0))
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
	for field := range s.indexes {
		if val, ok := fields[field]; ok {
			s.index(field, val, id)
		}
	}
	return id, nil
}

func (s *indexedStore) Update(id store.PrimaryKey, diff map[string]interface{}) error {
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

func (s *indexedStore) FindBy(filters map[string]interface{}) (map[store.PrimaryKey]map[string]interface{}, error) {
	matches := make(map[store.PrimaryKey]map[string]interface{})
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered, err := s.filterIndex(filters)
	if err != nil {
		return nil, err
	}
	sort.Slice(filtered, func(i, j int) bool { // Optimization: start with the smallest set
		return filtered[i].Size() < filtered[j].Size()
	})
	var filterSet store.Set
	for _, set := range filtered {
		if filterSet == nil {
			filterSet = set
			continue
		}
		filterSet.Intersect(set)
	}
	for id := range filterSet {
		if fields, err := s.Find(id); err == nil {
			matches[id] = fields
		}
	}
	return matches, nil
}

func (s *indexedStore) Delete(id store.PrimaryKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.mapStore.Find(id)
	if err == store.ErrNotFound {
		return nil
	}
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
