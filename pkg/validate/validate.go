// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
