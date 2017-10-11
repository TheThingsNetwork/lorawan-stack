// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"database/sql"
	"database/sql/driver"
	"reflect"
	"strconv"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/lib/pq"
)

// Array wraps an int32 slice (or a type whose underlying type is an int32 slice)
// into a new type that implements both driver.Valuer and sql.Scanner interfaces.
//
// Array is a wrapper for the pq (postgres golang driver) library Array method
// so if the passed argument it is not a pointer to int32 (or int32 like) slice
// it fallbacks into the pq Array method.
func Array(a interface{}) interface {
	driver.Valuer
	sql.Scanner
} {

	switch b := a.(type) {
	case *[]int32:
		return (*Int32Array)(b)
	default:
		typ := reflect.TypeOf(b)

		if typ.Kind() == reflect.Ptr && isInt32Slice(typ.Elem()) {
			return &IntLikeArray{
				val: b,
			}
		}
	}

	return pq.Array(a)
}

// IntLikeArray is the type used to wrap an []int32 like (i.e. a type whose
// underlying type is an int32 slice) into a type that implements both
// driver.Valuer and sql.Scanner interfaces.
type IntLikeArray struct {
	val interface{}
}

// Scan implements the sql.Scanner interface.
func (a *IntLikeArray) Scan(src interface{}) error {
	ints := Int32Array(make([]int32, 0))
	err := ints.Scan(src)
	if err != nil {
		return err
	}

	if ints != nil {
		set(a.val, ints)
	}

	return nil
}

// set will set the elements of ints in a (with the correct type).
func set(a interface{}, ints []int32) error {
	ptr := reflect.ValueOf(a)

	if ptr.Kind() != reflect.Ptr {
		return errors.Errorf("Expected pointer to slice but got %T", a)
	}

	value := reflect.ValueOf(ptr.Elem().Interface())

	if value.Kind() != reflect.Slice {
		return errors.Errorf("Expected to pointer to slice but got %T", a)
	}

	et := value.Type().Elem()
	if et.Kind() != reflect.Int32 {
		return errors.Errorf("Expected slice of int32 likes but got %T", a)
	}

	res := reflect.MakeSlice(value.Type(), len(ints), cap(ints))
	for i, v := range ints {
		el := reflect.New(et)
		el.Elem().Set(reflect.ValueOf(v).Convert(et))
		res.Index(i).Set(el.Elem())
	}

	ptr.Elem().Set(res)

	return nil
}

// Value implements driver.Valuer interface.
func (a IntLikeArray) Value() (driver.Value, error) {
	ints, err := int32Slice(a.val)
	if err != nil {
		return nil, err
	}
	return Int32Array(ints).Value()
}

// int32Slice converts and int32 like slice (i.e. a type whose underlying type
// is an int32 slice) into an int32 slice.
func int32Slice(in interface{}) ([]int32, error) {
	int32SliceType := reflect.ValueOf([]int32{}).Type()
	int32Type := int32SliceType.Elem()

	value := reflect.ValueOf(in)

	// dereference slice if necessary
	if value.Kind() == reflect.Ptr {
		return int32Slice(reflect.Indirect(value).Interface())
	}

	if value.Kind() != reflect.Slice {
		return nil, errors.Errorf("Expected slice but got %T", in)
	}

	if value.Type().Elem().Kind() != reflect.Int32 {
		return nil, errors.Errorf("Expected slice of int32 (or int32-like) but got %T", in)
	}

	result := reflect.MakeSlice(int32SliceType, value.Len(), value.Cap())

	for i := 0; i < value.Len(); i++ {
		el := value.Index(i)
		iel := el.Convert(int32Type)
		result.Index(i).Set(iel)
	}

	return result.Interface().([]int32), nil
}

// Int32Array is a type that wraps an []int32 and implements both sql.Scanner
// and driver.Valuer interface.
type Int32Array []int32

// Scan implements the sql.Scanner interface.
func (a *Int32Array) Scan(src interface{}) error {
	switch src := src.(type) {
	case string:
		return a.scan(src)
	case nil:
		*a = nil
		return nil
	}

	return errors.Errorf("Failed to convert %T to Int32Array", src)
}

func (a *Int32Array) scan(str string) error {
	str = strings.Replace(str, "{", "", 1)
	str = strings.Replace(str, "}", "", 1)
	elems := strings.Split(str, ",")

	b := make([]int32, 0, len(elems))
	if len(elems) >= 1 && len(elems[0]) != 0 {
		for _, elem := range elems {
			n, err := strconv.ParseInt(elem, 10, 32)
			if err != nil {
				return err
			}
			b = append(b, int32(n))
		}
	}
	*a = b

	return nil
}

// Value implements the driver.Valuer interface.
func (a Int32Array) Value() (driver.Value, error) {
	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, int64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, int64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// isInt32Slice checks if typ is an int32 slice (or a type whose underlying type
// is an int32 slice).
func isInt32Slice(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Int32
}
