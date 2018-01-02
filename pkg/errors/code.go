// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
)

// Code represents a unique error code
type Code uint32

// NoCode is a missing code
const NoCode Code = 0

// String implmenents stringer
func (c Code) String() string {
	return fmt.Sprintf("%v", uint32(c))
}
