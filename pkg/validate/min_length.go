// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"fmt"
	"reflect"
)

// MinLength checks whether the input value has a minimum length in the following cases:
//     - For strings checks the length
//     - For slices checks that aren't nil and its length
//     - For other types returns error
func MinLength(length int) validateFn {
	return func(v interface{}) error {
		typ := reflect.TypeOf(v)

		if typ.Kind() == reflect.String {
			str, _ := v.(string)
			return minLengthString(str, length)
		}

		if typ.Kind() == reflect.Slice {
			return minLengthSlice(reflect.ValueOf(v), length)
		}

		return fmt.Errorf("Unsupported input type: `%T`", v)
	}
}

func minLengthSlice(v reflect.Value, length int) error {
	if v.IsNil() || v.Len() < length {
		return fmt.Errorf("Must be non-empty and have at least a length of value %d", length)
	}

	return nil
}

func minLengthString(v string, length int) error {
	if len(v) < length {
		return fmt.Errorf("Got string of length %d but minimum required is %d", len(v), length)
	}

	return nil
}
