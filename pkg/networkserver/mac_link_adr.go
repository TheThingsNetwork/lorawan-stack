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
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtMACLinkADRRequest = events.Define("ns.mac.adr.request", "request ADR") // TODO(#988): publish when requesting
	evtMACLinkADRAccept  = events.Define("ns.mac.adr.accept", "device accepted ADR request")
	evtMACLinkADRReject  = events.Define("ns.mac.adr.reject", "device rejected ADR request")
)

func handleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns, fps *frequencyplans.Store) (err error) {
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

	logger := log.FromContext(ctx)

	dev.MACState.PendingRequests, err = handleMACResponseBlock(ttnpb.CID_LINK_ADR, func(cmd *ttnpb.MACCommand) {
		if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
			// TODO: Handle NACK, modify desired state
			// (https://github.com/TheThingsIndustries/ttn/issues/834)
			events.Publish(evtMACLinkADRReject(ctx, dev.EndDeviceIdentifiers, pld))
			return
		}

		req := cmd.GetLinkADRReq()

		// TODO: Ensure LoRaWAN1.0* compatibility (https://github.com/TheThingsIndustries/ttn/issues/870)

		if req.NbTrans > 15 || len(req.ChannelMask) != 16 || req.ChannelMaskControl > 7 {
			logger.Error("Network Server scheduled an invalid LinkADR command, assuming device dropped the request")
			return
		}

		if req.NbTrans > 0 {
			dev.MACState.ADRNbTrans = req.NbTrans
		}

		var m map[int]bool
		if band.ChanelMask == nil {
			// TODO: This check should probably be removed once all band structs contain ChannelMask field.
			m = make(map[int]bool, 16)
			for i, v := range req.ChannelMask {
				m[i] = v
			}
		} else {
			var mask [16]bool
			for i, v := range req.ChannelMask {
				mask[i] = v
			}

			// NOTE: err references the error outside the scope of this function.
			m, err = band.ChanelMask(mask, uint8(req.ChannelMaskControl))
			if err != nil {
				logger.WithError(err).Error("Failed to determine channel mask")
				return
			}
		}

		for i, masked := range m {
			if i >= len(dev.MACState.Channels) {
				dev.MACState.MACParameters.Channels = append(dev.MACState.MACParameters.Channels, make([]*ttnpb.MACParameters_Channel, 1+i-len(dev.MACState.MACParameters.Channels))...)
			}

			ch := dev.MACState.MACParameters.Channels[i]
			if ch == nil {
				ch = &ttnpb.MACParameters_Channel{}
				dev.MACState.MACParameters.Channels[i] = ch
			}
			ch.UplinkEnabled = masked
		}

		dev.MACState.ADRDataRateIndex = req.DataRateIndex
		dev.MACState.ADRTxPowerIndex = req.TxPowerIndex

		events.Publish(evtMACLinkADRAccept(ctx, dev.EndDeviceIdentifiers, req))
	}, dev.MACState.PendingRequests...)
	return
}
