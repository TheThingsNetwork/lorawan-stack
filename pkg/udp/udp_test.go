// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp_test

import (
	"fmt"
	"net"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/udp"
	"github.com/TheThingsNetwork/ttn/pkg/udp/validation"
)

var (
	downlink ttnpb.DownlinkMessage

	converter = dummyConverter{}
	store     = udp.GatewayStore{}
	validator = validation.InMemoryValidator(validation.DefaultWaitDuration)
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

			// Verify the packet against the IP-lock policy
			if !validator.Valid(*packet) {
				continue
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
			case udp.PullData:
				// Update the downlink address in the mapping
				UpdateDownlinkAddress(*packet.GatewayEUI, conn)
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

		addr, isAddr := GetDownlinkAddress(&packet)
		if !isAddr {
			fmt.Println("No address found for this gateway")
			return
		}

		if err := conn.WriteTo(&packet, addr); err != nil {
			fmt.Println("Error when sending downlink: ", err)
		}
	}()
}

func UpdateDownlinkAddress(eui types.EUI64, conn *udp.Conn) {
	store.Lock()
	store.Store[eui] = conn.UDPConn.LocalAddr()
	store.Unlock()
}

func GetDownlinkAddress(packet *udp.Packet) (net.Addr, bool) {
	store.Lock()
	address, isAddress := store.Store[*packet.GatewayEUI]
	store.Unlock()
	return address, isAddress
}

func Forward(interface{}) {}
