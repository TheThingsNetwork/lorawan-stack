// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPairs(t *testing.T) {
	a := assertions.New(t)

	{
		n, err := pairsToMap("a", 10, "b", true)
		a.So(err, should.BeNil)
		a.So(n["a"], should.Equal, 10)
		a.So(n["b"], should.Equal, true)
	}

	{
		n, err := pairsToMap()
		a.So(err, should.BeNil)
		a.So(n, should.BeEmpty)
	}

	{
		_, err := pairsToMap("a")
		a.So(err, should.NotBeNil)
	}

	{
		n, err := pairsToMap(10, 20, true, "OK")
		a.So(err, should.BeNil)
		a.So(n["10"], should.Equal, 20)
		a.So(n["true"], should.Equal, "OK")
	}
}

func TestFields(t *testing.T) {
	a := assertions.New(t)

	f := Fields()
	g := f.WithField("a", 10)
	h := f.WithField("a", 20)

	a.So(f, should.NotEqual, g)
	a.So(f, should.NotEqual, h)
	a.So(g, should.NotEqual, h)

	got, ok := f.Get("a")
	a.So(ok, should.BeFalse)
	a.So(got, should.Equal, nil)

	got, ok = g.Get("a")
	a.So(ok, should.BeTrue)
	a.So(got, should.Equal, 10)

	got, ok = h.Get("a")
	a.So(ok, should.BeTrue)
	a.So(got, should.Equal, 20)

	i := g.WithField("b", 20)

	got, ok = g.Get("a")
	a.So(ok, should.BeTrue)
	a.So(got, should.Equal, 10)

	got, ok = i.Get("b")
	a.So(ok, should.BeTrue)
	a.So(got, should.Equal, 20)

	a.So(i.Fields(), should.Resemble, map[string]interface{}{
		"a": 10,
		"b": 20,
	})

	a.So(g.Fields(), should.Resemble, map[string]interface{}{
		"a": 10,
	})
}
