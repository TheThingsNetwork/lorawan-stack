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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueDutyCycleRequest = defineEnqueueMACRequestEvent("duty_cycle", "maximum aggregated transmit duty-cycle change")()
	evtReceiveDutyCycleAnswer  = defineReceiveMACAnswerEvent("duty_cycle", "maximum aggregated transmit duty-cycle change")()
)

func enqueueDutyCycleReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) (uint16, uint16, bool) {
	if dev.MACState.DesiredParameters.MaxDutyCycle == dev.MACState.CurrentParameters.MaxDutyCycle {
		return maxDownLen, maxUpLen, true
	}

	var ok bool
	dev.MACState.PendingRequests, maxDownLen, maxUpLen, ok = enqueueMACCommand(ttnpb.CID_DUTY_CYCLE, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, false
		}
		pld := &ttnpb.MACCommand_DutyCycleReq{
			MaxDutyCycle: dev.MACState.DesiredParameters.MaxDutyCycle,
		}
		events.Publish(evtEnqueueDutyCycleRequest(ctx, dev.EndDeviceIdentifiers, pld))
		return []*ttnpb.MACCommand{pld.MACCommand()}, 1, true
	}, dev.MACState.PendingRequests...)
	return maxDownLen, maxUpLen, ok
}

func handleDutyCycleAns(ctx context.Context, dev *ttnpb.EndDevice) ([]events.DefinitionDataClosure, error) {
	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DUTY_CYCLE, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetDutyCycleReq()

		dev.MACState.CurrentParameters.MaxDutyCycle = req.MaxDutyCycle
		return nil
	}, dev.MACState.PendingRequests...)
	return []events.DefinitionDataClosure{
		evtReceiveDutyCycleAnswer.BindData(nil),
	}, err
}
