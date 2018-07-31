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
	evtMACLinkADRRequest = events.Define("ns.mac.adr.request", "request ADR") // TODO(#988): publish when requesting
	evtMACLinkADRAccept  = events.Define("ns.mac.adr.accept", "device accepted ADR request")
	evtMACLinkADRReject  = events.Define("ns.mac.adr.reject", "device rejected ADR request")
)

func handleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns) (err error) {
	if pld == nil {
		return errMissingPayload
	}

	dev.MACState.PendingRequests, err = handleMACResponseBlock(ttnpb.CID_LINK_ADR, func(cmd *ttnpb.MACCommand) {
		if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
			// TODO: Handle NACK, modify desired state
			// (https://github.com/TheThingsIndustries/ttn/issues/834)
			events.Publish(evtMACLinkADRReject(ctx, dev.EndDeviceIdentifiers, pld))
			return
		}

		req := cmd.GetLinkADRReq()

		// TODO: Ensure LoRaWAN1.0* compatibility (https://github.com/TheThingsIndustries/ttn/issues/870)

		// TODO: Modify channels in MACState (https://github.com/TheThingsIndustries/ttn/issues/292)
		_ = req.NbTrans
		_ = req.ChannelMask
		_ = req.ChannelMaskControl

		dev.MACState.ADRDataRateIndex = req.DataRateIndex
		dev.MACState.ADRTxPowerIndex = req.TxPowerIndex

		events.Publish(evtMACLinkADRAccept(ctx, dev.EndDeviceIdentifiers, req))
	}, dev.MACState.PendingRequests...)
	return
}
