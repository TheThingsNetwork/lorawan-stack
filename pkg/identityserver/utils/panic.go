// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Copyright © 2017 The Things Industries B.V.

package utils

// Panic panics if the error is not nil
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}
