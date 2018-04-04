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
	"github.com/smartystreets/assertions/should"
)

func testSelect(t *testing.T, q QueryContext) {
	a := assertions.New(t)

	{
		res := make([]*foo, 0)
		err := q.Select(&res, "SELECT * FROM foo")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, len(data))
	}

	// into a slice of struct ptr
	{
		res := make([]*foo, 0)
		err := q.Select(&res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
	}

	// into a slice of struct
	{
		res := make([]foo, 0)
		err := q.Select(&res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
	}

	// into a slice of values
	{
		res := make([]string, 0)
		err := q.Select(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0], should.Equal, data[1].Bar)
	}

	// cannot use struct directly
	{
		res := foo{}
		err := q.Select(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.NotBeNil)
	}

	// cannot use slice directly
	{
		res := make([]string, 0)
		err := q.Select(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.NotBeNil)
	}
}

func testSelectOne(t *testing.T, q QueryContext) {
	a := assertions.New(t)

	// into struct ptr
	{
		res := new(foo)
		err := q.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := q.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := q.SelectOne(&res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := q.SelectOne(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.Equal, data[1].Bar)
	}

	// into ptr to value
	{
		res := new(string)
		err := q.SelectOne(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(*res, should.Equal, data[1].Bar)
	}

	// cannot use struct directly
	{
		res := foo{}
		err := q.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.NotBeNil)
	}

	// cannot use value directly
	{
		res := ""
		err := q.SelectOne(res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.NotBeNil)
	}
}
