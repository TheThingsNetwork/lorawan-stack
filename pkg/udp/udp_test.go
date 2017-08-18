// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp_test

import (
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/udp"
)

var downlink *udp.Packet

func ExampleServer() {
	conn, err := udp.Listen(":1700")
	if err != nil {
		panic(err)
	}

	go func() {
		var packet *udp.Packet
		for {
			packet, err = conn.Read()
			if err != nil {
				panic(err)
			}
			err = packet.Ack()
			if err != nil {
				panic(err)
			}

			switch packet.PacketType {
			case udp.PushData:
				// Verify the source address of the packet against the "IP-lock" policy
				VerifySourceAddress(packet.GatewayEUI, conn.RemoteAddr())
				if packet.Data == nil {
					continue
				}
				// Handle the data
				for _, packet := range packet.Data.RxPacket {
					uplink := Convert(packet)
					Forward(uplink)
				}
				if packet.Data.Stat != nil {
					status := Convert(packet)
					Forward(status)
				}
			case udp.PullData:
				// Verify the source address of the packet against the "IP-lock" policy
				VerifySourceAddress(packet.GatewayEUI, conn.RemoteAddr())
				// Update the downlink address in the mapping
				UpdateDownlinkAddress(packet.GatewayEUI, conn.RemoteAddr())
			case udp.TxAck:
				// Verify the source address of the packet against the "IP-lock" policy
				VerifySourceAddress(packet.GatewayEUI, conn.RemoteAddr())
				// Handle the data
				if packet.Data.TxPacketAck != nil {
					txInfo := Convert(packet)
					Forward(txInfo)
				}
			}
		}
	}()

	go func() {
		packet := Convert(downlink)
		conn.WriteTo(packet, GetDownlinkAddress(downlink.GatewayEUI))
	}()
}

func VerifySourceAddress(interface{}, interface{})   {}
func UpdateDownlinkAddress(interface{}, interface{}) {}
func GetDownlinkAddress(interface{}) net.Addr        { return nil }
func Convert(interface{}) *udp.Packet                { return nil }
func Forward(interface{})                            {}
