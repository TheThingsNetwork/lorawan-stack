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
	evtMACDeviceStatusRequest = events.Define("ns.mac.device_status.request", "request device status") // TODO(#988): publish when requesting
	evtMACDeviceStatus        = events.Define("ns.mac.device_status", "handled device status")
)

func enqueueDevStatusReq(ctx context.Context, dev *ttnpb.EndDevice) {
	if dev.LastStatusReceivedAt == nil ||
		dev.MACSettings.StatusCountPeriodicity > 0 && dev.NextStatusAfter == 0 ||
		dev.MACSettings.StatusTimePeriodicity > 0 && dev.LastStatusReceivedAt.Add(dev.MACSettings.StatusTimePeriodicity).Before(time.Now()) {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, ttnpb.CID_DEV_STATUS.MACCommand())
		dev.NextStatusAfter = dev.MACSettings.StatusCountPeriodicity

	} else if dev.NextStatusAfter != 0 {
		dev.NextStatusAfter--
	}
}

func handleDevStatusAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DevStatusAns, recvAt time.Time) (err error) {
	if pld == nil {
		return errNoPayload
	}

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
		dev.LastStatusReceivedAt = &recvAt

		events.Publish(evtMACDeviceStatus(ctx, dev.EndDeviceIdentifiers, pld))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
