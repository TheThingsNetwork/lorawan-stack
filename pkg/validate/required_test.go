// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"reflect"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var isZeroCases = []struct {
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
		"",
		true,
	},
	{
		"42",
		false,
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
	{
		unsafe.Pointer(nil),
		true,
	},
	{
		unsafe.Pointer(&([]byte{42})[0]),
		false,
	},
}

func TestIsZero(t *testing.T) {
	for i, tc := range isZeroCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assertions.New(t).So(isZero(tc.v), should.Equal, tc.isZero)
		})
	}
}

func TestRequired(t *testing.T) {
	a := assertions.New(t)

	a.So(Field("", Required), should.NotBeNil)
	a.So(Field("f", Required), should.BeNil)

	a.So(Field("", NotRequired), should.BeNil)
	a.So(Field("f", NotRequired), should.BeNil)
}

func BenchmarkIsZero(b *testing.B) {
	for i, tc := range isZeroCases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				isZero(tc.v)
			}
		})
	}
}

func BenchmarkIsZeroValue(b *testing.B) {
	for i, tc := range isZeroCases {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				isZeroValue(reflect.ValueOf(tc.v))
			}
		})
	}
}
