// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"bytes"
	"encoding/gob"
	"os"
	"reflect"
	"strconv"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func gobEncoded(v interface{}) []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(v); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

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
			[]interface{}{1, 2},
			append([]byte{byte(GobEncoding)}, gobEncoded([]interface{}{1, 2})...),
		},
		{
			nil,
			append([]byte{byte(RawEncoding)}),
		},
		{
			struct{ a, b int }{},
			append([]byte{byte(RawEncoding)}),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			b, err := ToBytes(tc.v)
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, tc.expected)
			}

			rv := reflect.ValueOf(tc.v)
			if !rv.IsValid() {
				return
			}

			ptr := reflect.New(rv.Type())
			ptr.Elem().Set(rv)

			b, err = ToBytes(ptr.Interface())
			if a.So(err, should.BeNil) {
				a.So(b, should.Resemble, tc.expected)
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

func TestFlattenedValue(t *testing.T) {
	for _, tc := range []struct {
		in  map[string]reflect.Value
		out map[string]reflect.Value
	}{
		{
			map[string]reflect.Value{
				"foo": reflect.ValueOf(map[string]reflect.Value{
					"bar": reflect.ValueOf(os.Stdout),
					"baz": reflect.ValueOf(map[string]string{"test": "foo"}),
					"recursive": reflect.ValueOf(map[string]reflect.Value{
						"hello": reflect.ValueOf(struct{ hi string }{"hello"}),
					}),
				}),
				"42": reflect.ValueOf(map[string]reflect.Value{
					"foo": reflect.ValueOf(42),
					"baz": reflect.ValueOf("baz"),
				}),
			},
			map[string]reflect.Value{
				"foo.bar":             reflect.ValueOf(os.Stdout),
				"foo.baz":             reflect.ValueOf(map[string]string{"test": "foo"}),
				"foo.recursive.hello": reflect.ValueOf(struct{ hi string }{"hello"}),
				"42.foo":              reflect.ValueOf(42),
				"42.baz":              reflect.ValueOf("baz"),
			},
		},
	} {
		a := assertions.New(t)
		ret := FlattenedValue(tc.in)
		if !a.So(ret, should.HaveLength, len(tc.out)) {
			return
		}
		for k, v := range tc.out {
			a.So(ret[k].Interface(), should.Resemble, v.Interface())
			a.So(ret[k].Type(), should.Equal, v.Type())
		}
	}
}

func TestMarshalMap(t *testing.T) {
	for i, v := range values {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			m, err := MarshalMap(v.unmarshaled)
			a.So(err, should.BeNil)
			if !a.So(m, should.Resemble, v.marshaled) {
				pretty.Ldiff(t, m, v.marshaled)
			}
		})
	}
}

func TestMarshalByteMap(t *testing.T) {
	for i, v := range values {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			m, err := MarshalByteMap(v.unmarshaled)
			a.So(err, should.BeNil)
			if !a.So(m, should.Resemble, v.bytes) {
				pretty.Ldiff(t, m, v.bytes)
			}
		})
	}
}
