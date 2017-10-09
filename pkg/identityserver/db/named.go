// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"

	"github.com/gomezjdaniel/sqlx"
)

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
// It uses named items from args to fill the query.
func namedSelectOne(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, arg interface{}) error {
	bound, args, err := compileNamedQuery(query, arg)
	if err != nil {
		return wrap(err)
	}
	return selectOne(context, q, dest, bound, args...)
}

// namedSelectAll selects multiple items from the database and writes them to dest,
// which can be a slice of map[string]interface{} or a slice of structs, or a slice
// of scannable values. It uses the items from arg to fill the query.
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

	for i, arg := range args {
		// continue if implements driver.Valuer
		if _, ok := arg.(driver.Valuer); ok {
			continue
		}

		// wrap into Array if it's an int32 (or int32 like) slice
		typ := reflect.TypeOf(arg)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if isInt32Slice(typ) {
			args[i] = Array(arg)
		}
	}

	return bound, args, nil
}
