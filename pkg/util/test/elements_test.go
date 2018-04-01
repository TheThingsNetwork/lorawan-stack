// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSameElements(t *testing.T) {
	for _, tc := range []struct {
		A    interface{}
		B    interface{}
		Same bool
	}{
		{
			[][]byte{{42}, {43}},
			[][]byte{{43}, {44}},
			false,
		},
		{
			[][]byte{{43}, {43}},
			[][]byte{{43}, {44}},
			false,
		},
		{
			[][]byte{{43}, {43}, {43}},
			[][]byte{{43}, {44}},
			false,
		},
		{
			[][]byte{{42}, {43}, {43}},
			[][]byte{{43}, {42}, {43}},
			true,
		},
		{
			[][]byte{},
			[][]byte{{43}, {42}, {43}},
			false,
		},
		{
			[][]byte{{43}, {42}, {43}},
			[][]byte{},
			false,
		},
		{
			[]int{42},
			[]int{42},
			true,
		},
	} {
		a := assertions.New(t)

		a.So(SameElements(tc.A, tc.B), should.Equal, tc.Same)
	}
}
