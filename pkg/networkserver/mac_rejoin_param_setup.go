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
	evtMACRejoinParamRequest = events.Define("ns.mac.rejoin_param.request", "request rejoin parameter setup") // TODO(#988): publish when requesting
	evtMACRejoinParamAccept  = events.Define("ns.mac.rejoin_param.accept", "device accepted rejoin parameter setup request")
)

func handleRejoinParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RejoinParamSetupAns) (err error) {
	if pld == nil {
		return errMissingPayload
	}

	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_REJOIN_PARAM_SETUP, func(cmd *ttnpb.MACCommand) {
		req := cmd.GetRejoinParamSetupReq()
		var acked ttnpb.MACCommand_RejoinParamSetupReq

		// TODO: Handle (https://github.com/TheThingsIndustries/ttn/issues/834)
		acked.MaxCountExponent = req.MaxCountExponent
		if pld.MaxTimeExponentAck {
			acked.MaxTimeExponent = req.MaxTimeExponent
		}

		events.Publish(evtMACRejoinParamAccept(ctx, dev.EndDeviceIdentifiers, &acked))
	}, dev.MACState.PendingRequests...)
	return
}
