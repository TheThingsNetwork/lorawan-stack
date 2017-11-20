// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestIn(t *testing.T) {
	a := assertions.New(t)

	in := In([]int{0, 1, 2})

	a.So(in(3), should.NotBeNil)
	a.So(in(""), should.NotBeNil)
	a.So(in(0), should.BeNil)
	a.So(in([]int{0, 1}), should.BeNil)
	a.So(in([]int{2, 3, 4, 5}), should.NotBeNil)
}
