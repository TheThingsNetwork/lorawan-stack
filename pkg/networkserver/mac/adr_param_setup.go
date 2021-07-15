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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueADRParamSetupRequest = defineEnqueueMACRequestEvent(
		"adr_param_setup", "ADR parameter setup",
		events.WithDataType(&ttnpb.MACCommand_ADRParamSetupReq{}),
	)()
	EvtReceiveADRParamSetupAnswer = defineReceiveMACAnswerEvent(
		"adr_param_setup", "ADR parameter setup",
	)()
)

func deviceADRAckLimit(dev *ttnpb.EndDevice, phy *band.Band) ttnpb.ADRAckLimitExponent {
	if dev.MACState.CurrentParameters.AdrAckLimitExponent != nil {
		return dev.MACState.CurrentParameters.AdrAckLimitExponent.Value
	}
	return phy.ADRAckLimit
}

func deviceADRAckDelay(dev *ttnpb.EndDevice, phy *band.Band) ttnpb.ADRAckDelayExponent {
	if dev.MACState.CurrentParameters.AdrAckDelayExponent != nil {
		return dev.MACState.CurrentParameters.AdrAckDelayExponent.Value
	}
	return phy.ADRAckDelay
}

func DeviceNeedsADRParamSetupReq(dev *ttnpb.EndDevice, phy *band.Band) bool {
	if dev.GetMulticast() ||
		dev.GetMACState() == nil ||
		dev.MACState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		return false
	}
	return dev.MACState.DesiredParameters.AdrAckLimitExponent != nil &&
		deviceADRAckLimit(dev, phy) != dev.MACState.DesiredParameters.AdrAckLimitExponent.Value ||
		dev.MACState.DesiredParameters.AdrAckDelayExponent != nil &&
			deviceADRAckDelay(dev, phy) != dev.MACState.DesiredParameters.AdrAckDelayExponent.Value
}

func EnqueueADRParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy *band.Band) EnqueueState {
	if !DeviceNeedsADRParamSetupReq(dev, phy) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var desiredLimit ttnpb.ADRAckLimitExponent
	if dev.MACState.DesiredParameters.AdrAckLimitExponent != nil {
		desiredLimit = dev.MACState.DesiredParameters.AdrAckLimitExponent.Value
	} else {
		desiredLimit = deviceADRAckLimit(dev, phy)
	}

	var desiredDelay ttnpb.ADRAckDelayExponent
	if dev.MACState.DesiredParameters.AdrAckDelayExponent != nil {
		desiredDelay = dev.MACState.DesiredParameters.AdrAckDelayExponent.Value
	} else {
		desiredDelay = deviceADRAckDelay(dev, phy)
	}

	var st EnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_ADR_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}

		req := &ttnpb.MACCommand_ADRParamSetupReq{
			AdrAckLimitExponent: desiredLimit,
			AdrAckDelayExponent: desiredDelay,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"ack_limit_exponent", req.AdrAckLimitExponent,
			"ack_delay_exponent", req.AdrAckDelayExponent,
		)).Debug("Enqueued ADRParamSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueADRParamSetupRequest.With(events.WithData(req)),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func HandleADRParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (events.Builders, error) {
	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_ADR_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetAdrParamSetupReq()

		dev.MACState.CurrentParameters.AdrAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: req.AdrAckLimitExponent}
		dev.MACState.CurrentParameters.AdrAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: req.AdrAckDelayExponent}

		return nil
	}, dev.MACState.PendingRequests...)
	return events.Builders{
		EvtReceiveADRParamSetupAnswer,
	}, err
}
