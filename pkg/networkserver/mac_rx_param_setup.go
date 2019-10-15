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

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueRxParamSetupRequest = defineEnqueueMACRequestEvent("rx_param_setup", "Rx parameter setup")()
	evtReceiveRxParamSetupAccept  = defineReceiveMACAcceptEvent("rx_param_setup", "Rx parameter setup")()
	evtReceiveRxParamSetupReject  = defineReceiveMACRejectEvent("rx_param_setup", "Rx parameter setup")()
)

func needsRxParamSetupReq(dev *ttnpb.EndDevice) bool {
	return dev.MACState != nil &&
		(dev.MACState.DesiredParameters.Rx1DataRateOffset != dev.MACState.CurrentParameters.Rx1DataRateOffset ||
			dev.MACState.DesiredParameters.Rx2DataRateIndex != dev.MACState.CurrentParameters.Rx2DataRateIndex ||
			dev.MACState.DesiredParameters.Rx2Frequency != dev.MACState.CurrentParameters.Rx2Frequency)
}

func enqueueRxParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) macCommandEnqueueState {
	if !needsRxParamSetupReq(dev) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_RX_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_RxParamSetupReq{
			Rx1DataRateOffset: dev.MACState.DesiredParameters.Rx1DataRateOffset,
			Rx2DataRateIndex:  dev.MACState.DesiredParameters.Rx2DataRateIndex,
			Rx2Frequency:      dev.MACState.DesiredParameters.Rx2Frequency,
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
			[]events.DefinitionDataClosure{
				evtEnqueueRxParamSetupRequest.BindData(req),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleRxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RxParamSetupAns) ([]events.DefinitionDataClosure, error) {
	if pld == nil {
		return nil, errNoPayload
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_RX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		if !pld.Rx1DataRateOffsetAck || !pld.Rx2DataRateIndexAck || !pld.Rx2FrequencyAck {
			return nil
		}

		req := cmd.GetRxParamSetupReq()

		dev.MACState.CurrentParameters.Rx1DataRateOffset = req.Rx1DataRateOffset
		dev.MACState.CurrentParameters.Rx2DataRateIndex = req.Rx2DataRateIndex
		dev.MACState.CurrentParameters.Rx2Frequency = req.Rx2Frequency
		return nil
	}, dev.MACState.PendingRequests...)
	evt := evtReceiveRxParamSetupAccept
	if !pld.Rx1DataRateOffsetAck || !pld.Rx2DataRateIndexAck || !pld.Rx2FrequencyAck {
		evt = evtReceiveRxParamSetupReject
	}
	return []events.DefinitionDataClosure{
		evt.BindData(pld),
	}, err
}
