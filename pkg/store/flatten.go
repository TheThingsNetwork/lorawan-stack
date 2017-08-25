// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

// Flatten the map.
//
// If the map contains sub-maps, the values of these sub-maps are set under the root map, each level separated by a dot
func Flatten(s map[string]interface{}) map[string]interface{} {
	for k, v := range s {
		if sub, ok := v.(map[string]interface{}); ok {
			flattened := Flatten(sub)
			for j, v := range flattened {
				s[k+"."+j] = v
			}
			delete(s, k)
		}
	}
	return s
}
