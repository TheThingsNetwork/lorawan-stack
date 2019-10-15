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
	evtEnqueueDLChannelRequest = defineEnqueueMACRequestEvent("dl_channel", "downlink Rx1 channel frequency modification")()
	evtReceiveDLChannelAccept  = defineReceiveMACAcceptEvent("dl_channel", "downlink Rx1 channel frequency modification")()
	evtReceiveDLChannelReject  = defineReceiveMACRejectEvent("dl_channel", "downlink Rx1 channel frequency modification")()
)

func needsDLChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.MACState == nil || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		return false
	}
	for i := 0; i < len(dev.MACState.DesiredParameters.Channels) && i < len(dev.MACState.CurrentParameters.Channels); i++ {
		if dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency != dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency {
			return true
		}
	}
	return false
}

func enqueueDLChannelReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) macCommandEnqueueState {
	if !needsDLChannelReq(dev) {
		return macCommandEnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st macCommandEnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_DL_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, []events.DefinitionDataClosure, bool) {
		var cmds []*ttnpb.MACCommand
		var evs []events.DefinitionDataClosure
		for i := 0; i < len(dev.MACState.DesiredParameters.Channels) && i < len(dev.MACState.CurrentParameters.Channels); i++ {
			if dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency == dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency {
				continue
			}
			if nDown < 1 || nUp < 1 {
				return cmds, uint16(len(cmds)), nil, false
			}
			nDown--
			nUp--

			req := &ttnpb.MACCommand_DLChannelReq{
				ChannelIndex: uint32(i),
				Frequency:    dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency,
			}
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, evtEnqueueDLChannelRequest.BindData(req))
			log.FromContext(ctx).WithFields(log.Fields(
				"channel_index", req.ChannelIndex,
				"frequency", req.Frequency,
			)).Debug("Enqueued DLChannelReq")
		}
		return cmds, uint16(len(cmds)), evs, true
	}, dev.MACState.PendingRequests...)
	return st
}

func handleDLChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DLChannelAns) ([]events.DefinitionDataClosure, error) {
	if pld == nil {
		return nil, errNoPayload
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DL_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		if !pld.ChannelIndexAck || !pld.FrequencyAck {
			return nil
		}
		req := cmd.GetDLChannelReq()

		if uint(req.ChannelIndex) >= uint(len(dev.MACState.CurrentParameters.Channels)) || dev.MACState.CurrentParameters.Channels[req.ChannelIndex] == nil {
			return errCorruptedMACState.WithCause(errUnknownChannel)
		}
		dev.MACState.CurrentParameters.Channels[req.ChannelIndex].DownlinkFrequency = req.Frequency
		return nil
	}, dev.MACState.PendingRequests...)
	evt := evtReceiveDLChannelAccept
	if !pld.ChannelIndexAck || !pld.FrequencyAck {
		evt = evtReceiveDLChannelReject
	}
	return []events.DefinitionDataClosure{
		evt.BindData(pld),
	}, err
}
