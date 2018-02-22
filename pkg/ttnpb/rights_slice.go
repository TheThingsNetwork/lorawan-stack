// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
