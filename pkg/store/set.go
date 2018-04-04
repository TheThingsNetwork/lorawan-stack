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
