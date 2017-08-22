// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestStaticStore(t *testing.T) {
	a := assertions.New(t)

	addr := &net.UDPAddr{
		IP: net.IP("8.8.8.8"),
	}
	staticStore := StaticStore(addr)
	eui := types.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	staticStore.Set(eui, &net.UDPAddr{
		IP: net.IP("8.8.2.2"),
	})

	newAddr, found := staticStore.Get(eui)
	a.So(found, should.BeTrue)
	a.So(newAddr.String(), should.Equal, addr.String())
}
