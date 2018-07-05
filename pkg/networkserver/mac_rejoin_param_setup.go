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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func handleRejoinParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RejoinParamSetupAns) (err error) {
	if pld == nil {
		return errMissingPayload
	}

	dev.PendingMACRequests, err = handleMACResponse(ttnpb.CID_REJOIN_PARAM_SETUP, func(cmd *ttnpb.MACCommand) {
		req := cmd.GetRejoinParamSetupReq()

		// TODO: Handle (https://github.com/TheThingsIndustries/ttn/issues/834)
		_ = req.MaxCountExponent
		if pld.MaxTimeExponentAck {
			_ = req.MaxTimeExponent
		}

	}, dev.PendingMACRequests...)
	return
}
