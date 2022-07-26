// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import "strings"

// FieldMask is used to specify applicable fields in SELECT or UPDATE queries.
type FieldMask []string

// Contains returns true if the given field is present in the field mask.
func (fm FieldMask) Contains(search string) bool {
	for _, f := range fm {
		if f == search {
			return true
		}
	}
	return false
}

// TopLevel returns the top-level fields from the field mask.
func (fm FieldMask) TopLevel() FieldMask {
	out := make(FieldMask, 0, len(fm))
	for _, f := range fm {
		before, _, _ := strings.Cut(f, ".")
		if !out.Contains(before) {
			out = append(out, before)
		}
	}
	return out
}

// TrimPrefix returns a field mask with all fields of fm that contain prefix, but then having that prefix removed.
func (fm FieldMask) TrimPrefix(prefix string) FieldMask {
	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	out := make([]string, 0, len(fm))
	for _, f := range fm {
		if strings.HasPrefix(f, prefix) {
			out = append(out, strings.TrimPrefix(f, prefix))
		}
	}
	return out
}
