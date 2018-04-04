// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		IP:   net.IP("8.8.8.8"),
		Port: 1700,
	}
	staticStore := StaticStore(addr)
	eui := types.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	staticStore.SetDownlinkAddress(eui, &net.UDPAddr{
		IP:   net.IP("8.8.2.2"),
		Port: 1701,
	})
	staticStore.SetUplinkAddress(eui, &net.UDPAddr{
		IP:   net.IP("8.8.2.2"),
		Port: 1702,
	})

	newAddr, found := staticStore.GetDownlinkAddress(eui)
	a.So(found, should.BeTrue)
	a.So(newAddr.String(), should.Equal, addr.String())
}
