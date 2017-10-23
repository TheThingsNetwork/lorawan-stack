// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestIsZero(t *testing.T) {
	a := assertions.New(t)

	var slice []int
	a.So(isZero(slice), should.BeTrue)
	slice = append(slice, 1)
	a.So(isZero(slice), should.BeFalse)

	var mp map[string]string
	a.So(isZero(mp), should.BeTrue)
	mp = make(map[string]string)
	a.So(isZero(mp), should.BeFalse)

	type Bar struct {
		DevID string
	}
	var bar Bar
	a.So(isZero(bar), should.BeTrue)
	bar.DevID = "foo"
	a.So(isZero(bar), should.BeFalse)

	var str string
	a.So(isZero(str), should.BeTrue)
	str = "a"
	a.So(isZero(str), should.BeFalse)

	var b *Bar
	a.So(isZero(b), should.BeTrue)
	b = new(Bar)
	a.So(isZero(b), should.BeFalse)
}

func TestRequired(t *testing.T) {
	a := assertions.New(t)

	a.So(Field("", Required), should.NotBeNil)
	a.So(Field("f", Required), should.BeNil)

	a.So(Field("", NotRequired), should.BeNil)
	a.So(Field("f", NotRequired), should.BeNil)
}
