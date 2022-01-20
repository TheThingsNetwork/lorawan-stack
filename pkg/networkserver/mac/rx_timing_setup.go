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
	EvtEnqueueRxTimingSetupRequest = defineEnqueueMACRequestEvent(
		"rx_timing_setup", "Rx timing setup",
		events.WithDataType(&ttnpb.MACCommand_RxTimingSetupReq{}),
	)()
	EvtReceiveRxTimingSetupAnswer = defineReceiveMACAnswerEvent(
		"rx_timing_setup", "Rx timing setup",
	)()
)

func DeviceNeedsRxTimingSetupReq(dev *ttnpb.EndDevice) bool {
	return !dev.GetMulticast() &&
		dev.GetMacState() != nil &&
		dev.MacState.DesiredParameters.Rx1Delay != dev.MacState.CurrentParameters.Rx1Delay
}

func EnqueueRxTimingSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsRxTimingSetupReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_RxTimingSetupReq{
			Delay: dev.MacState.DesiredParameters.Rx1Delay,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"delay", req.Delay,
		)).Debug("Enqueued RxTimingSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueueRxTimingSetupRequest.With(events.WithData(req)),
			},
			true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandleRxTimingSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (events.Builders, error) {
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetRxTimingSetupReq()

		dev.MacState.CurrentParameters.Rx1Delay = req.Delay
		return nil
	}, dev.MacState.PendingRequests...)
	return events.Builders{
		EvtReceiveRxTimingSetupAnswer,
	}, err
}
