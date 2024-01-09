// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
)

// EvtReceiveRelayNotifyNewEndDeviceIndication is emitted when a relay notify new end device
// indication is received.
var EvtReceiveRelayNotifyNewEndDeviceIndication = defineReceiveMACIndicationEvent(
	"relay_notify_new_end_device", "relay notify new end device",
)()

// HandleRelayNotifyNewEndDeviceReq handles a relay notify new end device request.
func HandleRelayNotifyNewEndDeviceReq(
	_ context.Context, _ *ttnpb.EndDevice, pld *ttnpb.MACCommand_RelayNotifyNewEndDeviceReq,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	return events.Builders{
		EvtReceiveRelayNotifyNewEndDeviceIndication.With(events.WithData(pld)),
	}, nil
}
