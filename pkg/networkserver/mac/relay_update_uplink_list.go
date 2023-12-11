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
	emptyRelayUplinkForwardingRule = &ttnpb.RelayUplinkForwardingRule{}

	// EvtEnqueueRelayUpdateUplinkListRequest is emitted when a relay update uplink list request is enqueued.
	EvtEnqueueRelayUpdateUplinkListRequest = defineEnqueueMACRequestEvent(
		"relay_update_uplink_list", "relay update uplink list",
		events.WithDataType(&ttnpb.MACCommand_RelayUpdateUplinkListReq{}),
	)()
	// EvtReceiveRelayUpdateUplinkListAnswer is emitted when a relay update uplink list request is answered.
	EvtReceiveRelayUpdateUplinkListAnswer = defineReceiveMACAnswerEvent(
		"relay_update_uplink_list", "relay update uplink list",
		events.WithDataType(&ttnpb.MACCommand_RelayUpdateUplinkListAns{}),
	)()
)

// DeviceNeedsRelayUpdateUplinkListReqAtIndex returns true iff the device needs a relay
// update uplink list request at the given index.
func DeviceNeedsRelayUpdateUplinkListReqAtIndex(dev *ttnpb.EndDevice, i int) bool {
	currentRules := dev.GetMacState().GetCurrentParameters().GetRelay().GetServing().GetUplinkForwardingRules()
	desiredRules := dev.GetMacState().GetDesiredParameters().GetRelay().GetServing().GetUplinkForwardingRules()
	switch {
	case i >= len(currentRules) && i >= len(desiredRules):
	case i >= len(desiredRules), proto.Equal(desiredRules[i], emptyRelayUplinkForwardingRule):
		// A rule is desired to be deleted.
		// NOTE: UpdateUplinkListReq cannot delete a forwarding rule.
	case i >= len(currentRules), proto.Equal(currentRules[i], emptyRelayUplinkForwardingRule):
		// A rule is desired to be created.
		return true
	default:
		// A rule is desired to be updated.
		return !proto.Equal(desiredRules[i], currentRules[i])
	}
	return false
}

// DeviceNeedsRelayUpdateUplinkListReq returns true iff the device needs a relay update uplink list request.
func DeviceNeedsRelayUpdateUplinkListReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	currentServing := dev.MacState.GetCurrentParameters().GetRelay().GetServing()
	desiredServing := dev.MacState.GetDesiredParameters().GetRelay().GetServing()
	if desiredServing == nil {
		return false
	}
	if len(desiredServing.GetUplinkForwardingRules()) > len(currentServing.GetUplinkForwardingRules()) {
		return true
	}
	for i := range desiredServing.GetUplinkForwardingRules() {
		if DeviceNeedsRelayUpdateUplinkListReqAtIndex(dev, i) {
			return true
		}
	}
	return false
}

