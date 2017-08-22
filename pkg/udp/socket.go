// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/pkg/errors"
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

// Ack sends the corresponding Ack back to the gateway
func (p *Packet) Ack() error {
	if p.GatewayConn == nil || p.GatewayAddr == nil {
		return errors.New("No gateway connection associated to this packet")
	}

	ackPacket, err := p.BuildAck()
	if err != nil {
		return errors.Wrap(err, "failed to build ack package")
	}
	if ackPacket == nil {
		return nil
	}

	binaryAckPacket, err := ackPacket.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "failed to convert ack packet to binary format")
	}

	_, err = p.GatewayConn.WriteToUDP(binaryAckPacket, p.GatewayAddr)
	return err
}
