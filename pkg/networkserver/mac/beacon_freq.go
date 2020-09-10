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
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueBeaconFreqRequest = defineEnqueueMACRequestEvent(
		"beacon_freq", "beacon frequency change",
		events.WithDataType(&ttnpb.MACCommand_BeaconFreqReq{}),
	)()
	EvtReceiveBeaconFreqReject = defineReceiveMACRejectEvent(
		"beacon_freq", "beacon frequency change",
		events.WithDataType(&ttnpb.MACCommand_BeaconFreqAns{}),
	)()
	EvtReceiveBeaconFreqAccept = defineReceiveMACAcceptEvent(
		"beacon_freq", "beacon frequency change",
		events.WithDataType(&ttnpb.MACCommand_BeaconFreqAns{}),
	)()
)

func DeviceNeedsBeaconFreqReq(dev *ttnpb.EndDevice) bool {
	return !dev.GetMulticast() &&
		dev.GetMACState().GetDeviceClass() == ttnpb.CLASS_B &&
		dev.MACState.DesiredParameters.BeaconFrequency != dev.MACState.CurrentParameters.BeaconFrequency
}

func EnqueueBeaconFreqReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsBeaconFreqReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_BEACON_FREQ, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}

		req := &ttnpb.MACCommand_BeaconFreqReq{
			Frequency: dev.MACState.DesiredParameters.BeaconFrequency,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"frequency", req.Frequency,
		)).Debug("Enqueued BeaconFreqReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueBeaconFreqRequest.With(events.WithData(req)),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func HandleBeaconFreqAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_BeaconFreqAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	Evt := EvtReceiveBeaconFreqAccept
	if !pld.FrequencyAck {
		Evt = EvtReceiveBeaconFreqReject
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_BEACON_FREQ, func(cmd *ttnpb.MACCommand) error {
		if !pld.FrequencyAck {
			return nil
		}
		req := cmd.GetBeaconFreqReq()

		dev.MACState.CurrentParameters.BeaconFrequency = req.Frequency
		return nil
	}, dev.MACState.PendingRequests...)
	return events.Builders{
		Evt.With(events.WithData(pld)),
	}, err
}
