// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	"google.golang.org/protobuf/proto"
)

var (
	// EvtEnqueueRelayConfRequest is emitted when a relay configuration request is enqueued.
	EvtEnqueueRelayConfRequest = defineEnqueueMACRequestEvent(
		"relay_conf", "relay configuration",
		events.WithDataType(&ttnpb.MACCommand_RelayConfReq{}),
	)()
	// EvtReceiveRelayConfAccept is emitted when a relay configuration request is accepted.
	EvtReceiveRelayConfAccept = defineReceiveMACAcceptEvent(
		"relay_conf", "relay configuration",
		events.WithDataType(&ttnpb.MACCommand_RelayConfAns{}),
	)()
	// EvtReceiveRelayConfReject is emitted when a relay configuration request is rejected.
	EvtReceiveRelayConfReject = defineReceiveMACRejectEvent(
		"relay_conf", "relay configuration",
		events.WithDataType(&ttnpb.MACCommand_RelayConfAns{}),
	)()
)

// DeviceNeedsRelayConfReq returns true iff the device needs a relay configuration request.
func DeviceNeedsRelayConfReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	currentServing := dev.MacState.GetCurrentParameters().GetRelay().GetServing()
	desiredServing := dev.MacState.GetDesiredParameters().GetRelay().GetServing()
	if desiredServing == nil && currentServing == nil {
		return false
	}
	if (desiredServing == nil) != (currentServing == nil) {
		return true
	}
	// NOTE: The forwarding rules are handled by UpdateUplinkListReq, not RelayConfReq.
	// NOTE: The limits are handled by ConfigureFwdLimitReq, not RelayConfReq.
	return !proto.Equal(&ttnpb.ServingRelayParameters{
		SecondChannel:       desiredServing.SecondChannel,
		DefaultChannelIndex: desiredServing.DefaultChannelIndex,
		CadPeriodicity:      desiredServing.CadPeriodicity,
	}, &ttnpb.ServingRelayParameters{
		SecondChannel:       currentServing.SecondChannel,
		DefaultChannelIndex: currentServing.DefaultChannelIndex,
		CadPeriodicity:      currentServing.CadPeriodicity,
	})
}

// EnqueueRelayConfReq enqueues a relay configuration request if needed.
func EnqueueRelayConfReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsRelayConfReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}
	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(
		ttnpb.MACCommandIdentifier_CID_RELAY_CONF,
		maxDownLen, maxUpLen,
		func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
			if nDown < 1 || nUp < 1 {
				return nil, 0, nil, false
			}
			desiredServing := dev.MacState.GetDesiredParameters().GetRelay().GetServing()
			req := &ttnpb.MACCommand_RelayConfReq{}
			if desiredServing != nil {
				req.Configuration = &ttnpb.MACCommand_RelayConfReq_Configuration{
					CadPeriodicity:      desiredServing.CadPeriodicity,
					DefaultChannelIndex: desiredServing.DefaultChannelIndex,
					SecondChannel:       desiredServing.SecondChannel,
				}
			}
			log.FromContext(ctx).WithFields(servingRelayFields(desiredServing)).Debug("Enqueued RelayConfReq")
			return []*ttnpb.MACCommand{
					req.MACCommand(),
				},
				1,
				events.Builders{
					EvtEnqueueRelayConfRequest.With(events.WithData(req)),
				},
				true
		},
		dev.MacState.PendingRequests...,
	)
	return st
}

// HandleRelayConfAns handles a relay configuration answer.
func HandleRelayConfAns(
	_ context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RelayConfAns,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	rejected := !pld.SecondChannelFrequencyAck ||
		!pld.SecondChannelAckOffsetAck ||
		!pld.SecondChannelDataRateIndexAck ||
		!pld.SecondChannelIndexAck ||
		!pld.DefaultChannelIndexAck ||
		!pld.CadPeriodicityAck
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(
		ttnpb.MACCommandIdentifier_CID_RELAY_CONF,
		false,
		func(cmd *ttnpb.MACCommand) error {
			if rejected {
				return nil
			}
			req := cmd.GetRelayConfReq()
			currentParameters := dev.MacState.CurrentParameters
			if req.Configuration == nil {
				currentParameters.Relay = nil
				return nil
			}
			relay := currentParameters.Relay
			if relay == nil {
				relay = &ttnpb.RelayParameters{}
				currentParameters.Relay = relay
			}
			serving := relay.GetServing()
			if serving == nil {
				serving = &ttnpb.ServingRelayParameters{}
				relay.Mode = &ttnpb.RelayParameters_Serving{
					Serving: serving,
				}
			}
			serving.CadPeriodicity = req.Configuration.CadPeriodicity
			serving.DefaultChannelIndex = req.Configuration.DefaultChannelIndex
			serving.SecondChannel = req.Configuration.SecondChannel
			return nil
		},
		dev.MacState.PendingRequests...,
	)
	ev := EvtReceiveRelayConfAccept
	if rejected {
		ev = EvtReceiveRelayConfReject
	}
	return events.Builders{
		ev.With(events.WithData(pld)),
	}, err
}
