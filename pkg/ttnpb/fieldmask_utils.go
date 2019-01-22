// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import "strings"

// TopLevelFields returns the unique top level fields of the given paths.
func TopLevelFields(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		parts := strings.SplitN(path, ".", 2)
		if _, ok := seen[parts[0]]; ok {
			continue
		}
		seen[parts[0]] = struct{}{}
		out = append(out, parts[0])
	}
	return out
}

// HasOnlyAllowedFields returns whether the given requested paths only contains paths that are allowed.
// The requested fields (i.e. `a.b`) may be of a lower level than the allowed path (i.e. `a`).
func HasOnlyAllowedFields(requested []string, allowed ...string) bool {
nextRequested:
	for _, requested := range requested {
		for _, allowed := range allowed {
			if requested == allowed || strings.HasPrefix(requested, allowed+".") {
				continue nextRequested
			}
		}
		return false
	}
	return true
}

// HasAnyField returns whether the given requested paths contain any of the given fields.
// The requested fields (i.e. `a.b`) may be of a higher level than the search path (i.e. `a.b.c`).
func HasAnyField(requested []string, search ...string) bool {
	for _, requested := range requested {
		for _, search := range search {
			if requested == search || strings.HasPrefix(search, requested+".") {
				return true
			}
		}
	}
	return false
}
