// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"fmt"
	"regexp"
)

var idRegex = regexp.MustCompile("^[a-z0-9](?:[_-]?[a-z0-9]){1,35}$")

// ID checks whether the input value is a valid ID according:
//		- Length must be between 2 and 36
//		- It consists only of numbers, dashs, underscores and lowercase letters
//		- Must start by a number or lowercase letter
//		- It cannot match any of the blacklisted IDs
func ID(v interface{}) error {
	id, ok := v.(string)
	if !ok {
		return fmt.Errorf("ID validator: got %T instead of string", v)
	}

	if !idRegex.MatchString(id) {
		return fmt.Errorf("`%s` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers, dashes and underscores. It may not start or end with a dash or an underscore", id)

	}

	return nil
}
