// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"strconv"
)

// Code represents a unique error code
type Code uint32

// NoCode is a missing code
const NoCode Code = 0

// String implmenents stringer
func (c Code) String() string {
	return fmt.Sprintf("%v", uint32(c))
}

// pareCode parses a string into a Code or returns 0 if the parse failed
func parseCode(str string) Code {
	code, err := strconv.Atoi(str)
	if err != nil {
		return Code(0)
	}
	return Code(code)
}

// Range is a utility function that creates a code builder.
//
// Example:
//	var code = Range(1000, 2000)
//  var ErrSomethingWasWrong := &ErrDescriptor{
//		// ...
//		Code: code(77),
//  }
//
// This can be used to create disjunct code ranges and be strict about it.
// The codes created by the returned function will range from start (inclusive)
// to end (exclusive) or the function will panic otherwise.
func Range(start uint32, end uint32) func(uint32) Code {
	if end <= start {
		panic("Range end <= start")
	}

	return func(i uint32) Code {
		if i >= (end - start) {
			panic(fmt.Sprintf("Code %v does not fit in range [%v, %v[", i, start, end))
		}

		return Code(start + i)
	}
}
