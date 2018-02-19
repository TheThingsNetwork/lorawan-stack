// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package util

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// RightsIntersection returns the intersection between two slices of rights.
func RightsIntersection(a, b []ttnpb.Right) []ttnpb.Right {
	res := make([]ttnpb.Right, 0)

	for _, elemA := range a {
		for _, elemB := range b {
			if elemA == elemB {
				res = append(res, elemA)
			}
		}
	}

	return res
}

// RightsDifference returns the set of rights of `a` that are not contained in `b`.
func RightsDifference(a, b []ttnpb.Right) []ttnpb.Right {
	mapped := make(map[ttnpb.Right]bool)
	for _, elemB := range b {
		mapped[elemB] = true
	}

	res := make([]ttnpb.Right, 0)
	for _, elemA := range a {
		if _, exists := mapped[elemA]; !exists {
			res = append(res, elemA)
		}
	}

	return res
}
