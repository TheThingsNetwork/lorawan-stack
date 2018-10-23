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
	"time"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueDevStatusRequest = defineEnqueueMACRequestEvent("dev_status", "device status")
	evtReceiveDevStatusAnswer  = defineReceiveMACAnswerEvent("dev_status", "device status")
)

func enqueueDevStatusReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) (uint16, uint16, bool) {
	if dev.LastDevStatusReceivedAt != nil &&
		(dev.MACSettings.StatusCountPeriodicity == 0 || dev.MACState.LastDevStatusFCntUp+dev.MACSettings.StatusCountPeriodicity > dev.Session.LastFCntUp) &&
		(dev.MACSettings.StatusTimePeriodicity == 0 || dev.LastDevStatusReceivedAt.Add(dev.MACSettings.StatusTimePeriodicity).After(time.Now())) {
		return maxDownLen, maxUpLen, true
	}

	var ok bool
	dev.MACState.PendingRequests, maxDownLen, maxUpLen, ok = enqueueMACCommand(ttnpb.CID_DEV_STATUS, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, false
		}

		events.Publish(evtEnqueueDevStatusRequest(ctx, dev.EndDeviceIdentifiers, nil))
		return []*ttnpb.MACCommand{ttnpb.CID_DEV_STATUS.MACCommand()}, 1, true

	}, dev.MACState.PendingRequests...)
	return maxDownLen, maxUpLen, ok
}

func handleDevStatusAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DevStatusAns, fCntUp uint32, recvAt time.Time) (err error) {
	if pld == nil {
		return errNoPayload
	}

	events.Publish(evtReceiveDevStatusAnswer(ctx, dev.EndDeviceIdentifiers, pld))

	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DEV_STATUS, func(*ttnpb.MACCommand) error {
		switch pld.Battery {
		case 0:
			dev.PowerState = ttnpb.PowerState_POWER_EXTERNAL
			dev.BatteryPercentage = -1
		case 255:
			dev.PowerState = ttnpb.PowerState_POWER_UNKNOWN
			dev.BatteryPercentage = -1
		default:
			dev.PowerState = ttnpb.PowerState_POWER_BATTERY
			dev.BatteryPercentage = float32(pld.Battery-2) / 253
		}
		dev.DownlinkMargin = pld.Margin
		dev.LastDevStatusReceivedAt = &recvAt
		dev.MACState.LastDevStatusFCntUp = fCntUp
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
