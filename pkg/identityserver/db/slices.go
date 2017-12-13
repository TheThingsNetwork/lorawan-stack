// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

// NewInt32Slice returns an Int32Slice type from an int32 (or int32-like) slice.
func NewInt32Slice(from interface{}) (Int32Slice, error) {
	var dest Int32Slice
	if err := set(from, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// Value returns the slice marshaled into a string representation.
func (i Int32Slice) Value() (driver.Value, error) {
	return value(i)
}

// Scan unmarshals value into the receiver.
func (i *Int32Slice) Scan(src interface{}) error {
	return scan(src, i)
}

// SetInto copies the content of the receiver into dest with the correct type.
// dest must be an int32 (or int32-like) slice.
func (i *Int32Slice) SetInto(dest interface{}) error {
	return set(i, dest)
}

// StringSlice wraps a string slice and implements sql.Scanner and driver.Valuer interfaces.
type StringSlice []string

// NewStringSlice returns a StrignSlice type from a string (or string-like) slice.
func NewStringSlice(from interface{}) (StringSlice, error) {
	var dest StringSlice
	if err := set(from, &dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// Value returns the slice marshaled into a string representation.
func (s StringSlice) Value() (driver.Value, error) {
	return value(s)
}

// Scan unmarshals value into the receiver.
func (s *StringSlice) Scan(src interface{}) error {
	return scan(src, s)
}

// SetInto copies the content of the receiver into dest with the correct type.
// dest must be a string (or string-like) slice.
func (s *StringSlice) SetInto(dest interface{}) error {
	return set(s, dest)
}

func value(data interface{}) (driver.Value, error) {
	m, err := json.Marshal(data)
	if err != nil {
		return nil, errors.NewWithCause("Failed to marshal an Int32Slice type", err)
	}

	return string(m[:]), nil
}

func scan(src, to interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.Errorf("Expected argument to Int32Slice.Scan to be string but got %T", src)
	}

	if err := json.Unmarshal([]byte(str), to); err != nil {
		return errors.NewWithCause("Failed to unmarshal text into an Int32Slice", err)
	}

	return nil
}

// set copies the content of from to dest.
// Both arguments must be a slice with the same underlying type.
func set(from, dest interface{}) error {
	destPtr := reflect.ValueOf(dest)

	if destPtr.Kind() != reflect.Ptr {
		return errors.Errorf("Expected dest to be pointer to slice but got %T", dest)
	}

	destValue := reflect.ValueOf(destPtr.Elem().Interface())

	if destValue.Kind() != reflect.Slice {
		return errors.Errorf("Expected dest to be pointer to slice but got %T", dest)
	}

	destValueType := destValue.Type().Elem()

	fromValue := reflect.Indirect(reflect.ValueOf(from))

	if fromValue.Kind() != reflect.Slice {
		return errors.Errorf("Expected from to be slice but got %T", from)
	}

	if fromValue.Type().Elem().Kind() != destValueType.Kind() {
		return errors.Errorf("Expected from (%T) and dest (%T) to be same type but not", from, dest)
	}

	res := reflect.MakeSlice(destValue.Type(), fromValue.Len(), fromValue.Cap())
	for i := 0; i < fromValue.Len(); i++ {
		el := reflect.New(destValueType)
		el.Elem().Set(fromValue.Index(i).Convert(destValueType))
		res.Index(i).Set(el.Elem())
	}

	destPtr.Elem().Set(res)

	return nil
}
