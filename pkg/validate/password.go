// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"errors"
	"fmt"
	"regexp"
)

var passwordRegex = regexp.MustCompile("^.{8,}$")

// Password checks whether the input value is a string and is at least 8 characters long.
func Password(v interface{}) error {
	password, ok := v.(string)
	if !ok {
		return fmt.Errorf("Password validator: got %T instead of string", v)
	}

	if !passwordRegex.MatchString(password) {
		return errors.New("Password must be at least 8 characters long")
	}

	return nil
}
