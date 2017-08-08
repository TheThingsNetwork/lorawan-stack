// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNetID(t *testing.T) {
	a := assertions.New(t)

	a.So((NetID{0x12, 0x34, 0x56}).NwkID(), should.Equal, 0x56)
}
