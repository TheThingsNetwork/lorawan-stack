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

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func handleResetInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_ResetInd, fp *ttnpb.FrequencyPlan) error {
	if pld == nil {
		return common.ErrMissingPayload.New(nil)
	}

	band, err := band.GetByID(fp.GetBandID())
	if err != nil {
		return err
	}

	dev.MACState = newMACState(&band, uint32(dev.GetMaxTxPower()), fp.DwellTime != nil)
	dev.MACStateDesired = dev.MACState
	dev.QueuedMACCommands = append(
		dev.GetQueuedMACCommands(),
		(&ttnpb.MACCommand_ResetConf{
			MinorVersion: pld.GetMinorVersion(),
		}).MACCommand(),
	)
	return nil
}
