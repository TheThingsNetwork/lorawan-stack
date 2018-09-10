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
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtMACLinkADRRequest = events.Define("ns.mac.adr.request", "request ADR") // TODO(#988): publish when requesting
	evtMACLinkADRAccept  = events.Define("ns.mac.adr.accept", "device accepted ADR request")
	evtMACLinkADRReject  = events.Define("ns.mac.adr.reject", "device rejected ADR request")
)

func handleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns, dupCount uint, fps *frequencyplans.Store) (err error) {
	if pld == nil {
		return errMissingPayload
	}

	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return err
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return err
	}

	handler := handleMACResponseBlock
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		handler = handleMACResponse
	}

	if dev.MACState.LoRaWANVersion != ttnpb.MAC_V1_0_2 && dupCount != 0 {
		return errInvalidPayload
	}

	var n uint
	var req *ttnpb.MACCommand_LinkADRReq
	dev.MACState.PendingRequests, err = handler(ttnpb.CID_LINK_ADR, func(cmd *ttnpb.MACCommand) error {
		if dev.MACState.LoRaWANVersion == ttnpb.MAC_V1_0_2 && n > dupCount+1 {
			return errInvalidPayload
		}
		n++

		if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
			// TODO: Handle NACK, modify desired state
			// (https://github.com/TheThingsIndustries/ttn/issues/834)
			events.Publish(evtMACLinkADRReject(ctx, dev.EndDeviceIdentifiers, pld))
			return nil
		}

		req = cmd.GetLinkADRReq()

		if req.NbTrans > 15 || len(req.ChannelMask) != 16 || req.ChannelMaskControl > 7 {
			panic("Network Server scheduled an invalid LinkADR command")
		}

		if req.NbTrans > 0 {
			dev.MACState.ADRNbTrans = req.NbTrans
		}

		var mask [16]bool
		for i, v := range req.ChannelMask {
			mask[i] = v
		}

		m, err := band.ChannelMask(mask, uint8(req.ChannelMaskControl))
		if err != nil {
			return err
		}

		for i, masked := range m {
			if i >= len(dev.MACState.Channels) || dev.MACState.MACParameters.Channels[i] == nil {
				if !masked {
					continue
				}
				return errCorruptedMACState.WithCause(errUnknownChannel)
			}
			dev.MACState.MACParameters.Channels[i].UplinkEnabled = masked
		}

		events.Publish(evtMACLinkADRAccept(ctx, dev.EndDeviceIdentifiers, req))
		return nil

	}, dev.MACState.PendingRequests...)
	if err != nil {
		return err
	}

	dev.MACState.ADRDataRateIndex = req.DataRateIndex
	dev.MACState.ADRTxPowerIndex = req.TxPowerIndex
	return nil
}
