// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"fmt"
	"net/url"
)

func URL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("Invalid input type, got %T instead of string", v)
	}

	_, err := url.ParseRequestURI(str)
	if err != nil {
		return fmt.Errorf("`%s` is not a valid URL", str)
	}

	return nil
}
