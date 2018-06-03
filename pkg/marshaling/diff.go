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

package marshaling

import (
	"bytes"
	"reflect"
)

// Diff the new and the old value and return the changed fields
func Diff(new, old map[string]interface{}) (diff map[string]interface{}) {
	keys := make(map[string]struct{}, len(new))
	for k := range old {
		keys[k] = struct{}{}
	}
	for k := range new {
		keys[k] = struct{}{}
	}

	diff = make(map[string]interface{}, len(keys))
	for k := range keys {
		newVal, newOK := new[k]
		oldVal, oldOK := old[k]
		switch {
		case newOK && oldOK:
			if !reflect.DeepEqual(newVal, oldVal) {
				diff[k] = newVal
			}
		case newOK:
			diff[k] = newVal
		case oldOK:
			diff[k] = nil
		}
	}
	return diff
}

// ByteDiff the new and the old value and return the changed fields
func ByteDiff(new, old map[string][]byte) (diff map[string][]byte) {
	keys := make(map[string]struct{}, len(new))
	for k := range old {
		keys[k] = struct{}{}
	}
	for k := range new {
		keys[k] = struct{}{}
	}

	diff = make(map[string][]byte, len(keys))
	for k := range keys {
		newVal, newOK := new[k]
		oldVal, oldOK := old[k]
		switch {
		case newOK && oldOK:
			if !bytes.Equal(newVal, oldVal) {
				diff[k] = newVal
			}
		case newOK:
			diff[k] = newVal
		case oldOK:
			diff[k] = nil
		}
	}
	return diff
}

// DiffFields returns the fieldpaths of fields of a and b that differ in value.
func DiffFields(a, b interface{}) (fields []string, err error) {
	ma, err := MarshalMap(a)
	if err != nil {
		return nil, err
	}

	mb, err := MarshalMap(b)
	if err != nil {
		return nil, err
	}

	for k := range Diff(ma, mb) {
		fields = append(fields, k)
	}
	return fields, nil
}
