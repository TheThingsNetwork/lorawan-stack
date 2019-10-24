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
	evtEnqueueRxTimingSetupRequest = defineEnqueueMACRequestEvent("rx_timing_setup", "Rx timing setup")()
	evtReceiveRxTimingSetupAnswer  = defineReceiveMACAnswerEvent("rx_timing_setup", "Rx timing setup")()
)

func deviceNeedsRxTimingSetupReq(dev *ttnpb.EndDevice) bool {
	return dev.MACState != nil &&
		dev.MACState.DesiredParameters.Rx1Delay != dev.MACState.CurrentParameters.Rx1Delay
}

func enqueueRxTimingSetupReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) macCommandEnqueueState {
	if !deviceNeedsRxTimingSetupReq(dev) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_RX_TIMING_SETUP, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_RxTimingSetupReq{
			Delay: dev.MACState.DesiredParameters.Rx1Delay,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"delay", req.Delay,
		)).Debug("Enqueued RxTimingSetupReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			[]events.DefinitionDataClosure{
				evtEnqueueRxTimingSetupRequest.BindData(req),
			},
			true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleRxTimingSetupAns(ctx context.Context, dev *ttnpb.EndDevice) ([]events.DefinitionDataClosure, error) {
	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_RX_TIMING_SETUP, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetRxTimingSetupReq()

		dev.MACState.CurrentParameters.Rx1Delay = req.Delay
		return nil
	}, dev.MACState.PendingRequests...)
	return []events.DefinitionDataClosure{
		evtReceiveRxTimingSetupAnswer.BindData(nil),
	}, err
}
