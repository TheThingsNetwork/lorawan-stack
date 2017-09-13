// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db/migrations"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "db_tests"
	schema   = `
	CREATE TABLE IF NOT EXISTS foo (
		id       SERIAL,
		created  TIMESTAMP DEFAULT current_timestamp(),
		bar      TEXT,
		baz      BOOL,
		quu      INTEGER
	);
	`
)

var data = []foo{
	foo{
		Bar: "bar-1",
		Baz: true,
		Quu: 42,
	},
	foo{
		Bar: "bar-2",
		Baz: false,
		Quu: 392,
	},
}

type foo struct {
	ID      int64     `db:"id"`
	Created time.Time `db:"created"`
	Bar     string    `db:"bar"`
	Baz     bool      `db:"baz"`
	Quu     int       `db:"quu"`
}

var db Database

func getInstance() Database {
	if db == nil {
		db = clean()
	}

	return db
}

func clean() Database {
	registry := migrations.NewRegistry()
	registry.Register(1, "1_foo_schema", schema, "DROP TABLE IF EXISTS foo")

	// open database connection
	db, err := Open(
		context.Background(),
		fmt.Sprintf(address, database),
		registry,
	)
	if err != nil {
		panic(err)
	}

	// drop database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", database))
	if err != nil {
		panic(err)
	}

	// create it again
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", database))
	if err != nil {
		panic(err)
	}

	err = db.MigrateAll()
	if err != nil {
		panic(err)
	}

	for _, f := range data {
		_, err = db.Exec(`INSERT INTO foo (bar, baz, quu) VALUES ($1, $2, $3) RETURNING *`, f.Bar, f.Baz, f.Quu)
		if err != nil {
			panic(err)
		}
	}

	return db
}

func TestSelect(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	{
		res := make([]*foo, 0)
		err := db.Select(&res, "SELECT * FROM foo")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, len(data))
	}

	// into a slice of struct ptr
	{
		res := make([]*foo, 0)
		err := db.Select(&res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
	}

	// into a slice of struct
	{
		res := make([]foo, 0)
		err := db.Select(&res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
	}

	// into a slice of values
	{
		res := make([]string, 0)
		err := db.Select(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, 1)
		a.So(res[0], should.Equal, data[1].Bar)
	}
}

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

func TestNamedSelect(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	{
		res := make([]*foo, 0)
		err := db.NamedSelect(&res, "SELECT * FROM foo", map[string]interface{}{})
		a.So(err, should.BeNil)
		a.So(res, should.HaveLength, len(data))
	}

	// into a slice of ptr to struct
	{
		res := make([]*foo, 0)
		err := db.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
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
		err := db.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
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
		err := db.NamedSelect(&res, `SELECT * FROM foo WHERE bar = :bar`, foo{
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

	// into a slice of values
	{
		res := make([]string, 0)
		err := db.NamedSelect(&res, `SELECT bar FROM foo WHERE bar = :bar`, foo{
			Bar: "bar-2",
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

func TestSelectOne(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	// into struct ptr
	{
		res := new(foo)
		err := db.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := db.SelectOne(res, `SELECT * FROM foo WHERE bar = $1`, "bar-2")
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
		err := db.SelectOne(&res, `SELECT bar FROM foo WHERE bar = $1`, "bar-2")
		a.So(err, should.BeNil)
		a.So(res, should.Equal, data[1].Bar)
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

func TestNamedSelectOne(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	// into struct ptr
	{
		res := new(foo)
		err := db.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
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
		err := db.NamedSelectOne(res, `SELECT * FROM foo WHERE bar = :bar`, map[string]interface{}{
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
		err := db.NamedSelectOne(&res, `SELECT bar FROM foo WHERE bar = :bar`, map[string]interface{}{
			"bar": "bar-2",
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
