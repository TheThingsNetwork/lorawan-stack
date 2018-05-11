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
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func newUDPAddress(ip string, port int) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IP(ip),
		Port: port,
	}
}

func TestInvalidUplinkGatewayStore(t *testing.T) {
	v := NewGatewayStore(DefaultWaitDuration)
	var eui = new(types.EUI64)

	ip1 := newUDPAddress("8.8.8.8", 1700)
	udpPacket1 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := newUDPAddress("8.8.4.4", 1700)
	udpPacket2 := Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	udpPacket3 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip3 := newUDPAddress("192.168.1.1", 1700)
	var eui2 = types.EUI64([8]byte{1, 2, 3, 4, 6, 7, 2, 1})
	udpPacket4 := Packet{
		GatewayAddr: ip3,
		GatewayEUI:  &eui2,
	}

	v.SetUplinkAddress(*eui, ip1)
	v.SetDownlinkAddress(*eui, ip1)

	a := assertions.New(t)
	a.So(v.ValidUplink(udpPacket1), should.BeTrue)  // First packet should be valid
	a.So(v.ValidUplink(udpPacket2), should.BeFalse) // Second packet with the same EUI should be invalid, since it has a different IP but the same EUI
	a.So(v.ValidUplink(udpPacket3), should.BeTrue)  // Third packet should be valid, since it has the same IP as the first packet
	a.So(v.ValidUplink(udpPacket4), should.BeTrue)  // Fourth packet should be valid, since it has an unseen EUI
}

func TestInvalidIPv6UplinkGatewayStore(t *testing.T) {
	v := NewGatewayStore(DefaultWaitDuration)
	var eui = new(types.EUI64)
	eui2 := types.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	ip1 := newUDPAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334", 1700)
	udpPacket1 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := newUDPAddress("0db8:a32e:0890:3a1d:0000:8a2e:0370:7334", 1700)
	udpPacket2 := Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	udpPacket3 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	udpPacket4 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  &eui2,
	}

	v.SetUplinkAddress(*eui, ip1)

	a := assertions.New(t)
	a.So(v.ValidUplink(udpPacket1), should.BeTrue)  // First packet should be valid
	a.So(v.ValidUplink(udpPacket2), should.BeFalse) // Second packet with the same EUI should be invalid, since it has a different IP but the same EUI
	a.So(v.ValidUplink(udpPacket3), should.BeTrue)  // Third packet should be valid, since it has the same IP as the first packet
	a.So(v.ValidUplink(udpPacket4), should.BeTrue)  // Fourth packet should be valid since the EUI has not been seen before

	v.SetUplinkAddress(*eui, ip2)

	a.So(v.ValidUplink(udpPacket2), should.BeTrue) // Second packet with the same EUI should be invalid, since it has a different IP but the same EUI
}

func TestInvalidIPv6DownlinkGatewayStore(t *testing.T) {
	v := NewGatewayStore(DefaultWaitDuration)
	var eui = new(types.EUI64)
	eui2 := types.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	ip1 := newUDPAddress("2001:0db8:85a3:0000:0000:8a2e:0370:7334", 1700)
	udpPacket1 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := newUDPAddress("0db8:a32e:0890:3a1d:0000:8a2e:0370:7334", 1700)
	udpPacket2 := Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	udpPacket3 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  &eui2,
	}

	v.SetDownlinkAddress(*eui, ip2)

	a := assertions.New(t)
	a.So(v.ValidDownlink(udpPacket1), should.BeFalse) // First packet should be invalid, since it has a different IP but the same EUI
	a.So(v.ValidDownlink(udpPacket2), should.BeTrue)  // Second packet with the same EUI should be valid
	a.So(v.ValidDownlink(udpPacket3), should.BeTrue)  // Third packet should be valid since the EUI is unseen
}

func TestDataCoherence(t *testing.T) {
	v := NewGatewayStore(time.Duration(0))
	eui := types.EUI64([8]byte{1, 2, 3, 4, 0, 9, 8, 7})
	eui2 := types.EUI64([8]byte{2, 4, 9, 3, 5, 9, 8, 9})
	addr := newUDPAddress("8.8.8.8", 1700)
	v.SetDownlinkAddress(eui, addr)
	addr2, found := v.GetDownlinkAddress(eui)
	_, found2 := v.GetDownlinkAddress(eui2)

	a := assertions.New(t)
	a.So(found, should.BeTrue)      // Address associated to that EUI should be found
	a.So(found2, should.BeFalse)    // Address associated to that EUI should not be found
	a.So(addr2, should.Equal, addr) // The two addresses should be identical
}
func TestValidInMemory(t *testing.T) {
	v := NewGatewayStore(time.Duration(0))
	var eui = new(types.EUI64)

	ip1 := newUDPAddress("8.8.8.8", 1700)
	udpPacket1 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := newUDPAddress("8.8.4.4", 1700)
	udpPacket2 := Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	v.SetUplinkAddress(*eui, ip1)

	a := assertions.New(t)
	a.So(v.ValidUplink(udpPacket1), should.BeTrue)   // First packet should be valid
	a.So(v.ValidUplink(udpPacket2), should.BeTrue)   // Second packet with the same EUI should be valid, since expiration is set to 0seconds
	a.So(v.ValidDownlink(udpPacket1), should.BeTrue) // First packet should be valid
	a.So(v.ValidDownlink(udpPacket2), should.BeTrue) // Second packet with the same EUI should be valid, since expiration is set to 0seconds
}

func TestInvalidUDPPackets(t *testing.T) {
	v := NewGatewayStore(time.Duration(0))
	var eui = new(types.EUI64)

	ip1 := newUDPAddress("8.8.8.8", 1700)
	udpPacket1 := Packet{
		GatewayAddr: ip1,
	}

	udpPacket2 := Packet{
		GatewayEUI: eui,
	}

	a := assertions.New(t)
	a.So(v.ValidUplink(udpPacket1), should.BeFalse)   // Packets with no EUI should be invalid
	a.So(v.ValidUplink(udpPacket2), should.BeFalse)   // Packets with no IP should be invalid
	a.So(v.ValidDownlink(udpPacket2), should.BeFalse) // Packets with no IP should be invalid
}
