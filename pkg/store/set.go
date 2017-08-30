// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

// Set implements "set" data structure that contains strings
type Set map[PrimaryKey]struct{}

// NewSet creates a new set that holds primary keys
func NewSet() Set {
	return make(Set)
}

// Size returns the size of the set
func (s Set) Size() int {
	return len(s)
}

// IsEmpty returns whether the set is empty
func (s Set) IsEmpty() bool {
	return len(s) == 0
}

// Add adds the PrimaryKey to the set
func (s Set) Add(u PrimaryKey) {
	s[u] = struct{}{}
}

// Remove removes the PrimaryKey from the set
func (s Set) Remove(u PrimaryKey) {
	delete(s, u)
}

// Contains returns whether the set contains the given PrimaryKey
func (s Set) Contains(u PrimaryKey) bool {
	_, exists := s[u]
	return exists
}

// Slice returns the set as a slice
func (s Set) Slice() []PrimaryKey {
	result := make([]PrimaryKey, 0, len(s))
	for u := range s {
		result = append(result, u)
	}
	return result
}

// Union returns the union of the two sets
func (s Set) Union(other Set) Set {
	result := make(Set, len(other))
	for value := range s {
		result.Add(value)
	}
	for value := range other {
		result.Add(value)
	}
	return result
}

// Intersect returns the intersection of the two sets
func (s Set) Intersect(other Set) Set {
	result := make(Set, len(other))
	for u := range s {
		if other.Contains(u) {
			result.Add(u)
		}
	}
	return result
}
