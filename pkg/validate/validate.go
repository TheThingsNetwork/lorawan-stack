// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

type validateFn func(v interface{}) error

// All concatenates all the errors of the different fields validations and returns it.
func All(results ...Errors) error {
	errors := make(Errors, 0, 0)
	for _, result := range results {
		if len(result) != 0 {
			errors = append(errors, result)
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// Field applies to the input value all the provided validator functions and
// returns the concatenation of the returned errors.
func Field(v interface{}, fns ...validateFn) Errors {
	errors := make(Errors, 0)
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
