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

func deviceNeedsNewChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.MACState == nil {
		return false
	}
	for i, ch := range dev.MACState.DesiredParameters.Channels {
		if i >= len(dev.MACState.CurrentParameters.Channels) ||
			ch.UplinkFrequency != dev.MACState.CurrentParameters.Channels[i].UplinkFrequency ||
			ch.MinDataRateIndex != dev.MACState.CurrentParameters.Channels[i].MinDataRateIndex ||
			ch.MaxDataRateIndex != dev.MACState.CurrentParameters.Channels[i].MaxDataRateIndex {
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
		var cmds []*ttnpb.MACCommand
		var evs []events.DefinitionDataClosure
		for i, ch := range dev.MACState.DesiredParameters.Channels {
			if i < len(dev.MACState.CurrentParameters.Channels) &&
				ch.UplinkFrequency == dev.MACState.CurrentParameters.Channels[i].UplinkFrequency &&
				ch.MinDataRateIndex == dev.MACState.CurrentParameters.Channels[i].MinDataRateIndex &&
				ch.MaxDataRateIndex == dev.MACState.CurrentParameters.Channels[i].MaxDataRateIndex {
				continue
			}
			if nDown < 1 || nUp < 1 {
				return cmds, uint16(len(cmds)), nil, false
			}
			nDown--
			nUp--

			req := &ttnpb.MACCommand_NewChannelReq{
				ChannelIndex:     uint32(i),
				Frequency:        ch.UplinkFrequency,
				MinDataRateIndex: ch.MinDataRateIndex,
				MaxDataRateIndex: ch.MaxDataRateIndex,
			}
			log.FromContext(ctx).WithFields(log.Fields(
				"index", req.ChannelIndex,
				"frequency", req.Frequency,
				"min_data_rate_index", req.MinDataRateIndex,
				"max_data_rate_index", req.MaxDataRateIndex,
			)).Debug("Enqueued NewChannelReq")
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, evtEnqueueNewChannelRequest.BindData(req))
		}
		return cmds, uint16(len(cmds)), nil, true
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
