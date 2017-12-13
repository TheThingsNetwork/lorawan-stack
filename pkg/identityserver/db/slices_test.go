// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSlices(t *testing.T) {
	a := assertions.New(t)

	// empty int32 slice
	{
		s := Int32Slice{}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[]")
		a.So(s.Scan("[]"), should.BeNil)
		a.So(s, should.Resemble, s)

	}

	// filled int32 slice
	{
		s := Int32Slice{3, 3, 3}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[3,3,3]")
		a.So(s.Scan("[3,3,3]"), should.BeNil)
		a.So(s, should.Resemble, s)
	}

	type Foo int32

	// filled int32-like slice
	{
		s, err := NewInt32Slice([]Foo{1, 2, 3})
		a.So(err, should.BeNil)
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[1,2,3]")
		a.So(s.Scan("[1,2,3]"), should.BeNil)
		a.So(s, should.Resemble, s)

		dest := make([]Foo, 0)
		a.So(s.SetInto(&dest), should.BeNil)
		a.So(dest, should.Resemble, []Foo{1, 2, 3})
	}

	// empty string slice
	{
		s := StringSlice{}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[]")
		a.So(s.Scan("[]"), should.BeNil)
		a.So(s, should.Resemble, s)

	}

	// filled string slice
	{
		s := StringSlice{"a", "b", "c"}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, `["a","b","c"]`)
		a.So(s.Scan(`["a"]`), should.BeNil)
		a.So(s, should.Resemble, StringSlice{"a"})
	}

	type FooStr string

	// filled string-like slice
	{
		s, err := NewStringSlice([]FooStr{"bar"})
		a.So(err, should.BeNil)
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, `["bar"]`)
		a.So(s.Scan(`["qux"]`), should.BeNil)
		a.So(s, should.Resemble, StringSlice{"qux"})

		dest := make([]FooStr, 0)
		a.So(s.SetInto(&dest), should.BeNil)
		a.So(dest, should.Resemble, []FooStr{"qux"})
	}
}
