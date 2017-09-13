// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSelectTx(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	// into a slice of struct ptr
	{
		res := make([]*foo, 0)
		err := db.Transact(func(tx *Tx) error {
			return tx.Select(&res, "SELECT * FROM foo")
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, len(data))
	}

	// into a slice of struct
	{
		res := make([]*foo, 0)
		err := db.Transact(func(tx *Tx) error {
			return tx.Select(&res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
	}

	// into a slice of values
	{
		res := make([]string, 0)
		err := db.Transact(func(tx *Tx) error {
			return tx.Select(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0], should.Equal, data[1].Bar)
	}
}
func TestNamedSelectTx(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	// into a slice of ptr to struct
	{
		res := make([]*foo, 0)
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
				Bar: "bar-2",
			})
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
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
				Bar: "bar-2",
			})
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
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
				Bar: "bar-2",
			})
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0]["id"], should.NotBeNil)
		a.So(res[0]["created"], should.NotBeNil)
		a.So(res[0]["bar"], should.Equal, data[1].Bar)
		a.So(res[0]["baz"], should.Equal, data[1].Baz)
		a.So(res[0]["quu"], should.Equal, data[1].Quu)
	}

	// into a slice of values
	{
		res := make([]string, 0)
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelect(&res, `SELECT bar FROM foo WHERE bar = :bar`, foo{
				Bar: "bar-2",
			})
		})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0], should.Equal, data[1].Bar)
	}
}
func TestSelectOneTx(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	// into struct ptr
	{
		res := new(foo)
		err := db.Transact(func(tx *Tx) error {
			return tx.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := db.Transact(func(tx *Tx) error {
			return tx.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := db.Transact(func(tx *Tx) error {
			return tx.SelectOne(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		})
		a.So(err, should.BeNil)
		a.So(res, should.Equal, data[1].Bar)
	}
}
func TestNamedSelectOneTx(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	// into struct ptr
	{
		res := new(foo)
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
				"bar": "bar-2",
			})
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
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
				"bar": "bar-2",
			})
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
		err := db.Transact(func(tx *Tx) error {
			return tx.NamedSelectOne(&res, `SELECT bar FROM foo WHERE bar = :bar`, map[string]interface{}{
				"bar": "bar-2",
			})
		})
		a.So(err, should.BeNil)
		a.So(res, should.Equal, data[1].Bar)
	}
}
