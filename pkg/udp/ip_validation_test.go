// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestAlwaysValid(t *testing.T) {
	validator := AlwaysValid()

	a := assertions.New(t)
	a.So(validator.Valid(Packet{}), should.BeTrue)
}
