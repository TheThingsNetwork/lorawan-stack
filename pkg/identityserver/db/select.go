// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
)

// selectOne selects one item from the database and writes it to dest, which can
// be a map[string]interface{}, a struct or a scannable value.
func selectOne(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) (err error) {
	defer func() { err = wrap(err) }()

	// perform query
	row := q.QueryRowxContext(context, query, args...)
	if err := row.Err(); err != nil && !IsNoRows(err) {
		return err
	}

	value := reflect.ValueOf(dest)
	typ := value.Type()

	isPtr := value.Kind() == reflect.Ptr

	if isPtr && value.IsNil() {
		return errors.New("Nil pointer passed to selectOne")
	}

	if isPtr {
		typ = value.Type().Elem()
		value = reflect.Indirect(value)
	}

	// try map
	if typ.Kind() == reflect.Map {
		m, ok := value.Interface().(map[string]interface{})
		if !ok {
			return fmt.Errorf("Expected map[string]interface{} but got %s", typ)
		}
		return row.MapScan(m)
	}

	// try struct
	if typ.Kind() == reflect.Struct {
		return row.StructScan(dest)
	}

	// try scannable
	return row.Scan(dest)
}

// selectAll selects multiple items from the database and writes them to dest, which can
// be a slice of map[string]interface or a slice of structs, or a slice of scannable values.
func selectAll(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) (err error) {
	defer func() { err = wrap(err) }()

	rows, err := q.QueryxContext(context, query, args...)
	if err != nil && !IsNoRows(err) {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	defer rows.Close()

	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected pointer to slice but got %s", value.Type())
	}

	if value.IsNil() {
		return errors.New("Nil pointer passed to selectAll")
	}

	slice := reflect.Indirect(value)
	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("Expected pointer to slice of map or struct, but got %s", value.Type())
	}

	base := slice.Type().Elem()
	isPtr := base.Kind() == reflect.Ptr
	if isPtr {
		base = base.Elem()
	}

	// try map
	if base.Kind() == reflect.Map {
		_, ok := reflect.New(base).Elem().Interface().(map[string]interface{})
		if !ok {
			return fmt.Errorf("Expected []map[string]interface{} but got []%s", base)
		}

		for rows.Next() {
			res := make(map[string]interface{})
			err := rows.MapScan(res)
			if err != nil {
				return err
			}

			if isPtr {
				slice.Set(reflect.Append(slice, reflect.ValueOf(&res)))
			} else {
				slice.Set(reflect.Append(slice, reflect.ValueOf(res)))
			}
		}

		return nil
	}

	// try struct
	if base.Kind() == reflect.Struct {
		for rows.Next() {
			vp := reflect.New(base)
			res := vp.Interface()
			err := rows.StructScan(res)
			if err != nil {
				return err
			}

			if isPtr {
				slice.Set(reflect.Append(slice, vp))
			} else {
				slice.Set(reflect.Append(slice, reflect.Indirect(vp)))
			}
		}
		return nil
	}

	// try scannable
	for rows.Next() {
		vp := reflect.New(base)
		res := vp.Interface()
		err := rows.Scan(res)
		if err != nil {
			return err
		}

		if isPtr {
			slice.Set(reflect.Append(slice, vp))
		} else {
			slice.Set(reflect.Append(slice, reflect.Indirect(vp)))
		}
	}

	return nil
}

// namedSelectOne selects one row from the database and writes the result to the dest, which can be a map[string]interface{},
// a struct, or a scannable value. It uses named items from args to fill the query.
func namedSelectOne(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, arg interface{}) error {
	bound, args, err := sqlx.Named(query, arg)
	if err != nil {
		return wrap(err)
	}
	bound = sqlx.Rebind(sqlx.DOLLAR, bound)
	return selectOne(context, q, dest, bound, args...)
}

// namedSelectAll selects multiple items from the database and writes them to dest, which can
// be a slice of map[string]interface{} or a slice of structs, or a slice of scannable values.
// It uses the items from arg to fill the query.
func namedSelectAll(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, arg interface{}) error {
	bound, args, err := sqlx.Named(query, arg)
	if err != nil {
		return wrap(err)
	}
	bound = sqlx.Rebind(sqlx.DOLLAR, bound)
	return selectAll(context, q, dest, bound, args...)
}
