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
	evtMACADRParamRequest = events.Define("ns.mac.adr_param.request", "request ADR parameter setup") // TODO(#988): publish when requesting
	evtMACADRParamAccept  = events.Define("ns.mac.adr_param.accept", "device accepted ADR parameter setup request")
)

func enqueueADRParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice) {
	if dev.MACState.DesiredParameters.ADRAckLimit == dev.MACState.CurrentParameters.ADRAckLimit &&
		dev.MACState.DesiredParameters.ADRAckDelay == dev.MACState.CurrentParameters.ADRAckDelay {
		return
	}

	dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_ADRParamSetupReq{
		ADRAckLimitExponent: ttnpb.Uint32ToADRAckLimitExponent(dev.MACState.DesiredParameters.ADRAckLimit),
		ADRAckDelayExponent: ttnpb.Uint32ToADRAckDelayExponent(dev.MACState.DesiredParameters.ADRAckDelay),
	}).MACCommand())
}

func handleADRParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (err error) {
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_ADR_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetADRParamSetupReq()

		dev.MACState.CurrentParameters.ADRAckDelay = ttnpb.ADRAckDelayExponentToUint32(req.ADRAckDelayExponent)
		dev.MACState.CurrentParameters.ADRAckLimit = ttnpb.ADRAckLimitExponentToUint32(req.ADRAckLimitExponent)

		if ttnpb.Uint32ToADRAckDelayExponent(dev.MACState.DesiredParameters.ADRAckDelay) == req.ADRAckDelayExponent {
			dev.MACState.DesiredParameters.ADRAckDelay = dev.MACState.CurrentParameters.ADRAckDelay
		}

		if ttnpb.Uint32ToADRAckLimitExponent(dev.MACState.DesiredParameters.ADRAckLimit) == req.ADRAckLimitExponent {
			dev.MACState.DesiredParameters.ADRAckLimit = dev.MACState.CurrentParameters.ADRAckLimit
		}

		events.Publish(evtMACADRParamAccept(ctx, dev.EndDeviceIdentifiers, req))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
