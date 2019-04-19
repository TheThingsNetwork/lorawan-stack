// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtReceiveResetIndication   = defineReceiveMACIndicationEvent("reset", "device reset")()
	evtEnqueueResetConfirmation = defineEnqueueMACConfirmationEvent("reset", "device reset")()
)

func handleResetInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_ResetInd, fps *frequencyplans.Store, defaults ttnpb.MACSettings) error {
	if pld == nil {
		return errNoPayload
	}

	events.Publish(evtReceiveResetIndication(ctx, dev.EndDeviceIdentifiers, pld))

	if dev.SupportsJoin {
		return nil
	}

	macState, err := newMACState(dev, fps, defaults)
	if err != nil {
		return err
	}
	dev.MACState = macState
	dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1

	conf := &ttnpb.MACCommand_ResetConf{
		MinorVersion: pld.MinorVersion,
	}
	dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, conf.MACCommand())

	events.Publish(evtEnqueueResetConfirmation(ctx, dev.EndDeviceIdentifiers, conf))
	return nil
}
