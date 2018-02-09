// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestToBytes(t *testing.T) {
	for i, tc := range []struct {
		v        interface{}
		expected []byte
	}{
		{
			int(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		},
		{
			int8(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		},
		{
			int16(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		},
		{
			int32(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		},
		{
			int64(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		},
		{
			uint(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		},
		{
			uint8(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		},
		{
			uint16(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		},
		{
			uint32(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		},
		{
			uint64(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		},
		{
			float32(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatFloat(42, 'f', -1, 32))...),
		},
		{
			float64(42),
			append([]byte{byte(RawEncoding)}, []byte(strconv.FormatFloat(42, 'f', -1, 64))...),
		},
		{
			[]byte("42"),
			append([]byte{byte(RawEncoding)}, '4', '2'),
		},
		{
			"42",
			append([]byte{byte(RawEncoding)}, '4', '2'),
		},
		{
			nil,
			append([]byte{byte(RawEncoding)}),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			got, err := ToBytes(tc.v)
			if a.So(err, should.BeNil) {
				a.So(got, should.Resemble, tc.expected)
			}

			rv := reflect.ValueOf(tc.v)
			if !rv.IsValid() {
				return
			}

			ptr := reflect.New(rv.Type())
			ptr.Elem().Set(rv)

			got, err = ToBytes(ptr.Interface())
			if a.So(err, should.BeNil) {
				a.So(got, should.Resemble, tc.expected)
			}
		})
	}
}

func TestFlattened(t *testing.T) {
	for _, tc := range []struct {
		in  map[string]interface{}
		out map[string]interface{}
	}{
		{
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
			map[string]interface{}{
				"foo.bar":             os.Stdout,
				"foo.baz":             map[string]string{"test": "foo"},
				"foo.recursive.hello": struct{ hi string }{"hello"},
				"42.foo":              42,
				"42.baz":              "baz",
			},
		},
	} {
		assertions.New(t).So(Flattened(tc.in), should.Resemble, tc.out)
	}
}

func TestIsZero(t *testing.T) {
	for i, tc := range []struct {
		v      interface{}
		isZero bool
	}{
		{
			(*time.Time)(nil),
			true,
		},
		{
			time.Time{},
			true,
		},
		{
			&time.Time{},
			true,
		},
		{
			[]int{},
			true,
		},
		{
			[]int{0},
			false,
		},
		{
			[]interface{}{nil, nil, nil},
			false,
		},
		{
			map[string]interface{}{
				"empty": struct{}{},
				"map":   nil,
			},
			false,
		},
		{
			map[string]interface{}{
				"nonempty": struct{ A int }{42},
				"map":      nil,
			},
			false,
		},
		{
			([]int)(nil),
			true,
		},
		{
			nil,
			true,
		},
		{
			(interface{})(nil),
			true,
		},
		{
			struct{ a int }{42},
			false,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assertions.New(t).So(isZero(reflect.ValueOf(tc.v)), should.Equal, tc.isZero)
		})
	}
}

func TestMapify(t *testing.T) {
	for i, tc := range []struct {
		input  interface{}
		keep   func(rv reflect.Value) bool
		output interface{}
	}{
		{
			nil,
			nil,
			nil,
		},
		{
			[]int{1, 2, 3, 0, 5},
			nil,
			map[string]interface{}{
				"0": 1,
				"1": 2,
				"2": 3,
				"3": 0,
				"4": 5,
			},
		},
		{
			[]interface{}{
				nil,
				nil,
				struct{}{},
				time.Time{},
				"hello",
				(*time.Time)(nil),
				(*struct{})(nil),
				[]interface{}{
					1,
					2,
					[]interface{}{},
				},
			},
			nil,
			map[string]interface{}{
				"0": nil,
				"1": nil,
				"2": struct{}{},
				"3": time.Time{},
				"4": "hello",
				"5": (*time.Time)(nil),
				"6": (*struct{})(nil),
				"7": map[string]interface{}{
					"0": 1,
					"1": 2,
					"2": map[string]interface{}{},
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assertions.New(t).So(mapify(tc.input, tc.keep), should.Resemble, tc.output)
		})
	}
}
