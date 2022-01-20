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

package mac

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	EvtReceiveRekeyIndication = defineReceiveMACIndicationEvent(
		"rekey", "device rekey",
		events.WithDataType(&ttnpb.MACCommand_RekeyInd{}),
	)()
	EvtEnqueueRekeyConfirmation = defineEnqueueMACConfirmationEvent(
		"rekey", "device rekey",
		events.WithDataType(&ttnpb.MACCommand_RekeyConf{}),
	)()
)

func HandleRekeyInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RekeyInd, devAddr types.DevAddr) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	evs := events.Builders{
		EvtReceiveRekeyIndication.With(events.WithData(pld)),
	}
	if !dev.SupportsJoin {
		return evs, nil
	}
	if dev.PendingSession != nil && dev.MacState.PendingJoinRequest != nil && dev.PendingSession.DevAddr == devAddr {
		dev.Ids.DevAddr = &dev.PendingSession.DevAddr
		dev.Session = dev.PendingSession
	}
	dev.MacState.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
	dev.MacState.PendingJoinRequest = nil
	dev.PendingMacState = nil
	dev.PendingSession = nil

	conf := &ttnpb.MACCommand_RekeyConf{
		MinorVersion: pld.MinorVersion,
	}
	dev.MacState.QueuedResponses = append(dev.MacState.QueuedResponses, conf.MACCommand())
	return append(evs,
		EvtEnqueueRekeyConfirmation.With(events.WithData(conf)),
	), nil
}
