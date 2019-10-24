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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueDevStatusRequest = defineEnqueueMACRequestEvent("dev_status", "device status")()
	evtReceiveDevStatusAnswer  = defineReceiveMACAnswerEvent("dev_status", "device status")()
)

const (
	DefaultStatusCountPeriodicity uint32 = 20
	DefaultStatusTimePeriodicity         = time.Hour
)

func deviceStatusCountPeriodicity(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint32 {
	if dev.MACSettings != nil && dev.MACSettings.StatusCountPeriodicity != nil {
		return dev.MACSettings.StatusCountPeriodicity.Value
	}
	if defaults.StatusCountPeriodicity != nil {
		return defaults.StatusCountPeriodicity.Value
	}
	return DefaultStatusCountPeriodicity
}

func deviceStatusTimePeriodicity(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if dev.MACSettings != nil && dev.MACSettings.StatusTimePeriodicity != nil {
		return *dev.MACSettings.StatusTimePeriodicity
	}
	if defaults.StatusTimePeriodicity != nil {
		return *defaults.StatusTimePeriodicity
	}
	return DefaultStatusTimePeriodicity
}

func deviceNeedsDevStatusReqAt(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) (time.Time, bool) {
	if dev.MACState == nil {
		return time.Time{}, false
	}
	tp := deviceStatusTimePeriodicity(dev, defaults)
	if tp == 0 {
		return time.Time{}, false
	}
	if dev.LastDevStatusReceivedAt == nil {
		return time.Time{}, true
	}
	return dev.LastDevStatusReceivedAt.Add(tp).UTC(), true
}

func deviceNeedsDevStatusReq(dev *ttnpb.EndDevice, scheduleAt time.Time, defaults ttnpb.MACSettings) bool {
	if dev.MACState == nil {
		return false
	}
	timedAt, timeBound := deviceNeedsDevStatusReqAt(dev, defaults)
	cp := deviceStatusCountPeriodicity(dev, defaults)
	return (cp != 0 || timeBound) && dev.LastDevStatusReceivedAt == nil ||
		cp != 0 && dev.MACState.LastDevStatusFCntUp+cp <= dev.Session.LastFCntUp ||
		timeBound && timedAt.Before(timeNow())
}

func enqueueDevStatusReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, scheduleAt time.Time, defaults ttnpb.MACSettings) macCommandEnqueueState {
	if !deviceNeedsDevStatusReq(dev, scheduleAt, defaults) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_DEV_STATUS, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		log.FromContext(ctx).Debug("Enqueued DevStatusReq")
		return []*ttnpb.MACCommand{
				ttnpb.CID_DEV_STATUS.MACCommand(),
			},
			1,
			[]events.DefinitionDataClosure{
				evtEnqueueDevStatusRequest.BindData(nil),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleDevStatusAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DevStatusAns, fCntUp uint32, recvAt time.Time) ([]events.DefinitionDataClosure, error) {
	if pld == nil {
		return nil, errNoPayload
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DEV_STATUS, func(*ttnpb.MACCommand) error {
		switch pld.Battery {
		case 0:
			dev.PowerState = ttnpb.PowerState_POWER_EXTERNAL
			dev.BatteryPercentage = nil
		case 255:
			dev.PowerState = ttnpb.PowerState_POWER_UNKNOWN
			dev.BatteryPercentage = nil
		default:
			dev.PowerState = ttnpb.PowerState_POWER_BATTERY
			dev.BatteryPercentage = &pbtypes.FloatValue{Value: float32(pld.Battery-1) / 253}
		}
		dev.DownlinkMargin = pld.Margin
		dev.LastDevStatusReceivedAt = &recvAt
		dev.MACState.LastDevStatusFCntUp = fCntUp
		return nil
	}, dev.MACState.PendingRequests...)
	return []events.DefinitionDataClosure{
		evtReceiveDevStatusAnswer.BindData(pld),
	}, err
}
