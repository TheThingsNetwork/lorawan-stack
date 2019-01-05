// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validate

import (
	"fmt"
	"reflect"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errNoSliceOrArray = errors.DefineInvalidArgument("no_slice_or_array", "must be slice or array, is `{type}`")
	errNotPresent     = errors.DefineInvalidArgument("not_present", "`{value}` is not present in `{allowed_values}`")
	errNotSubset      = errors.DefineInvalidArgument("not_subset", "`{value}` is not a subset of `{reference}`")
)

// In checks whether:
//     - an element is contained in an array or slice
//     - a slice is a subset of other slice
func In(slice interface{}) validateFn { // nolint: golint, returns unexported type on purpose
	return func(v interface{}) error {
		sliceVal := reflect.ValueOf(slice)

		if sliceVal.Kind() != reflect.Slice && sliceVal.Kind() != reflect.Array {
			return errNoSliceOrArray.WithAttributes("type", fmt.Sprintf("%T", v))
		}

		fn := func(v interface{}, slice reflect.Value) error {
			var found bool
			for i := 0; i < slice.Len() && !found; i++ {
				found = reflect.DeepEqual(v, slice.Index(i).Interface())
			}

			if !found {
				return errNotPresent.WithAttributes("value", fmt.Sprintf("%v", v), "allowed_values", fmt.Sprintf("%v", slice))
			}

			return nil
		}

		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			if val.Len() > sliceVal.Len() {
				return errNotSubset.WithAttributes("value", fmt.Sprintf("%v", v), "reference", fmt.Sprintf("%v", slice))
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
