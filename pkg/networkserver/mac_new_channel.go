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
	evtMACNewChannelRequest = events.Define("ns.mac.new_channel.request", "request new channel") // TODO(#988): publish when requesting
	evtMACNewChannelAccept  = events.Define("ns.mac.new_channel.accept", "device accepted new channel request")
	evtMACNewChannelReject  = events.Define("ns.mac.new_channel.reject", "device rejected new channel request")
)

func handleNewChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_NewChannelAns) (err error) {
	if pld == nil {
		return errMissingPayload
	}

	dev.MACState.PendingRequests, err = handleMACResponse(ttnpb.CID_NEW_CHANNEL, func(cmd *ttnpb.MACCommand) {
		if !pld.DataRateAck || !pld.FrequencyAck {
			// TODO: Handle NACK, modify desired state
			// (https://github.com/TheThingsIndustries/ttn/issues/834)
			events.Publish(evtMACNewChannelReject(ctx, dev.EndDeviceIdentifiers, pld))
			return
		}

		req := cmd.GetNewChannelReq()

		if uint(req.ChannelIndex) >= uint(len(dev.MACState.Channels)) {
			dev.MACState.MACParameters.Channels = append(dev.MACState.MACParameters.Channels, make([]*ttnpb.MACParameters_Channel, 1+int(req.ChannelIndex-uint32(len(dev.MACState.MACParameters.Channels))))...)
		}

		ch := dev.MACState.MACParameters.Channels[req.ChannelIndex]
		if ch == nil {
			ch = &ttnpb.MACParameters_Channel{
				DownlinkFrequency: req.Frequency,
			}
			dev.MACState.MACParameters.Channels[req.ChannelIndex] = ch
		}

		ch.UplinkFrequency = req.Frequency
		ch.MinDataRateIndex = req.MinDataRateIndex
		ch.MaxDataRateIndex = req.MaxDataRateIndex

		events.Publish(evtMACNewChannelAccept(ctx, dev.EndDeviceIdentifiers, req))
	}, dev.MACState.PendingRequests...)
	return
}
