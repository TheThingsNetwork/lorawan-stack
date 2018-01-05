// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

// flattened returns a copy of m with keys 'flattened'.
// If the map contains sub-maps, the values of these sub-maps are set under the root map, each level separated by a dot
func flattened(m map[string]interface{}) (out map[string]interface{}) {
	out = make(map[string]interface{}, len(m))
	for k, v := range m {
		if sm, ok := v.(map[string]interface{}); ok {
			sm = flattened(sm)
			for sk, sv := range sm {
				out[k+Separator+sk] = sv
			}
		} else {
			out[k] = v
		}
	}
	return
}

func keepBytes(rv reflect.Value) bool {
	return (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Type().Elem().Kind() == reflect.Uint8
}

// mapify recursively replaces (sub-)slices in v by maps. If keep is set and returns true for encountered value, the value will be kept as-is. Look at keepBytes for an example of where this is useful.
func mapify(v interface{}, keep func(rv reflect.Value) bool) interface{} {
	rv := reflect.ValueOf(v)
	if isZero(rv) || keep != nil && keep(rv) {
		return v
	}

	switch rv.Kind() {
	case reflect.Map:
		m := make(map[string]interface{}, rv.Len())
		for _, sk := range rv.MapKeys() {
			m[sk.String()] = mapify(rv.MapIndex(sk).Interface(), keep)
		}
		return m
	case reflect.Slice, reflect.Array:
		m := make(map[string]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			m[strconv.Itoa(i)] = mapify(rv.Index(i).Interface(), keep)
		}
		return m
	default:
		return v
	}
}

// marshalNested retrieves recursively all types for the given value
// and returns the marshaled nested value.
func marshalNested(v reflect.Value) interface{} {
	if m, ok := v.Interface().(MapMarshaler); ok {
		return m.MarshalMap()
	}
	if isZero(v) {
		return v.Interface()
	}

	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.Struct:
		m := marshal(v.Interface())
		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) != 0 {
			return m
		}
	case reflect.Map:
		vt := v.Type()
		if vt.Key().Kind() == reflect.String {
			return marshal(v.Interface())
		}
		if vt.Key().Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
			m := make(map[string]interface{})
			for _, vk := range v.MapKeys() {
				m[vk.Interface().(fmt.Stringer).String()] = v.MapIndex(vk)
			}
			return marshal(m)
		}
	case reflect.Slice, reflect.Array:
		switch e := v.Type().Elem(); e.Kind() {
		case reflect.Ptr:
			switch se := e.Elem(); se.Kind() {
			case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			default:
				n := v.Len()
				sl := reflect.MakeSlice(se, n, n)
				for i := 0; i < n; i++ {
					sl.Index(i).Set(reflect.Indirect(v.Index(i)))
				}
				return sl.Interface()
			}
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		default:
			return v.Interface()
		}

		n := v.Len()
		s := make([]interface{}, n, n)
		for i := 0; i < n; i++ {
			s[i] = marshalNested(v.Index(i))
		}
		return s
	}
	return v.Interface()
}

// isNil is safe alternative to IsNil.
func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func isZero(v reflect.Value) bool {
	return !v.IsValid() || isNil(v) || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// marhshal converts the given struct s to a map[string]interface{}
func marshal(s interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(s))
	t := v.Type()

	vals := make(map[string]reflect.Value)
	switch t.Kind() {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic(errors.Errorf("github.com/TheThingsNetwork/ttn/pkg/store.marshal: expected the map key kind to be string, got %s", t.Elem().Kind()))
		}
		for _, k := range v.MapKeys() {
			// https://stackoverflow.com/questions/14142667/reflect-value-mapindex-returns-a-value-different-from-reflect-valueof
			vals[k.String()] = reflect.ValueOf(v.MapIndex(k).Interface())
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}

			fv := v.FieldByName(f.Name)
			if isZero(fv) && !f.Type.Implements(reflect.TypeOf((*MapMarshaler)(nil)).Elem()) {
				continue
			}
			vals[f.Name] = fv
		}
	default:
		panic(errors.Errorf("github.com/TheThingsNetwork/ttn/pkg/store.marshal: expected argument to be a struct or map with string keys, got %s", t.Kind()))
	}

	out := make(map[string]interface{}, len(vals))
	for k, v := range vals {
		out[k] = marshalNested(v)
	}
	return out
}

