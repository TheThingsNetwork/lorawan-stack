// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gogo/protobuf/proto"
)

var (
	stringType       = reflect.TypeOf("")
	reflectValueType = reflect.TypeOf(reflect.Value{})

	fmtStringerType    = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	protoMarshalerType = reflect.TypeOf((*proto.Marshaler)(nil)).Elem()
	gobGobEncoderType  = reflect.TypeOf((*gob.GobEncoder)(nil)).Elem()
	jsonMarshalerType  = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	mapMarshalerType   = reflect.TypeOf((*MapMarshaler)(nil)).Elem()
)

// NOTE: Adapted from gob package source
// isZero reports whether the value is the zero of its type.
func isZero(val reflect.Value) bool {
	if !val.IsValid() {
		return true
	}

	switch val.Kind() {
	case reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if !isZero(val.Index(i)) {
				return false
			}
		}
		return true
	case reflect.String:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Complex64, reflect.Complex128:
		return val.Complex() == 0
	case reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
		return val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			if !isZero(val.Field(i)) {
				return false
			}
		}
		return true
	}
	panic("unknown type in isZero " + val.Type().String())
}

// FlattenedValue is like Flattened, but it operates on maps containing reflect.Value.
func FlattenedValue(m map[string]reflect.Value) (out map[string]reflect.Value) {
	out = make(map[string]reflect.Value, len(m))
	for k, v := range m {
		if !v.IsValid() {
			out[k] = v
			continue
		}

		vt := v.Type()
		if vt.Kind() == reflect.Map && vt.Key().Kind() == reflect.String && vt.Elem() == reflectValueType {
			for sk, sv := range FlattenedValue(v.Interface().(map[string]reflect.Value)) {
				out[k+Separator+sk] = sv
			}
		} else {
			out[k] = v
		}
	}
	return out
}

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
	return out
}

// ToValueMap converts the input map m into a map[string]reflect.Value by calling reflect.ValueOf with each value in m.
func ToValueMap(m map[string]interface{}) map[string]reflect.Value {
	vm := make(map[string]reflect.Value, len(m))
	for k, iv := range m {
		vm[k] = reflect.ValueOf(iv)
	}
	return vm
}

// marshalNested retrieves recursively all types for the given value
// and returns the marshaled nested value.
func marshalNested(v reflect.Value) (reflect.Value, error) {
	iv := reflect.Indirect(v)
	if !iv.IsValid() {
		return reflect.Value{}, nil
	}
	if v.Type().Implements(mapMarshalerType) {
		im, err := v.Interface().(MapMarshaler).MarshalMap()
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(ToValueMap(im)), nil
	}
	v = iv
	t := v.Type()

	switch t.Kind() {
	case reflect.Map:
		m := reflect.MakeMapWithSize(reflect.MapOf(stringType, reflectValueType), v.Len())
		switch kt := t.Key(); {
		case kt.Kind() == reflect.String:
			for _, k := range v.MapKeys() {
				nv, err := marshalNested(v.MapIndex(k))
				if err != nil {
					return reflect.Value{}, err
				}
				m.SetMapIndex(k, reflect.ValueOf(nv))
			}
			return m, nil
		case kt.Implements(fmtStringerType):
			for _, k := range v.MapKeys() {
				nv, err := marshalNested(v.MapIndex(k))
				if err != nil {
					return reflect.Value{}, err
				}
				m.SetMapIndex(reflect.ValueOf(k.Interface().(fmt.Stringer).String()), reflect.ValueOf(nv))
			}
		}
		return m, nil
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).PkgPath == "" {
				// Only attempt to marshal structs with exported fields
				m, err := marshal(v)
				if err != nil {
					return reflect.Value{}, err
				}
				return reflect.ValueOf(m), nil
			}
		}
		return v, nil
	default:
		return v, nil
	}
}

// marhshal converts the given struct s to a map[string]reflect.Value
func marshal(v reflect.Value) (m map[string]reflect.Value, err error) {
	if v.Type().Implements(mapMarshalerType) {
		im, err := v.Interface().(MapMarshaler).MarshalMap()
		if err != nil {
			return nil, err
		}
		return ToValueMap(im), nil
	}

	v = reflect.Indirect(v)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil, errors.Errorf("Expected argument to be a struct, got %s (kind: %s)", t, t.Kind())
	}

	m = make(map[string]reflect.Value, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		m[f.Name], err = marshalNested(v.FieldByName(f.Name))
		if err != nil {
			return nil, err
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
		return mm.MarshalMap()
	}

	vm, err := marshal(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	if len(vm) == 0 {
		return nil, nil
	}

	vm = FlattenedValue(vm)

	m = make(map[string]interface{}, len(vm))
	for k, v := range vm {
		if isZero(v) {
			continue
		}
		switch v.Kind() {
		case reflect.Ptr, reflect.Struct, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func, reflect.Slice, reflect.Array:
			if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
				m[k] = v.Interface()
				continue
			}

			bv, err := ToBytesValue(v)
			if err != nil {
				return nil, err
			}
			m[k] = bv
		default:
			m[k] = v.Interface()
		}
	}
	return m, nil
}

