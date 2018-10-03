// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtMACDLChannelAccept = events.Define("ns.mac.dl_channel.accept", "device accepted downlink channel request")
	evtMACDLChannelReject = events.Define("ns.mac.dl_channel.reject", "device rejected downlink channel request")
)

func enqueueDLChannelReq(ctx context.Context, dev *ttnpb.EndDevice) {
	for i := 0; i < len(dev.MACState.DesiredParameters.Channels) && i < len(dev.MACState.CurrentParameters.Channels); i++ {
		if dev.MACState.DesiredParameters.Channels[i].UplinkFrequency == dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency &&
			dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency == dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency {
			continue
		}

		dev.MACState.PendingRequests = append(dev.MACState.PendingRequests, (&ttnpb.MACCommand_DLChannelReq{
			ChannelIndex: uint32(i),
			Frequency:    dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency,
		}).MACCommand())
	}
}

func handleDLChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_DLChannelAns) (err error) {
	if pld == nil {
		return errNoPayload
	}

	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_DL_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		if !pld.ChannelIndexAck && !pld.FrequencyAck {
			// TODO: Handle NACK, modify desired state
			// (https://github.com/TheThingsIndustries/ttn/issues/834)
			events.Publish(evtMACDLChannelReject(ctx, dev.EndDeviceIdentifiers, pld))
			return nil
		}

		req := cmd.GetDlChannelReq()

		if uint(req.ChannelIndex) >= uint(len(dev.MACState.CurrentParameters.Channels)) || dev.MACState.CurrentParameters.Channels[req.ChannelIndex] == nil {
			return errCorruptedMACState.WithCause(errUnknownChannel)
		}
		dev.MACState.CurrentParameters.Channels[req.ChannelIndex].DownlinkFrequency = req.Frequency

		events.Publish(evtMACDLChannelAccept(ctx, dev.EndDeviceIdentifiers, req))
		return nil

	}, dev.MACState.PendingRequests...)
	return
}
