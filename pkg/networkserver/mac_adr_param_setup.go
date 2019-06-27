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

	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueADRParamSetupRequest = defineEnqueueMACRequestEvent("adr_param_setup", "ADR parameter setup")()
	evtReceiveADRParamSetupAnswer  = defineReceiveMACAnswerEvent("adr_param_setup", "ADR parameter setup")()
)

func enqueueADRParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) (uint16, uint16, bool) {
	if dev.MACState.DesiredParameters.ADRAckLimit == dev.MACState.CurrentParameters.ADRAckLimit &&
		dev.MACState.DesiredParameters.ADRAckDelay == dev.MACState.CurrentParameters.ADRAckDelay {
		return maxDownLen, maxUpLen, true
	}

	var ok bool
	dev.MACState.PendingRequests, maxDownLen, maxUpLen, ok = enqueueMACCommand(ttnpb.CID_ADR_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, false
		}

		req := &ttnpb.MACCommand_ADRParamSetupReq{
			ADRAckLimitExponent: lorawan.Uint32ToADRAckLimitExponent(dev.MACState.DesiredParameters.ADRAckLimit),
			ADRAckDelayExponent: lorawan.Uint32ToADRAckDelayExponent(dev.MACState.DesiredParameters.ADRAckDelay),
		}
		events.Publish(evtEnqueueADRParamSetupRequest(ctx, dev.EndDeviceIdentifiers, req))
		return []*ttnpb.MACCommand{req.MACCommand()}, 1, true
	}, dev.MACState.PendingRequests...)
	return maxDownLen, maxUpLen, ok
}

func handleADRParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) ([]events.DefinitionDataClosure, error) {
	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_ADR_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetADRParamSetupReq()

		dev.MACState.CurrentParameters.ADRAckDelay = lorawan.ADRAckDelayExponentToUint32(req.ADRAckDelayExponent)
		dev.MACState.CurrentParameters.ADRAckLimit = lorawan.ADRAckLimitExponentToUint32(req.ADRAckLimitExponent)

		if lorawan.Uint32ToADRAckDelayExponent(dev.MACState.DesiredParameters.ADRAckDelay) == req.ADRAckDelayExponent {
			dev.MACState.DesiredParameters.ADRAckDelay = dev.MACState.CurrentParameters.ADRAckDelay
		}

		if lorawan.Uint32ToADRAckLimitExponent(dev.MACState.DesiredParameters.ADRAckLimit) == req.ADRAckLimitExponent {
			dev.MACState.DesiredParameters.ADRAckLimit = dev.MACState.CurrentParameters.ADRAckLimit
		}
		return nil
	}, dev.MACState.PendingRequests...)
	return []events.DefinitionDataClosure{
		evtReceiveADRParamSetupAnswer.BindData(nil),
	}, err
}
