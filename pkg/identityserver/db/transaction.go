// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"database/sql"

	"github.com/gomezjdaniel/sqlx"
)

// Tx is the type of a transaction.
type Tx struct {
	tx      *sqlx.Tx
	context context.Context
}

// Exec implements QueryContext.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	res, err := tx.tx.Exec(query, args...)
	return res, wrap(err)
}

// SelectOne implements QueryContext.
func (tx *Tx) SelectOne(dest interface{}, query string, args ...interface{}) error {
	return selectOne(tx.context, tx.tx, dest, query, args...)
}

// Select implements QueryContext.
func (tx *Tx) Select(dest interface{}, query string, args ...interface{}) error {
	return selectAll(tx.context, tx.tx, dest, query, args...)
}

// NamedExec implements QueryContext.
func (tx *Tx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return namedExec(tx.context, tx.tx, query, arg)
}

// NamedSelectOne implements QueryContext.
func (tx *Tx) NamedSelectOne(dest interface{}, query string, arg interface{}) error {
	return namedSelectOne(tx.context, tx.tx, dest, query, arg)
}

// NamedSelect implements QueryContext.
func (tx *Tx) NamedSelect(dest interface{}, query string, arg interface{}) error {
	return namedSelectAll(tx.context, tx.tx, dest, query, arg)
}
