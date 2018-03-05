// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

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
