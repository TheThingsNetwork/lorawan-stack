// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestRange(t *testing.T) {
	a := assertions.New(t)
	code := Range(1000, 2000)

	a.So(code(0), assertions.ShouldEqual, 1000)
	a.So(code(1), assertions.ShouldEqual, 1001)

	a.So(func() { code(1000) }, assertions.ShouldPanic)
	a.So(func() { code(1001) }, assertions.ShouldPanic)
	a.So(func() { _ = Range(2, 1) }, assertions.ShouldPanic)
}
