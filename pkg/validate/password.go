// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"errors"
	"fmt"
	"unicode"
)

var mustHave = []func(rune) bool{
	unicode.IsUpper,
	unicode.IsLower,
	unicode.IsDigit,
}

var errInvalidPassword = errors.New("Password must be at least 8 characters long and have contain at least one lowercase letter, one uppercase letter and one digit")

// Password checks wether the input value is a string and a valid password according:
//		-  Length must be 8 at least
//		- It must contain at least a lower case letter
//		- It must contain at least an upper case letter
//		- It must contain at least one digit
func Password(v interface{}) error {
	password, ok := v.(string)
	if !ok {
		return fmt.Errorf("Invalid input type, got %T instead of string", v)
	}

	if len(password) < 8 {
		return errInvalidPassword
	}

	for _, fn := range mustHave {
		found := false

		for _, c := range password {
			if fn(c) {
				found = true
				break
			}
		}

		if !found {
			return errInvalidPassword
		}
	}

	return nil
}
