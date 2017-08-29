// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "reflect"

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
