// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPairs(t *testing.T) {
	a := assertions.New(t)

	{
		n := pairsToMap("a", 10, "b", true)
		a.So(n["a"], should.Equal, 10)
		a.So(n["b"], should.Equal, true)
	}

	{
		n := pairsToMap()
		a.So(n, should.BeEmpty)
	}

	{
		a.So(func() { pairsToMap("a") }, should.Panic)
	}

	{
		n := pairsToMap(10, 20, true, "OK")
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
