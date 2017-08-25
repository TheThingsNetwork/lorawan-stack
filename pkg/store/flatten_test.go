// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestFlatten(t *testing.T) {
	a := assertions.New(t)
	toFlatten := map[string]interface{}{
		"foo": "foo",
		"bar": map[string]interface{}{
			"baz": "baz",
			"bar": map[string]interface{}{
				"qux": "qux",
			},
		},
	}
	a.So(Flatten(toFlatten), should.Resemble, map[string]interface{}{
		"foo":         "foo",
		"bar.baz":     "baz",
		"bar.bar.qux": "qux",
	})
}
