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

package ttnpb

// IntersectRights returns the set of rights that are contained in both input sets.
func IntersectRights(a, b []Right) []Right {
	mapped := mapRights(a)

	res := make([]Right, 0)
	for _, right := range b {
		if _, exists := mapped[right]; exists {
			res = append(res, right)
		}
	}

	return res
}

// DifferenceRights returns the set of rights in `a` that are not contained in `b`.
func DifferenceRights(a, b []Right) []Right {
	mapped := mapRights(b)

	res := make([]Right, 0)
	for _, right := range a {
		if _, exists := mapped[right]; !exists {
			res = append(res, right)
		}
	}

	return res
}

// IncludesRights returns true if and only if all search rights are contained in list.
func IncludesRights(list []Right, search ...Right) bool {
	mapped := mapRights(list)

	for _, right := range search {
		if _, exists := mapped[right]; !exists {
			return false
		}
	}

	return true
}

func mapRights(list []Right) map[Right]bool {
	res := make(map[Right]bool)
	for _, right := range list {
		res[right] = true
	}
	return res
}
