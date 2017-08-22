// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRange(t *testing.T) {
	a := assertions.New(t)
	code := Range(1000, 2000)

	a.So(code(0), should.Equal, 1000)
	a.So(code(1), should.Equal, 1001)

	a.So(func() { code(1000) }, should.Panic)
	a.So(func() { code(1001) }, should.Panic)
	a.So(func() { _ = Range(2, 1) }, should.Panic)
}
