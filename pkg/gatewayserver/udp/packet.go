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
	"encoding/json"
	"io"
	"net"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// ProtocolVersion of the forwarder
type ProtocolVersion uint8

// ProtocolVersions that we know
const (
	Version1 ProtocolVersion = 0x01
	Version2 ProtocolVersion = 0x02
)

// PacketType of the packet
type PacketType uint8

// PacketTypes that we know
const (
	PushData PacketType = iota
	PushAck
	PullData
	PullResp
	PullAck
	TxAck
)

// HasGatewayEUI returns true iff a packet of this type has a GatewayEUI field
func (p PacketType) HasGatewayEUI() bool {
	switch p {
	case PushData, PullData, TxAck:
		return true
	}
	return false
}

// HasData returns true iff a packet of this type has a Data field
func (p PacketType) HasData() bool {
	switch p {
	case PushData, PullResp, TxAck:
		return true
	}
	return false
}

// String implements the fmt.Stringer interface
func (p PacketType) String() string {
	switch p {
	case PushData:
		return "PushData"
	case PushAck:
		return "PushAck"
	case PullData:
		return "PullData"
	case PullResp:
		return "PullResp"
	case PullAck:
		return "PullAck"
	case TxAck:
		return "TxAck"
	}
	return "?"
}

// Packet struct
type Packet struct {
	ProtocolVersion ProtocolVersion
	Token           [2]byte
	PacketType      PacketType
	GatewayEUI      *types.EUI64
	Data            *Data

	GatewayAddr *net.UDPAddr
	GatewayConn *Conn
}

// BuildAck builds the corresponding Ack back to the gateway. If the received packet does not require an Ack, the function returns nil without an error.
func (p Packet) BuildAck() *Packet {
	ack := Packet{
		ProtocolVersion: p.ProtocolVersion,
		Token:           p.Token,
	}
	switch p.PacketType {
	case PushData:
		ack.PacketType = PushAck
	case PullData:
		ack.PacketType = PullAck
	default:
		return nil
	}

	return &ack
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler
func (p *Packet) UnmarshalBinary(b []byte) (err error) {
	if len(b) < 4 {
		return io.EOF
	}
	p.ProtocolVersion = ProtocolVersion(b[0])
	copy(p.Token[:], b[1:3])
	p.PacketType = PacketType(b[3])

	i := 4 // keep track of our position in the slice

	if p.PacketType.HasGatewayEUI() {
		if len(b) < i+8 {
			return errTooSmallToHaveGatewayEUI
		}

		p.GatewayEUI = new(types.EUI64)
		err = p.GatewayEUI.UnmarshalBinary(b[i : i+8])
		if err != nil {
			return errUnmarshalEUI.WithCause(err)
		}
		i += 8
	}

	if p.PacketType.HasData() {
		p.Data = new(Data)
		err = json.Unmarshal(b[i:], p.Data)
		if err != nil {
			return err
		}
	}

	return
}

// MarshalBinary implements the encoding.BinaryMarshaler
func (p Packet) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	b[0] = byte(p.ProtocolVersion)
	copy(b[1:3], p.Token[:])
	b[3] = byte(p.PacketType)
	if p.PacketType.HasGatewayEUI() && p.GatewayEUI != nil {
		b = append(b, p.GatewayEUI[:]...)
	}
	if p.PacketType.HasData() && p.Data != nil {
		data, _ := json.Marshal(p.Data)
		b = append(b, data...)
	}
	return b, nil
}

// Extractor is an interface representing objects capable of transforming packets between the Semtech structures format, and the TTN protobuf format.
//
// The TTN representation of those objects requires information from foreign actors, such as the UplinkMessage.TxSettings.DataRateIndex value - implementation of this Extracter should thus depend on the context of use.
type Extractor interface {
	RxPacket(p RxPacket) (ttnpb.UplinkMessage, error)
	Status(p Stat) (ttnpb.GatewayStatus, error)
	TxPacket(downlink ttnpb.DownlinkMessage) (Packet, error)
	TxPacketAck(p TxPacketAck) (ttnpb.UplinkMessage, error)
}
