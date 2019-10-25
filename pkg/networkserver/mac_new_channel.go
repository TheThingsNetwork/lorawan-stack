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

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtEnqueueNewChannelRequest = defineEnqueueMACRequestEvent("new_channel", "new channel")()
	evtReceiveNewChannelAccept  = defineReceiveMACAcceptEvent("new_channel", "new channel")()
	evtReceiveNewChannelReject  = defineReceiveMACRejectEvent("new_channel", "new channel")()
)

func channelNeedsNewChannelReq(desiredCh, currentCh *ttnpb.MACParameters_Channel) bool {
	return (desiredCh != nil || currentCh != nil) &&
		(desiredCh == nil ||
			currentCh == nil ||
			desiredCh.UplinkFrequency != currentCh.UplinkFrequency ||
			desiredCh.MaxDataRateIndex != currentCh.MaxDataRateIndex ||
			desiredCh.MinDataRateIndex != currentCh.MinDataRateIndex)
}

func deviceNeedsNewChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.MACState == nil {
		return false
	}
	if len(dev.MACState.DesiredParameters.Channels) != len(dev.MACState.CurrentParameters.Channels) {
		return true
	}
	for i := range dev.MACState.DesiredParameters.Channels {
		if channelNeedsNewChannelReq(dev.MACState.DesiredParameters.Channels[i], dev.MACState.CurrentParameters.Channels[i]) {
			return true
		}
	}
	return false
}

func enqueueNewChannelReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) macCommandEnqueueState {
	if !deviceNeedsNewChannelReq(dev) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_NEW_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		var reqs []*ttnpb.MACCommand_NewChannelReq
		for i := 0; i < len(dev.MACState.DesiredParameters.Channels) || i < len(dev.MACState.CurrentParameters.Channels); i++ {
			if i >= len(dev.MACState.DesiredParameters.Channels) {
				for j := range dev.MACState.CurrentParameters.Channels[i:] {
					reqs = append(reqs, &ttnpb.MACCommand_NewChannelReq{
						ChannelIndex: uint32(i + j),
					})
				}
				break
			}

			desiredCh := dev.MACState.DesiredParameters.Channels[i]
			if i >= len(dev.MACState.CurrentParameters.Channels) ||
				channelNeedsNewChannelReq(desiredCh, dev.MACState.CurrentParameters.Channels[i]) {
				reqs = append(reqs, &ttnpb.MACCommand_NewChannelReq{
					ChannelIndex:     uint32(i),
					Frequency:        desiredCh.GetUplinkFrequency(),
					MinDataRateIndex: desiredCh.GetMinDataRateIndex(),
					MaxDataRateIndex: desiredCh.GetMaxDataRateIndex(),
				})
			}
		}

		var cmds []*ttnpb.MACCommand
		var evs []events.DefinitionDataClosure
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
			evs = append(evs, evtEnqueueNewChannelRequest.BindData(req))
		}
		return cmds, uint16(len(cmds)), evs, true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleNewChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_NewChannelAns) ([]events.DefinitionDataClosure, error) {
	if pld == nil {
		return nil, errNoPayload
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_NEW_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		if !pld.DataRateAck || !pld.FrequencyAck {
			return nil
		}
		req := cmd.GetNewChannelReq()

		if uint(req.ChannelIndex) >= uint(len(dev.MACState.CurrentParameters.Channels)) {
			dev.MACState.CurrentParameters.Channels = append(dev.MACState.CurrentParameters.Channels, make([]*ttnpb.MACParameters_Channel, 1+int(req.ChannelIndex-uint32(len(dev.MACState.CurrentParameters.Channels))))...)
		}

		ch := dev.MACState.CurrentParameters.Channels[req.ChannelIndex]
		if ch == nil {
			ch = &ttnpb.MACParameters_Channel{
				DownlinkFrequency: req.Frequency,
			}
			dev.MACState.CurrentParameters.Channels[req.ChannelIndex] = ch
		}
		ch.UplinkFrequency = req.Frequency
		ch.MinDataRateIndex = req.MinDataRateIndex
		ch.MaxDataRateIndex = req.MaxDataRateIndex
		ch.EnableUplink = true
		return nil
	}, dev.MACState.PendingRequests...)
	evt := evtReceiveNewChannelAccept
	if !pld.DataRateAck || !pld.FrequencyAck {
		evt = evtReceiveNewChannelReject
	}
	return []events.DefinitionDataClosure{
		evt.BindData(pld),
	}, err
}
