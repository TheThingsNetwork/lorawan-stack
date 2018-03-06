// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

type randyMessages interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func NewPopulatedUplinkMessage(r randyMessages, easy bool) *UplinkMessage {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedUplinkMessage: %s", r)
		}
	}()

	out := &UplinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.RxMetadata = make([]RxMetadata, 1+r.Intn(5))
	for i := 0; i < len(out.RxMetadata); i++ {
		out.RxMetadata[i] = *NewPopulatedRxMetadata(r, false)
	}

	msg := NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(out.Settings.DataRateIndex), uint8(out.RxMetadata[0].ChannelIndex), r.Intn(2) == 1)
	out.Payload = *msg

	var err error
	out.RawPayload, err = msg.AppendLoRaWAN(out.RawPayload)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to encode uplink message to LoRaWAN"))
	}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, false)
	devAddr := msg.GetMACPayload().DevAddr
	out.EndDeviceIdentifiers.DevAddr = &devAddr
	return out
}

func NewPopulatedDownlinkMessage(r randyMessages, easy bool) *DownlinkMessage {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedDownlinkMessage: %s", r)
		}
	}()

	out := &DownlinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.TxMetadata = *NewPopulatedTxMetadata(r, false)

	msg := NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), r.Intn(2) == 1)
	out.Payload = *msg

	var err error
	out.RawPayload, err = msg.AppendLoRaWAN(out.RawPayload)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to encode downlink message to LoRaWAN"))
	}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, false)
	devAddr := msg.GetMACPayload().DevAddr
	out.EndDeviceIdentifiers.DevAddr = &devAddr
	return out
}
