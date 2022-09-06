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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
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

	EvtGenerateLinkADRFail = events.Define(
		"ns.mac.link_adr.request.fail",
		"link ADR request generation failure",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
		events.WithErrorDataType(),
	)
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

// generateUplinkChannelMask generates the enabled uplink channel mask from the provided parameters.
func generateUplinkChannelMask(name string, channels []*ttnpb.MACParameters_Channel, phy *band.Band) ([]bool, error) {
	mask := make([]bool, phy.MaxUplinkChannels)
	for i, ch := range channels {
		enabled := ch.GetEnableUplink()
		if enabled && ch.UplinkFrequency == 0 {
			return nil, internal.ErrNoUplinkFrequency.WithAttributes(
				"parameters", name,
				"i", i,
			)
		}
		mask[i] = enabled
	}
	return mask, nil
}

// generateLinkADRReq attempts to generate a `LinkADRReq` command payload in order to reconcile the drift
// between the current and desired MAC state parameters. The parameters affected are the data rate index,
// transmission power index, number of transmissions and the channels enabled for uplink usage.
//
// The desired parameters are validated against the provided band.
//
// The generated command will attempt to patch the channel mask even if the desired data rate index or
// transmission power index have been rejected, using the no-change data rate indices if the session
// LoRaWAN version supports them.
// If the generated command is expected to be invalid (because the data rate index or transmission power
// index has been rejected before), the generation will be skipped.
func generateLinkADRReq( //nolint:gocyclo
	_ context.Context, dev *ttnpb.EndDevice, phy *band.Band,
) (linkADRReqParameters, bool, error) {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return linkADRReqParameters{}, false, nil
	}

	macState := dev.MacState
	desiredParameters, currentParameters := macState.DesiredParameters, macState.CurrentParameters
	if n := len(currentParameters.Channels); n > int(phy.MaxUplinkChannels) {
		return linkADRReqParameters{}, false, internal.ErrTooManyChannels.WithAttributes(
			"parameters", "current",
			"channels_len", n,
			"phy_max_uplink_channels", phy.MaxUplinkChannels,
		)
	}
	if n := len(desiredParameters.Channels); n > int(phy.MaxUplinkChannels) {
		return linkADRReqParameters{}, false, internal.ErrTooManyChannels.WithAttributes(
			"parameters", "desired",
			"channels_len", n,
			"phy_max_uplink_channels", phy.MaxUplinkChannels,
		)
	}

	currentChs, err := generateUplinkChannelMask("current", currentParameters.Channels, phy)
	if err != nil {
		return linkADRReqParameters{}, false, err
	}
	for i, ch := range desiredParameters.Channels {
		// The channel may not be enabled yet, but an enqueued `NewChannelReq` within
		// the same uplink may enable it.
		if DeviceNeedsNewChannelReqAtIndex(dev, i) {
			currentChs[i] = ch.GetUplinkFrequency() != 0
		}
	}
	desiredChs, err := generateUplinkChannelMask("desired", desiredParameters.Channels, phy)
	if err != nil {
		return linkADRReqParameters{}, false, err
	}

	switch {
	case !band.EqualChMasks(currentChs, desiredChs):
		// NOTE: LinkADRReq is scheduled regardless of ADR settings if channel mask is required, which often
		// is the case with ABP devices or when ChMask CFList is not supported/used.
	case desiredParameters.AdrNbTrans != currentParameters.AdrNbTrans,
		desiredParameters.AdrDataRateIndex != currentParameters.AdrDataRateIndex,
		desiredParameters.AdrTxPowerIndex != currentParameters.AdrTxPowerIndex:
	default:
		return linkADRReqParameters{}, false, nil
	}
	desiredMasks, err := phy.GenerateChMasks(currentChs, desiredChs)
	if err != nil {
		return linkADRReqParameters{}, false, err
	}

	var (
		drIdx      ttnpb.DataRateIndex
		txPowerIdx uint32
		nbTrans    uint32
	)
	_, _, allowedDataRates, ok := channelDataRateRange(desiredParameters.Channels...)
	if !ok {
		return linkADRReqParameters{}, false, internal.ErrChannelDataRateRange.New()
	}

	if desiredParameters.AdrTxPowerIndex != currentParameters.AdrTxPowerIndex {
		if desiredParameters.AdrTxPowerIndex > uint32(phy.MaxTxPowerIndex()) {
			return linkADRReqParameters{}, false, internal.ErrTxPowerIndexTooHigh.WithAttributes(
				"desired_adr_tx_power_index", desiredParameters.AdrTxPowerIndex,
				"phy_max_tx_power_index", phy.MaxTxPowerIndex(),
			)
		}
	}
	if desiredParameters.AdrDataRateIndex != currentParameters.AdrDataRateIndex {
		if _, ok := allowedDataRates[desiredParameters.AdrDataRateIndex]; !ok {
			return linkADRReqParameters{}, false, internal.ErrInvalidDataRateIndex.WithAttributes(
				"desired_adr_data_rate_index", desiredParameters.AdrDataRateIndex,
			)
		}
	}

	drIdx = desiredParameters.AdrDataRateIndex
	txPowerIdx = desiredParameters.AdrTxPowerIndex
	nbTrans = desiredParameters.AdrNbTrans
	switch {
	case !deviceRejectedADRDataRateIndex(dev, drIdx) && !deviceRejectedADRTXPowerIndex(dev, txPowerIdx):
		// Only send the desired DataRateIndex and TXPowerIndex if neither of them were rejected.

	case len(desiredMasks) == 0 && desiredParameters.AdrNbTrans == currentParameters.AdrNbTrans:
		// Either desired data rate index or TX power output index have been rejected and there are
		// no channel mask and NbTrans changes desired.
		return linkADRReqParameters{}, false, internal.ErrRejectedParameters.WithAttributes(
			"parameters", "desired",
			"data_rate_index", drIdx,
			"tx_power_index", txPowerIdx,
		)

	default:
		// The desired data rate index and/or transmission power index have been rejected, and we have to
		// either change the channel mask or the number of transmissions. We will attempt to maintain the
		// current data rate index and transmission power index while changing the enabled channels mask
		// and the number of transmissions.

		drIdx = currentParameters.AdrDataRateIndex
		txPowerIdx = currentParameters.AdrTxPowerIndex

		// It can be the case that the new channel mask is no longer compatible with the old data rate index.
		// We will not attempt to generate a `LinkADRReq` in such cases.
		if _, ok := allowedDataRates[drIdx]; !ok {
			return linkADRReqParameters{}, false, internal.ErrIncompatibleChannelMask.WithAttributes(
				"data_rate_index", drIdx,
			)
		}

		dataRateIndexRejected := deviceRejectedADRDataRateIndex(dev, drIdx)
		txPowerIndexRejected := deviceRejectedADRTXPowerIndex(dev, txPowerIdx)
		if macspec.HasNoChangeADRIndices(macState.LorawanVersion) {
			dataRateIndexRejected = dataRateIndexRejected &&
				deviceRejectedADRDataRateIndex(dev, noChangeDataRateIndex)
			txPowerIndexRejected = txPowerIndexRejected &&
				deviceRejectedADRTXPowerIndex(dev, noChangeTXPowerIndex)
		}
		if !dataRateIndexRejected && !txPowerIndexRejected {
			break
		}

		// At this point we do cannot reuse the current data rate index and transmission power index in order
		// to change the channel mask or number of transmissions.
		return linkADRReqParameters{}, false, internal.ErrRejectedParameters.WithAttributes(
			"parameters", "current",
			"data_rate_index", drIdx,
			"tx_power_index", txPowerIdx,
		)
	}
	if macspec.HasNoChangeADRIndices(macState.LorawanVersion) {
		if drIdx == currentParameters.AdrDataRateIndex &&
			!deviceRejectedADRDataRateIndex(dev, noChangeDataRateIndex) {
			drIdx = noChangeDataRateIndex
		}
		if txPowerIdx == currentParameters.AdrTxPowerIndex &&
			!deviceRejectedADRTXPowerIndex(dev, noChangeTXPowerIndex) {
			txPowerIdx = noChangeTXPowerIndex
		}
	}

	return linkADRReqParameters{
		Masks:         desiredMasks,
		DataRateIndex: drIdx,
		TxPowerIndex:  txPowerIdx,
		NbTrans:       nbTrans,
	}, true, nil
}

