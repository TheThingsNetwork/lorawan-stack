// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goproto

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/ptypes/struct"
)

// Map returns the Struct proto as a map[string]interface{}.
func Map(p *structpb.Struct) (map[string]interface{}, error) {
	if p == nil || len(p.Fields) == 0 {
		return nil, nil
	}
	m := make(map[string]interface{}, len(p.Fields))
	for k, v := range p.Fields {
		if v == nil {
			continue
		}
		gv, err := Interface(v)
		if err != nil {
			return nil, err
		}
		m[k] = gv
	}
	return m, nil
}

// Slice returns the ListValue proto as a []interface{}.
func Slice(l *structpb.ListValue) ([]interface{}, error) {
	if l == nil || len(l.Values) == 0 {
		return nil, nil
	}
	s := make([]interface{}, len(l.Values))
	for i, v := range l.Values {
		gv, err := Interface(v)
		if err != nil {
			return nil, err
		}
		s[i] = gv
	}
	return s, nil
}

// Interface returns the Value proto as an interface{}.
func Interface(v *structpb.Value) (interface{}, error) {
	switch v := v.GetKind().(type) {
	case *structpb.Value_NullValue:
		return nil, nil
	case *structpb.Value_NumberValue:
		return v.NumberValue, nil
	case *structpb.Value_StringValue:
		return v.StringValue, nil
	case *structpb.Value_BoolValue:
		return v.BoolValue, nil
	case *structpb.Value_StructValue:
		return Map(v.StructValue)
	case *structpb.Value_ListValue:
		return Slice(v.ListValue)
	default:
		return nil, fmt.Errorf("unmatched structpb type: %T", v)
	}
}

// Struct returns the map as a Struct proto.
func Struct(m map[string]interface{}) (*structpb.Struct, error) {
	p := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}
	for k, v := range m {
		pv, err := Value(v)
		if err != nil {
			return nil, err
		}
		p.Fields[k] = pv
	}
	return p, nil
}

// List returns the slice as a ListValue proto.
func List(s []interface{}) (*structpb.ListValue, error) {
	l := &structpb.ListValue{
		Values: make([]*structpb.Value, len(s)),
	}
	for i, v := range s {
		pv, err := Value(v)
		if err != nil {
			return nil, err
		}
		l.Values[i] = pv
	}
	return l, nil
}

func valueFromReflect(rv reflect.Value) (*structpb.Value, error) {
	switch k := rv.Kind(); k {
	case reflect.Ptr:
		if rv.IsNil() {
			return &structpb.Value{Kind: &structpb.Value_NullValue{}}, nil
		}
		return valueFromReflect(rv.Elem())
	case reflect.String:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: rv.String()}}, nil

	case reflect.Bool:
		return &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: rv.Bool()}}, nil

	case reflect.Slice, reflect.Array:
		if k == reflect.Slice && rv.IsNil() {
			return &structpb.Value{Kind: &structpb.Value_NullValue{}}, nil
		}
		s := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			s[i] = rv.Index(i).Interface()
		}
		pv, err := List(s)
		if err != nil {
			return nil, err
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: pv}}, nil

	case reflect.Map:
		if rv.IsNil() {
			return &structpb.Value{Kind: &structpb.Value_NullValue{}}, nil
		}
		m := make(map[string]interface{}, rv.Len())
		for _, key := range rv.MapKeys() {
			m[fmt.Sprint(key.Interface())] = rv.MapIndex(key).Interface()
		}
		pv, err := Struct(m)
		if err != nil {
			return nil, err
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: pv}}, nil

	case reflect.Struct:
		n := rv.NumField()
		fields := make(map[string]*structpb.Value, n)
		for i := 0; i < n; i++ {
			f := rv.Field(i)
			ft := f.Type()
			if f.Type().PkgPath() != "" {
				continue
			}
			pv, err := valueFromReflect(f)
			if err != nil {
				return nil, err
			}
			fields[ft.Name()] = pv
		}
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: fields}}}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(rv.Int())}}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: float64(rv.Uint())}}, nil

	case reflect.Float32, reflect.Float64:
		return &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: rv.Float()}}, nil

	case reflect.Complex64, reflect.Complex128:
		return &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: fmt.Sprint(rv.Complex())}}, nil

	default:
		// either Invalid, Chan, Func, Interface or UnsafePointer.
		return nil, fmt.Errorf("can not map a value of kind %s to a *structpb.Value", k)
	}
}

// Value returns the value as a Value proto.
func Value(v interface{}) (*structpb.Value, error) {
	if v == nil {
		return &structpb.Value{Kind: &structpb.Value_NullValue{}}, nil
	}
	pv, err := valueFromReflect(reflect.Indirect(reflect.ValueOf(v)))
	if err != nil {
		return nil, err
	}
	return pv, nil
}
