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

package oauth

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ParseScope parses a scope string to a list of rights.
func ParseScope(scope string) ([]ttnpb.Right, error) {
	split := strings.Fields(scope)
	res := make([]ttnpb.Right, 0, len(split))
	for _, str := range split {
		right, err := ttnpb.ParseRight(str)
		if err != nil {
			return nil, fmt.Errorf("Invalid right: %s", str)
		}
		res = append(res, right)
	}
	return res, nil
}

// Scope takes a list of rights and returns a string representing the
// scope that contains those rights.
func Scope(rights []ttnpb.Right) string {
	switch len(rights) {
	case 0:
		return ""
	case 1:
		return rights[0].String()
	default:
		rightStrings := make([]string, len(rights))
		for i, right := range rights {
			rightStrings[i] = right.String()
		}
		return strings.Join(rightStrings, " ")
	}
}

// Subtract subtracts the rights in set from the rights in from, returning only the rights in
// from that are not in set.
func Subtract(from []ttnpb.Right, set []ttnpb.Right) []ttnpb.Right {
	res := make([]ttnpb.Right, 0, len(from))
	for _, right := range from {
		if !isMember(right, set) {
			res = append(res, right)
		}
	}

	return res
}

func isMember(right ttnpb.Right, set []ttnpb.Right) bool {
	for _, el := range set {
		if right == el {
			return true
		}
	}

	return false
}
