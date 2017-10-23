// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPassword(t *testing.T) {
	a := assertions.New(t)
	a.So(Password("_Foo1__BaR"), should.BeNil)
	a.So(Password("_Foo1."), should.NotBeNil)
	a.So(Password("hhHiHIHIii1555"), should.BeNil)
	a.So(Password("Hi12//i12ddddd"), should.BeNil)
	a.So(Password(1), should.NotBeNil)
	a.So(Password(""), should.NotBeNil)
}
