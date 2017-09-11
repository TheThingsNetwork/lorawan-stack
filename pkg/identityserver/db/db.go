// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"database/sql"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db/migrations"
	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/jmoiron/sqlx"

	// include pq for the postgres driver
	_ "github.com/lib/pq"
)

// QueryContext is the interface of contexts where a query can be run.
// Typically this means either the global database scope, or a transaction.
type QueryContext interface {
	// NamedExec executes the query in the specified context replacing the
	// placeholder parameters with fields from arg
	NamedExec(query string, arg interface{}) (sql.Result, error)

	// NamedSelectOne selects the query replacing the placeholder paramenters
	// with fields from arg and expects one result. It writes the result
	// into dest, which can be a map[string]interface, a struct, or a scannable.
	NamedSelectOne(dest interface{}, query string, arg interface{}) error

	// Select selects the query replacing the placeholder paramenters
	// with fields from arg and expects one result. It writes the results
	// into dest, which can be a []map[string]interface, or a slice of struct.
	NamedSelect(dest interface{}, query string, arg interface{}) error

	// Exec executes the query in the specified context and returns no result.
	Exec(query string, args ...interface{}) (sql.Result, error)

	// SelectOne selects the query and expects one result. It writes the result
	// into dest, which can be a map[string]interface, a struct, or a scannable.
	SelectOne(dest interface{}, query string, args ...interface{}) error

	// Select selects the query and expects one result. It writes the results
	// into dest, which can be a []map[string]interface, or a slice of struct.
	Select(dest interface{}, query string, args ...interface{}) error
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
	Migrator

	// A DB is a QueryContext that performs the queries at the top level
	QueryContext

	// Transact begins a transaction and runs the function in it, it returns the
	// error the function returns (and retries or rolls back automatically).
	Transact(func(*Tx) error) error
}

// DB implments Database
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
		return nil, err
	}

	return &DB{
		db:         db,
		context:    context,
		migrations: migrations,
	}, nil
}

// Close closes the connection to the database
func (db *DB) Close() error {
	return db.db.Close()
}

// WithContext returns a new DB with the same migratons registry and with the
// provided context as base context for all queries and transactions
func (db *DB) WithContext(context context.Context) *DB {
	return &DB{
		db:         db.db,
		context:    context,
		migrations: db.migrations,
	}
}

// NamedExec implements QueryContext
func (db *DB) NamedExec(query string, arg interface{}) (sql.Result, error) {
	nstmt, err := db.db.PrepareNamedContext(db.context, query)
	if err != nil {
		return nil, wrap(err)
	}
	res, err := nstmt.ExecContext(db.context, arg)
	return res, wrap(err)
}

// NamedSelectOne implements QueryContext
func (db *DB) NamedSelectOne(dest interface{}, query string, arg interface{}) error {
	nstmt, err := db.db.PrepareNamedContext(db.context, query)
	if err != nil {
		return wrap(err)
	}
	return namedSelectOne(db.context, nstmt, dest, arg)
}

// NamedSelect implements QueryContext
func (db *DB) NamedSelect(dest interface{}, query string, arg interface{}) error {
	nstmt, err := db.db.PrepareNamedContext(db.context, query)
	if err != nil {
		return wrap(err)
	}
	return namedSelectAll(db.context, nstmt, dest, arg)
}

func namedSelectOne(context context.Context, nstmt *sqlx.NamedStmt, dest, arg interface{}) error {
	var err error
	switch v := dest.(type) {
	case *map[string]interface{}:
		err = nstmt.QueryRowxContext(context, arg).MapScan(*v)
	case map[string]interface{}:
		err = nstmt.QueryRowxContext(context, arg).MapScan(v)
	default:
		err = nstmt.GetContext(context, dest, arg)
	}

	return wrap(err)
}

func namedSelectAll(context context.Context, nstmt *sqlx.NamedStmt, dest, arg interface{}) error {
	var err error
	switch v := dest.(type) {
	case *[]map[string]interface{}:
		rows, err := nstmt.QueryxContext(context, arg)
		if err != nil {
			return wrap(err)
		}

		aggr := make([]map[string]interface{}, 0)
		for rows.Next() {
			res := make(map[string]interface{})
			err = rows.MapScan(res)
			if err != nil {
				break
			}
			aggr = append(aggr, res)
		}
		*v = aggr
	default:
		err = nstmt.SelectContext(context, dest, arg)
	}

	return wrap(err)
}

// Exec implements QueryContext
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	res, err := db.db.ExecContext(db.context, query, args...)
	return res, wrap(err)
}

// SelectOne implements QueryContext
func (db *DB) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return selectOne(db.context, db.db, dest, query, args...)
}

// Select implements QueryContext
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return selectAll(db.context, db.db, dest, query, args...)
}

// Transact implements Database
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

	return wrap(crdb.ExecuteInTx(db.context, txx, func() error {
		return fn(tx)
	}))
}

// Tx is the type of a transaction
type Tx struct {
	tx      *sqlx.Tx
	context context.Context
}

// NamedExec implements QueryContext
func (tx *Tx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	res, err := tx.tx.NamedExec(query, arg)
	return res, wrap(err)
}

// NamedSelectOne implements QueryContext
func (tx *Tx) NamedSelectOne(dest interface{}, query string, arg interface{}) error {
	nstmt, err := tx.tx.PrepareNamed(query)
	if err != nil {
		return wrap(err)
	}
	return namedSelectOne(tx.context, nstmt, dest, arg)
}

// NamedSelect implements QueryContext
func (tx *Tx) NamedSelect(dest interface{}, query string, arg interface{}) error {
	nstmt, err := tx.tx.PrepareNamed(query)
	if err != nil {
		return wrap(err)
	}
	return namedSelectAll(tx.context, nstmt, dest, arg)
}

// Exec implements QueryContext
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	res, err := tx.tx.Exec(query, args...)
	return res, wrap(err)
}

// SelectOne implements QueryContext
func (tx *Tx) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return selectOne(tx.context, tx.tx, dest, query, args...)
}

// Select implements QueryContext
func (tx *Tx) Select(dest interface{}, query string, args ...interface{}) error {
	return selectAll(tx.context, tx.tx, dest, query, args...)
}

// selectOne selects one item from the database and writes it to dest, which can
// be a map[string]interface{} or a struct.
func selectOne(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) error {
	var err error
	switch v := dest.(type) {
	case *map[string]interface{}:
		err = q.QueryRowxContext(context, query, args...).MapScan(*v)
	case map[string]interface{}:
		err = q.QueryRowxContext(context, query, args...).MapScan(v)
	default:
		err = sqlx.GetContext(context, q, dest, query, args...)
	}

	return wrap(err)
}

// selectAll selects multiple items from the database and writes them to dest, which can
// be a slice of map[string]interface or a slice of structs.
func selectAll(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) error {
	var err error
	switch v := dest.(type) {
	case *[]map[string]interface{}:
		rows, err := q.QueryxContext(context, query, args...)
		if err != nil {
			return wrap(err)
		}

		aggr := make([]map[string]interface{}, 0)
		for rows.Next() {
			res := make(map[string]interface{})
			err = rows.MapScan(res)
			if err != nil {
				break
			}
			aggr = append(aggr, res)
		}
		*v = aggr
	default:
		err = sqlx.SelectContext(context, q, dest, query, args...)
	}

	return wrap(err)
}
