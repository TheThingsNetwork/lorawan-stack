// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
)

// MinLength checks whether the input value has a minimum length in the following cases:
//     - For strings checks the length
//     - For slices checks that aren't nil and its length
//     - For other types returns error
func MinLength(length int) validateFn { // nolint: golint, returns unexported type on purpose
	return func(v interface{}) error {
		if v == nil {
			return fmt.Errorf("MinLength validator: got %T instead of string or slice", v)
		}

		typ := reflect.TypeOf(v)

		if typ.Kind() == reflect.String {
			str, _ := v.(string)
			return minLengthString(str, length)
		}

		if typ.Kind() == reflect.Slice {
			return minLengthSlice(reflect.ValueOf(v), length)
		}

		return fmt.Errorf("MinLength validator: got %T instead of string or slice", v)
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
