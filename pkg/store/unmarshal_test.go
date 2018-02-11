// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"fmt"
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

			switch v.unmarshaled.(type) {
			case map[string]interface{}, []interface{}, struct{ Interfaces []interface{} }:
				t.Skipf("Skipping special case, when unmarshaled value is %T as we don't know the type of values to unmarshal to", v.unmarshaled)
			}
			err := UnmarshalMap(v.marshaled, rv.Interface(), v.decodeHooks...)
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

			switch v.unmarshaled.(type) {
			case map[string]interface{}, []interface{}, struct{ Interfaces []interface{} }:
				t.Skip(fmt.Sprintf("Skipping special case, when unmarshaled value is %T as we don't know the type of values to unmarshal to", v.unmarshaled))
			}
			err := UnmarshalByteMap(v.bytes, rv.Interface(), v.decodeHooks...)
			a.So(err, should.BeNil)
			if !a.So(rv.Elem().Interface(), should.Resemble, v.unmarshaled) {
				pretty.Ldiff(t, rv.Elem().Interface(), v.unmarshaled)
			}
		})
	}
}