// MapMarshaler is the interface implemented by an object that can
// marshal itself into a map[string]interface{}
//
// MarshalMap encodes the receiver into map[string]interface{} and returns the result.
type MapMarshaler interface {
	MarshalMap() map[string]interface{}
}

// MarshalMap returns the map encoding of v, where v is either a struct or a map with string keys.
//
// MarshalMap traverses the value v recursively. If v implements the MapMarshaler interface, MarshalMap calls its MarshalMap method to produce map[string]interface{}.
// Otherwise, MarshalMap first encodes the value v as a map[string]interface{}. Default marshaler marshals slices as maps with string keys, where all keys represent integers.
// The map produced by any of the methods will be flattened by joining sub-map values with a dot(note that slices produced by custom MarshalMap implementations won't be flattened).
func MarshalMap(v interface{}) map[string]interface{} {
	var im map[string]interface{}
	m, ok := v.(MapMarshaler)
	if ok {
		im = m.MarshalMap()
	} else {
		im = mapify(marshal(v), nil).(map[string]interface{})
	}
	return flattened(im)
}

func toBytes(v interface{}) (b []byte, err error) {
	var enc Encoding
	defer func() {
		if err != nil {
			return
		}
		b = append([]byte{byte(enc)}, b...)
	}()

	rv := reflect.Indirect(reflect.ValueOf(v))
	switch k := rv.Kind(); k {
	case reflect.String:
		enc = RawEncoding
		return []byte(rv.String()), nil
	case reflect.Bool:
		enc = RawEncoding
		return []byte(strconv.FormatBool(rv.Bool())), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		enc = RawEncoding
		return []byte(strconv.FormatInt(rv.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		enc = RawEncoding
		return []byte(strconv.FormatUint(rv.Uint(), 10)), nil
	case reflect.Float32:
		enc = RawEncoding
		return []byte(strconv.FormatFloat(rv.Float(), 'f', -1, 32)), nil
	case reflect.Float64:
		enc = RawEncoding
		return []byte(strconv.FormatFloat(rv.Float(), 'f', -1, 64)), nil
	case reflect.Slice, reflect.Array:
		elem := rv.Type().Elem()
		if elem.Kind() == reflect.Uint8 {
			enc = RawEncoding

			// Handle byte slices/arrays directly
			if k == reflect.Slice {
				return rv.Bytes(), nil
			}
			var byt byte
			out := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(byt)), rv.Len(), rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out.Index(i).Set(rv.Index(i))
			}
			return out.Bytes(), nil
		}
	}

	switch v := v.(type) {
	case encoding.BinaryMarshaler:
		enc = BinaryEncoding
		return v.MarshalBinary()
	case encoding.TextMarshaler:
		enc = TextEncoding
		return v.MarshalText()
	case proto.Marshaler:
		enc = ProtoEncoding
		return v.Marshal()
	case json.Marshaler:
		enc = JSONEncoding
		return v.MarshalJSON()
	}
	enc = UnknownEncoding
	return []byte(fmt.Sprint(v)), nil
}

// ByteMapMarshaler is the interface implemented by an object that can
// marshal itself into a map[string][]byte.
//
// MarshalByteMap encodes the receiver into map[string][]byte and returns the result.
type ByteMapMarshaler interface {
	MarshalByteMap() (map[string][]byte, error)
}

// MarshalByteMap returns the byte map encoding of v.
//
// MarshalByteMap traverses map returned by Marshal and converts all values to bytes.
func MarshalByteMap(v interface{}) (map[string][]byte, error) {
	var im map[string]interface{}
	switch v := v.(type) {
	case ByteMapMarshaler:
		return v.MarshalByteMap()
	case MapMarshaler:
		im = v.MarshalMap()
	default:
		im = mapify(marshal(v), keepBytes).(map[string]interface{})
	}

	bm := make(map[string][]byte, len(im))
	for k, v := range flattened(im) {
		b, err := toBytes(v)
		if err != nil {
			return nil, err
		}
		bm[k] = b
	}
	return bm, nil
}
