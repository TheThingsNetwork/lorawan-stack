// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package goproto

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/ptypes/struct"
)

// MapFromProto returns the Struct proto as a map[string]interface{}
func MapFromProto(p *structpb.Struct) (m map[string]interface{}) {
	m = make(map[string]interface{})
	for k, v := range p.Fields {
		if v == nil {
			continue
		}
		m[k] = ValueFromProto(v)
	}
	return
}

// ValueFromProto returns the Value proto as an interface{}
func ValueFromProto(v *structpb.Value) (i interface{}) {
	switch v := v.GetKind().(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_NumberValue:
		return v.NumberValue
	case *structpb.Value_StringValue:
		return v.StringValue
	case *structpb.Value_BoolValue:
		return v.BoolValue
	case *structpb.Value_StructValue:
		return MapFromProto(v.StructValue)
	case *structpb.Value_ListValue:
		return ListFromProto(v.ListValue)
	}
	return v.String()
}

// ListFromProto returns the ListValue proto as a []interface{}
func ListFromProto(l *structpb.ListValue) (s []interface{}) {
	s = make([]interface{}, len(l.Values))
	for i, v := range l.Values {
		s[i] = ValueFromProto(v)
	}
	return
}

// MapProto returns the map as a Struct proto
func MapProto(m map[string]interface{}) (p *structpb.Struct) {
	p = &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}
	for k, v := range m {
		p.Fields[k] = ValueProto(v)
	}
	return
}

// ValueProto returns the value as a Value proto
func ValueProto(i interface{}) (v *structpb.Value) {
	v = &structpb.Value{}
	if i == nil {
		v.Kind = &structpb.Value_NullValue{}
		return
	}
	rVal := reflect.ValueOf(i)
	if rVal.Type().Kind() == reflect.Ptr {
		rVal = rVal.Elem()
		i = rVal.Interface()
	}
	switch val := i.(type) {
	case bool:
		v.Kind = &structpb.Value_BoolValue{BoolValue: val}
		return
	case int, int8, int16, int32, int64:
		v.Kind = &structpb.Value_NumberValue{NumberValue: float64(rVal.Int())}
		return
	case uint, uint8, uint16, uint32, uint64:
		v.Kind = &structpb.Value_NumberValue{NumberValue: float64(rVal.Uint())}
		return
	case float32, float64:
		v.Kind = &structpb.Value_NumberValue{NumberValue: float64(rVal.Float())}
		return
	case string:
		v.Kind = &structpb.Value_StringValue{StringValue: val}
		return
	default:
		switch rVal.Type().Kind() {
		case reflect.Slice:
			s := make([]interface{}, rVal.Len())
			for i := 0; i < rVal.Len(); i++ {
				s[i] = rVal.Index(i).Interface()
			}
			v.Kind = &structpb.Value_ListValue{ListValue: ListProto(s)}
			return
		case reflect.Map:
			m := make(map[string]interface{})
			for _, key := range rVal.MapKeys() {
				m[fmt.Sprint(key.Interface())] = rVal.MapIndex(key).Interface()
			}
			v.Kind = &structpb.Value_StructValue{StructValue: MapProto(m)}
			return
		}
	}
	v.Kind = &structpb.Value_StringValue{StringValue: fmt.Sprint(i)}
	return
}

// ListProto returns the slice as a ListValue proto
func ListProto(s []interface{}) (l *structpb.ListValue) {
	l = &structpb.ListValue{
		Values: make([]*structpb.Value, len(s)),
	}
	for i, v := range s {
		l.Values[i] = ValueProto(v)
	}
	return
}
