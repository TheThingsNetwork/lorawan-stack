// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package gogoproto

import (
	"fmt"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gogo/protobuf/types"
)

// Map returns the Struct proto as a map[string]interface{}.
func Map(p *types.Struct) (map[string]interface{}, error) {
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
func Slice(l *types.ListValue) ([]interface{}, error) {
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
func Interface(v *types.Value) (interface{}, error) {
	switch v := v.GetKind().(type) {
	case *types.Value_NullValue:
		return nil, nil
	case *types.Value_NumberValue:
		return v.NumberValue, nil
	case *types.Value_StringValue:
		return v.StringValue, nil
	case *types.Value_BoolValue:
		return v.BoolValue, nil
	case *types.Value_StructValue:
		return Map(v.StructValue)
	case *types.Value_ListValue:
		return Slice(v.ListValue)
	default:
		return nil, errors.Errorf("unmatched types type: %T", v)
	}
}

// Struct returns the map as a Struct proto.
func Struct(m map[string]interface{}) (*types.Struct, error) {
	p := &types.Struct{
		Fields: make(map[string]*types.Value),
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
func List(s []interface{}) (*types.ListValue, error) {
	l := &types.ListValue{
		Values: make([]*types.Value, len(s)),
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

func valueFromReflect(rv reflect.Value) (*types.Value, error) {
	switch k := rv.Kind(); k {
	case reflect.Ptr:
		if rv.IsNil() {
			return &types.Value{Kind: &types.Value_NullValue{}}, nil
		}
		return valueFromReflect(rv.Elem())
	case reflect.String:
		return &types.Value{Kind: &types.Value_StringValue{rv.String()}}, nil

	case reflect.Bool:
		return &types.Value{Kind: &types.Value_BoolValue{rv.Bool()}}, nil

	case reflect.Slice, reflect.Array:
		if k == reflect.Slice && rv.IsNil() {
			return &types.Value{Kind: &types.Value_NullValue{}}, nil
		}
		s := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			s[i] = rv.Index(i).Interface()
		}
		pv, err := List(s)
		if err != nil {
			return nil, err
		}
		return &types.Value{Kind: &types.Value_ListValue{pv}}, nil

	case reflect.Map:
		if rv.IsNil() {
			return &types.Value{Kind: &types.Value_NullValue{}}, nil
		}
		m := make(map[string]interface{}, rv.Len())
		for _, key := range rv.MapKeys() {
			m[fmt.Sprint(key.Interface())] = rv.MapIndex(key).Interface()
		}
		pv, err := Struct(m)
		if err != nil {
			return nil, err
		}
		return &types.Value{Kind: &types.Value_StructValue{pv}}, nil

	case reflect.Struct:
		n := rv.NumField()
		fields := make(map[string]*types.Value, n)
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
		return &types.Value{Kind: &types.Value_StructValue{&types.Struct{fields}}}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &types.Value{Kind: &types.Value_NumberValue{float64(rv.Int())}}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &types.Value{Kind: &types.Value_NumberValue{float64(rv.Uint())}}, nil

	case reflect.Float32, reflect.Float64:
		return &types.Value{Kind: &types.Value_NumberValue{rv.Float()}}, nil

	case reflect.Complex64, reflect.Complex128:
		return &types.Value{Kind: &types.Value_StringValue{fmt.Sprint(rv.Complex())}}, nil

	default:
		// either Invalid, Chan. Func, Interface or UnsafePointer.
		return nil, errors.Errorf("Can not map a value of kind %s to a *types.Value", k)
	}
}

// Value returns the value as a Value proto
func Value(v interface{}) (*types.Value, error) {
	if v == nil {
		return &types.Value{Kind: &types.Value_NullValue{}}, nil
	}
	pv, err := valueFromReflect(reflect.Indirect(reflect.ValueOf(v)))
	if err != nil {
		return nil, err
	}
	return pv, nil
}
