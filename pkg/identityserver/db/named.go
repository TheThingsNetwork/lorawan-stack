// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"database/sql"

	"github.com/gomezjdaniel/sqlx"
)

// namedExec executes the named query using the provided argument and returns
// a sql.Result.
func namedExec(context context.Context, q sqlx.ExecerContext, query string, arg interface{}) (res sql.Result, err error) {
	defer func() { err = wrap(err) }()
	bound, args, err := compileNamedQuery(query, arg)
	if err != nil {
		return
	}
	return q.ExecContext(context, bound, args...)
}

// namedSelectOne selects one row from the database and writes the result to the
// dest, which can be a map[string]interface{}, a struct, or a scannable value.
// It construct the query using the named parameters and the argument.
func namedSelectOne(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, arg interface{}) error {
	bound, args, err := compileNamedQuery(query, arg)
	if err != nil {
		return wrap(err)
	}
	return selectOne(context, q, dest, bound, args...)
}

// namedSelectAll selects multiple items from the database and writes them to dest,
// which can be a slice of map[string]interface{} or a slice of structs, or a slice
// of scannable values. It construct the query using the named parameters and the argument.
func namedSelectAll(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, arg interface{}) error {
	bound, args, err := compileNamedQuery(query, arg)
	if err != nil {
		return wrap(err)
	}
	return selectAll(context, q, dest, bound, args...)
}

func compileNamedQuery(query string, arg interface{}) (string, []interface{}, error) {
	bound, args, err := sqlx.Named(query, arg)
	if err != nil {
		return "", nil, err
	}
	bound = sqlx.Rebind(sqlx.DOLLAR, bound)

	return bound, args, nil
}
