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

package store_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestBytesToType(t *testing.T) {
	for i, tc := range byteValues {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			if tc.value == nil {
				return
			}
			v, err := BytesToType(tc.bytes, reflect.TypeOf(tc.value))
			if a.So(err, should.BeNil) && !a.So(v, should.Resemble, tc.value) {
				pretty.Ldiff(t, tc.value, v)
			}
		})
	}
	t.Run("MsgPack", func(t *testing.T) {
		a := assertions.New(t)

		type subStruct struct {
			StringArray2 [2]string
			Time         time.Time
			ZeroTime     time.Time
			EmptyMap     map[string]interface{}
			NilMap       map[string]interface{}
			skip         interface{}
		}

		expected := struct {
			Int          int
			Int32        int32
			Int64        int64
			Float64      float64
			Uint16       uint16
			String       string
			Bytes        []byte
			Ints         []int
			MapStringInt map[string]int
			SubStruct    subStruct
		}{
			Int:     42,
			Int32:   42,
			Int64:   -42,
			Float64: 42,
			Uint16:  42,
			String:  "42",
			Bytes:   []byte("42"),
			Ints:    []int{4, 2},
			MapStringInt: map[string]int{
				"42": 42,
				"43": 43,
			},
			SubStruct: subStruct{
				StringArray2: [2]string{"foo", "bar"},
				Time:         time.Unix(42, 42),
				ZeroTime:     time.Time{},
				EmptyMap:     map[string]interface{}{},
				NilMap:       nil,
			},
		}

		v, err := BytesToType(append([]byte{byte(MsgPackEncoding)}, msgPackEncoded(expected)...), reflect.TypeOf(expected))
		if a.So(err, should.BeNil) && !a.So(v, should.Resemble, expected) {
			pretty.Ldiff(t, expected, v)
		}
	})
}

func TestUnflattened(t *testing.T) {
	for _, tc := range []struct {
		in  map[string]interface{}
		out map[string]interface{}
	}{
		{
			map[string]interface{}{
				"foo.bar":             os.Stdout,
				"foo.baz":             map[string]string{"test": "foo"},
				"foo.recursive.hello": struct{ hi string }{"hello"},
				"42.foo":              42,
				"42.baz":              "baz",
			},
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": os.Stdout,
					"baz": map[string]string{"test": "foo"},
					"recursive": map[string]interface{}{
						"hello": struct{ hi string }{"hello"},
					},
				},
				"42": map[string]interface{}{
					"foo": 42,
					"baz": "baz",
				},
			},
		},
	} {
		assertions.New(t).So(Unflattened(tc.in), should.Resemble, tc.out)
	}
}

func TestUnmarshalMap(t *testing.T) {
	for i, v := range structValues {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			rv := reflect.New(reflect.TypeOf(v.unmarshaled))

			switch v.unmarshaled.(type) {
			case map[string]interface{}, []interface{}, struct{ Interfaces []interface{} }:
				t.Skipf("Skipping special case, when unmarshaled value is %T as we don't know the type of values to unmarshal to", v.unmarshaled)
			}
			err := UnmarshalMap(v.marshaled, rv.Interface())
			if !a.So(err, should.BeNil) {
				t.Log(errors.Cause(err))
				return
			}
			if !a.So(rv.Elem().Interface(), should.Resemble, v.unmarshaled) {
				pretty.Ldiff(t, rv.Elem().Interface(), v.unmarshaled)
			}
		})
	}
}

func TestUnmarshalByteMap(t *testing.T) {
	for i, v := range structValues {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			rv := reflect.New(reflect.TypeOf(v.unmarshaled))

			switch v.unmarshaled.(type) {
			case map[string]interface{}, []interface{}, struct{ Interfaces []interface{} }:
				t.Skip(fmt.Sprintf("Skipping special case, when unmarshaled value is %T as we don't know the type of values to unmarshal to", v.unmarshaled))
			}
			err := UnmarshalByteMap(v.bytes, rv.Interface())
			if !a.So(err, should.BeNil) {
				t.Log(errors.Cause(err))
				return
			}
			if !a.So(rv.Elem().Interface(), should.Resemble, v.unmarshaled) {
				pretty.Ldiff(t, rv.Elem().Interface(), v.unmarshaled)
			}
		})
	}
}
