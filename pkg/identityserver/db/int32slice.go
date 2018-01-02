// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// Int32Slice is a type that wraps a int32 slice and implements sql.Scanner
// and driver.Valuer interfaces.
type Int32Slice []int32

// Value returns the slice marshaled into a string representation.
func (i Int32Slice) Value() (driver.Value, error) {
	m, err := json.Marshal(i)
	if err != nil {
		return nil, errors.NewWithCause("Failed to marshal an Int32Slice type", err)
	}

	return string(m[:]), nil
}

// Scan unmarshals value into the receiver.
func (i *Int32Slice) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.Errorf("Expected argument to Int32Slice.Scan to be string but got %T", src)
	}

	if err := json.Unmarshal([]byte(str), i); err != nil {
		return errors.NewWithCause("Failed to unmarshal text into an Int32Slice", err)
	}

	return nil
}

// NewInt32Slice returns an Int32Slice type from an int32 (or int32-like) slice.
func NewInt32Slice(from interface{}) (Int32Slice, error) {
	var dest Int32Slice
	if err := set(from, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// SetInto copies the content of the receiver into dest with the correct type.
// dest must be an int32 (or int32-like) slice.
func (i *Int32Slice) SetInto(dest interface{}) error {
	return set(i, dest)
}

// set copies the content of from to dest.
// Both arguments must be an int32 (or int32-like) slice.
func set(from, dest interface{}) error {
	ptr := reflect.ValueOf(dest)

	if ptr.Kind() != reflect.Ptr {
		return errors.Errorf("Expected pointer to slice but got %T", dest)
	}

	value := reflect.ValueOf(ptr.Elem().Interface())

	if value.Kind() != reflect.Slice {
		return errors.Errorf("Expected pointer to slice but got %T", dest)
	}

	et := value.Type().Elem()
	if et.Kind() != reflect.Int32 {
		return errors.Errorf("Expected slice of int32 (or int32-like) but got %T", dest)
	}

	fromValue := reflect.Indirect(reflect.ValueOf(from))

	if fromValue.Kind() != reflect.Slice {
		return errors.Errorf("Expected from to be slice but got %T", from)
	}

	if fromValue.Type().Elem().Kind() != reflect.Int32 {
		return errors.Errorf("Expected fromValue to be slice of int32 (or int32-like) but got %T", dest)
	}

	res := reflect.MakeSlice(value.Type(), fromValue.Len(), fromValue.Cap())
	for i := 0; i < fromValue.Len(); i++ {
		el := reflect.New(et)
		el.Elem().Set(fromValue.Index(i).Convert(et))
		res.Index(i).Set(el.Elem())
	}

	ptr.Elem().Set(res)

	return nil
}