// DeviceNeedsLinkADRReq returns if the end device needs a `LinkADRReq` command to be enqueued.
func DeviceNeedsLinkADRReq(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band) bool {
	_, required, err := generateLinkADRReq(ctx, dev, phy)
	return err == nil && required
}

// EnqueueLinkADRReq enqueues a `LinkADRReq` request if needed.
func EnqueueLinkADRReq(
	ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, phy *band.Band,
) EnqueueState {
	params, required, err := generateLinkADRReq(ctx, dev, phy)
	if err != nil {
		return EnqueueState{
			MaxDownLen:   maxDownLen,
			MaxUpLen:     maxUpLen,
			QueuedEvents: events.Builders{EvtGenerateLinkADRFail.BindData(err)},
			Ok:           true,
		}
	}
	if !required {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}
	macState := dev.MacState
	f := func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if int(nDown) < len(params.Masks) {
			return nil, 0, nil, false
		}

		uplinksNeeded := uint16(len(params.Masks))
		if macspec.SingularLinkADRAns(macState.LorawanVersion) {
			uplinksNeeded = 1
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
	}
	var st EnqueueState
	macState.PendingRequests, st = enqueueMACCommand(
		ttnpb.MACCommandIdentifier_CID_LINK_ADR, maxDownLen, maxUpLen, f, macState.PendingRequests...,
	)
	return st
}

