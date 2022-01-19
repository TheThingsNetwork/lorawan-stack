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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueDevStatusRequest = defineEnqueueMACRequestEvent(
		"dev_status", "device status",
	)()
	EvtReceiveDevStatusAnswer = defineReceiveMACAnswerEvent(
		"dev_status", "device status",
		events.WithDataType(&ttnpb.MACCommand_DevStatusAns{}),
	)()
)

const (
	DefaultStatusCountPeriodicity uint32 = 200
	DefaultStatusTimePeriodicity         = 24 * time.Hour
)

func DeviceStatusCountPeriodicity(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint32 {
	if v := dev.GetMacSettings().GetStatusCountPeriodicity(); v != nil {
		return v.Value
	}
	if defaults.StatusCountPeriodicity != nil {
		return defaults.StatusCountPeriodicity.Value
	}
	return DefaultStatusCountPeriodicity
}

func DeviceStatusTimePeriodicity(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if v := dev.GetMacSettings().GetStatusTimePeriodicity(); v != nil {
		return ttnpb.StdDurationOrZero(v)
	}
	if defaults.StatusTimePeriodicity != nil {
		return ttnpb.StdDurationOrZero(defaults.StatusTimePeriodicity)
	}
	return DefaultStatusTimePeriodicity
}

func DeviceNeedsDevStatusReqAt(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) (time.Time, bool) {
	if dev.MacState == nil {
		return time.Time{}, false
	}
	tp := DeviceStatusTimePeriodicity(dev, defaults)
	if tp == 0 {
		return time.Time{}, false
	}
	if dev.LastDevStatusReceivedAt == nil {
		return time.Time{}, true
	}
	return ttnpb.StdTime(dev.LastDevStatusReceivedAt).Add(tp).UTC(), true
}

func DeviceNeedsDevStatusReq(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings, transmitAt time.Time) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	timedAt, timeBound := DeviceNeedsDevStatusReqAt(dev, defaults)
	cp := DeviceStatusCountPeriodicity(dev, defaults)
	return (cp != 0 || timeBound) && dev.LastDevStatusReceivedAt == nil ||
		cp != 0 && dev.MacState.LastDevStatusFCntUp+cp <= dev.Session.LastFCntUp ||
		timeBound && !timedAt.After(transmitAt)
}

func EnqueueDevStatusReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, defaults ttnpb.MACSettings, transmitAt time.Time) EnqueueState {
	if !DeviceNeedsDevStatusReq(dev, defaults, transmitAt) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_DEV_STATUS, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		log.FromContext(ctx).Debug("Enqueued DevStatusReq")
		return []*ttnpb.MACCommand{
				ttnpb.MACCommandIdentifier_CID_DEV_STATUS.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueDevStatusRequest,
			},
			true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandleDevStatusAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DevStatusAns, fCntUp uint32, recvAt time.Time) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_DEV_STATUS, func(*ttnpb.MACCommand) error {
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
		dev.LastDevStatusReceivedAt = ttnpb.ProtoTimePtr(recvAt)
		dev.MacState.LastDevStatusFCntUp = fCntUp
		return nil
	}, dev.MacState.PendingRequests...)
	return events.Builders{
		EvtReceiveDevStatusAnswer.With(events.WithData(pld)),
	}, err
}
