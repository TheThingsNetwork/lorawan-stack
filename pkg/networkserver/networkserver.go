// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
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
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// StopServingGateway is called by the gateway server to indicate that it is no longer serving a gateway.
func (ns *NetworkServer) StopServingGateway(ctx context.Context, gtwID *ttnpb.GatewayIdentifiers) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the gateway server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, uplink *ttnpb.UplinkMessage) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// LinkApplication is called by the application server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(appID *ttnpb.ApplicationIdentifiers, stream ttnpb.AsNs_LinkApplicationServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueReplace is called by the application server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueuePush is called by the application server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueList is called by the application server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueClear is called by the application server to clear the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueClear(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
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
