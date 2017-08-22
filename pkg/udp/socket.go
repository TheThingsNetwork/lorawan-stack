// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

import (
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/pkg/errors"
)

// ErrGatewayNotConnected is returned when trying to send a packet to a
// udp.Conn that has never interacted with the packet's gateway.
var ErrGatewayNotConnected = errors.New("Not connected to the specified gateway")

// Listen on a port. Requires:
//
// - A Validator object, that represents application-level validation of an
// incoming packet. When polling for packets, invalid packets won't be
// transmitted.
//
// - An AddressStore object, used to store gateway IP addresses, and to
// retrieve the addresses when sending downlinks.
func Listen(addr string, validator Validator, addrStore AddressStore) (*Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &Conn{
		addrStore:   addrStore,
		gatewayAddr: make(map[types.EUI64]net.UDPAddr),
		UDPConn:     conn,
		validator:   validator,
	}, nil
}

// Conn wraps the net.UDPConn
type Conn struct {
	buf [65507]byte
	*net.UDPConn
	gatewayAddr map[types.EUI64]net.UDPAddr
	addrStore   AddressStore
	validator   Validator
}

// Read a packet from the conn. The poll system ignores UDP packets that are not
// compliant to the Semtech network protocol, and only returns once a valid
// Semtech-protocol packet has been received, or that there was an error when
// interacting with the socket.
func (c *Conn) Read() (*Packet, error) {
	for {
		n, addr, err := c.UDPConn.ReadFromUDP(c.buf[:])
		if err != nil {
			return nil, err
		}

		packet := &Packet{
			GatewayConn: c,
			GatewayAddr: addr,
		}

		err = packet.UnmarshalBinary(c.buf[:n])
		if err != nil {
			continue
		}

		if !c.validator.Valid(*packet) {
			continue
		}
		c.addrStore.Set(*packet.GatewayEUI, addr)

		return packet, nil
	}
}

// Write sends a packet. It returns ErrGatewayNotConnected if the gateway was
// never connected, and otherwise, returns the result of the marshalling and
// socket operation.
func (c *Conn) Write(packet *Packet) error {
	buf, err := packet.MarshalBinary()
	if err != nil {
		return err
	}

	addr, hasAddr := c.addrStore.Get(*packet.GatewayEUI)
	if !hasAddr {
		return ErrGatewayNotConnected
	}

	_, err = c.UDPConn.WriteTo(buf, addr)
	return err
}

// Ack builds the corresponding Ack packet, using p.BuildAck(), and sends it to
// the gateway.
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
