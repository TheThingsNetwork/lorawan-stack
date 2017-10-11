// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestArray(t *testing.T) {
	a := assertions.New(t)

	// empty int32
	{
		s := make([]int32, 0)
		array := Array(&s)
		value, err := array.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "{}")
		a.So(array.Scan("{1,2,3}"), should.BeNil)
		a.So(s, should.Resemble, []int32{1, 2, 3})
	}

	// int32 filled
	{
		s := []int32{3, 3, 3}
		array := Array(&s)
		value, err := array.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "{3,3,3}")
		a.So(array.Scan("{}"), should.BeNil)
		a.So(s, should.Resemble, []int32{})
	}

	type Foo int32

	// empty Foo slice
	{
		s := make([]Foo, 0)
		array := Array(&s)
		value, err := array.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "{}")
		a.So(array.Scan("{1,2,3}"), should.BeNil)
		a.So(s, should.Resemble, []Foo{1, 2, 3})
	}

	// Foo slice filled
	{
		s := []Foo{3, 3, 3}
		array := Array(&s)
		value, err := array.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "{3,3,3}")
		a.So(array.Scan("{}"), should.BeNil)
		a.So(s, should.Resemble, []Foo{})
	}
}
