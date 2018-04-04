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
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
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
