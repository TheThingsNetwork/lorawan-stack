// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package goproto

import (
	"fmt"
	"reflect"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cast"
)

// Map returns the Struct proto as a map[string]interface{}
func Map(p *pbtypes.Struct) map[string]interface{} {
	m := make(map[string]interface{}, len(p.Fields))
	for k, v := range p.Fields {
		if v == nil {
			continue
		}
		m[k] = Interface(v)
	}
	return m
}

// Slice returns the ListValue proto as a []interface{}
func Slice(l *pbtypes.ListValue) []interface{} {
	s := make([]interface{}, len(l.Values))
	for i, v := range l.Values {
		s[i] = Interface(v)
	}
	return s
}

// Interface returns the Value proto as an interface{}
func Interface(v *pbtypes.Value) interface{} {
	switch v := v.Kind.(type) {
	case *pbtypes.Value_NullValue:
		return nil
	case *pbtypes.Value_NumberValue:
		return v.NumberValue
	case *pbtypes.Value_StringValue:
		return v.StringValue
	case *pbtypes.Value_BoolValue:
		return v.BoolValue
	case *pbtypes.Value_StructValue:
		return Map(v.StructValue)
	case *pbtypes.Value_ListValue:
		return Slice(v.ListValue)
	}
	return nil
}

// Struct returns the map as a Struct proto
func Struct(m map[string]interface{}) *pbtypes.Struct {
	p := &pbtypes.Struct{
		Fields: make(map[string]*pbtypes.Value),
	}
	for k, v := range m {
		p.Fields[k] = Value(v)
	}
	return p
}

// List returns the slice as a ListValue proto
func List(s []interface{}) *pbtypes.ListValue {
	l := &pbtypes.ListValue{
		Values: make([]*pbtypes.Value, len(s)),
	}
	for i, v := range s {
		l.Values[i] = Value(v)
	}
	return l
}

func valueFromReflect(rv reflect.Value) *pbtypes.Value {
	switch k := rv.Kind(); k {
	case reflect.Ptr:
		if rv.IsNil() {
			return &pbtypes.Value{Kind: &pbtypes.Value_NullValue{}}
		}
		return valueFromReflect(rv.Elem())
	case reflect.String:
		return &pbtypes.Value{Kind: &pbtypes.Value_StringValue{rv.String()}}

	case reflect.Bool:
		return &pbtypes.Value{Kind: &pbtypes.Value_BoolValue{rv.Bool()}}

	case reflect.Slice, reflect.Array:
		if k == reflect.Slice && rv.IsNil() {
			return &pbtypes.Value{Kind: &pbtypes.Value_NullValue{}}
		}
		s := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			s[i] = rv.Index(i).Interface()
		}
		return &pbtypes.Value{Kind: &pbtypes.Value_ListValue{ListValue: List(s)}}

	case reflect.Map:
		if rv.IsNil() {
			return &pbtypes.Value{Kind: &pbtypes.Value_NullValue{}}
		}
		m := make(map[string]interface{}, rv.Len())
		for _, key := range rv.MapKeys() {
			m[fmt.Sprint(key.Interface())] = rv.MapIndex(key).Interface()
		}
		return &pbtypes.Value{Kind: &pbtypes.Value_StructValue{Struct(m)}}

	case reflect.Struct:
		n := rv.NumField()
		fields := make(map[string]*pbtypes.Value, n)
		for i := 0; i < n; i++ {
			f := rv.Field(i)
			ft := f.Type()
			if f.Type().PkgPath() != "" {
				continue
			}
			fields[ft.Name()] = valueFromReflect(f)
		}
		return &pbtypes.Value{Kind: &pbtypes.Value_StructValue{&pbtypes.Struct{fields}}}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:

		return &pbtypes.Value{Kind: &pbtypes.Value_NumberValue{cast.ToFloat64(rv.Interface())}}
	}
	return &pbtypes.Value{Kind: &pbtypes.Value_NullValue{}}
}

// Value returns the value as a Value proto
func Value(v interface{}) *pbtypes.Value {
	if v == nil {
		return &pbtypes.Value{Kind: &pbtypes.Value_NullValue{}}
	}
	return valueFromReflect(reflect.Indirect(reflect.ValueOf(v)))
}
