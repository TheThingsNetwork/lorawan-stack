// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package db implements a reusable interface around sql databases.
package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db/migrations"
	"github.com/cockroachdb/cockroach-go/crdb"

	"github.com/jmoiron/sqlx"

	// include pq for the postgres driver
	_ "github.com/lib/pq"
)

// make sure DB implements Database.
var _ Database = &DB{}

// QueryContext is the interface of contexts where a query can be run.
// Typically this means either the global database scope, or a transaction.
type QueryContext interface {
	// Exec executes the query in the specified context and returns no result.
	Exec(query string, args ...interface{}) (sql.Result, error)

	// NamedExec executes the query in the specified context replacing the
	// placeholder parameters with fields from arg.
	NamedExec(query string, arg interface{}) (sql.Result, error)

	// SelectOne selects the query and expects one result. It writes the result
	// into dest, which can be a map[string]interface, a struct, or a scannable.
	SelectOne(dest interface{}, query string, args ...interface{}) error

	// NamedSelectOne selects the query replacing the placeholder paramenters
	// with fields from arg and expects one result. It writes the result
	// into dest, which can be a map[string]interface, a struct, or a scannable.
	NamedSelectOne(dest interface{}, query string, arg interface{}) error

	// Select selects the query and expects one result. It writes the results
	// into dest, which can be a []map[string]interface, or a slice of struct.
	Select(dest interface{}, query string, args ...interface{}) error

	// Select selects the query replacing the placeholder paramenters
	// with fields from arg and expects one result. It writes the results
	// into dest, which can be a []map[string]interface, or a slice of struct.
	NamedSelect(dest interface{}, query string, arg interface{}) error
}

// Migrator is the interface that provides methods to manage the database schema
// through incremental but also reversible migrations.
type Migrator interface {
	// Migrate applies migrations  forwards or backwards until the target
	// migration is reached.
	Migrate(target int) error

	// MigrateAll applies all migrations forwards until the final migration is
	// reached.
	MigrateAll() error
}

// Database is the interface of an sql database, it can run global queries or
// start a transaction.
type Database interface {
	// A DB is a QueryContext that performs the queries at the top level.
	QueryContext

	// Transact begins a transaction and runs the function in it, it returns the
	// error the function returns (and retries or rolls back automatically).
	Transact(func(*Tx) error, ...TxOption) error

	// Close closes the database connection, releasing any open resources.
	Close() error
}

// DB implements Database.
type DB struct {
	db         *sqlx.DB
	context    context.Context
	migrations migrations.Registry
}

// Open opens a new database connection to the specified address.
func Open(context context.Context, address string, migrations migrations.Registry) (*DB, error) {
	db, err := sqlx.Open("postgres", address)
	if err != nil {
		return nil, err
	}

	err = db.PingContext(context)
	if err != nil {
		return nil, errors.NewWithCause(fmt.Sprintf("Failed to ping the CockroachDB instance at `%s`. Are you sure it is running?", address), err)
	}

	return &DB{
		db:         db,
		context:    context,
		migrations: migrations,
	}, nil
}

// Close closes the connection to the database.
func (db *DB) Close() error {
	return db.db.Close()
}

// WithContext returns a new DB with the same migratons registry and with the
// provided context as base context for all queries and transactions.
func (db *DB) WithContext(context context.Context) *DB {
	return &DB{
		db:         db.db,
		context:    context,
		migrations: db.migrations,
	}
}

// Exec implements QueryContext.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	res, err := db.db.ExecContext(db.context, query, args...)
	return res, wrap(err)
}

// NamedExec implements QueryContext.
func (db *DB) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return namedExec(db.context, db.db, query, arg)
}

// SelectOne implements QueryContext.
func (db *DB) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return selectOne(db.context, db.db, dest, query, args...)
}

// NamedSelectOne implements QueryContext.
func (db *DB) NamedSelectOne(dest interface{}, query string, arg interface{}) error {
	return namedSelectOne(db.context, db.db, dest, query, arg)
}

// Select implements QueryContext.
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return selectAll(db.context, db.db, dest, query, args...)
}

// NamedSelect implements QueryContext.
func (db *DB) NamedSelect(dest interface{}, query string, arg interface{}) error {
	return namedSelectAll(db.context, db.db, dest, query, arg)
}

// Transact implements Database.
func (db *DB) Transact(fn func(*Tx) error, options ...TxOption) error {
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}

	for _, option := range options {
		if option != nil {
			option(opts)
		}
	}

	txx, err := db.db.BeginTxx(db.context, opts)
	if err != nil {
		return wrap(err)
	}

	tx := &Tx{
		tx:      txx,
		context: db.context,
	}

	return wrap(crdb.ExecuteInTx(tx.context, txx, func() error {
		return fn(tx)
	}))
}
