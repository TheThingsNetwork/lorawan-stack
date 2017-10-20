// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import "fmt"

// Errors is a slice of errors. It it is used to concatenate the different
// validation errors than can happen in a single field.
type Errors []error

// Error implements error.
func (e Errors) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		/*msg := "multiple errors:"
		for _, err := range e {
			msg += "\n"
			if _, ok := err.(Errors); ok {
				msg += "\n\t" + err.Error()
			} else {
				msg += "\t" + err.Error()
			}
		}
		return msg*/
		msg := e[0].Error()
		for i := 1; i < len(e); i++ {
			msg += "\n" + e[i].Error()
		}
		return msg
	}
}

// DescribeFieldName allows to prefix the errors with the name of the field.
func (e Errors) DescribeFieldName(fieldName string) Errors {
	if len(e) == 0 {
		return nil
	}
	for i := 0; i < len(e); i++ {
		e[i] = fmt.Errorf("%s: %s", fieldName, e[i].Error())
	}
	return e
}
