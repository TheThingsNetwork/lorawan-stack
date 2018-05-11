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

package applicationserver

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
func New(c *component.Component, conf *Config) (*ApplicationServer, error) {
	as := &ApplicationServer{
		Component: c,
		registry:  conf.Registry,
	}

	registryRPC, err := deviceregistry.NewRPC(c, conf.Registry, deviceregistry.ForComponents(ttnpb.PeerInfo_APPLICATION_SERVER))
	if err != nil {
		return nil, err
	}
	as.RegistryRPC = registryRPC

	c.RegisterGRPC(as)
	return as, nil
}

// Subscribe subscribes to application uplink messages for an EndDevice filter.
func (as *ApplicationServer) Subscribe(req *ttnpb.ApplicationIdentifiers, stream ttnpb.As_SubscribeServer) error {
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
	ttnpb.RegisterAsApplicationDownlinkQueueHandler(as.Context(), s, conn)
	ttnpb.RegisterAsDeviceRegistryHandler(as.Context(), s, conn)
}

// Roles returns the roles that the application server fulfils
func (as *ApplicationServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_APPLICATION_SERVER}
}