// HandleLinkADRAns applies the update of the associated `LinkADRReq` request to the end device, if applicable.
func HandleLinkADRAns( //nolint:gocyclo
	_ context.Context,
	dev *ttnpb.EndDevice,
	pld *ttnpb.MACCommand_LinkADRAns,
	dupCount uint,
	fCntUp uint32,
	fps *frequencyplans.Store,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	macState := dev.MacState
	allowDuplicateLinkADRAns := macspec.AllowDuplicateLinkADRAns(macState.LorawanVersion)
	if !allowDuplicateLinkADRAns && dupCount != 0 {
		return nil, internal.ErrInvalidPayload.New()
	}

	ev := EvtReceiveLinkADRAccept
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		ev = EvtReceiveLinkADRReject
	}
	evs := events.Builders{ev.With(events.WithData(pld))}

	phy, err := internal.DeviceBand(dev, fps)
	if err != nil {
		return evs, err
	}

	currentParameters := macState.CurrentParameters
	handler := handleMACResponseBlock
	if !allowDuplicateLinkADRAns && !macspec.SingularLinkADRAns(macState.LorawanVersion) {
		handler = handleMACResponse
	}
	var n uint
	var req *ttnpb.MACCommand_LinkADRReq
	macState.PendingRequests, err = handler(
		ttnpb.MACCommandIdentifier_CID_LINK_ADR,
		false,
		func(cmd *ttnpb.MACCommand) error {
			if allowDuplicateLinkADRAns && n > dupCount+1 {
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
			copy(mask[:], req.ChannelMask)
			m, err := phy.ParseChMask(mask, uint8(req.ChannelMaskControl))
			if err != nil {
				return err
			}
			channels := currentParameters.Channels
			for i, masked := range m {
				if int(i) >= len(channels) || channels[i] == nil {
					if !masked {
						continue
					}
					return internal.ErrCorruptedMACState.
						WithAttributes(
							"i", i,
							"channels_len", len(channels),
						).
						WithCause(internal.ErrUnknownChannel)
				}
				channels[i].EnableUplink = masked
			}
			return nil
		},
		macState.PendingRequests...,
	)
	if err != nil || req == nil {
		return evs, err
	}

	if !pld.DataRateIndexAck {
		i := searchDataRateIndex(req.DataRateIndex, macState.RejectedAdrDataRateIndexes...)
		if i == len(macState.RejectedAdrDataRateIndexes) ||
			macState.RejectedAdrDataRateIndexes[i] != req.DataRateIndex {
			macState.RejectedAdrDataRateIndexes = append(
				macState.RejectedAdrDataRateIndexes, ttnpb.DataRateIndex_DATA_RATE_0,
			)
			copy(macState.RejectedAdrDataRateIndexes[i+1:], macState.RejectedAdrDataRateIndexes[i:])
			macState.RejectedAdrDataRateIndexes[i] = req.DataRateIndex
		}
	}
	if !pld.TxPowerIndexAck {
		i := searchUint32(req.TxPowerIndex, macState.RejectedAdrTxPowerIndexes...)
		if i == len(macState.RejectedAdrTxPowerIndexes) ||
			macState.RejectedAdrTxPowerIndexes[i] != req.TxPowerIndex {
			macState.RejectedAdrTxPowerIndexes = append(macState.RejectedAdrTxPowerIndexes, 0)
			copy(macState.RejectedAdrTxPowerIndexes[i+1:], macState.RejectedAdrTxPowerIndexes[i:])
			macState.RejectedAdrTxPowerIndexes[i] = req.TxPowerIndex
		}
	}
	if !pld.ChannelMaskAck || !pld.DataRateIndexAck || !pld.TxPowerIndexAck {
		return evs, nil
	}
	if !macspec.HasNoChangeADRIndices(macState.LorawanVersion) || req.DataRateIndex != noChangeDataRateIndex {
		currentParameters.AdrDataRateIndex = req.DataRateIndex
		macState.LastAdrChangeFCntUp = fCntUp
	}
	if !macspec.HasNoChangeADRIndices(macState.LorawanVersion) || req.TxPowerIndex != noChangeTXPowerIndex {
		currentParameters.AdrTxPowerIndex = req.TxPowerIndex
		macState.LastAdrChangeFCntUp = fCntUp
	}
	if req.NbTrans > 0 && currentParameters.AdrNbTrans != req.NbTrans {
		currentParameters.AdrNbTrans = req.NbTrans
		macState.LastAdrChangeFCntUp = fCntUp
	}
	return evs, nil
}
