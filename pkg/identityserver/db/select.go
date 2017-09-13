// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

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
		defer rows.Close()

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

func namedSelectOne(context context.Context, e sqlx.ExtContext, dest interface{}, query string, arg interface{}) error {
	q, args, err := sqlx.Named(query, arg)
	if err != nil {
		return wrap(err)
	}
	q = sqlx.Rebind(sqlx.DOLLAR, q)
	return selectOne(context, e, dest, q, args...)
}

func namedSelectAll(context context.Context, e sqlx.ExtContext, dest interface{}, query string, arg interface{}) error {
	q, args, err := sqlx.Named(query, arg)
	if err != nil {
		return wrap(err)
	}
	q = sqlx.Rebind(sqlx.DOLLAR, q)
	return selectAll(context, e, dest, q, args...)
}
