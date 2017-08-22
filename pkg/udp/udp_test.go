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

	extracter    = dummyExtracter{}
	gatewayStore = udp.NewGatewayStore(udp.DefaultWaitDuration)
)

type dummyExtracter struct{}

func (d dummyExtracter) RxPacket(p udp.RxPacket) (ttnpb.UplinkMessage, error) {
	return ttnpb.UplinkMessage{}, nil
}
func (d dummyExtracter) Status(p udp.Stat) (ttnpb.GatewayStatus, error) {
	return ttnpb.GatewayStatus{}, nil
}
func (d dummyExtracter) TxPacket(downlink ttnpb.DownlinkMessage) (udp.Packet, error) {
	return udp.Packet{}, nil
}
func (d dummyExtracter) TxPacketAck(p udp.TxPacketAck) (ttnpb.UplinkMessage, error) {
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
			fmt.Println("Couldn't convert TX packet")
			return
		}

		if err := conn.Send(&packet); err != nil {
			fmt.Println("Error when sending downlink: ", err)
		}
	}()
}

func Forward(interface{}) {}
