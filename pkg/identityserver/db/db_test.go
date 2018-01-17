// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db/migrations"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var dsn = &DataSourceName{
	DatabaseHostname: "localhost",
	DatabasePort:     26257,
	DatabaseName:     "is_db_tests",
	DatabaseUser:     "root",
}

const schema = `
	CREATE TABLE IF NOT EXISTS foo (
		id       SERIAL PRIMARY KEY,
		created  TIMESTAMP DEFAULT current_timestamp(),
		bar      TEXT,
		baz      BOOL,
		quu      INTEGER
	);
	`

var data = []foo{
	{
		Bar: "bar-1",
		Baz: true,
		Quu: 42,
	},
	{
		Bar: "bar-2",
		Baz: false,
		Quu: 392,
	},
}

type foo struct {
	ID      int64
	Created time.Time
	Bar     string
	Baz     bool
	Quu     int
}

var db Database

func getInstance(t testing.TB) Database {
	if db == nil {
		db = clean(t)
	}

	return db
}

func clean(t testing.TB) Database {
	logger := test.GetLogger(t).WithField("tag", "Identity Server")

	registry := migrations.NewRegistry()
	registry.Register(1, "1_foo_schema", schema, "DROP TABLE IF EXISTS foo")

	// open database connection
	db, err := Open(context.Background(), dsn, registry)
	if err != nil {
		logger.WithError(err).Fatal("Failed to establish a connection with the CockroachDB instance")
		return nil
	}

	// drop database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s CASCADE", dsn.DatabaseName))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to delete database `%s`", dsn.DatabaseName)
		return nil
	}

	// create it again
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dsn.DatabaseName))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create database `%s`", dsn.DatabaseName)
		return nil
	}

	// apply all migrations
	err = db.MigrateAll()
	if err != nil {
		logger.WithError(err).Fatal("Failed to apply migrations from the registry")
		return nil
	}

	for _, f := range data {
		_, err = db.Exec(`INSERT INTO foo (bar, baz, quu) VALUES ($1, $2, $3) RETURNING *`, f.Bar, f.Baz, f.Quu)
		if err != nil {
			logger.WithError(err).Fatalf("Failed to feed the test database `%s` with some data", dsn.DatabaseName)
			return nil
		}
	}

	return db
}

func testExec(t *testing.T, q QueryContext) {
	a := assertions.New(t)

	id := int64(1234)

	_, err := q.Exec(`INSERT INTO foo (id, bar) VALUES ($1, $2)`, id, "bar")
	a.So(err, should.BeNil)

	_, err = q.Exec(`DELETE FROM foo WHERE id = $1`, id)
	a.So(err, should.BeNil)
}

func TestExec(t *testing.T) {
	testExec(t, getInstance(t))
}

func TestNamedExec(t *testing.T) {
	testNamedExec(t, getInstance(t))
}

func TestSelectOne(t *testing.T) {
	testSelectOne(t, getInstance(t))
}

func TestNamedSelectOne(t *testing.T) {
	testNamedSelectOne(t, getInstance(t))
}

func TestSelect(t *testing.T) {
	testSelect(t, getInstance(t))
}

func TestNamedSelect(t *testing.T) {
	testNamedSelect(t, getInstance(t))
}
