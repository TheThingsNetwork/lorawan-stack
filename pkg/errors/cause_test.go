// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
)

func ExampleNewWithCause() {
	cause := fmt.Errorf("Underlying cause")
	err := NewWithCause(cause, "Something went wrong!")
	fmt.Println(err)
	err = NewWithCausef(cause, "Something went wrong for the %s time!", "second")
	fmt.Println(err)
	// Output: errors: Something went wrong! (Underlying cause)
	// errors: Something went wrong for the second time! (Underlying cause)
}
