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

// Code is in parts adapted from https://github.com/fatih/structs/blob/master/structs.go, which is MIT licensed

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
		}
		return m
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

	if t.Kind() != reflect.Struct {
		panic(errors.Errorf("github.com/TheThingsNetwork/ttn/pkg/store.marshal: expected argument to be a struct, got %s", t.Kind()))
	}

	out := make(map[string]interface{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if ft.PkgPath != "" {
			continue
		}

		fv := v.FieldByName(ft.Name)
		fk := fv.Kind()

		if reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()) {
			continue
		}

		val := marshalNested(fv)

		if fv.Kind() == reflect.Ptr {
			fk = fv.Elem().Kind()
		}

		if fk == reflect.Struct || fk == reflect.Map {
			if m, ok := val.(map[string]interface{}); ok {
				for k, v := range m {
					out[ft.Name+Separator+k] = v
				}
				continue
			}
		}
		out[ft.Name] = val
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

// MarshalMap returns the map encoding of v.
//
// MarshalMap traverses the value v recursively. If v implements the MapMarshaler interface, MarshalMap calls its MarshalMap method to produce map.
// Otherwise, MarshalMap first encodes the value v as a map[string]interface{} and flattens it, by joining sub-map values with a dot. Structs are encoded as map[string]inteface{}
func MarshalMap(v interface{}) map[string]interface{} {
	switch v := v.(type) {
	case MapMarshaler:
		return v.MarshalMap()
	case map[string]interface{}:
		return flattened(v)
	default:
		return marshal(v)
	}
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
	switch v := v.(type) {
	case ByteMapMarshaler:
		return v.MarshalByteMap()
	case map[string][]byte:
		return v, nil
	}
	im := MarshalMap(v)
	bm := make(map[string][]byte, len(im))
	for k, v := range im {
		b, err := toBytes(v)
		if err != nil {
			return nil, err
		}
		bm[k] = b
	}
	return bm, nil
}
