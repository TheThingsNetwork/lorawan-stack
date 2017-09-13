// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package goproto

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/ptypes/struct"
	"github.com/spf13/cast"
)

// Map returns the Struct proto as a map[string]interface{}
func Map(p *structpb.Struct) map[string]interface{} {
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
func Slice(l *structpb.ListValue) []interface{} {
	s := make([]interface{}, len(l.Values))
	for i, v := range l.Values {
		s[i] = Interface(v)
	}
	return s
}

// Interface returns the Value proto as an interface{}
func Interface(v *structpb.Value) interface{} {
	switch v := v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return v.NumberValue
	case *structpb.Value_StringValue:
		return v.StringValue
	case *structpb.Value_BoolValue:
		return v.BoolValue
	case *structpb.Value_StructValue:
		return Map(v.StructValue)
	case *structpb.Value_ListValue:
		return Slice(v.ListValue)
	}
	return nil
}

// Struct returns the map as a Struct proto
func Struct(m map[string]interface{}) *structpb.Struct {
	p := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}
	for k, v := range m {
		p.Fields[k] = Value(v)
	}
	return p
}

// List returns the slice as a ListValue proto
func List(s []interface{}) *structpb.ListValue {
	l := &structpb.ListValue{
		Values: make([]*structpb.Value, len(s)),
	}
	for i, v := range s {
		l.Values[i] = Value(v)
	}
	return l
}

func valueFromReflect(rv reflect.Value) *structpb.Value {
	switch k := rv.Kind(); k {
	case reflect.Ptr:
		if rv.IsNil() {
			return &structpb.Value{Kind: &structpb.Value_NullValue{}}
		}
		return valueFromReflect(rv.Elem())
	case reflect.String:
		return &structpb.Value{Kind: &structpb.Value_StringValue{rv.String()}}

	case reflect.Bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{rv.Bool()}}

	case reflect.Slice, reflect.Array:
		if k == reflect.Slice && rv.IsNil() {
			return &structpb.Value{Kind: &structpb.Value_NullValue{}}
		}
		s := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			s[i] = rv.Index(i).Interface()
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: List(s)}}

	case reflect.Map:
		if rv.IsNil() {
			return &structpb.Value{Kind: &structpb.Value_NullValue{}}
		}
		m := make(map[string]interface{}, rv.Len())
		for _, key := range rv.MapKeys() {
			m[fmt.Sprint(key.Interface())] = rv.MapIndex(key).Interface()
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{Struct(m)}}

	case reflect.Struct:
		n := rv.NumField()
		fields := make(map[string]*structpb.Value, n)
		for i := 0; i < n; i++ {
			f := rv.Field(i)
			ft := f.Type()
			if f.Type().PkgPath() != "" {
				continue
			}
			fields[ft.Name()] = valueFromReflect(f)
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{&structpb.Struct{fields}}}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:

		return &structpb.Value{Kind: &structpb.Value_NumberValue{cast.ToFloat64(rv.Interface())}}
	}
	return &structpb.Value{Kind: &structpb.Value_NullValue{}}
}

// Value returns the value as a Value proto
func Value(v interface{}) *structpb.Value {
	if v == nil {
		return &structpb.Value{Kind: &structpb.Value_NullValue{}}
	}
	return valueFromReflect(reflect.Indirect(reflect.ValueOf(v)))
}
