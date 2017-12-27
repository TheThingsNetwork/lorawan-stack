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

// Package networkserver provides a LoRaWAN 1.1-compliant network server implementation.
package networkserver

import (
	"bytes"
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/networkserver/mac"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NetworkServer implements the network server component.
//
// The network server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component
	*deviceregistry.RegistryRPC
	registry deviceregistry.Interface
}

// Config represents the NetworkServer configuration.
type Config struct {
	Registry deviceregistry.Interface
}

// New returns new *NetworkServer.
func New(c *component.Component, conf *Config) *NetworkServer {
	ns := &NetworkServer{
		Component:   c,
		RegistryRPC: deviceregistry.NewRPC(c, conf.Registry), // TODO: Add checks
		registry:    conf.Registry,
	}
	c.RegisterGRPC(ns)
	return ns
}

// StartServingGateway is called by the gateway server to indicate that it is serving a gateway.
func (ns *NetworkServer) StartServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*ptypes.Empty, error) {
	// A GatewayServer (or proxy) links to ALL NetworkServers, so this better be efficient
	// TODO: Announce to cluster
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// StopServingGateway is called by the gateway server to indicate that it is no longer serving a gateway.
func (ns *NetworkServer) StopServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*ptypes.Empty, error) {
	// TODO: Announce to cluster
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// matchDevice tries to match the uplink message with a device.
// If the uplink message matches a fallback session, that fallback session is recovered.
// If successful, the FCnt in the uplink message is set to the full FCnt. The NextFCntUp in the session is updated accordingly.
func (ns *NetworkServer) matchDevice(ctx context.Context, uplink *ttnpb.UplinkMessage) (match *ttnpb.EndDevice, err error) {
	pld := uplink.Payload.GetMACPayload()

	if pld.DevAddr.IsZero() {
		return nil, ErrMissingDevAddr.New(nil)
	}

	// TODO: Not sure we can do this. We have to check both the DevAddr that is in the Session and the one that is in the FallbackSession.
	// Maybe the DevAddr field in EndDeviceIdentifiers should become a slice?
	devices, err := ns.registry.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevAddr: &pld.DevAddr,
	})
	if err != nil {
		return nil, err
	}

	// TODO: Sort (by FCnt) so that most likely matches come first

	for _, device := range devices {
		var session *ttnpb.Session
		if device.Session != nil && device.Session.DevAddr.Equal(pld.DevAddr) {
			session = device.Session
		} else if device.SessionFallback != nil && device.SessionFallback.DevAddr.Equal(pld.DevAddr) {
			session = device.SessionFallback
		}
		if session == nil {
			continue
		}

		fullFCnt, err := validFCnt(pld.FCnt, session.NextFCntUp)
		if err != nil {
			continue
		}

		switch device.LoRaWANVersion {
		case ttnpb.MAC_V1_1:
			ke := session.GetSNwkSIntKey()
			if ke == nil {
				return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("TODO: SNwkSIntKey not present"))
			}
			if ke.Key == nil || ke.Key.IsZero() {
				return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("TODO: SNwkSIntKey not present"))
			}
			sNwkSIntKey := *ke.Key
			ke = session.GetFNwkSIntKey()
			if ke == nil {
				return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("TODO: SNwkSIntKey not present"))
			}
			if ke.Key == nil || ke.Key.IsZero() {
				return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("TODO: SNwkSIntKey not present"))
			}
			fNwkSIntKey := *ke.Key
			var confFCnt uint32
			if pld.Ack && len(device.RecentDownlinks) > 0 {
				lastDownlink := device.RecentDownlinks[len(device.RecentDownlinks)-1].GetPayload()
				confFCnt = lastDownlink.GetMACPayload().GetFCnt()
			}
			var txDRIdx uint8 // TODO: determine
			var txChIdx uint8 // TODO: determine
			expectedMIC, err := crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, confFCnt, txDRIdx, txChIdx, pld.DevAddr, fullFCnt, pld.FRMPayload)
			if err != nil {
				return nil, errors.New("TODO: ErrMICComputeFailed")
			}
			if !bytes.Equal(uplink.Payload.MIC, expectedMIC[:]) {
				continue
			}
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			ke := session.GetFNwkSIntKey()
			if ke == nil {
				return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("TODO: SNwkSIntKey not present"))
			}
			if ke.Key == nil || ke.Key.IsZero() {
				return nil, ErrCorruptRegistry.NewWithCause(nil, errors.New("TODO: SNwkSIntKey not present"))
			}
			fNwkSIntKey := *ke.Key
			expectedMIC, err := crypto.ComputeLegacyUplinkMIC(fNwkSIntKey, pld.DevAddr, fullFCnt, pld.FRMPayload)
			if err != nil {
				return nil, errors.New("TODO: ErrMICComputeFailed")
			}
			if !bytes.Equal(uplink.Payload.MIC, expectedMIC[:]) {
				continue
			}
		}

		pld.FCnt = fullFCnt
		device.Session = session
		device.Session.NextFCntUp = pld.FCnt + 1
		device.SessionFallback = nil

		return device.EndDevice, nil
	}

	return nil, errors.New("TODO: ErrNotFound")
}

