// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// HandleDownlink handles downlink MAC commands
func HandleDownlink(ctx context.Context, dev *ttnpb.EndDevice, downlink *ttnpb.DownlinkMessage) error {
	if downlink.Payload.MType != ttnpb.MType_UNCONFIRMED_DOWN && downlink.Payload.MType != ttnpb.MType_CONFIRMED_DOWN {
		return nil // Only handle downlink messages
	}

	if len(dev.QueuedMACCommands) == 0 {
		return nil // Nothing to do
	}

	var macCommands ttnpb.MACCommands
	for _, macCommand := range dev.QueuedMACCommands {
		macCommands = append(macCommands, macCommand)
	}

	// TODO: Sort macCommands on LoRaWAN version support

	macCommandBytes, err := macCommands.MarshalLoRaWAN()
	if err != nil {
		return err
	}

	macPayload := downlink.Payload.GetMACPayload()

	if len(macCommandBytes) > 15 {
		// If we have more MAC commands than fit in the FOpts, or if there is no application downlink, we send MAC commands in the FRMPayload.
		macPayload.FPort = 0
		macPayload.FRMPayload = macCommandBytes
		macCommandBytes, err = crypto.EncryptDownlink(*dev.Session.NwkSEncKey.Key, macPayload.DevAddr, macPayload.GetFCnt(), macPayload.GetFRMPayload())
		if err != nil {
			return err
		}
	} else {
		macPayload.FOpts = macCommandBytes
	}

	// TODO: Remove response messages from dev.QueuedMACCommands (only keep requests)

	return nil
}
