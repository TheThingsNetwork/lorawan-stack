// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"fmt"
	"sort"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/oklog/ulid"
)

// NewIndexed returns a new MapStore that keeps indexes for the given fields
func NewIndexed(indexed ...string) store.Store {
	internal := New().(*mapStore)
	s := &indexedStore{
		mapStore: internal,
		indexes:  make(map[string]map[string]store.ULIDSet),
	}
	for _, field := range indexed {
		s.indexes[field] = make(map[string]store.ULIDSet)
	}
	return s
}

type indexedStore struct {
	*mapStore
	mu      sync.RWMutex
	indexes map[string]map[string]store.ULIDSet
}

func (s *indexedStore) transform(i interface{}) string {
	return fmt.Sprint(i)
}

func (s *indexedStore) index(field string, val interface{}, id ulid.ULID) {
	index := s.indexes[field]
	ik := s.transform(val)
	if _, ok := index[ik]; !ok {
		index[ik] = store.NewULIDSet()
	}
	index[ik].Add(id)
}

func (s *indexedStore) deindex(field string, val interface{}, id ulid.ULID) {
	index := s.indexes[field]
	ik := s.transform(val)
	if idx, ok := index[ik]; ok {
		idx.Remove(id)
		if idx.IsEmpty() {
			delete(index, ik)
		}
	}
}

func (s *indexedStore) filterIndex(filters map[string]interface{}) ([]store.ULIDSet, error) {
	filtered := make([]store.ULIDSet, len(filters))
	var i int
	for filterK, filterV := range filters {
		index, ok := s.indexes[filterK]
		if !ok {
			return nil, fmt.Errorf(`no index "%s"`, filterK)
		}
		filtered[i], _ = index[s.transform(filterV)]
		i++
	}
	return filtered, nil
}

func (s *indexedStore) Create(obj map[string]interface{}) (ulid.ULID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, err := s.mapStore.Create(obj)
	if err != nil {
		return id, err
	}
	for field := range s.indexes {
		if val, ok := obj[field]; ok {
			s.index(field, val, id)
		}
	}
	return id, nil
}

func (s *indexedStore) Update(id ulid.ULID, new, old map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.mapStore.Update(id, new, old)
	if err != nil {
		return err
	}
	for field := range s.indexes {
		oldVal, oldOK := old[field]
		newVal, newOK := new[field]
		if s.transform(oldVal) == s.transform(newVal) {
			continue
		}
		if oldOK {
			s.deindex(field, oldVal, id)
		}
		if newOK {
			s.index(field, newVal, id)
		}
	}
	return nil
}

func (s *indexedStore) FindBy(filters map[string]interface{}) (map[ulid.ULID]map[string]interface{}, error) {
	matches := make(map[ulid.ULID]map[string]interface{})
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered, err := s.filterIndex(filters)
	if err != nil {
		return nil, err
	}
	sort.Slice(filtered, func(i, j int) bool { // Optimization: start with the smallest set
		return filtered[i].Size() < filtered[j].Size()
	})
	var filterSet store.ULIDSet
	for _, set := range filtered {
		if filterSet == nil {
			filterSet = set
			continue
		}
		filterSet.Intersect(set)
	}
	for id := range filterSet {
		if obj, err := s.Find(id); err == nil {
			matches[id] = obj
		}
	}
	return matches, nil
}

func (s *indexedStore) Delete(id ulid.ULID) error {
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
