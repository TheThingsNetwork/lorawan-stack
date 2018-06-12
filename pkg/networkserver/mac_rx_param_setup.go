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

	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func handleRxParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RxParamSetupAns) error {
	if pld == nil {
		return common.ErrMissingPayload.New(nil)
	}

	cmds := dev.GetPendingMACCommands()
	for i, cmd := range cmds {
		if cmd.CID() != ttnpb.CID_RX_PARAM_SETUP {
			continue
		}

		req := cmd.GetRxParamSetupReq()
		if pld.GetRx1DataRateOffsetAck() {
			dev.MACState.Rx1DataRateOffset = req.GetRx1DataRateOffset()
		}
		if pld.GetRx2DataRateIndexAck() {
			dev.MACState.Rx2DataRateIndex = req.GetRx2DataRateIndex()
		}
		if pld.GetRx2FrequencyAck() {
			dev.MACState.Rx2Frequency = req.GetRx2Frequency()
		}

		dev.PendingMACCommands = append(cmds[:i], cmds[i+1:]...)
		return nil
	}
	return ErrMACRequestNotFound.New(nil)
}
