// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

func unflattened(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		skeys := strings.Split(k, ".")
		parent := out
		for _, sk := range skeys[:len(skeys)-1] {
			sm, ok := parent[sk]
			if !ok {
				sm = make(map[string]interface{})
				parent[sk] = sm
			}
			parent = sm.(map[string]interface{})
		}
		parent[skeys[len(skeys)-1]] = v
	}
	return out
}

// MapUnmarshaler is capable of unmarshaling itself from a nested map[string]interface{}
type MapUnmarshaler interface {
	UnmarshalMap(map[string]interface{}) error
}

func Unmarshal(m map[string]interface{}, v interface{}) error {
	m = unflattened(m)
	switch t := v.(type) {
	case MapUnmarshaler:
		return t.UnmarshalMap(m)
	case map[string]interface{}:
		for k, v := range m {
			t[k] = v
		}
		return nil
	default:
		return mapstructure.Decode(m, v)
	}
}
