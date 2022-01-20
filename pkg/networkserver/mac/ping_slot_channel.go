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
	EvtEnqueuePingSlotChannelRequest = defineEnqueueMACRequestEvent(
		"ping_slot_channel", "ping slot channel",
		events.WithDataType(&ttnpb.MACCommand_PingSlotChannelReq{}),
	)()
	EvtReceivePingSlotChannelAnswer = defineReceiveMACAcceptEvent(
		"ping_slot_channel", "ping slot channel",
		events.WithDataType(&ttnpb.MACCommand_PingSlotChannelAns{}),
	)()
)

func DeviceNeedsPingSlotChannelReq(dev *ttnpb.EndDevice) bool {
	switch {
	case dev.GetMulticast(),
		dev.GetMacState() == nil:
		return false
	case dev.MacState.DesiredParameters.PingSlotFrequency != dev.MacState.CurrentParameters.PingSlotFrequency:
		return true
	case dev.MacState.DesiredParameters.PingSlotDataRateIndexValue == nil:
		return false
	case dev.MacState.CurrentParameters.PingSlotDataRateIndexValue == nil,
		dev.MacState.DesiredParameters.PingSlotDataRateIndexValue.Value != dev.MacState.CurrentParameters.PingSlotDataRateIndexValue.Value:
		return true
	}
	return false
}

func EnqueuePingSlotChannelReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsPingSlotChannelReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}

	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(ttnpb.MACCommandIdentifier_CID_PING_SLOT_CHANNEL, maxDownLen, maxUpLen, func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
		if nDown < 1 || nUp < 1 {
			return nil, 0, nil, false
		}
		req := &ttnpb.MACCommand_PingSlotChannelReq{
			Frequency:     dev.MacState.DesiredParameters.PingSlotFrequency,
			DataRateIndex: dev.MacState.DesiredParameters.PingSlotDataRateIndexValue.Value,
		}
		log.FromContext(ctx).WithFields(log.Fields(
			"frequency", req.Frequency,
			"data_rate_index", req.DataRateIndex,
		)).Debug("Enqueued PingSlotChannelReq")
		return []*ttnpb.MACCommand{
				req.MACCommand(),
			},
			1,
			events.Builders{
				EvtEnqueuePingSlotChannelRequest.With(events.WithData(req)),
			},
			true
	}, dev.MacState.PendingRequests...)
	return st
}

func HandlePingSlotChannelAns(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_PingSlotChannelAns) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}

	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(ttnpb.MACCommandIdentifier_CID_PING_SLOT_CHANNEL, func(cmd *ttnpb.MACCommand) error {
		req := cmd.GetPingSlotChannelReq()

		dev.MacState.CurrentParameters.PingSlotFrequency = req.Frequency
		dev.MacState.CurrentParameters.PingSlotDataRateIndexValue = &ttnpb.DataRateIndexValue{Value: req.DataRateIndex}
		return nil
	}, dev.MacState.PendingRequests...)
	return events.Builders{
		EvtReceivePingSlotChannelAnswer.With(events.WithData(pld)),
	}, err
}
