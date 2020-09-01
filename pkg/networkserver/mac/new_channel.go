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

func DeviceNeedsNewChannelReqAtIndex(dev *ttnpb.EndDevice, i int) bool {
	if i >= len(dev.MACState.CurrentParameters.Channels) {
		// A channel is desired to be created.
		return i < len(dev.MACState.DesiredParameters.Channels)
	}
	if i >= len(dev.MACState.DesiredParameters.Channels) {
		// A channel is desired to be deleted.
		return !deviceRejectedFrequency(dev, 0)
	}
	// NOTE: i < len(dev.MACState.CurrentParameters.Channels) && i < len(dev.MACState.DesiredParameters.Channels) at this point
	desiredCh, currentCh := dev.MACState.DesiredParameters.Channels[i], dev.MACState.CurrentParameters.Channels[i]
	if desiredCh == nil || currentCh == nil {
		return desiredCh != currentCh
	}
	if desiredCh.UplinkFrequency != currentCh.UplinkFrequency || desiredCh.MaxDataRateIndex != currentCh.MaxDataRateIndex || desiredCh.MinDataRateIndex != currentCh.MinDataRateIndex {
		return !deviceRejectedFrequency(dev, desiredCh.UplinkFrequency)
	}
	return false
}

func DeviceNeedsNewChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMACState() == nil {
		return false
	}
	if len(dev.MACState.DesiredParameters.Channels) != len(dev.MACState.CurrentParameters.Channels) {
		return true
	}
	for i := range dev.MACState.DesiredParameters.Channels {
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
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_NEW_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
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

			if i >= len(dev.MACState.CurrentParameters.Channels) || DeviceNeedsNewChannelReqAtIndex(dev, i) {
				desiredCh := dev.MACState.DesiredParameters.Channels[i]
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
	}, dev.MACState.PendingRequests...)
	return st
}

func HandleNewChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_NewChannelAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_NEW_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetNewChannelReq()
		// TODO: Handle DataRate nack. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2192)
		if !pld.FrequencyAck {
			if i := searchUint64(req.Frequency, dev.MACState.RejectedFrequencies...); i == len(dev.MACState.RejectedFrequencies) || dev.MACState.RejectedFrequencies[i] != req.Frequency {
				dev.MACState.RejectedFrequencies = append(dev.MACState.RejectedFrequencies, 0)
				copy(dev.MACState.RejectedFrequencies[i+1:], dev.MACState.RejectedFrequencies[i:])
				dev.MACState.RejectedFrequencies[i] = req.Frequency
			}
		}
		if !pld.DataRateAck || !pld.FrequencyAck {
			return nil
		}

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
	Evt := EvtReceiveNewChannelAccept
	if !pld.DataRateAck || !pld.FrequencyAck {
		Evt = EvtReceiveNewChannelReject
	}
	return events.Builders{
		Evt.With(events.WithData(pld)),
	}, err
}
