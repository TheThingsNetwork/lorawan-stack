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
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtReceiveResetIndication = defineReceiveMACIndicationEvent(
		"reset", "device reset",
		events.WithDataType(&ttnpb.MACCommand_ResetInd{}),
	)()
	EvtEnqueueResetConfirmation = defineEnqueueMACConfirmationEvent(
		"reset", "device reset",
		events.WithDataType(&ttnpb.MACCommand_ResetConf{}),
	)()
)

func HandleResetInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_ResetInd, fps *frequencyplans.Store, defaults ttnpb.MACSettings) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	evs := events.Builders{
		EvtReceiveResetIndication.With(events.WithData(pld)),
	}
	if dev.SupportsJoin {
		return evs, nil
	}

	macState, err := NewState(dev, fps, defaults)
	if err != nil {
		return evs, err
	}
	dev.MacState = macState
	dev.MacState.LorawanVersion = ttnpb.MACVersion_MAC_V1_1

	conf := &ttnpb.MACCommand_ResetConf{
		MinorVersion: pld.MinorVersion,
	}
	dev.MacState.QueuedResponses = append(dev.MacState.QueuedResponses, conf.MACCommand())
	return append(evs,
		EvtEnqueueResetConfirmation.With(events.WithData(conf)),
	), nil
}
