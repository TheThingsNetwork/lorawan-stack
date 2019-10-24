// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"math"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueLinkADRRequest = defineEnqueueMACRequestEvent("link_adr", "link ADR")()
	evtReceiveLinkADRAccept  = defineReceiveMACAcceptEvent("link_adr", "link ADR")()
	evtReceiveLinkADRReject  = defineReceiveMACRejectEvent("link_adr", "link ADR")()
)

func deviceNeedsLinkADRReq(dev *ttnpb.EndDevice) bool {
	if dev.MACState == nil {
		return false
	}
	if dev.MACState.DesiredParameters.ADRDataRateIndex != dev.MACState.CurrentParameters.ADRDataRateIndex ||
		dev.MACState.DesiredParameters.ADRNbTrans != dev.MACState.CurrentParameters.ADRNbTrans ||
		dev.MACState.DesiredParameters.ADRTxPowerIndex != dev.MACState.CurrentParameters.ADRTxPowerIndex {
		return true
	}
	for i := 0; i < len(dev.MACState.CurrentParameters.Channels); i++ {
		switch {
		case i >= len(dev.MACState.DesiredParameters.Channels):
			if dev.MACState.CurrentParameters.Channels[i].EnableUplink {
				return true
			}
		case dev.MACState.CurrentParameters.Channels[i].EnableUplink != dev.MACState.DesiredParameters.Channels[i].EnableUplink:
			return true
		}
	}
	return false
}

func enqueueLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy band.Band) (macCommandEnqueueState, error) {
	if !deviceNeedsLinkADRReq(dev) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}, nil
	}

	if len(dev.MACState.DesiredParameters.Channels) > int(phy.MaxUplinkChannels) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, errCorruptedMACState
	}

	desiredChs := make([]bool, phy.MaxUplinkChannels)
	for i, ch := range dev.MACState.DesiredParameters.Channels {
		desiredChs[i] = ch.EnableUplink
	}
	desiredMasks, err := phy.GenerateChMasks(desiredChs)
	if err != nil {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, err
	}
	if len(desiredMasks) > math.MaxUint16 {
		// Something is really wrong.
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, errCorruptedMACState
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_LINK_ADR, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		if int(nDown) < len(desiredMasks) {
			return nil, 0, nil, false
		}

		uplinksNeeded := uint16(1)
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			uplinksNeeded = uint16(len(desiredMasks))
		}
		if nUp < uplinksNeeded {
			return nil, 0, nil, false
		}
		evs := make([]events.DefinitionDataClosure, 0, len(desiredMasks))
		cmds := make([]*ttnpb.MACCommand, 0, len(desiredMasks))
		for i, m := range desiredMasks {
			req := &ttnpb.MACCommand_LinkADRReq{
				DataRateIndex:      dev.MACState.DesiredParameters.ADRDataRateIndex,
				NbTrans:            dev.MACState.DesiredParameters.ADRNbTrans,
				TxPowerIndex:       dev.MACState.DesiredParameters.ADRTxPowerIndex,
				ChannelMaskControl: uint32(m.Cntl),
				ChannelMask:        desiredMasks[i].Mask[:],
			}
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, evtEnqueueLinkADRRequest.BindData(req))
			log.FromContext(ctx).WithFields(log.Fields(
				"data_rate_index", req.DataRateIndex,
				"nb_trans", req.NbTrans,
				"tx_power_index", req.TxPowerIndex,
				"channel_mask_control", req.ChannelMaskControl,
				"channel_mask", req.ChannelMask,
			)).Debug("Enqueued LinkADRReq")
		}
		return cmds, uplinksNeeded, evs, true
	}, dev.MACState.PendingRequests...)
	return st, nil
}

func handleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns, dupCount uint, fps *frequencyplans.Store) ([]events.DefinitionDataClosure, error) {
	if pld == nil {
		return nil, errNoPayload
	}

	handler := handleMACResponseBlock
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		handler = handleMACResponse
	}

	if (dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0) && dupCount != 0 {
		return nil, errInvalidPayload
	}

	evt := evtReceiveLinkADRAccept
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		evt = evtReceiveLinkADRReject
	}
	evs := []events.DefinitionDataClosure{evt.BindData(pld)}

	_, phy, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return evs, err
	}

	var n uint
	var req *ttnpb.MACCommand_LinkADRReq
	dev.MACState.PendingRequests, err = handler(ttnpb.CID_LINK_ADR, func(cmd *ttnpb.MACCommand) error {
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && n > dupCount+1 {
			return errInvalidPayload
		}
		n++

		if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
			return nil
		}

		req = cmd.GetLinkADRReq()

		if req.NbTrans > 15 || len(req.ChannelMask) != 16 || req.ChannelMaskControl > 7 {
			panic("Network Server scheduled an invalid LinkADR command")
		}

		if req.NbTrans > 0 && dev.MACState.CurrentParameters.ADRNbTrans != req.NbTrans {
			dev.MACState.CurrentParameters.ADRNbTrans = req.NbTrans
			dev.RecentADRUplinks = nil
		}

		var mask [16]bool
		for i, v := range req.ChannelMask {
			mask[i] = v
		}

		m, err := phy.ParseChMask(mask, uint8(req.ChannelMaskControl))
		if err != nil {
			return err
		}

		for i, masked := range m {
			if int(i) >= len(dev.MACState.CurrentParameters.Channels) || dev.MACState.CurrentParameters.Channels[i] == nil {
				if !masked {
					continue
				}
				return errCorruptedMACState.WithCause(errUnknownChannel)
			}
			dev.MACState.CurrentParameters.Channels[i].EnableUplink = masked
		}
		return nil
	}, dev.MACState.PendingRequests...)
	if err != nil || req == nil {
		return evs, err
	}

	if dev.MACState.CurrentParameters.ADRDataRateIndex != req.DataRateIndex || dev.MACState.CurrentParameters.ADRTxPowerIndex != req.TxPowerIndex {
		dev.MACState.CurrentParameters.ADRDataRateIndex = req.DataRateIndex
		dev.MACState.CurrentParameters.ADRTxPowerIndex = req.TxPowerIndex
		dev.RecentADRUplinks = nil
	}
	return evs, nil
}