func validFCnt(messageFCnt, deviceNextFCnt uint32) (fullFCnt uint32, err error) {
	// TODO: change 16 MSB of messageFCnt if needed (see v2 https://github.com/TheThingsNetwork/ttn/blob/develop/utils/fcnt/fcnt.go)
	// TODO: compare to deviceFCnt (at least deviceNextFCnt but not too large), otherwise return ErrFCntTooLow or ErrFCntTooHigh
	return messageFCnt, nil // TODO: return new FCnt
}

func (ns *NetworkServer) handleUplink(ctx context.Context, uplink *ttnpb.UplinkMessage) (*ptypes.Empty, error) {
	pld := uplink.Payload.GetMACPayload()
	if pld == nil {
		return nil, ErrMissingPayload.New(nil)
	}

	// TODO: Deduplicate the message. We can assume that messages are correctly sharded most of the time and that we were
	// the only NS that receives it, but maybe we should have a distributed "claim-lock" that expires after a few seconds.
	// See v2 https://github.com/TheThingsNetwork/ttn/blob/develop/core/broker/deduplicator.go

	device, err := ns.matchDevice(ctx, uplink)
	if err != nil {
		return nil, errors.NewWithCause("Could not find matching device", err)
	}

	device.RecentUplinks = append(device.RecentUplinks, uplink)
	if len(device.RecentUplinks) > 20 { // TODO: should that 20 be configurable?
		device.RecentUplinks = device.RecentUplinks[len(device.RecentUplinks)-20:]
	}

	// TODO: Prepare Downlink
	var downlink *ttnpb.DownlinkMessage

	// TODO: Determine order of preferece of downlink gateways

	err = mac.HandleUplink(ctx, device, uplink)
	if err != nil {
		return nil, errors.NewWithCause("Could not handle uplink MAC commands", err)
	}

	err = mac.UpdateQueue(device)
	if err != nil {
		return nil, errors.NewWithCause("Could not determine MAC commands to queue", err)
	}

	// TODO: Save device

	// TODO: Send on ApplicationServer stream if any

	// TODO: Wait for response from ApplicationServer?

	err = mac.HandleDownlink(ctx, device, downlink)
	if err != nil {
		return nil, errors.NewWithCause("Could not set downlink MAC commands", err)
	}

	// TODO: Try scheduling downlink (if needed) on downlink gateways (ordered by preference)

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (ns *NetworkServer) handleJoin(ctx context.Context, uplink *ttnpb.UplinkMessage) (*ptypes.Empty, error) {
	pld := uplink.Payload.GetJoinRequestPayload()
	if pld == nil {
		return nil, ErrMissingPayload.New(nil)
	}

	if pld.DevEUI.IsZero() {
		return nil, errors.New("TODO: ErrMissingDevEUI")
	}
	if pld.JoinEUI.IsZero() {
		return nil, errors.New("TODO: ErrMissingJoinEUI")
	}

	var devAddr types.DevAddr // TODO: fill random

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &ttnpb.EndDeviceIdentifiers{
		DevEUI:  &pld.DevEUI,
		JoinEUI: &pld.JoinEUI,
	})
	if err == nil {
		if dev.Session != nil && dev.Session.DevAddr.Equal(devAddr) {
			// TODO: Don't want any confusion or possibility tracking devices between sessions, make sure it has a different DevAddr
		}
		// TODO: apply device preferences to the join request (MAC version, Rx timing)
	}

	// TODO: Save device

	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the gateway server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, uplink *ttnpb.UplinkMessage) (*ptypes.Empty, error) {
	rawPayload := uplink.GetRawPayload()
	if uplink.Payload.Payload == nil {
		if rawPayload == nil {
			return nil, ErrMissingPayload.New(nil)
		}
		if err := uplink.Payload.UnmarshalLoRaWAN(rawPayload); err != nil {
			return nil, ErrUnmarshalFailed.NewWithCause(nil, err)
		}
	}

	msg := uplink.GetPayload()
	if msg.GetMajor() != ttnpb.Major_LORAWAN_R1 {
		return nil, ErrUnsupportedLoRaWANMajorVersion.New(errors.Attributes{
			"major": msg.GetMajor(),
		})
	}

	switch mType := msg.GetMType(); mType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return ns.handleUplink(ctx, uplink)
	case ttnpb.MType_JOIN_REQUEST:
		return ns.handleJoin(ctx, uplink)
	default:
		return nil, ErrWrongPayloadType.New(errors.Attributes{
			"type": mType,
		})
	}
}

