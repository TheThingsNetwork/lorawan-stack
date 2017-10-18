// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var (
	length    = 10
	testcases = [][]byte{
		{},
		{0},
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 5},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	}
)

func TestPad(t *testing.T) {
	a := assertions.New(t)

	for _, original := range testcases {
		padded := pad(original, length)

		// first byte is the length
		a.So(int(padded[0]), should.Resemble, min(len(original), length))

		// total length should be what we asked for
		a.So(len(padded), should.Equal, length+1)

		// the original byte should be present
		m := min(length, len(original))
		a.So(padded[1:m+1], should.Resemble, original[:m])

		// the random bytes should be random
		random := padded[m+1:]
		if len(random) > 0 {
			a.So(random, should.NotResemble, make([]byte, len(random)))
		}

		// the unpadded slice should equal the (truncated) original slice
		unpadded := unpad(padded)
		a.So(unpadded, should.Resemble, original[:m])
	}
}
