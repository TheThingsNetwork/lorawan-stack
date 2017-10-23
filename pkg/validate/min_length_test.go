// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMinLength(t *testing.T) {
	a := assertions.New(t)

	minLength2 := MinLength(2)

	a.So(minLength2(""), should.NotBeNil)
	a.So(minLength2("a"), should.NotBeNil)
	a.So(minLength2("aa"), should.BeNil)

	var slice []int
	a.So(minLength2(slice), should.NotBeNil)
	slice = append(slice, 1)
	a.So(minLength2(slice), should.NotBeNil)
	slice = append(slice, 2)
	a.So(minLength2(slice), should.BeNil)

	a.So(minLength2(nil), should.NotBeNil)
	a.So(minLength2(1), should.NotBeNil)
}