// LinkApplication is called by the application server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(appID *ttnpb.ApplicationIdentifiers, stream ttnpb.AsNs_LinkApplicationServer) error {
	// An ApplicationServer (or proxy) links to ALL NetworkServers, so this better be efficient
	// TODO: Can there be more than one subscription?
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueReplace is called by the application server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	// TODO: authenticate
	dev.QueuedApplicationDownlinks = req.Downlinks
	// TODO: save
	return new(ptypes.Empty), nil
}

// DownlinkQueuePush is called by the application server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, &req.EndDeviceIdentifiers)
	if err != nil {
		return nil, err
	}
	// TODO: authenticate
	dev.QueuedApplicationDownlinks = append(dev.QueuedApplicationDownlinks, req.Downlinks...)
	// TODO: save
	return new(ptypes.Empty), nil
}

// DownlinkQueueList is called by the application server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, req)
	if err != nil {
		return nil, err
	}
	// TODO: authenticate
	return &ttnpb.ApplicationDownlinks{Downlinks: dev.QueuedApplicationDownlinks}, nil
}

// DownlinkQueueClear is called by the application server to clear the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueClear(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*ptypes.Empty, error) {
	dev, err := deviceregistry.FindOneDeviceByIdentifiers(ns.registry, req)
	if err != nil {
		return nil, err
	}
	// TODO: authenticate
	dev.QueuedApplicationDownlinks = nil
	// TODO: save
	return new(ptypes.Empty), nil
}

// RegisterServices registers services provided by ns at s.
func (ns *NetworkServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsNsServer(s, ns)
	ttnpb.RegisterAsNsServer(s, ns)
	ttnpb.RegisterNsApplicationDownlinkQueueServer(s, ns)
	ttnpb.RegisterNsDeviceRegistryServer(s, ns)
}

// RegisterHandlers registers gRPC handlers.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterNsDeviceRegistryHandler(ns.Context(), s, conn)
}

// Roles returns the roles that the network server fulfils
func (ns *NetworkServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_NETWORK_SERVER}
}
