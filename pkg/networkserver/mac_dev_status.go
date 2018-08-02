// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtMACDeviceStatusRequest = events.Define("ns.mac.device_status.request", "request device status") // TODO(#988): publish when requesting
	evtMACDeviceStatus        = events.Define("ns.mac.device_status", "handled device status")
)

func handleDevStatusAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DevStatusAns) (err error) {
	if pld == nil {
		return errMissingPayload
	}

	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DEV_STATUS, func(*ttnpb.MACCommand) error {
		// TODO: Modify status variables in MACState (https://github.com/TheThingsIndustries/ttn/issues/834)
		_ = pld.Battery
		_ = pld.Margin

		events.Publish(evtMACDeviceStatus(ctx, dev.EndDeviceIdentifiers, pld))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
