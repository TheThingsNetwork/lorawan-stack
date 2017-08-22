// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// Listen on a port
func Listen(addr string) (*Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &Conn{gatewayAddr: make(map[types.EUI64]net.UDPAddr), UDPConn: conn}, nil
}

// Conn wraps the net.UDPConn
type Conn struct {
	buf [65507]byte
	*net.UDPConn
	gatewayAddr map[types.EUI64]net.UDPAddr
}

// Read a packet from the conn
func (c *Conn) Read() (*Packet, error) {
	n, addr, err := c.UDPConn.ReadFromUDP(c.buf[:])
	if err != nil {
		return nil, err
	}
	packet := &Packet{
		GatewayConn: c,
		GatewayAddr: addr,
	}
	err = packet.UnmarshalBinary(c.buf[:n])
	return packet, err
}

// WriteTo writes a packet to the conn
func (c *Conn) WriteTo(packet *Packet, addr net.Addr) error {
	buf, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = c.UDPConn.WriteTo(buf, addr)
	return err
}
