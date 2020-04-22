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

func deviceNeedsLinkADRReq(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) bool {
	if dev.MACState == nil {
		return false
	}
	// TODO: Check that a LinkADRReq *can* be scheduled given the rejections received so far. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2192)
	for i, currentCh := range dev.MACState.CurrentParameters.Channels {
		switch {
		case i >= len(dev.MACState.DesiredParameters.Channels):
			if currentCh.GetEnableUplink() {
				return true
			}
		case currentCh.GetEnableUplink() != dev.MACState.DesiredParameters.Channels[i].GetEnableUplink():
			return true
		}
	}
	if dev.MACState.DesiredParameters.ADRNbTrans != dev.MACState.CurrentParameters.ADRNbTrans {
		return true
	}
	if !deviceUseADR(dev, defaults) {
		return false
	}
	return dev.MACState.DesiredParameters.ADRDataRateIndex != dev.MACState.CurrentParameters.ADRDataRateIndex ||
		dev.MACState.DesiredParameters.ADRTxPowerIndex != dev.MACState.CurrentParameters.ADRTxPowerIndex
}

const (
	noChangeDataRateIndex = ttnpb.DATA_RATE_15
	noChangeTXPowerIndex  = 15
)

func enqueueLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, defaults ttnpb.MACSettings, phy band.Band) (macCommandEnqueueState, error) {
	if !deviceNeedsLinkADRReq(dev, defaults) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}, nil
	}
	if len(dev.MACState.DesiredParameters.Channels) == 0 ||
		len(dev.MACState.DesiredParameters.Channels) > int(phy.MaxUplinkChannels) ||
		dev.MACState.DesiredParameters.ADRTxPowerIndex > uint32(phy.MaxTxPowerIndex()) ||
		dev.MACState.DesiredParameters.ADRDataRateIndex > phy.MaxADRDataRateIndex {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, errCorruptedMACState.New()
	}
	minDataRateIndex := dev.MACState.DesiredParameters.Channels[0].MinDataRateIndex
	maxDataRateIndex := dev.MACState.DesiredParameters.Channels[0].MaxDataRateIndex
	for _, ch := range dev.MACState.DesiredParameters.Channels {
		if ch.MinDataRateIndex < minDataRateIndex {
			minDataRateIndex = ch.MinDataRateIndex
		}
		if ch.MaxDataRateIndex < maxDataRateIndex {
			maxDataRateIndex = ch.MaxDataRateIndex
		}
	}
	if dev.MACState.DesiredParameters.ADRDataRateIndex < minDataRateIndex || dev.MACState.DesiredParameters.ADRDataRateIndex > maxDataRateIndex {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, errCorruptedMACState.New()
	}

	currentChs := make([]bool, phy.MaxUplinkChannels)
	for i, ch := range dev.MACState.CurrentParameters.Channels {
		currentChs[i] = ch.GetEnableUplink()
	}
	desiredChs := make([]bool, phy.MaxUplinkChannels)
	for i, ch := range dev.MACState.DesiredParameters.Channels {
		if ch.GetEnableUplink() && ch.UplinkFrequency == 0 {
			return macCommandEnqueueState{
				MaxDownLen: maxDownLen,
				MaxUpLen:   maxUpLen,
			}, errCorruptedMACState.New()
		}
		if deviceNeedsNewChannelReqAtIndex(dev, i) {
			currentChs[i] = ch != nil && ch.UplinkFrequency != 0
		}
		desiredChs[i] = ch.GetEnableUplink()
	}
	desiredMasks, err := phy.GenerateChMasks(currentChs, desiredChs)
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
		}, errCorruptedMACState.New()
	}

	drIdx := dev.MACState.DesiredParameters.ADRDataRateIndex
	txPowerIdx := dev.MACState.DesiredParameters.ADRTxPowerIndex
	switch {
	case !deviceRejectedADRDataRateIndex(dev, drIdx) && !deviceRejectedADRTXPowerIndex(dev, txPowerIdx):
		// Only send the desired DataRateIndex and TXPowerIndex if neither of them were rejected.

	case len(desiredMasks) == 0 && dev.MACState.DesiredParameters.ADRNbTrans == dev.MACState.CurrentParameters.ADRNbTrans:
		log.FromContext(ctx).Debug("Either desired data rate index or TX power output index have been rejected and there are no channel mask and NbTrans changes desired, avoid enqueueing LinkADRReq")
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, nil

	case dev.MACState.LoRaWANVersion.HasNoChangeDataRateIndex() && !deviceRejectedADRDataRateIndex(dev, noChangeDataRateIndex) &&
		dev.MACState.LoRaWANVersion.HasNoChangeTXPowerIndex() && !deviceRejectedADRTXPowerIndex(dev, noChangeTXPowerIndex):
		drIdx = noChangeDataRateIndex
		txPowerIdx = noChangeTXPowerIndex

	default:
		drIdx = minDataRateIndex
		for drIdx < maxDataRateIndex {
			if deviceRejectedADRDataRateIndex(dev, drIdx) {
				drIdx++
				continue
			}
		}
		txPowerIdx = 0
		for txPowerIdx < uint32(phy.MaxTxPowerIndex()) {
			if deviceRejectedADRTXPowerIndex(dev, txPowerIdx) {
				txPowerIdx++
				continue
			}
		}
		if deviceRejectedADRDataRateIndex(dev, drIdx) || deviceRejectedADRTXPowerIndex(dev, txPowerIdx) {
			log.FromContext(ctx).Warn("Device rejected either all available data rate indexes or all available TX power output indexes combinations and there are channel mask or NbTrans changes desired, avoid enqueueing LinkADRReq")
			return macCommandEnqueueState{
				MaxDownLen: maxDownLen,
				MaxUpLen:   maxUpLen,
			}, nil
		}
	}
	if drIdx == dev.MACState.CurrentParameters.ADRDataRateIndex && dev.MACState.LoRaWANVersion.HasNoChangeDataRateIndex() && !deviceRejectedADRDataRateIndex(dev, noChangeDataRateIndex) {
		drIdx = noChangeDataRateIndex
	}
	if txPowerIdx == dev.MACState.CurrentParameters.ADRTxPowerIndex && dev.MACState.LoRaWANVersion.HasNoChangeTXPowerIndex() && !deviceRejectedADRTXPowerIndex(dev, noChangeTXPowerIndex) {
		txPowerIdx = noChangeTXPowerIndex
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
				DataRateIndex:      drIdx,
				TxPowerIndex:       txPowerIdx,
				NbTrans:            dev.MACState.DesiredParameters.ADRNbTrans,
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
		return nil, errNoPayload.New()
	}
	if (dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0) && dupCount != 0 {
		return nil, errInvalidPayload.New()
	}

	evt := evtReceiveLinkADRAccept
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		evt = evtReceiveLinkADRReject

		// See "Table 6: LinkADRAns status bits signification" of LoRaWAN 1.1 specification
		if !pld.ChannelMaskAck {
			log.FromContext(ctx).Warn("Either Network Server sent a channel mask, which enables a yet undefined channel or requires all channels to be disabled, or device is malfunctioning.")
		}
	}
	evs := []events.DefinitionDataClosure{evt.BindData(pld)}

	_, phy, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return evs, err
	}

	handler := handleMACResponseBlock
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		handler = handleMACResponse
	}
	var n uint
	var req *ttnpb.MACCommand_LinkADRReq
	dev.MACState.PendingRequests, err = handler(ttnpb.CID_LINK_ADR, func(cmd *ttnpb.MACCommand) error {
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && n > dupCount+1 {
			return errInvalidPayload.New()
		}
		n++

		req = cmd.GetLinkADRReq()
		if req.NbTrans > 15 || len(req.ChannelMask) != 16 || req.ChannelMaskControl > 7 {
			panic("Network Server scheduled an invalid LinkADR command")
		}
		if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
			return nil
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

	if !pld.DataRateIndexAck {
		if i := searchDataRateIndex(req.DataRateIndex, dev.MACState.RejectedADRDataRateIndexes...); i == len(dev.MACState.RejectedADRDataRateIndexes) || dev.MACState.RejectedADRDataRateIndexes[i] != req.DataRateIndex {
			dev.MACState.RejectedADRDataRateIndexes = append(dev.MACState.RejectedADRDataRateIndexes, ttnpb.DATA_RATE_0)
			copy(dev.MACState.RejectedADRDataRateIndexes[i+1:], dev.MACState.RejectedADRDataRateIndexes[i:])
			dev.MACState.RejectedADRDataRateIndexes[i] = req.DataRateIndex
		}
	}
	if !pld.TxPowerIndexAck {
		if i := searchUint32(req.TxPowerIndex, dev.MACState.RejectedADRTxPowerIndexes...); i == len(dev.MACState.RejectedADRTxPowerIndexes) || dev.MACState.RejectedADRTxPowerIndexes[i] != req.TxPowerIndex {
			dev.MACState.RejectedADRTxPowerIndexes = append(dev.MACState.RejectedADRTxPowerIndexes, 0)
			copy(dev.MACState.RejectedADRTxPowerIndexes[i+1:], dev.MACState.RejectedADRTxPowerIndexes[i:])
			dev.MACState.RejectedADRTxPowerIndexes[i] = req.TxPowerIndex
		}
	}
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		return evs, nil
	}
	if !dev.MACState.LoRaWANVersion.HasNoChangeDataRateIndex() || req.DataRateIndex != noChangeDataRateIndex {
		dev.MACState.CurrentParameters.ADRDataRateIndex = req.DataRateIndex
		dev.RecentADRUplinks = nil
	}
	if !dev.MACState.LoRaWANVersion.HasNoChangeTXPowerIndex() || req.TxPowerIndex != noChangeTXPowerIndex {
		dev.MACState.CurrentParameters.ADRTxPowerIndex = req.TxPowerIndex
		dev.RecentADRUplinks = nil
	}
	if req.NbTrans > 0 && dev.MACState.CurrentParameters.ADRNbTrans != req.NbTrans {
		dev.MACState.CurrentParameters.ADRNbTrans = req.NbTrans
		dev.RecentADRUplinks = nil
	}
	return evs, nil
}
