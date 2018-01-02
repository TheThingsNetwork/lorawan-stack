// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"fmt"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/jmoiron/sqlx"
)

// selectOne selects one item from the database and writes it to dest, which can
// be a map[string]interface{}, a struct or a scannable value.
func selectOne(context context.Context, q sqlx.QueryerContext, dest interface{}, query string, args ...interface{}) (err error) {
	defer func() { err = wrap(err) }()

	value := reflect.ValueOf(dest)
	typ := value.Type()

	isPtr := value.Kind() == reflect.Ptr

	if isPtr && value.IsNil() {
		return errors.New("Nil pointer passed to selectOne")
	}

	if isPtr {
		typ = typ.Elem()
		value = reflect.Indirect(value)
	}

	row := q.QueryRowxContext(context, query, args...)
	if err := row.Err(); err != nil && !IsNoRows(err) {
		return err
	}

	// try map
	if typ.Kind() == reflect.Map {
		m, ok := value.Interface().(map[string]interface{})
		if !ok {
			return errors.Errorf("Expected map[string]interface{} but got %s in selectOne", typ)
		}
		return row.MapScan(m)
	}

	// try struct
	if typ.Kind() == reflect.Struct {
		return row.StructScan(dest)
	}

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
			if err := rows.StructScan(res); err != nil {
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
