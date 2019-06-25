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

package gogoproto_test

import (
	"bytes"
	"reflect"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/spf13/cast"
	. "go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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
	s, err := Struct(m)
	a.So(err, should.BeNil)
	sm, err := Map(s)
	a.So(err, should.BeNil)
	for k, v := range m {
		a.So(s.Fields, should.ContainKey, k)
		a.So(sm, should.ContainKey, k)
		if v == nil {
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &pbtypes.Value_NullValue{})
			a.So(sm[k], should.BeNil)
			continue
		}

		rv := reflect.Indirect(reflect.ValueOf(v))

		switch kind := rv.Kind(); kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:

			var vt float64
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &pbtypes.Value_NumberValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			a.So(sm[k], should.Equal, cast.ToFloat64(rv.Interface()))

		case reflect.Bool:
			var vt bool
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &pbtypes.Value_BoolValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			a.So(sm[k], should.Equal, rv.Bool())

		case reflect.String:
			var vt string
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &pbtypes.Value_StringValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			a.So(sm[k], should.Equal, rv.String())

		case reflect.Slice, reflect.Array:
			var vt []interface{}
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &pbtypes.Value_ListValue{})
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
			a.So(s.Fields[k].Kind, should.HaveSameTypeAs, &pbtypes.Value_StructValue{})
			a.So(sm[k], should.HaveSameTypeAs, vt)
			if kind == reflect.Map {
				a.So(sm[k], should.HaveLength, rv.Len())
			}

		default:
			panic("Unmatched kind: " + rv.Kind().String())
		}
		pv, err := Value(rv.Interface())
		if a.So(err, should.BeNil) {
			a.So(s.Fields[k], should.Resemble, pv)

			gv, err := Interface(pv)
			if a.So(err, should.BeNil) {
				a.So(sm[k], should.Resemble, gv)
			}
		}
	}
}