// EnqueueRelayUpdateUplinkListReq enqueues a relay update uplink list request.
func EnqueueRelayUpdateUplinkListReq(
	ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, keyService RelayKeyService,
) EnqueueState {
	if !DeviceNeedsRelayUpdateUplinkListReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}
	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(
		ttnpb.MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST,
		maxDownLen, maxUpLen,
		func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
			if nDown < 1 || nUp < 1 {
				return nil, 0, nil, false
			}
			currentRules := dev.MacState.GetCurrentParameters().GetRelay().GetServing().GetUplinkForwardingRules()
			desiredRules := dev.MacState.DesiredParameters.Relay.GetServing().UplinkForwardingRules
			pendingRuleIndices := make([]int, 0, len(desiredRules))
			pendingDeviceIDs := make([]string, 0, len(desiredRules))
			pendingSessionKeyIDs := make([][]byte, 0, len(desiredRules))
			var reqs []*ttnpb.MACCommand_RelayUpdateUplinkListReq
			for i := 0; i < len(desiredRules) || i < len(currentRules); i++ {
				if DeviceNeedsRelayUpdateUplinkListReqAtIndex(dev, i) {
					pendingRuleIndices = append(pendingRuleIndices, i)
					pendingDeviceIDs = append(pendingDeviceIDs, desiredRules[i].DeviceId)
					pendingSessionKeyIDs = append(pendingSessionKeyIDs, desiredRules[i].SessionKeyId)
				}
			}
			devAddrs, keys, err := keyService.BatchDeriveRootWorSKey(
				ctx, dev.Ids.ApplicationIds, pendingDeviceIDs, pendingSessionKeyIDs,
			)
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Root relay session keys derivation failed")
				return nil, 0, nil, true
			}
			for i, ruleIdx := range pendingRuleIndices {
				if devAddrs[i] == nil || keys[i] == nil {
					continue
				}
				desiredRule := desiredRules[ruleIdx]
				reqs = append(reqs, &ttnpb.MACCommand_RelayUpdateUplinkListReq{
					RuleIndex:     uint32(ruleIdx),
					ForwardLimits: desiredRule.Limits,
					DevAddr:       devAddrs[i][:],
					WFCnt:         desiredRule.LastWFCnt,
					RootWorSKey:   keys[i][:],

					DeviceId:     desiredRule.DeviceId,
					SessionKeyId: desiredRule.SessionKeyId,
				})
			}
			cmds := make([]*ttnpb.MACCommand, 0, len(reqs))
			evs := make(events.Builders, 0, len(reqs))
			for _, req := range reqs {
				if nDown < 1 || nUp < 1 {
					return cmds, uint16(len(cmds)), evs, false
				}
				nDown--
				nUp--
				log.FromContext(ctx).WithFields(relayUpdateUplinkListReqFields(req)).
					Debug("Enqueued RelayUpdateUplinkListReq")
				cmds = append(cmds, req.MACCommand())
				evs = append(evs, EvtEnqueueRelayUpdateUplinkListRequest.With(events.WithData(req.Sanitized())))
			}
			return cmds, uint16(len(cmds)), evs, true
		},
		dev.MacState.PendingRequests...,
	)
	return st
}

// HandleRelayUpdateUplinkListAns handles a relay update uplink list answer.
func HandleRelayUpdateUplinkListAns(
	_ context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RelayUpdateUplinkListAns,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(
		ttnpb.MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST,
		false,
		func(cmd *ttnpb.MACCommand) error {
			req := cmd.GetRelayUpdateUplinkListReq()
			currentServing := dev.MacState.GetCurrentParameters().GetRelay().GetServing()
			if currentServing == nil {
				// NOTE: EnqueueRelayUpdateUplinkListReq is optimistic and assumes that EnqueueRelayConfReq
				// has enqueued the desired relay parameters, and that the end device will accept them. If
				// either of these conditions is not true, the current serving parameters will be nil.
				return nil
			}
			if n := len(currentServing.UplinkForwardingRules); uint(req.RuleIndex) >= uint(n) {
				currentServing.UplinkForwardingRules = append(
					currentServing.UplinkForwardingRules,
					make(
						[]*ttnpb.RelayUplinkForwardingRule,
						1+int(req.RuleIndex-uint32(n)),
					)...,
				)
				for i := n; i < len(currentServing.UplinkForwardingRules); i++ {
					currentServing.UplinkForwardingRules[i] = &ttnpb.RelayUplinkForwardingRule{}
				}
			}
			rule := currentServing.UplinkForwardingRules[req.RuleIndex]
			rule.Limits = req.ForwardLimits
			rule.LastWFCnt = req.WFCnt
			rule.DeviceId = req.DeviceId
			rule.SessionKeyId = req.SessionKeyId
			return nil
		},
		dev.MacState.PendingRequests...,
	)
	return []events.Builder{
		EvtReceiveRelayUpdateUplinkListAnswer.With(events.WithData(pld)),
	}, err
}