// ToBytesValue is like ToBytes, but operates on values of type reflect.Value.
func ToBytesValue(v reflect.Value) (b []byte, err error) {
	var enc Encoding
	defer func() {
		if err != nil {
			return
		}
		b = append([]byte{byte(enc)}, b...)
	}()

	iv := reflect.Indirect(v)
	if !iv.IsValid() || isZero(iv) {
		enc = RawEncoding
		return []byte{}, nil
	}

	switch k := iv.Kind(); k {
	case reflect.String:
		enc = RawEncoding
		return []byte(iv.String()), nil
	case reflect.Bool:
		enc = RawEncoding
		return []byte(strconv.FormatBool(iv.Bool())), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		enc = RawEncoding
		return []byte(strconv.FormatInt(iv.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		enc = RawEncoding
		return []byte(strconv.FormatUint(iv.Uint(), 10)), nil
	case reflect.Float32:
		enc = RawEncoding
		return []byte(strconv.FormatFloat(iv.Float(), 'f', -1, 32)), nil
	case reflect.Float64:
		enc = RawEncoding
		return []byte(strconv.FormatFloat(iv.Float(), 'f', -1, 64)), nil
	case reflect.Slice, reflect.Array:
		elem := iv.Type().Elem()
		if elem.Kind() == reflect.Uint8 {
			enc = RawEncoding

			// Handle byte slices/arrays directly
			if k == reflect.Slice {
				return iv.Bytes(), nil
			}
			var byt byte
			out := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(byt)), iv.Len(), iv.Len())
			for i := 0; i < iv.Len(); i++ {
				out.Index(i).Set(iv.Index(i))
			}
			return out.Bytes(), nil
		}
	}

	t := v.Type()

outer:
	switch {
	case t.Implements(jsonMarshalerType):
		enc = JSONEncoding
		return v.Interface().(json.Marshaler).MarshalJSON()
	case t.Implements(protoMarshalerType):
		enc = ProtoEncoding
		return v.Interface().(proto.Marshaler).Marshal()
	case !t.Implements(gobGobEncoderType) && iv.Kind() == reflect.Struct:
		it := iv.Type()
		for i := 0; i < it.NumField(); i++ {
			if it.Field(i).PkgPath == "" {
				break outer
			}
		}
		// The struct can not be encoded using gob, if it does not implement gob.GobEncoder
		// and has no exported fields, hence we return an error
		return nil, errors.Errorf("Struct type %s should have exported fields or implement gob.GobEncoder to be encoded", t)
	case t.Kind() == reflect.Chan, t.Kind() == reflect.Func:
		return nil, errors.Errorf("Values of type %s (kind %s), which do not implement custom marshaling logic are not supported", t, t.Kind())
	}

	enc = GobEncoding

	// Encode the value as a pointer to include type info.
	pv := reflect.New(t)
	pv.Elem().Set(v)

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).EncodeValue(pv); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ToBytes marshals v into a []byte value and returns the result.
// Slices and arrays of bytes, strings, booleans and numeric types are stored in a human-readable
// format, if value implements proto.Marshaler, result of Marshal() method is stored, otherwise encoding/gob is used.
// Encoded values have the according Encoding byte prepended.
func ToBytes(v interface{}) (b []byte, err error) {
	return ToBytesValue(reflect.ValueOf(v))
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

	var vm map[string]reflect.Value
	if mm, ok := v.(MapMarshaler); ok {
		im, err := mm.MarshalMap()
		if err != nil {
			return nil, err
		}
		vm = ToValueMap(im)
	} else {
		vm, err = marshal(reflect.ValueOf(v))
		if err != nil {
			return nil, err
		}
		vm = FlattenedValue(vm)
	}
	if len(vm) == 0 {
		return nil, nil
	}

	bm = make(map[string][]byte, len(vm))
	for k, v := range vm {
		if isZero(v) {
			continue
		}

		b, err := ToBytesValue(v)
		if err != nil {
			return nil, err
		}
		bm[k] = b
	}
	return bm, nil
}
