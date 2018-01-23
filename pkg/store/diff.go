// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

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
func DiffFields(a, b interface{}) (fields []string) {
	for k := range Diff(MarshalMap(a), MarshalMap(b)) {
		fields = append(fields, k)
	}
	return fields
}
