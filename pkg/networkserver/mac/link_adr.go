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

package mac

import (
	"context"
	"math"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueLinkADRRequest = defineEnqueueMACRequestEvent(
		"link_adr", "link ADR",
		events.WithDataType(&ttnpb.MACCommand_LinkADRReq{}),
	)()
	EvtReceiveLinkADRAccept = defineReceiveMACAcceptEvent(
		"link_adr", "link ADR",
		events.WithDataType(&ttnpb.MACCommand_LinkADRAns{}),
	)()
	EvtReceiveLinkADRReject = defineReceiveMACRejectEvent(
		"link_adr", "link ADR",
		events.WithDataType(&ttnpb.MACCommand_LinkADRAns{}),
	)()
)

const (
	noChangeDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_15
	noChangeTXPowerIndex  = 15
)

type linkADRReqParameters struct {
	Masks         []band.ChMaskCntlPair
	DataRateIndex ttnpb.DataRateIndex
	TxPowerIndex  uint32
	NbTrans       uint32
}

func generateLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band) (linkADRReqParameters, bool, error) {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return linkADRReqParameters{}, false, nil
	}
	if len(dev.MacState.DesiredParameters.Channels) > int(phy.MaxUplinkChannels) {
		return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
			WithAttributes(
				"desired_channels_len", len(dev.MacState.DesiredParameters.Channels),
				"phy_max_uplink_channels", phy.MaxUplinkChannels,
			).
			WithCause(internal.ErrUnknownChannel)
	}

	currentChs := make([]bool, phy.MaxUplinkChannels)
	for i, ch := range dev.MacState.CurrentParameters.Channels {
		currentChs[i] = ch.GetEnableUplink()
	}
	desiredChs := make([]bool, phy.MaxUplinkChannels)
	for i, ch := range dev.MacState.DesiredParameters.Channels {
		isEnabled := ch.GetEnableUplink()
		if isEnabled && ch.UplinkFrequency == 0 {
			return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
				WithAttributes(
					"i", i,
					"enabled", isEnabled,
					"uplink_frequency", ch.UplinkFrequency,
				).
				WithCause(internal.ErrDownlinkChannel)
		}
		if DeviceNeedsNewChannelReqAtIndex(dev, i) {
			currentChs[i] = ch != nil && ch.UplinkFrequency != 0
		}
		desiredChs[i] = isEnabled
	}

	switch {
	case !band.EqualChMasks(currentChs, desiredChs):
		// NOTE: LinkADRReq is scheduled regardless of ADR settings if channel mask is required, which often is the case with ABP devices or when ChMask CFList is not supported/used.
	case dev.MacState.DesiredParameters.AdrNbTrans != dev.MacState.CurrentParameters.AdrNbTrans,
		dev.MacState.DesiredParameters.AdrDataRateIndex != dev.MacState.CurrentParameters.AdrDataRateIndex,
		dev.MacState.DesiredParameters.AdrTxPowerIndex != dev.MacState.CurrentParameters.AdrTxPowerIndex:
	default:
		return linkADRReqParameters{}, false, nil
	}
	desiredMasks, err := phy.GenerateChMasks(currentChs, desiredChs)
	if err != nil {
		return linkADRReqParameters{}, false, err
	}
	if len(desiredMasks) > math.MaxUint16 {
		// Something is really wrong.
		return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
			WithAttributes(
				"len", len(desiredMasks),
			).
			WithCause(internal.ErrChannelMask)
	}

	var (
		drIdx      ttnpb.DataRateIndex
		txPowerIdx uint32
		nbTrans    uint32
	)
	minDataRateIndex, maxDataRateIndex, ok := channelDataRateRange(dev.MacState.DesiredParameters.Channels...)
	if !ok {
		return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
			WithCause(internal.ErrChannelDataRateRange)
	}

	if dev.MacState.DesiredParameters.AdrTxPowerIndex != dev.MacState.CurrentParameters.AdrTxPowerIndex {
		attributes := []interface{}{
			"current_adr_tx_power_index", dev.MacState.CurrentParameters.AdrTxPowerIndex,
			"desired_adr_tx_power_index", dev.MacState.DesiredParameters.AdrTxPowerIndex,
		}
		switch {
		case dev.MacState.DesiredParameters.AdrTxPowerIndex > uint32(phy.MaxTxPowerIndex()):
			return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
				WithAttributes(append(attributes,
					"phy_max_tx_power_index", phy.MaxTxPowerIndex(),
				)...)
		}
	}
	if dev.MacState.DesiredParameters.AdrDataRateIndex != dev.MacState.CurrentParameters.AdrDataRateIndex {
		attributes := []interface{}{
			"current_adr_data_rate_index", dev.MacState.CurrentParameters.AdrDataRateIndex,
			"desired_adr_data_rate_index", dev.MacState.DesiredParameters.AdrDataRateIndex,
		}
		switch {
		case dev.MacState.DesiredParameters.AdrDataRateIndex > phy.MaxADRDataRateIndex:
			return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
				WithAttributes(append(attributes,
					"phy_max_adr_data_rate_index", phy.MaxADRDataRateIndex,
				)...)
		case dev.MacState.DesiredParameters.AdrDataRateIndex < minDataRateIndex:
			return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
				WithAttributes(append(attributes,
					"min_adr_data_rate_index", minDataRateIndex,
				)...)
		case dev.MacState.DesiredParameters.AdrDataRateIndex > maxDataRateIndex:
			return linkADRReqParameters{}, false, internal.ErrCorruptedMACState.
				WithAttributes(append(attributes,
					"max_adr_data_rate_index", maxDataRateIndex,
				)...)
		}
	}

	drIdx = dev.MacState.DesiredParameters.AdrDataRateIndex
	txPowerIdx = dev.MacState.DesiredParameters.AdrTxPowerIndex
	nbTrans = dev.MacState.DesiredParameters.AdrNbTrans
	resetDRTXToCurrent := func() {
		drIdx = dev.MacState.CurrentParameters.AdrDataRateIndex
		txPowerIdx = dev.MacState.CurrentParameters.AdrTxPowerIndex
	}
	switch {
	case !deviceRejectedADRDataRateIndex(dev, drIdx) && !deviceRejectedADRTXPowerIndex(dev, txPowerIdx):
		// Only send the desired DataRateIndex and TXPowerIndex if neither of them were rejected.

	case len(desiredMasks) == 0 && dev.MacState.DesiredParameters.AdrNbTrans == dev.MacState.CurrentParameters.AdrNbTrans:
		log.FromContext(ctx).Debug("Either desired data rate index or TX power output index have been rejected and there are no channel mask and NbTrans changes desired, avoid enqueueing LinkADRReq")
		return linkADRReqParameters{}, false, nil

	case dev.MacState.LorawanVersion.HasNoChangeDataRateIndex() && !deviceRejectedADRDataRateIndex(dev, noChangeDataRateIndex) &&
		dev.MacState.LorawanVersion.HasNoChangeTXPowerIndex() && !deviceRejectedADRTXPowerIndex(dev, noChangeTXPowerIndex):
		drIdx = noChangeDataRateIndex
		txPowerIdx = noChangeTXPowerIndex

	default:
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"current_adr_nb_trans", dev.MacState.CurrentParameters.AdrNbTrans,
			"desired_adr_nb_trans", dev.MacState.DesiredParameters.AdrNbTrans,
			"desired_mask_count", len(desiredMasks),
		))
		for deviceRejectedADRDataRateIndex(dev, drIdx) || deviceRejectedADRTXPowerIndex(dev, txPowerIdx) {
			if drIdx < minDataRateIndex {
				logger.Warn("Device desired data rate is under the minimum data rate for ADR. Avoiding data rate and TX power changes")
				resetDRTXToCurrent()
				break
			}
			// Since either data rate or TX power index (or both) were rejected by the device, undo the
			// desired ADR adjustments step-by-step until possibly fitting index pair is found.
			if drIdx == minDataRateIndex && (deviceRejectedADRDataRateIndex(dev, drIdx) || txPowerIdx == 0) {
				logger.Warn("Device rejected either all available data rate indexes or all available TX power output indexes. Avoiding data rate and TX power changes")
				resetDRTXToCurrent()
				break
			}
			for drIdx > minDataRateIndex && (deviceRejectedADRDataRateIndex(dev, drIdx) || txPowerIdx == 0 && deviceRejectedADRTXPowerIndex(dev, txPowerIdx)) {
				// Increase data rate until a non-rejected index is found.
				// Set TX power to maximum possible value.
				drIdx--
				txPowerIdx = uint32(phy.MaxTxPowerIndex())
			}
			for txPowerIdx > 0 && deviceRejectedADRTXPowerIndex(dev, txPowerIdx) {
				// Increase TX output power until a non-rejected index is found.
				txPowerIdx--
			}
		}
	}
	if drIdx == dev.MacState.CurrentParameters.AdrDataRateIndex && dev.MacState.LorawanVersion.HasNoChangeDataRateIndex() && !deviceRejectedADRDataRateIndex(dev, noChangeDataRateIndex) {
		drIdx = noChangeDataRateIndex
	}
	if txPowerIdx == dev.MacState.CurrentParameters.AdrTxPowerIndex && dev.MacState.LorawanVersion.HasNoChangeTXPowerIndex() && !deviceRejectedADRTXPowerIndex(dev, noChangeTXPowerIndex) {
		txPowerIdx = noChangeTXPowerIndex
	}
	return linkADRReqParameters{
		Masks:         desiredMasks,
		DataRateIndex: drIdx,
		TxPowerIndex:  txPowerIdx,
		NbTrans:       nbTrans,
	}, true, nil
}

func DeviceNeedsLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band) bool {
	_, required, err := generateLinkADRReq(ctx, dev, phy)
	return err == nil && required
}

func EnqueueLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy *band.Band) (EnqueueState, error) {
	params, required, err := generateLinkADRReq(ctx, dev, phy)
	if err != nil {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
		}, err
	}
	if !required {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}, nil
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_LINK_ADR, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if int(nDown) < len(params.Masks) {
			return nil, 0, nil, false
		}

		uplinksNeeded := uint16(1)
		if dev.MacState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
			uplinksNeeded = uint16(len(params.Masks))
		}
		if nUp < uplinksNeeded {
			return nil, 0, nil, false
		}
		evs := make(events.Builders, 0, len(params.Masks))
		cmds := make([]*ttnpb.MACCommand, 0, len(params.Masks))
		for i, m := range params.Masks {
			req := &ttnpb.MACCommand_LinkADRReq{
				DataRateIndex:      params.DataRateIndex,
				TxPowerIndex:       params.TxPowerIndex,
				NbTrans:            params.NbTrans,
				ChannelMaskControl: uint32(m.Cntl),
				ChannelMask:        params.Masks[i].Mask[:],
			}
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, EvtEnqueueLinkADRRequest.With(events.WithData(req)))
			log.FromContext(ctx).WithFields(log.Fields(
				"data_rate_index", req.DataRateIndex,
				"nb_trans", req.NbTrans,
				"tx_power_index", req.TxPowerIndex,
				"channel_mask_control", req.ChannelMaskControl,
				"channel_mask", req.ChannelMask,
			)).Debug("Enqueued LinkADRReq")
		}
		return cmds, uplinksNeeded, evs, true
	}, dev.MacState.PendingRequests...)
	return st, nil
}

func HandleLinkADRAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_LinkADRAns, dupCount uint, fCntUp uint32, fps *frequencyplans.Store) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	if (dev.MacState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_0_2) < 0 || dev.MacState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_1) >= 0) && dupCount != 0 {
		return nil, internal.ErrInvalidPayload.New()
	}

	ev := EvtReceiveLinkADRAccept
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		ev = EvtReceiveLinkADRReject

		// See "Table 6: LinkADRAns status bits signification" of LoRaWAN 1.1 specification
		if !pld.ChannelMaskAck {
			log.FromContext(ctx).Warn("Either Network Server sent a channel mask, which enables a yet undefined channel or requires all channels to be disabled, or device is malfunctioning.")
		}
	}
	evs := events.Builders{ev.With(events.WithData(pld))}

	phy, err := internal.DeviceBand(dev, fps)
	if err != nil {
		return evs, err
	}

	handler := handleMACResponseBlock
	if dev.MacState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_0_2) < 0 {
		handler = handleMACResponse
	}
	var n uint
	var req *ttnpb.MACCommand_LinkADRReq
	dev.MacState.PendingRequests, err = handler(ttnpb.MACCommandIdentifier_CID_LINK_ADR, func(cmd *ttnpb.MACCommand) error {
		if dev.MacState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_0_2) >= 0 && dev.MacState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 && n > dupCount+1 {
			return internal.ErrInvalidPayload.New()
		}
		n++

		req = cmd.GetLinkAdrReq()
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
			if int(i) >= len(dev.MacState.CurrentParameters.Channels) || dev.MacState.CurrentParameters.Channels[i] == nil {
				if !masked {
					continue
				}
				return internal.ErrCorruptedMACState.
					WithAttributes(
						"i", i,
						"channels_len", len(dev.MacState.CurrentParameters.Channels),
					).
					WithCause(internal.ErrUnknownChannel)
			}
			dev.MacState.CurrentParameters.Channels[i].EnableUplink = masked
		}
		return nil
	}, dev.MacState.PendingRequests...)
	if err != nil || req == nil {
		return evs, err
	}

	if !pld.DataRateIndexAck {
		if i := searchDataRateIndex(req.DataRateIndex, dev.MacState.RejectedAdrDataRateIndexes...); i == len(dev.MacState.RejectedAdrDataRateIndexes) || dev.MacState.RejectedAdrDataRateIndexes[i] != req.DataRateIndex {
			dev.MacState.RejectedAdrDataRateIndexes = append(dev.MacState.RejectedAdrDataRateIndexes, ttnpb.DataRateIndex_DATA_RATE_0)
			copy(dev.MacState.RejectedAdrDataRateIndexes[i+1:], dev.MacState.RejectedAdrDataRateIndexes[i:])
			dev.MacState.RejectedAdrDataRateIndexes[i] = req.DataRateIndex
		}
	}
	if !pld.TxPowerIndexAck {
		if i := searchUint32(req.TxPowerIndex, dev.MacState.RejectedAdrTxPowerIndexes...); i == len(dev.MacState.RejectedAdrTxPowerIndexes) || dev.MacState.RejectedAdrTxPowerIndexes[i] != req.TxPowerIndex {
			dev.MacState.RejectedAdrTxPowerIndexes = append(dev.MacState.RejectedAdrTxPowerIndexes, 0)
			copy(dev.MacState.RejectedAdrTxPowerIndexes[i+1:], dev.MacState.RejectedAdrTxPowerIndexes[i:])
			dev.MacState.RejectedAdrTxPowerIndexes[i] = req.TxPowerIndex
		}
	}
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		return evs, nil
	}
	if !dev.MacState.LorawanVersion.HasNoChangeDataRateIndex() || req.DataRateIndex != noChangeDataRateIndex {
		dev.MacState.CurrentParameters.AdrDataRateIndex = req.DataRateIndex
		dev.MacState.LastAdrChangeFCntUp = fCntUp
	}
	if !dev.MacState.LorawanVersion.HasNoChangeTXPowerIndex() || req.TxPowerIndex != noChangeTXPowerIndex {
		dev.MacState.CurrentParameters.AdrTxPowerIndex = req.TxPowerIndex
		dev.MacState.LastAdrChangeFCntUp = fCntUp
	}
	if req.NbTrans > 0 && dev.MacState.CurrentParameters.AdrNbTrans != req.NbTrans {
		dev.MacState.CurrentParameters.AdrNbTrans = req.NbTrans
		dev.MacState.LastAdrChangeFCntUp = fCntUp
	}
	return evs, nil
}
