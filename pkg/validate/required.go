// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"errors"
	"reflect"
)

var errZeroValue = errors.New("variable is empty")

// IsZeroer is an interface, which reports whether it represents a zero value.
type IsZeroer interface {
	IsZero() bool
}

var isZeroerType = reflect.TypeOf((*IsZeroer)(nil)).Elem()

// isZeroValue is like isZero, but acts on values of reflect.Value type.
func isZeroValue(v reflect.Value) bool {
	v = reflect.Indirect(v)
	if !v.IsValid() {
		return true
	}

	if v.Type().Implements(isZeroerType) {
		return v.Interface().(IsZeroer).IsZero()
	}

	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

// isZero reports whether the value is the zero of its type.
func isZero(v interface{}) bool {
	return isZeroValue(reflect.ValueOf(v))
}

// Empty returns error if v is set.
// It is meant to be used as the first validator function passed as argument to Field.
// It uses IsZero, if v implements IsZeroer interface.
func Empty(v interface{}) error {
	if !isZero(v) {
		return errors.New("Field must not be set")
	}
	return nil
}

// NotRequired returns an error, used internally in Field, if v is zero.
// It is meant to be used as the first validator function passed as argument to Field.
// It uses IsZero, if v implements IsZeroer interface.
func NotRequired(v interface{}) error {
	if isZero(v) {
		return errZeroValue
	}
	return nil
}

// Required returns error if v is empty.
// It is meant to be used as the first validator function passed as argument to Field.
// It uses IsZero, if v implements IsZeroer interface.
func Required(v interface{}) error {
	if isZero(v) {
		return errors.New("Field must not be empty")
	}
	return nil
}
