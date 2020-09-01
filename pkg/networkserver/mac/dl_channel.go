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
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	EvtEnqueueDLChannelRequest = defineEnqueueMACRequestEvent(
		"dl_channel", "downlink Rx1 channel frequency modification",
		events.WithDataType(&ttnpb.MACCommand_DLChannelReq{}),
	)()
	EvtReceiveDLChannelAccept = defineReceiveMACAcceptEvent(
		"dl_channel", "downlink Rx1 channel frequency modification",
		events.WithDataType(&ttnpb.MACCommand_DLChannelAns{}),
	)()
	EvtReceiveDLChannelReject = defineReceiveMACRejectEvent(
		"dl_channel", "downlink Rx1 channel frequency modification",
		events.WithDataType(&ttnpb.MACCommand_DLChannelAns{}),
	)()
)

func DeviceNeedsDLChannelReqAtIndex(dev *ttnpb.EndDevice, i int) bool {
	if i >= len(dev.MACState.CurrentParameters.Channels) || i >= len(dev.MACState.DesiredParameters.Channels) {
		return false
	}
	desiredCh, currentCh := dev.MACState.DesiredParameters.Channels[i], dev.MACState.CurrentParameters.Channels[i]
	if desiredCh == nil || currentCh == nil {
		return false
	}
	if DeviceNeedsNewChannelReqAtIndex(dev, i) {
		// Cannot define a downlink frequency for disabled channel.
		return desiredCh.UplinkFrequency != 0 && desiredCh.DownlinkFrequency != desiredCh.UplinkFrequency
	}
	return desiredCh.DownlinkFrequency != currentCh.DownlinkFrequency
}

func DeviceNeedsDLChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() ||
		dev.GetMACState() == nil ||
		dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		return false
	}
	for i := 0; i < len(dev.MACState.DesiredParameters.Channels) && i < len(dev.MACState.CurrentParameters.Channels); i++ {
		if DeviceNeedsDLChannelReqAtIndex(dev, i) {
			return true
		}
	}
	return false
}

func EnqueueDLChannelReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsDLChannelReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MACState.PendingRequests, st = enqueueMACCommand(ttnpb.CID_DL_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		var cmds []*ttnpb.MACCommand
		var evs events.Builders
		for i := 0; i < len(dev.MACState.DesiredParameters.Channels) && i < len(dev.MACState.CurrentParameters.Channels); i++ {
			if !DeviceNeedsDLChannelReqAtIndex(dev, i) {
				continue
			}
			if nDown < 1 || nUp < 1 {
				return cmds, uint16(len(cmds)), evs, false
			}
			nDown--
			nUp--

			req := &ttnpb.MACCommand_DLChannelReq{
				ChannelIndex: uint32(i),
				Frequency:    dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency,
			}
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, EvtEnqueueDLChannelRequest.With(events.WithData(req)))
			log.FromContext(ctx).WithFields(log.Fields(
				"channel_index", req.ChannelIndex,
				"frequency", req.Frequency,
			)).Debug("Enqueued DLChannelReq")
		}
		return cmds, uint16(len(cmds)), evs, true
	}, dev.MACState.PendingRequests...)
	return st
}

func HandleDLChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DLChannelAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	if !pld.ChannelIndexAck {
		// See "Table 10: DlChannelAns status bits signification" of LoRaWAN 1.1 specification
		log.FromContext(ctx).Warn("Network Server attempted to configure downlink frequency for a channel, for which uplink frequency is not defined or device is malfunctioning.")
	}

	var err error
	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DL_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetDLChannelReq()
		if !pld.FrequencyAck {
			if i := searchUint64(req.Frequency, dev.MACState.RejectedFrequencies...); i == len(dev.MACState.RejectedFrequencies) || dev.MACState.RejectedFrequencies[i] != req.Frequency {
				dev.MACState.RejectedFrequencies = append(dev.MACState.RejectedFrequencies, 0)
				copy(dev.MACState.RejectedFrequencies[i+1:], dev.MACState.RejectedFrequencies[i:])
				dev.MACState.RejectedFrequencies[i] = req.Frequency
			}
		}
		if !pld.FrequencyAck || !pld.ChannelIndexAck {
			return nil
		}

		if uint(req.ChannelIndex) >= uint(len(dev.MACState.CurrentParameters.Channels)) || dev.MACState.CurrentParameters.Channels[req.ChannelIndex] == nil {
			return ErrCorruptedMACState.WithCause(ErrUnknownChannel)
		}
		dev.MACState.CurrentParameters.Channels[req.ChannelIndex].DownlinkFrequency = req.Frequency
		return nil
	}, dev.MACState.PendingRequests...)
	Evt := EvtReceiveDLChannelAccept
	if !pld.ChannelIndexAck || !pld.FrequencyAck {
		Evt = EvtReceiveDLChannelReject
	}
	return events.Builders{
		Evt.With(events.WithData(pld)),
	}, err
}
