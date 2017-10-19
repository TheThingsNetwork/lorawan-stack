// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"fmt"
	"reflect"
)

func In(slice interface{}) validateFn {
	return func(v interface{}) error {
		value := reflect.ValueOf(slice)

		if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
			return fmt.Errorf("Expected a slice or array but got %T", v)
		}

		var found bool
		for i := 0; i < value.Len() && !found; i++ {
			if reflect.DeepEqual(v, value.Index(i).Interface()) {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("Expected `%v` to be in `%v`", v, slice)
		}

		return nil
	}
}
