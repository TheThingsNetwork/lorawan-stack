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
	evtMACRxParamRequest = events.Define("ns.mac.rx_param.request", "request rx parameter setup") // TODO(#988): publish when requesting
	evtMACRxParamAccept  = events.Define("ns.mac.rx_param.accept", "device accepted rx parameter setup request")
	evtMACRxParamReject  = events.Define("ns.mac.rx_param.reject", "device rejected rx parameter setup request")
)

func enqueueRxParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice) {
	if dev.MACState.DesiredParameters.Rx2Frequency != dev.MACState.CurrentParameters.Rx2Frequency ||
		dev.MACState.DesiredParameters.Rx2DataRateIndex != dev.MACState.CurrentParameters.Rx2DataRateIndex ||
		dev.MACState.DesiredParameters.Rx1DataRateOffset != dev.MACState.CurrentParameters.Rx1DataRateOffset {
		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_RxParamSetupReq{
			Rx2Frequency:      dev.MACState.DesiredParameters.Rx2Frequency,
			Rx2DataRateIndex:  dev.MACState.DesiredParameters.Rx2DataRateIndex,
			Rx1DataRateOffset: dev.MACState.DesiredParameters.Rx1DataRateOffset,
		}).MACCommand())
	}
}

func handleRxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RxParamSetupAns) (err error) {
	if pld == nil {
		return errNoPayload
	}

	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_RX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		if !pld.Rx1DataRateOffsetAck || !pld.Rx2DataRateIndexAck || !pld.Rx2FrequencyAck {
			// TODO: Handle NACK, modify desired state
			// (https://github.com/TheThingsIndustries/ttn/issues/834)
			events.Publish(evtMACRxParamReject(ctx, dev.EndDeviceIdentifiers, pld))
			return nil
		}

		req := cmd.GetRxParamSetupReq()

		dev.MACState.CurrentParameters.Rx1DataRateOffset = req.Rx1DataRateOffset
		dev.MACState.CurrentParameters.Rx2DataRateIndex = req.Rx2DataRateIndex
		dev.MACState.CurrentParameters.Rx2Frequency = req.Rx2Frequency

		events.Publish(evtMACRxParamAccept(ctx, dev.EndDeviceIdentifiers, req))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
