// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDevAddr(t *testing.T) {
	a := assertions.New(t)

	devAddr := DevAddr{0x26, 0x12, 0x34, 0x56}
	a.So(devAddr.NwkID(), should.Equal, 0x13)

	prefix := DevAddrPrefix{DevAddr{0x26}, 7}
	a.So(prefix.Matches(devAddr), should.BeTrue)

	addr := DevAddr{1, 2, 3, 4}
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{0, 0, 0, 0}, 0}), should.BeTrue)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 2, 3, 0}, 24}), should.BeTrue)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{2, 2, 3, 4}, 31}), should.BeFalse)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 1, 3, 4}, 31}), should.BeFalse)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 1, 1, 1}, 15}), should.BeFalse)
}
