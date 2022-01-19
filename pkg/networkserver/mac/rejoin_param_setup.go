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
	EvtEnqueueRejoinParamSetupRequest = defineEnqueueMACRequestEvent(
		"rejoin_param_setup", "rejoin parameter setup",
		events.WithDataType(&ttnpb.MACCommand_RejoinParamSetupReq{}),
	)()
	EvtReceiveRejoinParamSetupAnswer = defineReceiveMACAnswerEvent(
		"rejoin_param_setup", "rejoin parameter setup",
		events.WithDataType(&ttnpb.MACCommand_RejoinParamSetupAns{}),
	)()
)

func DeviceNeedsRejoinParamSetupReq(dev *ttnpb.EndDevice) bool {
	return !dev.GetMulticast() &&
		dev.GetMacState() != nil &&
		dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0 &&
		(dev.MacState.DesiredParameters.RejoinTimePeriodicity != dev.MacState.CurrentParameters.RejoinTimePeriodicity ||
			dev.MacState.DesiredParameters.RejoinCountPeriodicity != dev.MacState.CurrentParameters.RejoinCountPeriodicity)
}

func EnqueueRejoinParamSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsRejoinParamSetupReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_REJOIN_PARAM_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}

		req := &ttnpb.MACCommand_RejoinParamSetupReq{
			MaxTimeExponent:  dev.MacState.DesiredParameters.RejoinTimePeriodicity,
			MaxCountExponent: dev.MacState.DesiredParameters.RejoinCountPeriodicity,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"max_time_exponent", req.MaxTimeExponent,
			"max_count_exponent", req.MaxCountExponent,
		)).Debug("Enqueued RejoinParamSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueRejoinParamSetupRequest.With(events.WithData(req)),
			},
			true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandleRejoinParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RejoinParamSetupAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_REJOIN_PARAM_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetRejoinParamSetupReq()

		dev.MacState.CurrentParameters.RejoinCountPeriodicity = req.MaxCountExponent
		if pld.MaxTimeExponentAck {
			dev.MacState.CurrentParameters.RejoinTimePeriodicity = req.MaxTimeExponent
		}
		return nil
	}, dev.MacState.PendingRequests...)
	return events.Builders{
		EvtReceiveRejoinParamSetupAnswer.With(events.WithData(pld)),
	}, err
}
