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
	"math"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestPacket(t *testing.T) {
	eui := new(types.EUI64)
	data := new(Data)
	p := Packet{
		ProtocolVersion: Version1,
		Token:           [2]byte{},
		PacketType:      PushData,
		GatewayEUI:      eui,
		Data:            data,
	}

	res, err := p.MarshalBinary()
	if err != nil {
		t.Error("Failed to marshal packet:", err)
	}

	var p2 Packet
	err = p2.UnmarshalBinary(res)
	if err != nil {
		t.Error("Failed to unmarshal binary packet:", err)
	}

	p.BuildAck()
}

func TestFailedPackets(t *testing.T) {
	var p Packet

	a := assertions.New(t)

	b := []byte{}
	err := p.UnmarshalBinary(b)
	a.So(err, should.NotBeNil)

	b = []byte{0, 0, 0, 0}
	err = p.UnmarshalBinary(b)
	a.So(err, should.NotBeNil)

	b = []byte{}
	for i := 0; i < 12; i++ {
		b = append(b, 0)
	}
	err = p.UnmarshalBinary(b)
	a.So(err, should.NotBeNil)
}

func TestPacketType(t *testing.T) {
	a := assertions.New(t)

	eui := new(types.EUI64)
	data := new(Data)

	pTypes := []PacketType{PushAck, PushData, PullData, PullResp, PullAck, TxAck}
	for _, pType := range pTypes {
		a.So(pType.String(), should.NotEqual, "?")

		pType.HasGatewayEUI()
		pType.HasData()

		p := Packet{
			ProtocolVersion: Version1,
			Token:           [2]byte{},
			PacketType:      pType,
			GatewayEUI:      eui,
			Data:            data,
		}
		_, err := p.BuildAck()
		a.So(err, should.BeNil)
	}

	inexistantPacketType := PacketType(math.MaxUint8)
	a.So(inexistantPacketType.String(), should.Equal, "?")
}
