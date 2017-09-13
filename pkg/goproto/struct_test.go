// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package goproto

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/golang/protobuf/ptypes/struct"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"github.com/spf13/cast"
)

type jsonMarshaler struct {
	Text string
}

func (m jsonMarshaler) MarshalJSON() ([]byte, error) {
	return bytes.ToUpper([]byte(`"` + m.Text + `"`)), nil
}

func (m *jsonMarshaler) UnmarshalJSON(b []byte) error {
	m.Text = string(bytes.ToLower(bytes.Trim(b, `"`)))
	return nil
}

func TestStructProto(t *testing.T) {
	a := assertions.New(t)

	ptr := "ptr"
	m := map[string]interface{}{
		"foo":            "bar",
		"ptr":            &ptr,
		"answer":         42,
		"answer.precise": 42.0,
		"works":          true,
		"empty":          nil,
		"list":           []string{"a", "b", "c"},
		"map":            map[string]string{"foo": "bar"},
		"eui":            types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		"jsonMarshaler":  &jsonMarshaler{Text: "testtext"},
	}
	s := Struct(m)
	sm := Map(s)
	for k, v := range m {
		a.So(s.Fields, should.ContainKey, k)
		a.So(sm, should.ContainKey, k)
		if v == nil {
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &structpb.Value_NullValue{})
			a.So(sm[k], should.BeNil)
			continue
		}

		rv := reflect.Indirect(reflect.ValueOf(v))

		switch kind := rv.Kind(); kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:

			var vt float64
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &structpb.Value_NumberValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			a.So(sm[k], should.Equal, cast.ToFloat64(rv.Interface()))

		case reflect.Bool:
			var vt bool
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &structpb.Value_BoolValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			a.So(sm[k], should.Equal, rv.Bool())

		case reflect.String:
			var vt string
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &structpb.Value_StringValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			a.So(sm[k], should.Equal, rv.String())

		case reflect.Slice, reflect.Array:
			var vt []interface{}
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &structpb.Value_ListValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			if a.So(sm[k], should.HaveLength, rv.Len()) {
				// TODO find a way to compare these values
				//smv := reflect.ValueOf(sm[k])
				//for i := 0; i < rv.Len(); i++ {
				//a.So(smv.Index(i).Interface(), should.Resemble, rv.Index(i).Interface())
				//}
			}

		case reflect.Struct, reflect.Map:
			var vt map[string]interface{}
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &structpb.Value_StructValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			if kind == reflect.Map {
				a.So(sm[k], should.HaveLength, rv.Len())
			}

		default:
			panic("Unmatched kind: " + rv.Kind().String())
		}
		a.So(s.Fields[k], should.Resemble, Value(rv.Interface()))
		a.So(sm[k], should.Resemble, Interface(Value(rv.Interface())))
	}
}
