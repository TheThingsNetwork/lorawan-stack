// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestEmail(t *testing.T) {
	a := assertions.New(t)
	a.So(Email("daniel@daniel.me"), should.BeNil)
	a.So(Email(1), should.NotBeNil)
	a.So(Email("root@localhost"), should.BeNil)
	a.So(Email("rootlocalhost.com"), should.NotBeNil)
}
