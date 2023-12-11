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
	// EvtEnqueueRelayCtrlUplinkListRequest is emitted when a relay control uplink list request is enqueued.
	EvtEnqueueRelayCtrlUplinkListRequest = defineEnqueueMACRequestEvent(
		"relay_ctrl_uplink_list", "relay control uplink list",
		events.WithDataType(&ttnpb.MACCommand_RelayCtrlUplinkListReq{}),
	)()
	// EvtReceiveRelayCtrlUplinkListAccept is emitted when a relay control uplink list request is accepted.
	EvtReceiveRelayCtrlUplinkListAccept = defineReceiveMACAcceptEvent(
		"relay_ctrl_uplink_list", "relay control uplink list",
		events.WithDataType(&ttnpb.MACCommand_RelayCtrlUplinkListAns{}),
	)()
	// EvtReceiveRelayCtrlUplinkListReject is emitted when a relay control uplink list request is rejected.
	EvtReceiveRelayCtrlUplinkListReject = defineReceiveMACRejectEvent(
		"relay_ctrl_uplink_list", "relay control uplink list",
		events.WithDataType(&ttnpb.MACCommand_RelayCtrlUplinkListAns{}),
	)()
)

// DeviceNeedsRelayCtrlUplinkListReqAtIndex returns true iff the device needs a relay
// control uplink list request at the given index.
func DeviceNeedsRelayCtrlUplinkListReqAtIndex(dev *ttnpb.EndDevice, i int) bool {
	currentRules := dev.GetMacState().GetCurrentParameters().GetRelay().GetServing().GetUplinkForwardingRules()
	desiredRules := dev.GetMacState().GetDesiredParameters().GetRelay().GetServing().GetUplinkForwardingRules()
	switch {
	case i >= len(currentRules) && i >= len(desiredRules):
	case i >= len(desiredRules), proto.Equal(desiredRules[i], emptyRelayUplinkForwardingRule):
		// A rule is desired to be deleted.
		return !proto.Equal(currentRules[i], emptyRelayUplinkForwardingRule)
	case i >= len(currentRules), proto.Equal(currentRules[i], emptyRelayUplinkForwardingRule):
		// A rule is desired to be created.
		// NOTE: CtrlUplinkListReq cannot delete a forwarding rule.
	default:
		// NOTE: CtrlUplinkListReq cannot update a forwarding rule.
	}
	return false
}

// DeviceNeedsRelayCtrlUplinkListReq returns true iff the device needs a relay control uplink list request.
func DeviceNeedsRelayCtrlUplinkListReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	currentServing := dev.MacState.GetCurrentParameters().GetRelay().GetServing()
	desiredServing := dev.MacState.GetDesiredParameters().GetRelay().GetServing()
	if desiredServing == nil {
		return false
	}
	for i := range currentServing.GetUplinkForwardingRules() {
		if DeviceNeedsRelayCtrlUplinkListReqAtIndex(dev, i) {
			return true
		}
	}
	return false
}

// EnqueueRelayCtrlUplinkListReq enqueues a relay control uplink list request.
func EnqueueRelayCtrlUplinkListReq(
	ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16,
) EnqueueState {
	if !DeviceNeedsRelayCtrlUplinkListReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}
	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(
		ttnpb.MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST,
		maxDownLen, maxUpLen,
		func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
			if nDown < 1 || nUp < 1 {
				return nil, 0, nil, false
			}
			currentRules := dev.MacState.GetCurrentParameters().GetRelay().GetServing().GetUplinkForwardingRules()
			desiredRules := dev.MacState.DesiredParameters.Relay.GetServing().UplinkForwardingRules
			var reqs []*ttnpb.MACCommand_RelayCtrlUplinkListReq
			for i := 0; i < len(currentRules); i++ {
				switch {
				case !DeviceNeedsRelayCtrlUplinkListReqAtIndex(dev, i):
				case i >= len(desiredRules), proto.Equal(desiredRules[i], emptyRelayUplinkForwardingRule):
					reqs = append(reqs, &ttnpb.MACCommand_RelayCtrlUplinkListReq{
						RuleIndex: uint32(i),
						Action:    ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE,
					})
				}
			}
			cmds := make([]*ttnpb.MACCommand, 0, len(reqs))
			evs := make(events.Builders, 0, len(reqs))
			for _, req := range reqs {
				if nDown < 1 || nUp < 1 {
					return cmds, uint16(len(cmds)), evs, false
				}
				nDown--
				nUp--
				log.FromContext(ctx).WithFields(relayCtrlUplinkListReqFields(req)).
					Debug("Enqueued RelayCtrlUplinkListReq")
				cmds = append(cmds, req.MACCommand())
				evs = append(evs, EvtEnqueueRelayCtrlUplinkListRequest.With(events.WithData(req)))
			}
			return cmds, uint16(len(cmds)), evs, true
		},
		dev.MacState.PendingRequests...,
	)
	return st
}

// HandleRelayCtrlUplinkListAns handles a relay control uplink list answer.
func HandleRelayCtrlUplinkListAns(
	_ context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RelayCtrlUplinkListAns,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(
		ttnpb.MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST,
		false,
		func(cmd *ttnpb.MACCommand) error {
			if !pld.RuleIndexAck {
				return nil
			}
			req := cmd.GetRelayCtrlUplinkListReq()
			currentServing := dev.MacState.GetCurrentParameters().GetRelay().GetServing()
			if currentServing == nil || req.RuleIndex >= uint32(len(currentServing.UplinkForwardingRules)) {
				return nil
			}
			switch req.Action {
			case ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_READ_W_F_CNT:
			case ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE:
				currentServing.UplinkForwardingRules[req.RuleIndex] = &ttnpb.RelayUplinkForwardingRule{}
			default:
				panic("unreachable")
			}
			return nil
		},
		dev.MacState.PendingRequests...,
	)
	ev := EvtReceiveRelayCtrlUplinkListAccept
	if !pld.RuleIndexAck {
		ev = EvtReceiveRelayCtrlUplinkListReject
	}
	return []events.Builder{
		ev.With(events.WithData(pld)),
	}, err
}
