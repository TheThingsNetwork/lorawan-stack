// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import "fmt"

func Example() {

	// ErrSomeUserMistake is the description of the error some user made that costs the company some money.
	// The namespace is left blank and will be filled in automatically based on the package name.
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

	// Register the error so others can find it based on the error Code and Namespace.
	// do this in the init() function of your package.
	ErrSomeUserMistake.Register()

	// Create a new error based on the descriptor
	err := ErrSomeUserMistake.New(Attributes{
		"price": 7,
	})

	// You can get the error descriptor back based on any error.
	_ = Descriptor(err)

	// This will print the formatted error message.
	fmt.Println(err.Error())
	fmt.Println("namespace:", err.Namespace())
	fmt.Println("code:", err.Code())
	// Output:
	// errors[391]: You made a mistake cost us 7 dollars
	// namespace: errors
	// code: 391
}
