// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import "fmt"

func Example() {

	// ErrSomeUserMistake is the description of the error some user made
	// that costs the company some money
	ErrSomeUserMistake := &ErrDescriptor{
		// MessageFormat is the format the error message will be using
		// It is written in ICU message format
		MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",

		// Type is the general category of the error (like HTTP status codes it puts
		// the error in a category that clients can understand without knowing the
		// error).
		Type: InvalidArgument,

		// Code is the unique code of this error
		Code: 391,
	}

	// register the error so others can find it based on the error Code
	ErrSomeUserMistake.Register()

	// Create a new error based on the descriptor
	err := ErrSomeUserMistake.New(Attributes{
		"price": 7,
	})

	// this will print the formatted error message
	fmt.Println(err)
	// Output: You made a mistake cost us 7 dollars

	// You can get the error descriptor back based on any error
	_ = Descriptor(err)
}
