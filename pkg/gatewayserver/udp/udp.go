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

/*
Package udp contains a framework for interacting with UDP Semtech packet
forwarders. It provides structs and methods to do so, as well as some other
structs of convenience.
*/
package udp

import (
	"net"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// ErrGatewayNotConnected is returned when trying to send a packet to a udp.Conn that has never interacted with the packet's gateway.
var ErrGatewayNotConnected = errors.New("Not connected to the specified gateway")

// Handle returns a UDP packet socket from a raw UDP socket. Requires:
//
// - A Validator object, that represents application-level validation of an incoming packet. When polling for packets, invalid packets won't be transmitted.
//
// - An AddressStore object, used to store gateway IP addresses, and to retrieve the addresses when sending downlinks.
func Handle(conn *net.UDPConn, validator Validator, addrStore AddressStore) *Conn {
	return &Conn{
		addrStore:   addrStore,
		gatewayAddr: make(map[types.EUI64]net.UDPAddr),
		UDPConn:     conn,
		validator:   validator,
	}
}

// Conn wraps the net.UDPConn
type Conn struct {
	buf [65507]byte
	*net.UDPConn
	gatewayAddr map[types.EUI64]net.UDPAddr
	addrStore   AddressStore
	validator   Validator
}

// Read a packet from the conn. The poll system ignores UDP packets that are not compliant to the Semtech network protocol, and only returns once a valid Semtech-protocol packet has been received, or that there was an error when interacting with the socket.
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

		switch packet.PacketType {
		case PullData, TxAck:
			if c.validator != nil && !c.validator.ValidDownlink(*packet) {
				continue
			}
			c.addrStore.SetDownlinkAddress(*packet.GatewayEUI, addr)
		case PushData:
			if c.validator != nil && !c.validator.ValidUplink(*packet) {
				continue
			}
			c.addrStore.SetUplinkAddress(*packet.GatewayEUI, addr)
		}

		return packet, nil
	}
}

// Write sends a packet. It returns ErrGatewayNotConnected if the gateway was never connected, and otherwise, returns the result of the marshalling and socket operation.
func (c *Conn) Write(packet *Packet) error {
	buf, err := packet.MarshalBinary()
	if err != nil {
		return ErrMarshalFailed.New(nil)
	}

	addr, hasAddr := c.addrStore.GetDownlinkAddress(*packet.GatewayEUI)
	if !hasAddr || addr == nil {
		return ErrGatewayNotConnected
	}

	_, err = c.UDPConn.WriteTo(buf, addr)
	return err
}

// Ack builds the corresponding Ack packet, using p.BuildAck(), and sends it to the gateway.
func (p *Packet) Ack() error {
	if p.GatewayConn == nil || p.GatewayAddr == nil {
		return errors.New("No gateway connection associated to this packet")
	}

	ackPacket, err := p.BuildAck()
	if err != nil {
		return errors.NewWithCause(err, "failed to build ack package")
	}
	if ackPacket == nil {
		return nil
	}

	binaryAckPacket, err := ackPacket.MarshalBinary()
	if err != nil {
		return ErrMarshalFailed.NewWithCause(nil, err)
	}

	_, err = p.GatewayConn.WriteToUDP(binaryAckPacket, p.GatewayAddr)
	return err
}
