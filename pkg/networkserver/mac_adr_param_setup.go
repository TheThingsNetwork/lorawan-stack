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

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueADRParamSetupRequest = defineEnqueueMACRequestEvent("adr_param_setup", "ADR parameter setup")()
	evtReceiveADRParamSetupAnswer  = defineReceiveMACAnswerEvent("adr_param_setup", "ADR parameter setup")()
)

func needsADRParamSetupReq(dev *ttnpb.EndDevice, phy band.Band) bool {
	if dev.MACState == nil || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		return false
	}

	desiredDelay, currentDelay := dev.MACState.DesiredParameters.ADRAckDelayExponent, dev.MACState.CurrentParameters.ADRAckDelayExponent
	desiredLimit, currentLimit := dev.MACState.DesiredParameters.ADRAckLimitExponent, dev.MACState.CurrentParameters.ADRAckLimitExponent
	return desiredLimit != nil &&
		(currentLimit != nil && currentLimit.Value != desiredLimit.Value ||
			currentLimit == nil && phy.ADRAckLimit != desiredLimit.Value) ||
		desiredDelay != nil &&
			(currentDelay != nil && currentDelay.Value != desiredDelay.Value ||
				currentDelay == nil && phy.ADRAckDelay != desiredDelay.Value)
}

func enqueueADRParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy band.Band) macCommandEnqueueState {
	if !needsADRParamSetupReq(dev, phy) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var desiredLimit ttnpb.ADRAckLimitExponent
	if dev.MACState.DesiredParameters.ADRAckLimitExponent != nil {
		desiredLimit = dev.MACState.DesiredParameters.ADRAckLimitExponent.Value
	} else if dev.MACState.CurrentParameters.ADRAckLimitExponent != nil {
		desiredLimit = dev.MACState.CurrentParameters.ADRAckLimitExponent.Value
	} else {
		desiredLimit = phy.ADRAckLimit
	}

	var desiredDelay ttnpb.ADRAckDelayExponent
	if dev.MACState.DesiredParameters.ADRAckDelayExponent != nil {
		desiredDelay = dev.MACState.DesiredParameters.ADRAckDelayExponent.Value
	} else if dev.MACState.CurrentParameters.ADRAckDelayExponent != nil {
		desiredDelay = dev.MACState.CurrentParameters.ADRAckDelayExponent.Value
	} else {
		desiredDelay = phy.ADRAckDelay
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_ADR_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}

		req := &ttnpb.MACCommand_ADRParamSetupReq{
			ADRAckLimitExponent: desiredLimit,
			ADRAckDelayExponent: desiredDelay,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"ack_limit_exponent", req.ADRAckLimitExponent,
			"ack_delay_exponent", req.ADRAckDelayExponent,
		)).Debug("Enqueued ADRParamSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			[]events.DefinitionDataClosure{
				evtEnqueueADRParamSetupRequest.BindData(req),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleADRParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) ([]events.DefinitionDataClosure, error) {
	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_ADR_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetADRParamSetupReq()

		dev.MACState.CurrentParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: req.ADRAckLimitExponent}
		dev.MACState.CurrentParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: req.ADRAckDelayExponent}

		return nil
	}, dev.MACState.PendingRequests...)
	return []events.DefinitionDataClosure{
		evtReceiveADRParamSetupAnswer.BindData(nil),
	}, err
}
