// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"reflect"
	"sort"
)

func keys(v map[string]interface{}) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}

// Diff the new and the old value and return the changed fields
func Diff(new, old map[string]interface{}) (diff map[string]interface{}) {
	allKeys := append(keys(old), keys(new)...)

	// Remove duplicate keys
	if len(allKeys) > 1 {
		sort.Strings(allKeys)
		var p int
		for i := 1; i < len(allKeys); i++ {
			if allKeys[i] == allKeys[p] {
				continue
			}
			p++
			if p < i {
				allKeys[p], allKeys[i] = allKeys[i], allKeys[p]
			}
		}
		allKeys = allKeys[:p+1]
	}

	// The actual diff
	diff = make(map[string]interface{})
	for _, key := range allKeys {
		oldVal, oldOK := old[key]
		newVal, newOK := new[key]
		if oldOK && newOK {
			if !reflect.DeepEqual(oldVal, newVal) {
				diff[key] = newVal
			}
			continue
		}
		if newOK {
			diff[key] = newVal
		}
		if oldOK {
			diff[key] = nil
		}
	}

	return diff
}
