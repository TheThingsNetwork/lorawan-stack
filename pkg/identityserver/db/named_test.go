// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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

	// into a slice of ptr to struct
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

	// into a slice of struct
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

	// into a slice of maps
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

	// into a slice of ptr to maps
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

	// into a slice of values
	{
		res := make([]string, 0)
		err := q.NamedSelect(&res, `SELECT bar FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0], should.Equal, data[1].Bar)
	}

	// into a slice of ptr to values
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

	// into struct ptr
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

	// into map
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

	// into ptr to map
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

	// into value
	{
		res := ""
		err := q.NamedSelectOne(&res, `SELECT bar FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(res, should.Equal, data[1].Bar)
	}

	// into ptr to value
	{
		res := new(string)
		err := q.NamedSelectOne(&res, `SELECT bar FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.BeNil)
		a.So(*res, should.Equal, data[1].Bar)
	}

	// cannot use struct directly
	{
		res := foo{}
		err := q.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.NotBeNil)
	}

	// cannot use value directly
	{
		res := ""
		err := q.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
		})
		a.So(err, should.NotBeNil)
	}
}
