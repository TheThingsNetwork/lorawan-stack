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
)

var (
	EvtReceiveDeviceModeIndication = defineReceiveMACIndicationEvent(
		"device_mode", "device mode",
		events.WithDataType(&ttnpb.MACCommand_DeviceModeInd{}),
	)()
	EvtEnqueueDeviceModeConfirmation = defineEnqueueMACConfirmationEvent(
		"device_mode", "device mode",
		events.WithDataType(&ttnpb.MACCommand_DeviceModeConf{}),
	)()
)

func HandleDeviceModeInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DeviceModeInd) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	evs := events.Builders{
		EvtReceiveDeviceModeIndication.With(events.WithData(pld)),
	}
	switch {
	case pld.Class == ttnpb.Class_CLASS_C && dev.SupportsClassC && dev.MacState.DeviceClass != ttnpb.Class_CLASS_C:
		evs = append(evs, EvtClassCSwitch.With(events.WithData(dev.MacState.DeviceClass)))
		dev.MacState.DeviceClass = ttnpb.Class_CLASS_C

	case pld.Class == ttnpb.Class_CLASS_A && dev.MacState.DeviceClass != ttnpb.Class_CLASS_A:
		evs = append(evs, EvtClassASwitch.With(events.WithData(dev.MacState.DeviceClass)))
		dev.MacState.DeviceClass = ttnpb.Class_CLASS_A
	}
	conf := &ttnpb.MACCommand_DeviceModeConf{
		Class: dev.MacState.DeviceClass,
	}
	dev.MacState.QueuedResponses = append(dev.MacState.QueuedResponses, conf.MACCommand())
	return append(evs,
		EvtEnqueueDeviceModeConfirmation.With(events.WithData(conf)),
	), nil
}
