// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	// pkg/errors[391]: You made a mistake cost us 7 dollars
	// namespace: pkg/errors
	// code: 391
}
