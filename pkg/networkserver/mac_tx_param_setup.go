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

func handleTxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) (err error) {
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_TX_PARAM_SETUP, func(cmd *ttnpb.MACCommand) {
		req := cmd.GetTxParamSetupReq()

		dev.MACState.MACParameters.DownlinkDwellTime = req.DownlinkDwellTime
		dev.MACState.MACParameters.UplinkDwellTime = req.UplinkDwellTime
		dev.MACState.MACParameters.MaxEIRP = ttnpb.DeviceEIRPToFloat32(req.MaxEIRPIndex)

		if ttnpb.Float32ToDeviceEIRP(dev.MACState.DesiredMACParameters.MaxEIRP) == req.MaxEIRPIndex {
			dev.MACState.DesiredMACParameters.MaxEIRP = dev.MACState.MACParameters.MaxEIRP
		}

	}, dev.MACState.PendingRequests...)
	return
}
