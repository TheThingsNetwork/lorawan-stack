// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package applicationserver

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

// ApplicationServer implements the application server component.
//
// The application server exposes the As, DeviceRegistry and ApplicationDownlinkQueue services.
type ApplicationServer struct {
	*component.Component
	*deviceregistry.RegistryRPC
	registry deviceregistry.Interface
}

// Config represents the ApplicationServer configuration.
type Config struct {
	Registry deviceregistry.Interface
}

// New returns new *ApplicationServer.
func New(c *component.Component, conf *Config) *ApplicationServer {
	as := &ApplicationServer{
		Component:   c,
		RegistryRPC: deviceregistry.NewRPC(c, conf.Registry), // TODO: Add checks
		registry:    conf.Registry,
	}
	c.RegisterGRPC(as)
	return as
}

// Subscribe subscribes to application uplink messages for an EndDevice filter.
func (as *ApplicationServer) Subscribe(req *ttnpb.EndDeviceIdentifiers, stream ttnpb.As_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueReplace is called by the application server to completely replace the downlink queue for a device.
func (as *ApplicationServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueuePush is called by the application server to push a downlink to queue for a device.
func (as *ApplicationServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueList is called by the application server to get the current state of the downlink queue for a device.
func (as *ApplicationServer) DownlinkQueueList(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueClear is called by the application server to clear the downlink queue for a device.
func (as *ApplicationServer) DownlinkQueueClear(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	ttnpb.RegisterAsApplicationDownlinkQueueServer(s, as)
	ttnpb.RegisterAsDeviceRegistryServer(s, as)
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsDeviceRegistryHandler(as.Context(), s, conn)
}

// Roles returns the roles that the application server fulfils
func (as *ApplicationServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_APPLICATION_SERVER}
}
