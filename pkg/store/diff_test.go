// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDiff(t *testing.T) {
	a := assertions.New(t)
	old := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	new := map[string]interface{}{
		"foo": "baz",
		"bar": "bar",
		"qux": "qux",
	}
	a.So(Diff(new, old), should.Resemble, map[string]interface{}{
		"foo": "baz", // new value updated
		"qux": "qux", // new value added
		// bar unchanged
		"baz": nil, // old value removed
	})
}
