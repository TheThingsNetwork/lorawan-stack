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

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	evtReceiveRekeyIndication   = defineReceiveMACIndicationEvent("rekey", "device rekey")()
	evtEnqueueRekeyConfirmation = defineEnqueueMACConfirmationEvent("rekey", "device rekey")()
)

func handleRekeyInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RekeyInd, devAddr types.DevAddr) (events.Builders, error) {
	if pld == nil {
		return nil, errNoPayload.New()
	}

	evs := events.Builders{
		evtReceiveRekeyIndication.With(events.WithData(pld)),
	}
	if !dev.SupportsJoin {
		return evs, nil
	}
	if dev.PendingSession != nil && dev.MACState.PendingJoinRequest != nil && dev.PendingSession.DevAddr == devAddr {
		dev.EndDeviceIdentifiers.DevAddr = &dev.PendingSession.DevAddr
		dev.Session = dev.PendingSession
	}
	dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
	dev.MACState.PendingJoinRequest = nil
	dev.PendingMACState = nil
	dev.PendingSession = nil

	conf := &ttnpb.MACCommand_RekeyConf{
		MinorVersion: pld.MinorVersion,
	}
	dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, conf.MACCommand())
	return append(evs,
		evtEnqueueRekeyConfirmation.With(events.WithData(conf)),
	), nil
}
