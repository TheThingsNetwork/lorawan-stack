// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

// KeySet implements "set" data structure that contains instances of PrimaryKey
type KeySet map[PrimaryKey]struct{}

// NewSet creates a new set that holds primary keys
func NewSet() KeySet {
	return make(KeySet)
}

// Size returns the size of the set
func (s KeySet) Size() int {
	return len(s)
}

// IsEmpty returns whether the set is empty
func (s KeySet) IsEmpty() bool {
	return len(s) == 0
}

// Add adds the PrimaryKey to the set
func (s KeySet) Add(k PrimaryKey) {
	s[k] = struct{}{}
}

// Remove removes the PrimaryKey from the set
func (s KeySet) Remove(k PrimaryKey) {
	delete(s, k)
}

// Contains returns whether the set contains the given PrimaryKey
func (s KeySet) Contains(k PrimaryKey) bool {
	_, ok := s[k]
	return ok
}

// Slice returns the set as a slice
func (s KeySet) Slice() []PrimaryKey {
	result := make([]PrimaryKey, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	return result
}

// Union returns the union of the two sets
func (s KeySet) Union(other KeySet) KeySet {
	result := make(KeySet, len(other))
	for k := range s {
		result.Add(k)
	}
	for k := range other {
		result.Add(k)
	}
	return result
}

// Intersect returns the intersection of the two sets
func (s KeySet) Intersect(other KeySet) KeySet {
	result := make(KeySet, len(other))
	for k := range s {
		if other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}
