// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db/migrations"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "db_tests"
	schema   = `
	CREATE TABLE IF NOT EXISTS foo (
		id       SERIAL PRIMARY KEY,
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
	testSelect(t, getInstance())
}

func TestNamedSelect(t *testing.T) {
	testNamedSelect(t, getInstance())
}

func TestSelectOne(t *testing.T) {
	testSelectOne(t, getInstance())
}

func TestNamedSelectOne(t *testing.T) {
	testNamedSelectOne(t, getInstance())
}
