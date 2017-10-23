// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"fmt"
	"regexp"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Email checks whether the input value is a valid email or not.
func Email(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("Email validator: got %T instead of string", v)
	}

	if !emailRegex.MatchString(str) {
		return fmt.Errorf("`%s` is not a valid email", str)
	}

	return nil
}
