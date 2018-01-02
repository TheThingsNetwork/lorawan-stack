// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestID(t *testing.T) {
	a := assertions.New(t)
	a.So(ID("app-test"), should.BeNil)
	a.So(ID("_dd"), should.NotBeNil)
	a.So(ID("A"), should.NotBeNil)
	a.So(ID(1), should.NotBeNil)
	a.So(ID("dddsdddjjjjđsdddsdsdsdsdsdsdsdsdfddfdfsuif"), should.NotBeNil)
}
