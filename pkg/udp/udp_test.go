// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp_test

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/udp"
)

var (
	downlink ttnpb.DownlinkMessage

	converter    = dummyConverter{}
	gatewayStore = udp.NewGatewayStore(udp.DefaultWaitDuration)
)

type dummyConverter struct{}

func (d dummyConverter) RxPacket(p udp.RxPacket) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}
func (d dummyConverter) Status(p udp.Stat) (ttnpb.GatewayStatus, error) {
	return ttnpb.GatewayStatus{}, nil
}
func (d dummyConverter) TxPacket(downlink ttnpb.DownlinkMessage) (udp.Packet, error) {
	return udp.Packet{}, nil
}
func (d dummyConverter) TxPacketAck(p udp.TxPacketAck) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}

func Example() {
	conn, err := udp.Listen(":1700", gatewayStore, gatewayStore)
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
					uplink, err := converter.RxPacket(*packet)
					if err != nil {
						continue
					}

					Forward(uplink)
				}
				if packet.Data.Stat != nil {
					status, err := converter.Status(*packet.Data.Stat)
					if err != nil {
						continue
					}

					Forward(status)
				}
			case udp.TxAck:
				// Handle the data
				if packet.Data.TxPacketAck != nil {
					txInfo, err := converter.TxPacketAck(*packet.Data.TxPacketAck)
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
		packet, err := converter.TxPacket(downlink)
		if err != nil {
			fmt.Println("Couldn't convert TX packet")
			return
		}

		if err := conn.Send(&packet); err != nil {
			fmt.Println("Error when sending downlink: ", err)
		}
	}()
}

func Forward(interface{}) {}
