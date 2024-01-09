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
	// EvtEnqueueRelayEndDeviceConfRequest is emitted when a relay end device configuration request is enqueued.
	EvtEnqueueRelayEndDeviceConfRequest = defineEnqueueMACRequestEvent(
		"relay_end_device_conf", "relay end device configuration",
		events.WithDataType(&ttnpb.MACCommand_RelayEndDeviceConfReq{}),
	)()
	// EvtReceiveRelayEndDeviceConfAccept is emitted when a relay end device configuration request is accepted.
	EvtReceiveRelayEndDeviceConfAccept = defineReceiveMACAcceptEvent(
		"relay_end_device_conf", "relay end device configuration",
		events.WithDataType(&ttnpb.MACCommand_RelayEndDeviceConfAns{}),
	)()
	// EvtReceiveRelayEndDeviceConfReject is emitted when a relay end device configuration request is rejected.
	EvtReceiveRelayEndDeviceConfReject = defineReceiveMACRejectEvent(
		"relay_end_device_conf", "relay end device configuration",
		events.WithDataType(&ttnpb.MACCommand_RelayEndDeviceConfAns{}),
	)()
)

// DeviceNeedsRelayEndDeviceConfReq returns true iff the device needs a relay end device configuration request.
func DeviceNeedsRelayEndDeviceConfReq(dev *ttnpb.EndDevice) bool {
	if dev.GetMulticast() || dev.GetMacState() == nil {
		return false
	}
	currentServed := dev.MacState.GetCurrentParameters().GetRelay().GetServed()
	desiredServed := dev.MacState.GetDesiredParameters().GetRelay().GetServed()
	return !proto.Equal(
		&ttnpb.ServedRelayParameters{
			Mode:          desiredServed.GetMode(),
			Backoff:       desiredServed.GetBackoff(),
			SecondChannel: desiredServed.GetSecondChannel(),
		},
		&ttnpb.ServedRelayParameters{
			Mode:          currentServed.GetMode(),
			Backoff:       currentServed.GetBackoff(),
			SecondChannel: currentServed.GetSecondChannel(),
		},
	)
}

// EnqueueRelayEndDeviceConfReq enqueues a relay end device configuration request if needed.
func EnqueueRelayEndDeviceConfReq(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) EnqueueState {
	if !DeviceNeedsRelayEndDeviceConfReq(dev) {
		return EnqueueState{
			MaxDownLen: maxDownLen,
			MaxUpLen:   maxUpLen,
			Ok:         true,
		}
	}
	var st EnqueueState
	dev.MacState.PendingRequests, st = enqueueMACCommand(
		ttnpb.MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF,
		maxDownLen, maxUpLen,
		func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool) {
			if nDown < 1 || nUp < 1 {
				return nil, 0, nil, false
			}
			desiredServed := dev.MacState.GetDesiredParameters().GetRelay().GetServed()
			req := &ttnpb.MACCommand_RelayEndDeviceConfReq{}
			if desiredServed != nil {
				req.Configuration = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
					Backoff:         desiredServed.Backoff,
					SecondChannel:   desiredServed.SecondChannel,
					ServingDeviceId: desiredServed.ServingDeviceId,
				}
				switch {
				case desiredServed.GetAlways() != nil:
					req.Configuration.Mode = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always{
						Always: desiredServed.GetAlways(),
					}
				case desiredServed.GetDynamic() != nil:
					req.Configuration.Mode = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Dynamic{
						Dynamic: desiredServed.GetDynamic(),
					}
				case desiredServed.GetEndDeviceControlled() != nil:
					req.Configuration.Mode = &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_EndDeviceControlled{
						EndDeviceControlled: desiredServed.GetEndDeviceControlled(),
					}
				default:
					panic("unreachable")
				}
			}
			log.FromContext(ctx).WithFields(servedRelayFields(desiredServed)).Debug("Enqueued RelayEndDeviceConfReq")
			return []*ttnpb.MACCommand{
					req.MACCommand(),
				},
				1,
				events.Builders{
					EvtEnqueueRelayEndDeviceConfRequest.With(events.WithData(req)),
				},
				true
		},
		dev.MacState.PendingRequests...,
	)
	return st
}

// HandleRelayEndDeviceConfAns handles a relay end device configuration answer.
func HandleRelayEndDeviceConfAns(
	_ context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RelayEndDeviceConfAns,
) (events.Builders, error) {
	if pld == nil {
		return nil, ErrNoPayload.New()
	}
	rejected := !pld.SecondChannelFrequencyAck ||
		// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
		!pld.SecondChannelDataRateIndexAck ||
		!pld.SecondChannelIndexAck ||
		!pld.BackoffAck
	var err error
	dev.MacState.PendingRequests, err = handleMACResponse(
		ttnpb.MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF,
		false,
		func(cmd *ttnpb.MACCommand) error {
			if rejected {
				return nil
			}
			conf := cmd.GetRelayEndDeviceConfReq().Configuration
			currentParameters := dev.MacState.CurrentParameters
			if conf == nil {
				currentParameters.Relay = nil
				return nil
			}
			relay := currentParameters.Relay
			if relay == nil {
				relay = &ttnpb.RelayParameters{}
				currentParameters.Relay = relay
			}
			served := relay.GetServed()
			if served == nil {
				served = &ttnpb.ServedRelayParameters{}
				relay.Mode = &ttnpb.RelayParameters_Served{
					Served: served,
				}
			}
			switch {
			case conf.GetAlways() != nil:
				served.Mode = &ttnpb.ServedRelayParameters_Always{
					Always: conf.GetAlways(),
				}
			case conf.GetDynamic() != nil:
				served.Mode = &ttnpb.ServedRelayParameters_Dynamic{
					Dynamic: conf.GetDynamic(),
				}
			case conf.GetEndDeviceControlled() != nil:
				served.Mode = &ttnpb.ServedRelayParameters_EndDeviceControlled{
					EndDeviceControlled: conf.GetEndDeviceControlled(),
				}
			default:
				panic("unreachable")
			}
			served.Backoff = conf.Backoff
			served.SecondChannel = conf.SecondChannel
			served.ServingDeviceId = conf.ServingDeviceId
			return nil
		},
		dev.MacState.PendingRequests...,
	)
	ev := EvtReceiveRelayEndDeviceConfAccept
	if rejected {
		ev = EvtReceiveRelayEndDeviceConfReject
	}
	return events.Builders{
		ev.With(events.WithData(pld)),
	}, err
}
