// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package utils

// Float32Address returns the address of a float32 literal.
func Float32Address(f float32) *float32 {
	return &f
}

// Int32Address returns the address of a int32 literal.
func Int32Address(i int32) *int32 {
	return &i
}

// StringAddress returns the address of a string literal.
func StringAddress(s string) *string {
	return &s
}
