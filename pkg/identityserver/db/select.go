// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
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
	var err error
	rows, err := sqlx.NamedQueryContext(context, e, query, arg)
	if err != nil {
		return wrap(err)
	}

	defer rows.Close()

	if rows.Next() {
		switch v := dest.(type) {
		case *map[string]interface{}:
			return wrap(rows.MapScan(*v))
		case map[string]interface{}:
			return wrap(rows.MapScan(v))
		default:
			vv := reflect.ValueOf(dest)
			if vv.Kind() == reflect.Ptr && vv.Type().Elem().Kind() == reflect.Struct {
				return wrap(rows.StructScan(dest))
			}

			return wrap(rows.Scan(dest))
		}
	}
	return nil
}

func namedSelectAll(context context.Context, e sqlx.ExtContext, dest interface{}, query string, arg interface{}) error {
	var err error
	rows, err := sqlx.NamedQueryContext(context, e, query, arg)
	if err != nil {
		return wrap(err)
	}
	defer rows.Close()
	switch v := dest.(type) {
	case *[]map[string]interface{}:
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
		err = scanAll(rows, dest)
		if err != nil {
			return err
		}
	}

	return wrap(err)
}

var _scannerInterface = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

// scanAll scans all the rows into a slice of struct, slice of map[string]interface{} or slice of sql.Scanner.
// Copied and modified from https://github.com/jmoiron/sqlx
func scanAll(rows *sqlx.Rows, dest interface{}) error {
	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Ptr {
		return errors.New("Must pass a pointer, not a value to StructScan destination")
	}

	if value.IsNil() {
		return errors.New("nil pointer passed to StructScan destination")
	}

	slice := reflect.Indirect(value)

	if slice.Kind() != reflect.Slice {
		return fmt.Errorf("Expected a slice, but got %s", slice.Kind())
	}

	isPtr := slice.Type().Elem().Kind() == reflect.Ptr
	base := baseType(slice.Type().Elem())

	scannable := isScannable(base)
	if scannable {
		for rows.Next() {
			vp := reflect.New(base)
			err := rows.Scan(vp.Interface())
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

	if base.Kind() != reflect.Struct {
		return fmt.Errorf("Expected a slice of struct, but got %s", slice.Type())
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	fields := rows.Mapper.TraversalsByName(base, columns)
	values := make([]interface{}, len(columns))

	for rows.Next() {
		vp := reflect.New(base)
		v := reflect.Indirect(vp)

		err = fieldsByTraversal(v, fields, values, true)
		if err != nil {
			return err
		}

		err = rows.Scan(values...)
		if err != nil {
			return err
		}

		if isPtr {
			slice.Set(reflect.Append(slice, vp))
		} else {
			slice.Set(reflect.Append(slice, v))
		}
	}

	return nil
}

// isScannable takes the reflect.Type and the actual dest value and returns
// whether or not it's Scannable.  Something is scannable if:
// Copied and modified from https://github.com/jmoiron/sqlx
func isScannable(t reflect.Type) bool {
	if reflect.PtrTo(t).Implements(_scannerInterface) {
		return true
	}
	if t.Kind() != reflect.Struct {
		return true
	}

	return false
}

// fieldsByName fills a values interface with fields from the passed value based
// on the traversals in int.
// Copied and modified from https://github.com/jmoiron/sqlx
func fieldsByTraversal(v reflect.Value, traversals [][]int, values []interface{}, ptrs bool) error {
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return errors.New("argument not a struct")
	}

	for i, traversal := range traversals {
		if len(traversal) == 0 {
			values[i] = new(interface{})
			continue
		}
		f := reflectx.FieldByIndexes(v, traversal)
		if ptrs {
			values[i] = f.Addr().Interface()
		} else {
			values[i] = f.Interface()
		}
	}
	return nil
}

func baseType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}

	return t
}
