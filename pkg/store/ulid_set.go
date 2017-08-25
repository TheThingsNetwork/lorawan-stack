// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/oklog/ulid"

// ULIDSet implements "set" data structure that contains ULIDs
type ULIDSet map[ulid.ULID]bool

// NewULIDSet creates a new set that holds ULIDs
func NewULIDSet() ULIDSet {
	return make(ULIDSet)
}

// Size returns the size of the set
func (s ULIDSet) Size() int {
	return len(s)
}

// IsEmtpy returns whether the set is empty
func (s ULIDSet) IsEmpty() bool {
	return len(s) == 0
}

// Add adds the ULID to the set
func (s ULIDSet) Add(u ulid.ULID) {
	s[u] = true
}

// Remove removes the ULID from the set
func (s ULIDSet) Remove(u ulid.ULID) {
	delete(s, u)
}

// Contains returns whether the set contains the given ULID
func (s ULIDSet) Contains(u ulid.ULID) bool {
	_, exists := s[u]
	return exists
}

// Slice returns the set as a slice
func (s ULIDSet) Slice() []ulid.ULID {
	result := make([]ulid.ULID, 0, len(s))
	for u := range s {
		result = append(result, u)
	}
	return result
}

// Union returns the union of the two sets
func (s ULIDSet) Union(other ULIDSet) ULIDSet {
	result := make(ULIDSet)
	for value := range s {
		result[value] = true
	}
	for value := range other {
		result[value] = true
	}
	return result
}

// Intersect returns the intersection of the two sets
func (s ULIDSet) Intersect(other ULIDSet) ULIDSet {
	result := make(ULIDSet)
	for u := range s {
		if other.Contains(u) {
			result[u] = true
		}
	}
	return result
}
