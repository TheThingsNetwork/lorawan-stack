// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "reflect"

// Code is in parts adapted from https://goo.gl/MB5Sao, which is MIT licensed

// flattened returns a copy of m with keys 'flattened'.
// If the map contains sub-maps, the values of these sub-maps are set under the root map, each level separated by a dot
func flattened(m map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		if sub, ok := v.(map[string]interface{}); ok {
			flattened := flattened(sub)
			for j, v := range flattened {
				m[k+"."+j] = v
			}
			delete(m, k)
		}
	}
	return m
}

// marshalNested retrieves recursively all types for the given value
// and returns the marhshaled nested value.
func marshalNested(val reflect.Value) interface{} {
	var k reflect.Kind
	if val.Kind() == reflect.Ptr {
		k = val.Elem().Kind()
	} else {
		k = val.Kind()
	}

	switch k {
	case reflect.Struct:
		m := marshal(val.Interface())

		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) == 0 {
			return val.Interface()
		} else {
			return m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()
		switch val.Type().Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map,
			reflect.Slice, reflect.Chan:
			mapElem = val.Type().Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice &&
				mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]interface{}, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = marshalNested(val.MapIndex(k))
			}
			return m
		}

		return val.Interface()
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			return val.Interface()
		}

		if val.Type().Elem().Kind() != reflect.Struct &&
			!(val.Type().Elem().Kind() == reflect.Ptr &&
				val.Type().Elem().Elem().Kind() == reflect.Struct) {
			return val.Interface()
		}

		slices := make([]interface{}, val.Len(), val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = marshalNested(val.Index(x))
		}
		return slices
	default:
		return val.Interface()
	}
}

// marhshal converts the given struct s to a flattened map[string]interface{}
func marshal(s interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(s))
	t := v.Type()

	out := make(map[string]interface{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.PkgPath != "" {
			continue
		}

		fv := v.FieldByName(ft.Name)
		fk := fv.Kind()

		isZero := false
		switch fk {
		case reflect.Chan, reflect.Map, reflect.Slice:
			isZero = fv.IsNil() || fv.Len() == 0
		case reflect.Func, reflect.Interface, reflect.Ptr:
			isZero = fv.IsNil()
		case reflect.Array, reflect.String:
			isZero = fv.Len() == 0
		default:
			isZero = fv.Interface() == reflect.Zero(fv.Type()).Interface()
		}
		if isZero {
			continue
		}

		val := marshalNested(fv)

		if fv.Kind() == reflect.Ptr {
			fk = fv.Elem().Kind()
		}

		if fk == reflect.Struct || fk == reflect.Map {
			valMap := val.(map[string]interface{})
			for k := range valMap {
				out[ft.Name+"."+k] = valMap[k]
			}
		} else {
			out[ft.Name] = val
		}
	}
	return out
}

type MapMarshaler interface {
	MarshalMap() map[string]interface{}
}

func Marshal(v interface{}) map[string]interface{} {
	switch t := v.(type) {
	case MapMarshaler:
		return flattened(t.MarshalMap())
	case map[string]interface{}:
		return flattened(t)
	default:
		return marshal(v)
	}
}
