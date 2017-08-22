// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestFormatTypes(t *testing.T) {
	a := assertions.New(t)

	format := "{foo} - {bar} - {nil} - {list} - {map} - {complex} - {ptr}"
	{
		val := 10
		res := Format(format, Attributes{
			"foo":     10,
			"bar":     "bar",
			"nil":     nil,
			"list":    []int{1, 2, 3},
			"map":     map[string]int{"ok": 1},
			"complex": 3 + 4i,
			"ptr":     &val,
		})
		a.So(res, assertions.ShouldEqual, "10 - bar - <nil> - [1 2 3] - map[ok:1] - (3+4i) - 10")
	}
}

func TestFormat(t *testing.T) {
	a := assertions.New(t)

	format := "Found {foo, plural, =0 {no foos} =1 {# foo} other {# foos}}"
	{
		res := Format(format, Attributes{
			"foo": 1,
		})
		a.So(res, assertions.ShouldEqual, "Found 1 foo")
	}
	{
		res := Format(format, Attributes{
			"foo": 0,
		})
		a.So(res, assertions.ShouldEqual, "Found no foos")
	}
}
