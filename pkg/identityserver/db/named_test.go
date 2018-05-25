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

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func testNamedExec(t *testing.T, q QueryContext) {
	a := assertions.New(t)

	id := int64(1234)

	_, err := q.NamedExec(`INSERT INTO foo (id, bar) VALUES (:id, :bar)`, foo{
		ID:  id,
		Bar: "bar",
	})
	a.So(err, should.BeNil)

	_, err = q.NamedExec(`DELETE FROM foo WHERE id = :id`, map[string]interface{}{
		"id": id,
	})
	a.So(err, should.BeNil)
}

func testNamedSelect(t *testing.T, q QueryContext) {
	a := assertions.New(t)

	{
		res := make([]*foo, 0)
		err := q.NamedSelect(&res, "SELECT * FROM foo", map[string]interface{}{})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, len(data))
	}

	// Into a slice of ptr to struct.
	{
		res := make([]*foo, 0)
		err := q.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0].ID, should.NotBeNil)
		a.So(res[0].Created, should.NotBeNil)
		a.So(res[0].Bar, should.Equal, data[1].Bar)
		a.So(res[0].Baz, should.Equal, data[1].Baz)
		a.So(res[0].Quu, should.Equal, data[1].Quu)
	}

	// Into a slice of struct.
	{
		res := make([]foo, 0)
		err := q.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0].ID, should.NotBeNil)
		a.So(res[0].Created, should.NotBeNil)
		a.So(res[0].Bar, should.Equal, data[1].Bar)
		a.So(res[0].Baz, should.Equal, data[1].Baz)
		a.So(res[0].Quu, should.Equal, data[1].Quu)
	}

	// Into a slice of maps.
	{
		res := make([]map[string]interface{}, 0)
		err := q.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0]["id"], should.NotBeNil)
		a.So(res[0]["created"], should.NotBeNil)
		a.So(res[0]["bar"], should.Equal, data[1].Bar)
		a.So(res[0]["baz"], should.Equal, data[1].Baz)
		a.So(res[0]["quu"], should.Equal, data[1].Quu)
	}

	// Into a slice of ptr to maps.
	{
		res := make([]*map[string]interface{}, 0)
		err := q.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		r := *res[0]
		a.So(r["id"], should.NotBeNil)
		a.So(r["created"], should.NotBeNil)
		a.So(r["bar"], should.Equal, data[1].Bar)
		a.So(r["baz"], should.Equal, data[1].Baz)
		a.So(r["quu"], should.Equal, data[1].Quu)
	}

	// Into a slice of values.
	{
		res := make([]string, 0)
		err := q.NamedSelect(&res, `SELECT bar FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0], should.Equal, data[1].Bar)
	}

	// Into a slice of ptr to values.
	{
		res := make([]*string, 0)
		err := q.NamedSelect(&res, `SELECT bar FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(*res[0], should.Equal, data[1].Bar)
	}
}

func testNamedSelectOne(t *testing.T, q QueryContext) {
	a := assertions.New(t)

	// Into struct ptr.
	{
		res := new(foo)
		err := q.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res.ID, should.NotBeNil)
		a.So(res.Created, should.NotBeNil)
		a.So(res.Bar, should.Equal, data[1].Bar)
		a.So(res.Baz, should.Equal, data[1].Baz)
		a.So(res.Quu, should.Equal, data[1].Quu)
	}

	// Into map.
	{
		res := make(map[string]interface{})
		err := q.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res["id"], should.NotBeNil)
		a.So(res["created"], should.NotBeNil)
		a.So(res["bar"], should.Equal, data[1].Bar)
		a.So(res["baz"], should.Equal, data[1].Baz)
		a.So(res["quu"], should.Equal, data[1].Quu)
	}

	// Into ptr to map.
	{
		res := make(map[string]interface{})
		err := q.NamedSelectOne(&res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.NotBeNil)
		a.So(res["id"], should.NotBeNil)
		a.So(res["created"], should.NotBeNil)
		a.So(res["bar"], should.Equal, data[1].Bar)
		a.So(res["baz"], should.Equal, data[1].Baz)
		a.So(res["quu"], should.Equal, data[1].Quu)
	}

	// Into value.
	{
		res := ""
		err := q.NamedSelectOne(&res, `SELECT bar FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.Equal, data[1].Bar)
	}

	// Into ptr to value.
	{
		res := new(string)
		err := q.NamedSelectOne(&res, `SELECT bar FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(*res, should.Equal, data[1].Bar)
	}

	// Cannot use struct directly.
	{
		res := foo{}
		err := q.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.NotBeNil)
	}

	// Cannot use value directly.
	{
		res := ""
		err := q.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.NotBeNil)
	}
}
