// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestPairs(t *testing.T) {
	a := New(t)

	{
		n, err := pairsToMap("a", 10, "b", true)
		a.So(err, ShouldBeNil)
		a.So(n["a"], ShouldEqual, 10)
		a.So(n["b"], ShouldEqual, true)
	}

	{
		n, err := pairsToMap()
		a.So(err, ShouldBeNil)
		a.So(n, ShouldBeEmpty)
	}

	{
		_, err := pairsToMap("a")
		a.So(err, ShouldNotBeNil)
	}

	{
		n, err := pairsToMap(10, 20, true, "OK")
		a.So(err, ShouldBeNil)
		a.So(n["10"], ShouldEqual, 20)
		a.So(n["true"], ShouldEqual, "OK")
	}
}

func TestFields(t *testing.T) {
	a := New(t)

	f := Fields()
	g := f.Set("a", 10)
	h := f.Set("a", 20)

	a.So(f, ShouldNotEqual, g)
	a.So(f, ShouldNotEqual, h)
	a.So(g, ShouldNotEqual, h)

	got, ok := f.Get("a")
	a.So(ok, ShouldBeFalse)
	a.So(got, ShouldEqual, nil)

	got, ok = g.Get("a")
	a.So(ok, ShouldBeTrue)
	a.So(got, ShouldEqual, 10)

	got, ok = h.Get("a")
	a.So(ok, ShouldBeTrue)
	a.So(got, ShouldEqual, 20)

	i := g.Set("b", 20)

	got, ok = g.Get("a")
	a.So(ok, ShouldBeTrue)
	a.So(got, ShouldEqual, 10)

	got, ok = i.Get("b")
	a.So(ok, ShouldBeTrue)
	a.So(got, ShouldEqual, 20)

	a.So(i.Fields(), ShouldResemble, map[string]interface{}{
		"a": 10,
		"b": 20,
	})

	a.So(g.Fields(), ShouldResemble, map[string]interface{}{
		"a": 10,
	})
}
