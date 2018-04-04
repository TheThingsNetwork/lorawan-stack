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

type validateFn func(v interface{}) error

// All concatenates all the errors and appends them into an `Errors` type.
func All(results ...error) error {
	errors := make(Errors, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		if e, ok := result.(Errors); ok && len(e) == 0 {
			continue
		}

		errors = append(errors, result)
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// Field execute all the given validation functions on v and returns an `Errors`
// type containing all the errors of the validating functions.
func Field(v interface{}, fns ...validateFn) Errors {
	errors := make(Errors, 0, len(fns))
	for _, fn := range fns {
		err := fn(v)

		if err == errZeroValue {
			return nil
		}

		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}
