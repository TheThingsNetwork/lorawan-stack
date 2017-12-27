// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// HandleUplink handles uplink MAC state and commands
func HandleUplink(ctx context.Context, dev *ttnpb.EndDevice, uplink *ttnpb.UplinkMessage) (err error) {
	switch uplink.Payload.MType {
	case ttnpb.MType_UNCONFIRMED_UP,
		ttnpb.MType_CONFIRMED_UP:
	default:
		return nil // Only handle uplink messages
	}

	if dev.MACState == nil {
		band, err := band.GetByID(dev.FrequencyPlanID)
		if err != nil {
			return err
		}
		resetMACState(dev.MACState, band)
	}

	dev.MACState.AdrDataRateIndex = uplink.Settings.DataRateIndex

	macPayload := uplink.Payload.GetMACPayload()
	macCommandBytes := macPayload.GetFOpts()
	if macPayload.FPort == 0 {
		macCommandBytes, err = crypto.DecryptUplink(*dev.Session.NwkSEncKey.Key, macPayload.DevAddr, macPayload.GetFCnt(), macPayload.GetFRMPayload())
		if err != nil {
			return err
		}
	}

	if len(macCommandBytes) == 0 {
		return nil // Nothing to do
	}

	var macCommands ttnpb.MACCommands
	if err := macCommands.UnmarshalLoRaWAN(macCommandBytes, true); err != nil {
		return err
	}

	ctx = newContextWithUplinkMessage(ctx, uplink)

	for _, macCommand := range macCommands {
		if handler, ok := handlers[macCommand.CID()]; ok {
			if err := handler.HandleMACCommand(ctx, dev, macCommand); err != nil {
				return err
			}
		}
	}

	return nil
}
