// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"reflect"
	"strconv"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestUnmarshalMap(t *testing.T) {
	for i, v := range values {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			rv := reflect.New(reflect.TypeOf(v.unmarshaled))

			if _, ok := v.unmarshaled.(map[string]interface{}); ok {
				t.Skip("Skipping special case, when unmarshaled value is map[string]interface{} as we don't know the type of values to unmarshal to")
			}
			err := UnmarshalMap(v.marshaled, rv.Interface())
			a.So(err, should.BeNil)
			if !a.So(rv.Elem().Interface(), should.Resemble, v.unmarshaled) {
				pretty.Ldiff(t, rv.Elem().Interface(), v.unmarshaled)
			}
		})
	}
}

func TestUnmarshalByteMap(t *testing.T) {
	for i, v := range values {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			rv := reflect.New(reflect.TypeOf(v.unmarshaled))

			if _, ok := v.unmarshaled.(map[string]interface{}); ok {
				t.Skip("Skipping special case when unmarshaled value is map[string]interface{} as we don't know the type of values to unmarshal to")
			}
			err := UnmarshalByteMap(v.bytes, rv.Interface())
			a.So(err, should.BeNil)
			if !a.So(rv.Elem().Interface(), should.Resemble, v.unmarshaled) {
				pretty.Ldiff(t, rv.Elem().Interface(), v.unmarshaled)
			}
		})
	}
}
