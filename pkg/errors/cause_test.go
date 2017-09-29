// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
)

func ExampleNewWithCause() {
	cause := fmt.Errorf("Underlying cause")
	err := NewWithCause("Something went wrong!", cause)
	fmt.Println(err)
	// Output: errors: Something went wrong! (Underlying cause)
}
