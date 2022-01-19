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
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
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
	if i >= len(dev.MacState.DesiredParameters.Channels) {
		return false
	}
	desiredCh := dev.MacState.DesiredParameters.Channels[i]
	if desiredCh == nil || desiredCh.UplinkFrequency == 0 || deviceRejectedFrequency(dev, desiredCh.DownlinkFrequency) {
		return false
	}
	if DeviceNeedsNewChannelReqAtIndex(dev, i) {
		return desiredCh.DownlinkFrequency != desiredCh.UplinkFrequency
	}
	// NOTE: NewChannelReq may be needed, but parameters could have been rejected before.
	if i >= len(dev.MacState.CurrentParameters.Channels) || dev.MacState.CurrentParameters.Channels[i] == nil {
		return false
	}
	return desiredCh.DownlinkFrequency != dev.MacState.CurrentParameters.Channels[i].DownlinkFrequency
}

func DeviceNeedsDLChannelReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() ||
		dev.GetMacState() == nil ||
		dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_0_2) < 0 {
		return false
	}
	for i := range dev.MacState.DesiredParameters.Channels {
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
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_DL_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		var cmds []*ttnpb.MACCommand
		var evs events.Builders
		for i := 0; i < len(dev.MacState.DesiredParameters.Channels) && i < len(dev.MacState.CurrentParameters.Channels); i++ {
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
				Frequency:    dev.MacState.DesiredParameters.Channels[i].DownlinkFrequency,
			}
			cmds = append(cmds, req.MACCommand())
			evs = append(evs, EvtEnqueueDLChannelRequest.With(events.WithData(req)))
			log.FromContext(ctx).WithFields(log.Fields(
				"channel_index", req.ChannelIndex,
				"frequency", req.Frequency,
			)).Debug("Enqueued DLChannelReq")
		}
		return cmds, uint16(len(cmds)), evs, true
	}, dev.MacState.PendingRequests...)
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
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_DL_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetDlChannelReq()
		if !pld.FrequencyAck {
			if i := searchUint64(req.Frequency, dev.MacState.RejectedFrequencies...); i == len(dev.MacState.RejectedFrequencies) || dev.MacState.RejectedFrequencies[i] != req.Frequency {
				dev.MacState.RejectedFrequencies = append(dev.MacState.RejectedFrequencies, 0)
				copy(dev.MacState.RejectedFrequencies[i+1:], dev.MacState.RejectedFrequencies[i:])
				dev.MacState.RejectedFrequencies[i] = req.Frequency
			}
		}
		if !pld.FrequencyAck || !pld.ChannelIndexAck {
			return nil
		}

		if uint(req.ChannelIndex) >= uint(len(dev.MacState.CurrentParameters.Channels)) {
			return internal.ErrCorruptedMACState.
				WithAttributes(
					"request_channel_id", req.ChannelIndex,
					"channels_len", len(dev.MacState.CurrentParameters.Channels),
				).
				WithCause(internal.ErrUnknownChannel)
		}
		if dev.MacState.CurrentParameters.Channels[req.ChannelIndex] == nil {
			return internal.ErrCorruptedMACState.
				WithAttributes(
					"request_channel_id", req.ChannelIndex,
				).
				WithCause(internal.ErrUnknownChannel)
		}
		dev.MacState.CurrentParameters.Channels[req.ChannelIndex].DownlinkFrequency = req.Frequency
		return nil
	}, dev.MacState.PendingRequests...)
	ev := EvtReceiveDLChannelAccept
	if !pld.ChannelIndexAck || !pld.FrequencyAck {
		ev = EvtReceiveDLChannelReject
	}
	return events.Builders{
		ev.With(events.WithData(pld)),
	}, err
}
