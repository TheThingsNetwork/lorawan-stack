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

func handleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns) error {
	if pld == nil {
		return common.ErrMissingPayload.New(nil)
	}

	first := -1
	last := -1

	cmds := dev.GetQueuedMACCommands()
outer:
	for i, cmd := range cmds {
		last = i

		switch {
		case first >= 0 && cmd.CID() != ttnpb.CID_LINK_ADR:
			break outer
		case first < 0 && cmd.CID() != ttnpb.CID_LINK_ADR:
			continue
		case first < 0:
			first = i
		}

		req := cmd.GetLinkADRReq()
		if pld.GetChannelMaskAck() {
			// TODO: Modify channels in MACState (https://github.com/TheThingsIndustries/ttn/issues/292)
		}
		if pld.GetDataRateIndexAck() {
			dev.MACState.ADRDataRateIndex = req.GetDataRateIndex()
		}
		if pld.GetTxPowerIndexAck() {
			dev.MACState.ADRTXPowerIndex = req.GetTxPowerIndex()
		}
	}

	if first < 0 {
		return ErrMACRequestNotFound.New(nil)
	}

	dev.QueuedMACCommands = append(dev.QueuedMACCommands[:first], dev.QueuedMACCommands[last+1:]...)
	return nil
}
