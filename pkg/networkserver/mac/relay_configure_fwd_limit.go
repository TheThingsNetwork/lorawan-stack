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
	// defaultRelayForwardingLimits is the default relay forward limits, based on the contents of the
	// relay specification.
	defaultRelayForwardingLimits = &ttnpb.ServingRelayForwardingLimits{
		JoinRequests: &ttnpb.RelayForwardLimits{
			BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
			ReloadRate: 4,
		},
		Notifications: &ttnpb.RelayForwardLimits{
			BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
			ReloadRate: 4,
		},
		UplinkMessages: &ttnpb.RelayForwardLimits{
			BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
			ReloadRate: 8,
		},
		Overall: &ttnpb.RelayForwardLimits{
			BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
			ReloadRate: 8,
		},
	}

	// EvtEnqueueRelayConfigureFwdLimitRequest is emitted when a relay forward limits configuration request is enqueued.
	EvtEnqueueRelayConfigureFwdLimitRequest = defineEnqueueMACRequestEvent(
		"relay_configure_fwd_limit", "relay configure forward limit",
		events.WithDataType(&ttnpb.MACCommand_RelayConfigureFwdLimitReq{}),
	)()
	// EvtReceiveRelayConfigureFwdLimitAnswer is emitted when a relay forward limits configuration request is answered.
	EvtReceiveRelayConfigureFwdLimitAnswer = defineReceiveMACAnswerEvent(
		"relay_configure_fwd_limit", "relay configure forward limit",
		events.WithDataType(&ttnpb.MACCommand_RelayConfigureFwdLimitAns{}),
	)()
)

// DeviceNeedsRelayConfigureFwdLimitReq returns true iff the device needs a relay forward limits configuration request.
func DeviceNeedsRelayConfigureFwdLimitReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	currentLimits := dev.MacState.GetCurrentParameters().GetRelay().GetServing().GetLimits()
	desiredLimits := dev.MacState.GetDesiredParameters().GetRelay().GetServing().GetLimits()
	if desiredLimits == nil && currentLimits == nil {
		return false
	}
	if currentLimits == nil {
		currentLimits = defaultRelayForwardingLimits
	}
	if desiredLimits == nil {
		desiredLimits = defaultRelayForwardingLimits
	}
	return !proto.Equal(desiredLimits, currentLimits)
}

// EnqueueRelayConfigureFwdLimitReq enqueues a relay forward limits configuration request if needed.
func EnqueueRelayConfigureFwdLimitReq(
	ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16,
) EnqueueState {
	if !DeviceNeedsRelayConfigureFwdLimitReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}
	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(
		ttnpb.MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT,
		maxDownLen, maxUpLen,
		func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
			if nDown < 1 || nUp < 1 {
				return nil, 0, nil, false
			}
			desiredLimits := dev.MacState.GetDesiredParameters().GetRelay().GetServing().GetLimits()
			if desiredLimits == nil {
				desiredLimits = defaultRelayForwardingLimits
			}
			req := &ttnpb.MACCommand_RelayConfigureFwdLimitReq{
				ResetLimitCounter:  desiredLimits.ResetBehavior,
				JoinRequestLimits:  desiredLimits.JoinRequests,
				NotifyLimits:       desiredLimits.Notifications,
				GlobalUplinkLimits: desiredLimits.UplinkMessages,
				OverallLimits:      desiredLimits.Overall,
			}
			log.FromContext(ctx).WithFields(relayConfigureForwardLimitsFields(desiredLimits)).
				Debug("Enqueued RelayConfigureFwdLimitReq")
			return []*ttnpb.MACCommand{
					req.MACCommand(),
				},
				1,
				events.Builders{
					EvtEnqueueRelayConfigureFwdLimitRequest.With(events.WithData(req)),
				},
				true
		},
		dev.MacState.PendingRequests...,
	)
	return st
}

// HandleRelayConfigureFwdLimitAns handles a relay forward limits configuration answer.
func HandleRelayConfigureFwdLimitAns(
	_ context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RelayConfigureFwdLimitAns,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(
		ttnpb.MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT,
		false,
		func(cmd *ttnpb.MACCommand) error {
			req := cmd.GetRelayConfigureFwdLimitReq()
			currentServing := dev.MacState.GetCurrentParameters().GetRelay().GetServing()
			if currentServing == nil {
				// NOTE: EnqueueRelayConfigureFwdLimitReq is optimistic and assumes that EnqueueRelayConfReq
				// has enqueued the desired relay parameters, and that the end device will accept them. If
				// either of these conditions is not true, the current serving parameters will be nil.
				return nil
			}
			currentServing.Limits = &ttnpb.ServingRelayForwardingLimits{
				ResetBehavior:  req.ResetLimitCounter,
				JoinRequests:   req.JoinRequestLimits,
				Notifications:  req.NotifyLimits,
				UplinkMessages: req.GlobalUplinkLimits,
				Overall:        req.OverallLimits,
			}
			return nil
		},
		dev.MacState.PendingRequests...,
	)
	return events.Builders{
		EvtReceiveRelayConfigureFwdLimitAnswer.With(events.WithData(pld)),
	}, err
}
