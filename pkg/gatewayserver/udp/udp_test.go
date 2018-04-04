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

package udp_test

import (
	"fmt"
	"net"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var (
	downlink ttnpb.DownlinkMessage

	extracter    = dummyExtractor{}
	gatewayStore = udp.NewGatewayStore(udp.DefaultWaitDuration)
)

type dummyExtractor struct{}

func (d dummyExtractor) RxPacket(p udp.RxPacket) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}
func (d dummyExtractor) Status(p udp.Stat) (ttnpb.GatewayStatus, error) {
	return ttnpb.GatewayStatus{}, nil
}
func (d dummyExtractor) TxPacket(downlink ttnpb.DownlinkMessage) (udp.Packet, error) {
	return udp.Packet{}, nil
}
func (d dummyExtractor) TxPacketAck(p udp.TxPacketAck) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}

func Example() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":1700")
	if err != nil {
		panic(err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}

	conn := udp.Handle(udpConn, gatewayStore, nil)

	go func() {
		var packet *udp.Packet
		for {
			packet, err = conn.Read()
			if err != nil {
				panic(err)
			}
			err := packet.Ack()
			if err != nil {
				panic(err)
			}

			switch packet.PacketType {
			case udp.PushData:
				if packet.Data == nil {
					continue
				}
				// Handle the data
				for _, packet := range packet.Data.RxPacket {
					uplink, err := extracter.RxPacket(*packet)
					if err != nil {
						continue
					}

					Forward(uplink)
				}
				if packet.Data.Stat != nil {
					status, err := extracter.Status(*packet.Data.Stat)
					if err != nil {
						continue
					}

					Forward(status)
				}
			case udp.TxAck:
				// Handle the data
				if packet.Data.TxPacketAck != nil {
					txInfo, err := extracter.TxPacketAck(*packet.Data.TxPacketAck)
					if err != nil {
						continue
					}

					Forward(txInfo)
				}
			}
		}
	}()

	go func() {
		time.Sleep(10 * time.Second)
		packet, err := extracter.TxPacket(downlink)
		if err != nil {
			fmt.Println("Couldn't convert Tx packet")
			return
		}

		if err := conn.Write(&packet); err != nil {
			fmt.Println("Error when sending downlink: ", err)
		}
	}()
}

func Forward(interface{}) {}
