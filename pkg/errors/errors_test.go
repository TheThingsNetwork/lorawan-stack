// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNew(t *testing.T) {
	a := assertions.New(t)

	err := New("Something went wrong")

	a.So(err.Namespace(), should.Equal, "errors")
}
