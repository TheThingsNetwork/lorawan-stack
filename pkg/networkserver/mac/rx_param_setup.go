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
	EvtEnqueueRxParamSetupRequest = defineEnqueueMACRequestEvent(
		"rx_param_setup", "Rx parameter setup",
		events.WithDataType(&ttnpb.MACCommand_RxParamSetupReq{}),
	)()
	EvtReceiveRxParamSetupAccept = defineReceiveMACAcceptEvent(
		"rx_param_setup", "Rx parameter setup",
		events.WithDataType(&ttnpb.MACCommand_RxParamSetupAns{}),
	)()
	EvtReceiveRxParamSetupReject = defineReceiveMACRejectEvent(
		"rx_param_setup", "Rx parameter setup",
		events.WithDataType(&ttnpb.MACCommand_RxParamSetupAns{}),
	)()
)

func DeviceNeedsRxParamSetupReq(dev *ttnpb.EndDevice) bool {
	return !dev.GetMulticast() &&
		dev.GetMacState() != nil &&
		(dev.MacState.DesiredParameters.Rx1DataRateOffset != dev.MacState.CurrentParameters.Rx1DataRateOffset ||
			dev.MacState.DesiredParameters.Rx2DataRateIndex != dev.MacState.CurrentParameters.Rx2DataRateIndex ||
			dev.MacState.DesiredParameters.Rx2Frequency != dev.MacState.CurrentParameters.Rx2Frequency)
}

func EnqueueRxParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsRxParamSetupReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_RX_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_RxParamSetupReq{
			Rx1DataRateOffset: dev.MacState.DesiredParameters.Rx1DataRateOffset,
			Rx2DataRateIndex:  dev.MacState.DesiredParameters.Rx2DataRateIndex,
			Rx2Frequency:      dev.MacState.DesiredParameters.Rx2Frequency,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"rx1_data_rate_offset", req.Rx1DataRateOffset,
			"rx2_data_rate_index", req.Rx2DataRateIndex,
			"rx2_frequency", req.Rx2Frequency,
		)).Debug("Enqueued RxParamSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueRxParamSetupRequest.With(events.WithData(req)),
			},
			true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandleRxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RxParamSetupAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_RX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		if !pld.Rx1DataRateOffsetAck || !pld.Rx2DataRateIndexAck || !pld.Rx2FrequencyAck {
			return nil
		}

		req := cmd.GetRxParamSetupReq()

		dev.MacState.CurrentParameters.Rx1DataRateOffset = req.Rx1DataRateOffset
		dev.MacState.CurrentParameters.Rx2DataRateIndex = req.Rx2DataRateIndex
		dev.MacState.CurrentParameters.Rx2Frequency = req.Rx2Frequency
		return nil
	}, dev.MacState.PendingRequests...)
	ev := EvtReceiveRxParamSetupAccept
	if !pld.Rx1DataRateOffsetAck || !pld.Rx2DataRateIndexAck || !pld.Rx2FrequencyAck {
		ev = EvtReceiveRxParamSetupReject
	}
	return events.Builders{
		ev.With(events.WithData(pld)),
	}, err
}
