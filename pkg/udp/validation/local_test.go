// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validation

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/udp"
)

func TestInvalidInMemory(t *testing.T) {
	v := InMemoryValidator(DefaultWaitDuration)
	var eui = new(types.EUI64)

	ip1 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket1 := udp.Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := &net.UDPAddr{IP: net.IP("8.8.4.4")}
	udpPacket2 := udp.Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	ip3 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket3 := udp.Packet{
		GatewayAddr: ip3,
		GatewayEUI:  eui,
	}

	if v.Valid(udpPacket1) == false {
		t.Error("First packet should be valid")
	}
	if v.Valid(udpPacket2) == true {
		t.Error("Second packet with the same EUI should be invalid")
	}
	if v.Valid(udpPacket3) == false {
		t.Error("Third packet should be valid, since it has the same IP as the first packet")
	}
}
func TestValidInMemory(t *testing.T) {
	v := InMemoryValidator(time.Duration(0))
	var eui = new(types.EUI64)

	ip1 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket1 := udp.Packet{
		GatewayAddr: ip1,
		GatewayEUI:  eui,
	}

	ip2 := &net.UDPAddr{IP: net.IP("8.8.4.4")}
	udpPacket2 := udp.Packet{
		GatewayAddr: ip2,
		GatewayEUI:  eui,
	}

	if v.Valid(udpPacket1) == false {
		t.Error("First packet should be valid")
	}
	if v.Valid(udpPacket2) == false {
		t.Error("Second packet with the same EUI should be valid")
	}
}

func TestInvalidPacket(t *testing.T) {
	v := InMemoryValidator(time.Duration(0))
	var eui = new(types.EUI64)

	ip1 := &net.UDPAddr{IP: net.IP("8.8.8.8")}
	udpPacket1 := udp.Packet{
		GatewayAddr: ip1,
	}

	udpPacket2 := udp.Packet{
		GatewayEUI: eui,
	}

	if v.Valid(udpPacket1) == true {
		t.Error("Packets with no EUI should be invalid")
	}
	if v.Valid(udpPacket2) == true {
		t.Error("Packets with no IP should be invalid")
	}
}
