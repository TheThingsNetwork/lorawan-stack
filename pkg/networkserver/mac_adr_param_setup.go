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

func handleADRParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (err error) {
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_ADR_PARAM_SETUP, func(cmd *ttnpb.MACCommand) {
		req := cmd.GetADRParamSetupReq()

		dev.MACState.MACParameters.ADRAckDelay = ttnpb.ADRAckDelayExponentToUint32(req.ADRAckDelayExponent)
		dev.MACState.MACParameters.ADRAckLimit = ttnpb.ADRAckLimitExponentToUint32(req.ADRAckLimitExponent)

		if ttnpb.Uint32ToADRAckDelayExponent(dev.MACState.DesiredMACParameters.ADRAckDelay) == req.ADRAckDelayExponent {
			dev.MACState.DesiredMACParameters.ADRAckDelay = dev.MACState.MACParameters.ADRAckDelay
		}

		if ttnpb.Uint32ToADRAckLimitExponent(dev.MACState.DesiredMACParameters.ADRAckLimit) == req.ADRAckLimitExponent {
			dev.MACState.DesiredMACParameters.ADRAckLimit = dev.MACState.MACParameters.ADRAckLimit
		}

	}, dev.MACState.PendingRequests...)
	return
}
