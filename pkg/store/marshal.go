// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// Flattened returns a copy of m with keys 'Flattened'.
// If the map contains sub-maps, the values of these sub-maps are set under the root map, each level separated by Separator.
func Flattened(m map[string]interface{}) (out map[string]interface{}) {
	out = make(map[string]interface{}, len(m))
	for k, v := range m {
		if sm, ok := v.(map[string]interface{}); ok {
			sm = Flattened(sm)
			for sk, sv := range sm {
				out[k+Separator+sk] = sv
			}
		} else {
			out[k] = v
		}
	}
	return
}

// KeepBytes is a keep function intended to be used in Mapify.
func KeepBytes(rv reflect.Value) bool {
	return (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Type().Elem().Kind() == reflect.Uint8
}

// Mapify recursively replaces (sub-)slices in v by maps. If keep is set and returns true for encountered value, the value will be kept as-is. Look at keepBytes for an example of where this is useful.
func Mapify(v interface{}, keep func(rv reflect.Value) bool) interface{} {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	if keep != nil && keep(rv) {
		return v
	}

	switch rv.Kind() {
	case reflect.Map:
		m := make(map[string]interface{}, rv.Len())
		for _, sk := range rv.MapKeys() {
			sv := reflect.ValueOf(rv.MapIndex(sk).Interface())
			if !sv.IsValid() {
				m[sk.String()] = nil
				continue
			}
			m[sk.String()] = Mapify(sv.Interface(), keep)
		}
		return m
	case reflect.Slice, reflect.Array:
		m := make(map[string]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			sv := reflect.ValueOf(rv.Index(i).Interface())
			if !sv.IsValid() {
				m[strconv.Itoa(i)] = nil
				continue
			}
			m[strconv.Itoa(i)] = Mapify(sv.Interface(), keep)
		}
		return m
	default:
		return v
	}
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return isZero(v.Elem())
	case reflect.Func:
		return v.IsNil()
	case reflect.Map, reflect.Slice, reflect.Array, reflect.String, reflect.Chan:
		return v.Len() == 0
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// marshalNested retrieves recursively all types for the given value
// and returns the marshaled nested value.
func marshalNested(v reflect.Value) (interface{}, error) {
	if !v.IsValid() {
		return nil, nil
	}

	iv := reflect.Indirect(v)
	if !iv.IsValid() {
		return nil, nil
	}
	if m, ok := v.Interface().(MapMarshaler); ok {
		return m.MarshalMap()
	}
	v = iv

	var err error
	switch v.Kind() {
	case reflect.Map:
		m := make(map[string]interface{}, v.Len())
		switch kt := v.Type().Key(); {
		case kt.Kind() == reflect.String:
			for _, k := range v.MapKeys() {
				m[k.String()], err = marshalNested(v.MapIndex(k))
				if err != nil {
					return nil, err
				}
			}
			return m, nil
		case kt.Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()):
			for _, k := range v.MapKeys() {
				m[k.Interface().(fmt.Stringer).String()], err = marshalNested(v.MapIndex(k))
				if err != nil {
					return nil, err
				}
			}
			return m, nil
		default:
			return nil, errors.Errorf("Expected the map key kind to be string or implement Stringer, got %s", kt)
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).PkgPath == "" {
				// Only attempt to marshal structs with exported fields
				return marshal(v.Interface())
			}
		}
		return v.Interface(), nil
	case reflect.Slice, reflect.Array:
		switch e := v.Type().Elem(); e.Kind() {
		case reflect.Ptr:
			switch se := e.Elem(); se.Kind() {
			case reflect.Interface, reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			default:
				// return slices of pointers to simple types as slices of values of that type
				n := v.Len()
				sl := reflect.MakeSlice(se, n, n)
				for i := 0; i < n; i++ {
					sl.Index(i).Set(reflect.Indirect(v.Index(i)))
				}
				return sl.Interface(), nil
			}
		case reflect.Interface, reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		default:
			// return slices of simple types as-is
			return v.Interface(), nil
		}

		n := v.Len()
		s := make([]interface{}, n, n)
		for i := 0; i < n; i++ {
			sv := v.Index(i)
			if isZero(sv) {
				continue
			}
			s[i], err = marshalNested(sv)
			if err != nil {
				return nil, err
			}
		}
		return s, nil
	default:
		return v.Interface(), nil
	}
}

// marhshal converts the given struct s to a map[string]interface{}
func marshal(s interface{}) (m map[string]interface{}, err error) {
	if mm, ok := s.(MapMarshaler); ok {
		return mm.MarshalMap()
	}

	v := reflect.Indirect(reflect.ValueOf(s))
	if v.Kind() != reflect.Struct {
		return nil, errors.Errorf("Expected argument to be a struct, got %s(%s)", v.Type(), v.Kind())
	}

	t := v.Type()
	m = make(map[string]interface{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if fv := v.FieldByName(f.Name); !isZero(fv) {
			m[f.Name], err = marshalNested(fv)
			if err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

// MapMarshaler is the interface implemented by an object that can
// marshal itself into a flattened map[string]interface{}
//
// MarshalMap encodes the receiver into map[string]interface{} and returns the result.
type MapMarshaler interface {
	MarshalMap() (map[string]interface{}, error)
}

// MarshalMap returns the map encoding of v, where v is either a struct or a map with string keys.
//
// MarshalMap traverses the value v recursively. If v implements the MapMarshaler interface, MarshalMap calls its MarshalMap method to produce map[string]interface{}.
// Otherwise, MarshalMap first encodes the value v as a map[string]interface{}. Default marshaler marshals slices as maps with string keys, where all keys represent integers.
// The map produced by any of the methods will be flattened by joining sub-map values with a dot(note that slices produced by custom MarshalMap implementations won't be flattened).
func MarshalMap(v interface{}) (m map[string]interface{}, err error) {
	if mm, ok := v.(MapMarshaler); ok {
		m, err = mm.MarshalMap()
	} else {
		m, err = marshal(v)
	}
	if err != nil {
		return nil, err
	}
	m = Flattened(Mapify(m, KeepBytes).(map[string]interface{}))
	for k, v := range m {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Struct, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
			bv, err := ToBytes(v)
			if err != nil {
				return nil, err
			}
			m[k] = bv
		}
	}
	return m, nil
}

var jsonpbMarshaler = &jsonpb.Marshaler{}

func ToBytes(v interface{}) (b []byte, err error) {
	var enc Encoding
	defer func() {
		if err != nil {
			return
		}
		b = append([]byte{byte(enc)}, b...)
	}()

	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		enc = RawEncoding
		return []byte{}, nil
	}
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
	case jsonpb.JSONPBMarshaler:
		enc = JSONPBEncoding
		return v.MarshalJSONPB(jsonpbMarshaler)
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
func MarshalByteMap(v interface{}) (bm map[string][]byte, err error) {
	if bmm, ok := v.(ByteMapMarshaler); ok {
		return bmm.MarshalByteMap()
	}

	var im map[string]interface{}
	switch v := v.(type) {
	case MapMarshaler:
		im, err = v.MarshalMap()
		if err != nil {
			return nil, err
		}
	default:
		im, err = marshal(v)
		if err != nil {
			return nil, err
		}
		im = Flattened(Mapify(im, KeepBytes).(map[string]interface{}))
	}

	bm = make(map[string][]byte, len(im))
	for k, v := range im {
		b, err := ToBytes(v)
		if err != nil {
			return nil, err
		}
		bm[k] = b
	}
	return bm, nil
}
