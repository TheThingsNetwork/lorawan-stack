// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"github.com/TheThingsNetwork/ttn/pkg/random"
)

// pad pads a byte slice with random bytes until it has the desired length.
// The slice is prefixed with the length of the original byte slice to allow for unpadding so the length
// of the resulting slice will be to + 1.
// If the length of the byte slice is bigger than the desired length, it is returned unaltered.
func pad(a []byte, to int) []byte {
	if len(a) > to {
		return append([]byte{byte(len(a))}, a...)
	}

	res := make([]byte, to+1)
	res[0] = byte(len(a))

	for i := range a {
		if to > i {
			res[i+1] = a[i]
		}
	}

	if len(res) >= len(a)+1 {
		random.FillBytes(res[len(a)+1:])
	}

	return res
}

// unpad removes the random padding added by pad, or nil if the
// input is invalid.
func unpad(a []byte) []byte {
	length := int(a[0])

	if length > len(a)+1 {
		return nil
	}

	res := make([]byte, length)

	for i := range res {
		res[i] = a[i+1]
	}

	return res
}
