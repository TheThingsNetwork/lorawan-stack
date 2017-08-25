// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

// StringSet implements "set" data structure that contains strings
type StringSet map[string]bool

// NewStringSet creates a new set that holds ULIDs
func NewStringSet() StringSet {
	return make(StringSet)
}

// Size returns the size of the set
func (s StringSet) Size() int {
	return len(s)
}

// IsEmpty returns whether the set is empty
func (s StringSet) IsEmpty() bool {
	return len(s) == 0
}

// Add adds the ULID to the set
func (s StringSet) Add(u string) {
	s[u] = true
}

// Remove removes the ULID from the set
func (s StringSet) Remove(u string) {
	delete(s, u)
}

// Contains returns whether the set contains the given ULID
func (s StringSet) Contains(u string) bool {
	_, exists := s[u]
	return exists
}

// Slice returns the set as a slice
func (s StringSet) Slice() []string {
	result := make([]string, 0, len(s))
	for u := range s {
		result = append(result, u)
	}
	return result
}

// Union returns the union of the two sets
func (s StringSet) Union(other StringSet) StringSet {
	result := make(StringSet)
	for value := range s {
		result[value] = true
	}
	for value := range other {
		result[value] = true
	}
	return result
}

// Intersect returns the intersection of the two sets
func (s StringSet) Intersect(other StringSet) StringSet {
	result := make(StringSet)
	for u := range s {
		if other.Contains(u) {
			result[u] = true
		}
	}
	return result
}
