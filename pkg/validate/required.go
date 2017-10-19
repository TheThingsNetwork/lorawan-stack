// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"errors"
	"reflect"
)

var errZeroValue = errors.New("variable is empty")

// NotRequired checks whether the input value is empty.
func NotRequired(v interface{}) error {
	if isZero(v) {
		return errZeroValue
	}
	return nil
}

// Required checks whether the input value is empty or not.
func Required(v interface{}) error {
	if isZero(v) {
		return errors.New("Field cannot be empty")
	}
	return nil
}

func isZero(v interface{}) bool {
	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return value.IsNil()
	case reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if z := isZero(value.Index(i).Interface()); !z {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if z := isZero(value.Field(i).Interface()); !z {
				return false
			}
		}
		return true
	}
	return value.Interface() == reflect.Zero(value.Type()).Interface()
}
