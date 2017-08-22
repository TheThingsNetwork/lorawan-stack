// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestInvalidInMemory(t *testing.T) {
	v := NewGatewayStore(DefaultWaitDuration)
	var eui = new(types.EUI64)

	ip1 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket1 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := &net.UDPAddr{IP: net.IP("8.8.4.4")}
	udpPacket2 := Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	ip3 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket3 := Packet{
		GatewayAddr: ip3,
		GatewayEUI:  eui,
	}

	ip4 := &net.UDPAddr{IP: net.IP("192.168.0.1")}
	var eui2 = types.EUI64([8]byte{1, 2, 3, 4, 6, 7, 2, 1})
	udpPacket4 := Packet{
		GatewayAddr: ip4,
		GatewayEUI:  &eui2,
	}

	v.Set(*eui, ip1)

	a := assertions.New(t)
	a.So(v.Valid(udpPacket1), should.BeTrue)  // First packet should be valid
	a.So(v.Valid(udpPacket2), should.BeFalse) // Second packet with the same EUI should be invalid, since it has a different IP but the same EUI
	a.So(v.Valid(udpPacket3), should.BeTrue)  // Third packet should be valid, since it has the same IP as the first packet
	a.So(v.Valid(udpPacket4), should.BeTrue)  // Fourth packet should be valid, since it has an unset EUI
}

func TestDataCoherence(t *testing.T) {
	v := NewGatewayStore(time.Duration(0))
	eui := types.EUI64([8]byte{1, 2, 3, 4, 0, 9, 8, 7})
	eui2 := types.EUI64([8]byte{2, 4, 9, 3, 5, 9, 8, 9})
	addr := &net.UDPAddr{
		IP: net.IP("8.8.8.8"),
	}
	v.Set(eui, addr)
	addr2, found := v.Get(eui)
	_, found2 := v.Get(eui2)

	a := assertions.New(t)
	a.So(found, should.BeTrue)      // Address associated to that EUI should be found
	a.So(found2, should.BeFalse)    // Address associated to that EUI should not be found
	a.So(addr2, should.Equal, addr) // The two addresses should be identical
}
func TestValidInMemory(t *testing.T) {
	v := NewGatewayStore(time.Duration(0))
	var eui = new(types.EUI64)

	ip1 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket1 := Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := &net.UDPAddr{IP: net.IP("8.8.4.4")}
	udpPacket2 := Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	v.Set(*eui, ip1)

	a := assertions.New(t)
	a.So(v.Valid(udpPacket1), should.BeTrue) // First packet should be valid
	a.So(v.Valid(udpPacket2), should.BeTrue) // Second packet with the same EUI should be valid, since expiration is set to 0seconds
}

func TestInvalidPacket(t *testing.T) {
	v := NewGatewayStore(time.Duration(0))
	var eui = new(types.EUI64)

	ip1 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket1 := Packet{
		GatewayAddr: ip1,
	}

	udpPacket2 := Packet{
		GatewayEUI: eui,
	}

	a := assertions.New(t)
	a.So(v.Valid(udpPacket1), should.BeFalse) // Packets with no EUI should be invalid
	a.So(v.Valid(udpPacket2), should.BeFalse) // Packets with no IP should be invalid
}
