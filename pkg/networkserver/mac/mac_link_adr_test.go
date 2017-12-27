// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMACLinkADR(t *testing.T) {
	a := assertions.New(t)

	setADR := func(dev *ttnpb.EndDevice) {
		dev.QueuedMACCommands = []*ttnpb.MACCommand{
			(&ttnpb.MACCommand_LinkADRReq{
				DataRateIndex:      4,
				ChannelMask:        []bool{false, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
				ChannelMaskControl: 7,
				NbTrans:            1,
			}).MACCommand(),
			(&ttnpb.MACCommand_LinkADRReq{
				DataRateIndex:      2,
				ChannelMask:        []bool{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
				ChannelMaskControl: 0,
				NbTrans:            1,
			}).MACCommand(),
		}
	}

	ctx := context.Background()

	t.Run("LoRaWAN 1.0", func(t *testing.T) {
		dev := newDev()
		dev.LoRaWANVersion = ttnpb.MAC_V1_0
		setADR(dev)

		msg := newMsg()
		msg.GetMACPayload().FOpts = []byte{0x03, 0x07}

		err := HandleUplink(ctx, dev, &ttnpb.UplinkMessage{
			Payload: msg,
			Settings: ttnpb.TxSettings{
				Modulation:      ttnpb.Modulation_LORA,
				Bandwidth:       125,
				SpreadingFactor: 10,
			},
		})
		a.So(err, should.BeNil)
		// TODO:
	})
	t.Run("LoRaWAN 1.0.1", func(t *testing.T) {
		dev := newDev()
		dev.LoRaWANVersion = ttnpb.MAC_V1_0_1
		setADR(dev)

		msg := newMsg()
		msg.GetMACPayload().FOpts = []byte{0x03, 0x07}

		err := HandleUplink(ctx, dev, &ttnpb.UplinkMessage{
			Payload: msg,
			Settings: ttnpb.TxSettings{
				Modulation:      ttnpb.Modulation_LORA,
				Bandwidth:       125,
				SpreadingFactor: 10,
			},
		})
		a.So(err, should.BeNil)
		// TODO:
	})
	t.Run("LoRaWAN 1.0.2", func(t *testing.T) {
		dev := newDev()
		dev.LoRaWANVersion = ttnpb.MAC_V1_0_2
		setADR(dev)

		msg := newMsg()
		msg.GetMACPayload().FOpts = []byte{0x03, 0x07}

		err := HandleUplink(ctx, dev, &ttnpb.UplinkMessage{
			Payload: msg,
			Settings: ttnpb.TxSettings{
				Modulation:      ttnpb.Modulation_LORA,
				Bandwidth:       125,
				SpreadingFactor: 10,
			},
		})
		a.So(err, should.BeNil)
		// TODO
	})
	t.Run("LoRaWAN 1.1", func(t *testing.T) {
		dev := newDev()
		dev.LoRaWANVersion = ttnpb.MAC_V1_1
		setADR(dev)

		msg := newMsg()
		msg.GetMACPayload().FOpts = []byte{0x03, 0x07}

		err := HandleUplink(ctx, dev, &ttnpb.UplinkMessage{
			Payload: msg,
			Settings: ttnpb.TxSettings{
				Modulation:      ttnpb.Modulation_LORA,
				Bandwidth:       125,
				SpreadingFactor: 10,
			},
		})
		a.So(err, should.BeNil)
		// TODO
	})
}
