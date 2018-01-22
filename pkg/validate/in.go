// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"fmt"
	"reflect"
)

// In checks whether:
//     - an element is contained in an array or slice
//     - a slice is a subset of other slice
func In(slice interface{}) validateFn { // nolint: golint, returns unexported type on purpose
	return func(v interface{}) error {
		sliceVal := reflect.ValueOf(slice)

		if sliceVal.Kind() != reflect.Slice && sliceVal.Kind() != reflect.Array {
			return fmt.Errorf("In validator: got %T instead of an slice or array", v)
		}

		fn := func(v interface{}, slice reflect.Value) error {
			var found bool
			for i := 0; i < slice.Len() && !found; i++ {
				found = reflect.DeepEqual(v, slice.Index(i).Interface())
			}

			if !found {
				return fmt.Errorf("Expected `%v` to be in `%v` but not", v, slice)
			}

			return nil
		}

		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			if val.Len() > sliceVal.Len() {
				return fmt.Errorf("Expected `%v` (length %d) to be a subset of `%v` (length %d) but not", v, val.Len(), slice, sliceVal.Len())
			}

			for i := 0; i < val.Len(); i++ {
				err := fn(val.Index(i).Interface(), sliceVal)
				if err != nil {
					return err
				}
			}
		} else {
			return fn(v, sliceVal)
		}

		return nil
	}
}
