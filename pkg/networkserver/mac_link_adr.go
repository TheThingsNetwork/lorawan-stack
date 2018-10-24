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
	evtEnqueueLinkADRRequest = defineEnqueueMACRequestEvent("link_adr", "ADR request")()
	evtReceiveLinkADRAccept  = defineReceiveMACAcceptEvent("link_adr", "ADR request")()
	evtReceiveLinkADRReject  = defineReceiveMACRejectEvent("link_adr", "ADR request")()
)

func enqueueLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) (uint16, uint16, bool) {
	if dev.MACState.DesiredParameters.ADRDataRateIndex == dev.MACState.CurrentParameters.ADRDataRateIndex &&
		dev.MACState.DesiredParameters.ADRNbTrans == dev.MACState.CurrentParameters.ADRNbTrans &&
		dev.MACState.DesiredParameters.ADRTxPowerIndex == dev.MACState.CurrentParameters.ADRTxPowerIndex {
		return maxDownLen, maxUpLen, true
	}

	var ok bool
	dev.MACState.PendingRequests, maxDownLen, maxUpLen, ok = enqueueMACCommand(ttnpb.CID_LINK_ADR, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, false
		}
		pld := &ttnpb.MACCommand_LinkADRReq{
			DataRateIndex: dev.MACState.DesiredParameters.ADRDataRateIndex,
			NbTrans:       dev.MACState.DesiredParameters.ADRNbTrans,
			TxPowerIndex:  dev.MACState.DesiredParameters.ADRTxPowerIndex,
			// NOTE: This is invalid in most of the cases.
			// TODO: Generate proper ChannelMask (https://github.com/TheThingsIndustries/lorawan-stack/issues/1235)
			ChannelMask: []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		}
		events.Publish(evtEnqueueLinkADRRequest(ctx, dev.EndDeviceIdentifiers, pld))
		return []*ttnpb.MACCommand{pld.MACCommand()}, 1, true

	}, dev.MACState.PendingRequests...)
	return maxDownLen, maxUpLen, ok
}

func handleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns, dupCount uint, fps *frequencyplans.Store) (err error) {
	if pld == nil {
		return errNoPayload
	}

	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		events.Publish(evtReceiveLinkADRReject(ctx, dev.EndDeviceIdentifiers, pld))
	} else {
		events.Publish(evtReceiveLinkADRAccept(ctx, dev.EndDeviceIdentifiers, pld))
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
			return nil
		}

		req = cmd.GetLinkADRReq()

		if req.NbTrans > 15 || len(req.ChannelMask) != 16 || req.ChannelMaskControl > 7 {
			panic("Network Server scheduled an invalid LinkADR command")
		}

		if req.NbTrans > 0 {
			dev.MACState.CurrentParameters.ADRNbTrans = req.NbTrans
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
			if i >= len(dev.MACState.CurrentParameters.Channels) || dev.MACState.CurrentParameters.Channels[i] == nil {
				if !masked {
					continue
				}
				return errCorruptedMACState.WithCause(errUnknownChannel)
			}
			dev.MACState.CurrentParameters.Channels[i].EnableUplink = masked
		}

		return nil

	}, dev.MACState.PendingRequests...)
	if err != nil {
		return err
	}

	dev.MACState.CurrentParameters.ADRDataRateIndex = req.DataRateIndex
	dev.MACState.CurrentParameters.ADRTxPowerIndex = req.TxPowerIndex
	return nil
}
