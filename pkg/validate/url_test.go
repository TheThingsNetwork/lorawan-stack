// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestURL(t *testing.T) {
	a := assertions.New(t)

	a.So(URL("http:"), should.NotBeNil)
}
