// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/gomezjdaniel/sqlx"
	"github.com/gomezjdaniel/sqlx/reflectx"
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
		return structScan(row, dest)
	}

	// try int32 (or int32 like) slice
	if isInt32Slice(typ) {
		return row.Scan(Array(dest))
	}

	return row.Scan(dest)
}

// structScan scans a single Row into an struct. It supports to scan columns
// into struct fields that are int32 (or int32 like) slices.
func structScan(row *sqlx.Row, dest interface{}) error {
	defer row.Close()

	if err := row.Err(); err != nil {
		return err
	}

	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Ptr {
		return errors.New("Must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		return errors.New("Nil pointer passed to StructScan destination")
	}
	if value.Elem().Kind() != reflect.Struct {
		return errors.Errorf("Expected struct as StructScan destination but got %T", dest)
	}

	columns, err := row.Columns()
	if err != nil {
		return err
	}

	// try scannable
	if _, ok := dest.(sql.Scanner); ok {
		if len(columns) > 1 {
			return errors.Errorf("Scannable destination type %T with >1 columns (%d) in result", dest, columns)
		}

		return row.Scan(dest)
	}

	fields := row.Mapper.TraversalsByName(value.Type(), columns)
	if field, err := missingFields(fields); err != nil {
		return errors.Errorf("Missing destination name %s in %T", columns[field], dest)
	}

	values := make([]interface{}, len(columns))
	if err := fieldsByTraversal(value, fields, values, true); err != nil {
		return err
	}
	return row.Scan(values...)
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
		wrows := &wrows{Rows: rows}
		for rows.Next() {
			vp := reflect.New(base)
			res := vp.Interface()
			if err := wrows.StructScan(res); err != nil {
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

// wrows is the type that wraps an *sqlx.Rows type and allows to scan a set of
// rows into an struct supporting fields that are int32 (or int32 like) slices.
//
// In comparison with scanning a single row, for multiple rows it is needed to
// define this custom type in order to cache the reflect work of matching up
// column positions to fields to avoid that overhead per scan, which means it is
// not safe to run StructScan on the same wrows instance with different struct types.
type wrows struct {
	*sqlx.Rows
	started bool
	fields  [][]int
	values  []interface{}
}

// StructScan scan all rows into dest. It does exactly the same as sqlx.StructScan
// does except that it supports fields that are int32 (or int32 like) slices by
// wrapping them using the Array method provided in this package.
func (r *wrows) StructScan(dest interface{}) error {
	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Ptr {
		return errors.New("Must pass a pointer, not a value, to StructScan destination")
	}

	value = reflect.Indirect(value)

	if !r.started {
		columns, err := r.Rows.Columns()
		if err != nil {
			return err
		}

		r.fields = r.Mapper.TraversalsByName(value.Type(), columns)
		if field, err := missingFields(r.fields); err != nil {
			return errors.Errorf("Missing destination name %s in %T", columns[field], dest)
		}

		r.values = make([]interface{}, len(columns))
		r.started = true
	}

	if err := fieldsByTraversal(value, r.fields, r.values, true); err != nil {
		return err
	}

	if err := r.Rows.Scan(r.values...); err != nil {
		return err
	}

	return r.Rows.Err()
}

// fieldsByTraversal fills a values interface with fields from the passed value based
// on the traversals in int. If ptrs is true, return addresses instead of values.
// We write this instead of using FieldsByName to save allocations and map lookups
// when iterating over many rows. Empty traversals will get an interface pointer.
// Because of the necessity of requesting ptrs or values, it's considered a bit too
// specialized for inclusion in reflectx itself.
//
// This is a copy of the sqlx.fieldsByTraversal method but modifying to wrap a
// field with the Array method of this package if the field it is an int32 (or int32 like)
// slice.
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
			addr := f.Addr().Interface()
			if isInt32Slice(reflect.TypeOf(addr).Elem()) {
				values[i] = Array(addr)
			} else {
				values[i] = addr
			}
		} else {
			values[i] = f.Interface()
		}
	}
	return nil
}

// missingFields checks if the mapper has traversed all columns into the type,
// returning an error and the field number if it was not the case.
//
// This is an exact copy of missingFields method from sqlx package.
func missingFields(traversals [][]int) (field int, err error) {
	for i, t := range traversals {
		if len(t) == 0 {
			return i, errors.New("missing field")
		}
	}
	return 0, nil
}
