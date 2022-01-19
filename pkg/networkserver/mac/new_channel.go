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

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueNewChannelRequest = defineEnqueueMACRequestEvent(
		"new_channel", "new channel",
		events.WithDataType(&ttnpb.MACCommand_NewChannelReq{}),
	)()
	EvtReceiveNewChannelAccept = defineReceiveMACAcceptEvent(
		"new_channel", "new channel",
		events.WithDataType(&ttnpb.MACCommand_NewChannelAns{}),
	)()
	EvtReceiveNewChannelReject = defineReceiveMACRejectEvent(
		"new_channel", "new channel",
		events.WithDataType(&ttnpb.MACCommand_NewChannelAns{}),
	)()
)

func deviceRejectedNewChannelReq(dev *ttnpb.EndDevice, freq uint64, minDRIdx, maxDRIdx ttnpb.DataRateIndex) bool {
	return deviceRejectedFrequency(dev, freq) || deviceRejectedDataRateRange(dev, freq, minDRIdx, maxDRIdx)
}

func DeviceNeedsNewChannelReqAtIndex(dev *ttnpb.EndDevice, i int) bool {
	switch {
	case i >= len(dev.MacState.CurrentParameters.Channels) && i >= len(dev.MacState.DesiredParameters.Channels):
	case i >= len(dev.MacState.DesiredParameters.Channels),
		dev.MacState.DesiredParameters.Channels[i] == nil:
		// A channel is desired to be deleted.
		if currentCh := dev.MacState.CurrentParameters.Channels[i]; currentCh != nil && currentCh.UplinkFrequency > 0 {
			return !deviceRejectedNewChannelReq(dev, 0, ttnpb.DataRateIndex_DATA_RATE_0, ttnpb.DataRateIndex_DATA_RATE_0)
		}
	case i >= len(dev.MacState.CurrentParameters.Channels),
		dev.MacState.CurrentParameters.Channels[i] == nil:
		// A channel is desired to be created.
		if desiredCh := dev.MacState.DesiredParameters.Channels[i]; desiredCh != nil && desiredCh.UplinkFrequency > 0 {
			return !deviceRejectedNewChannelReq(dev, desiredCh.UplinkFrequency, desiredCh.MinDataRateIndex, desiredCh.MaxDataRateIndex)
		}
	default:
		desiredCh, currentCh := dev.MacState.DesiredParameters.Channels[i], dev.MacState.CurrentParameters.Channels[i]
		if desiredCh.UplinkFrequency != currentCh.UplinkFrequency || desiredCh.MaxDataRateIndex != currentCh.MaxDataRateIndex || desiredCh.MinDataRateIndex != currentCh.MinDataRateIndex {
			return !deviceRejectedNewChannelReq(dev, desiredCh.UplinkFrequency, desiredCh.MinDataRateIndex, desiredCh.MaxDataRateIndex)
		}
	}
	return false
}

func DeviceNeedsNewChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	if len(dev.MacState.DesiredParameters.Channels) != len(dev.MacState.CurrentParameters.Channels) {
		return true
	}
	for i := range dev.MacState.DesiredParameters.Channels {
		if DeviceNeedsNewChannelReqAtIndex(dev, i) {
			return true
		}
	}
	return false
}

func EnqueueNewChannelReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsNewChannelReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_NEW_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		var reqs []*ttnpb.MACCommand_NewChannelReq
		for i := 0; i < len(dev.MacState.DesiredParameters.Channels) || i < len(dev.MacState.CurrentParameters.Channels); i++ {
			switch {
			case !DeviceNeedsNewChannelReqAtIndex(dev, i):
			case i >= len(dev.MacState.DesiredParameters.Channels) || dev.MacState.DesiredParameters.Channels[i] == nil:
				reqs = append(reqs, &ttnpb.MACCommand_NewChannelReq{
					ChannelIndex: uint32(i),
				})
			default:
				desiredCh := dev.MacState.DesiredParameters.Channels[i]
				reqs = append(reqs, &ttnpb.MACCommand_NewChannelReq{
					ChannelIndex:     uint32(i),
					Frequency:        desiredCh.GetUplinkFrequency(),
					MinDataRateIndex: desiredCh.GetMinDataRateIndex(),
					MaxDataRateIndex: desiredCh.GetMaxDataRateIndex(),
				})
			}
		}

		var cmds []*ttnpb.MACCommand
		var evs events.Builders
		for _, req := range reqs {
			if nDown < 1 || nUp < 1 {
				return cmds, uint16(len(cmds)), evs, false
			}
			nDown--
			nUp--
			log.FromContext(ctx).WithFields(log.Fields(
				"channel_index", req.ChannelIndex,
				"frequency", req.Frequency,
				"max_data_rate_index", req.MaxDataRateIndex,
				"min_data_rate_index", req.MinDataRateIndex,
			)).Debug("Enqueued NewChannelReq")
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, EvtEnqueueNewChannelRequest.With(events.WithData(req)))
		}
		return cmds, uint16(len(cmds)), evs, true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandleNewChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_NewChannelAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_NEW_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetNewChannelReq()
		if !pld.DataRateAck {
			if dev.MacState.RejectedDataRateRanges == nil {
				dev.MacState.RejectedDataRateRanges = make(map[uint64]*ttnpb.MACState_DataRateRanges, 1)
			}
			r, ok := dev.MacState.RejectedDataRateRanges[req.Frequency]
			if !ok {
				r = &ttnpb.MACState_DataRateRanges{
					Ranges: make([]*ttnpb.MACState_DataRateRange, 0, 1),
				}
				dev.MacState.RejectedDataRateRanges[req.Frequency] = r
			}
			r.Ranges = append(r.Ranges, &ttnpb.MACState_DataRateRange{
				MinDataRateIndex: req.MinDataRateIndex,
				MaxDataRateIndex: req.MaxDataRateIndex,
			})
		}
		if !pld.FrequencyAck {
			if i := searchUint64(req.Frequency, dev.MacState.RejectedFrequencies...); i == len(dev.MacState.RejectedFrequencies) || dev.MacState.RejectedFrequencies[i] != req.Frequency {
				dev.MacState.RejectedFrequencies = append(dev.MacState.RejectedFrequencies, 0)
				copy(dev.MacState.RejectedFrequencies[i+1:], dev.MacState.RejectedFrequencies[i:])
				dev.MacState.RejectedFrequencies[i] = req.Frequency
			}
		}
		if !pld.DataRateAck || !pld.FrequencyAck {
			return nil
		}

		if uint(req.ChannelIndex) >= uint(len(dev.MacState.CurrentParameters.Channels)) {
			dev.MacState.CurrentParameters.Channels = append(dev.MacState.CurrentParameters.Channels, make([]*ttnpb.MACParameters_Channel, 1+int(req.ChannelIndex-uint32(len(dev.MacState.CurrentParameters.Channels))))...)
		}
		ch := dev.MacState.CurrentParameters.Channels[req.ChannelIndex]
		if ch == nil {
			ch = &ttnpb.MACParameters_Channel{
				DownlinkFrequency: req.Frequency,
			}
			dev.MacState.CurrentParameters.Channels[req.ChannelIndex] = ch
		}
		ch.UplinkFrequency = req.Frequency
		ch.MinDataRateIndex = req.MinDataRateIndex
		ch.MaxDataRateIndex = req.MaxDataRateIndex
		ch.EnableUplink = req.Frequency > 0
		return nil
	}, dev.MacState.PendingRequests...)
	ev := EvtReceiveNewChannelAccept
	if !pld.DataRateAck || !pld.FrequencyAck {
		ev = EvtReceiveNewChannelReject
	}
	return events.Builders{
		ev.With(events.WithData(pld)),
	}, err
}
