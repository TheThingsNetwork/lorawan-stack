// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func ToBytes(v interface{}) ([]byte, error) {
	return toBytes(v)
}

func TestToBytes(t *testing.T) {
	a := assertions.New(t)
	for v, expected := range map[interface{}][]byte{
		int(42):     append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		int8(42):    append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		int16(42):   append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		int32(42):   append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		int64(42):   append([]byte{byte(RawEncoding)}, []byte(strconv.FormatInt(42, 10))...),
		uint(42):    append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		uint8(42):   append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		uint16(42):  append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		uint32(42):  append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		uint64(42):  append([]byte{byte(RawEncoding)}, []byte(strconv.FormatUint(42, 10))...),
		float32(42): append([]byte{byte(RawEncoding)}, []byte(strconv.FormatFloat(42, 'f', -1, 32))...),
		float64(42): append([]byte{byte(RawEncoding)}, []byte(strconv.FormatFloat(42, 'f', -1, 64))...),
		"42":        append([]byte{byte(RawEncoding)}, '4', '2'),
	} {
		rv := reflect.ValueOf(v)
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		t.Run(rv.Type().String(), func(t *testing.T) {
			got, err := toBytes(v)
			if a.So(err, should.BeNil) {
				a.So(got, should.Resemble, expected)
			}

			got, err = toBytes(ptr.Interface())
			if a.So(err, should.BeNil) {
				a.So(got, should.Resemble, expected)
			}
		})
	}
	t.Run(reflect.TypeOf([]byte{}).String(), func(t *testing.T) {
		b := []byte("42")
		got, err := toBytes(b)
		a.So(err, should.BeNil)
		a.So(got, should.Resemble, append([]byte{byte(RawEncoding)}, b...))

		got, err = toBytes(&b)
		a.So(err, should.BeNil)
		a.So(got, should.Resemble, append([]byte{byte(RawEncoding)}, b...))
	})
}
