// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package deviceclaimingserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"google.golang.org/grpc"
)

// EndDeviceClaimingUpstream provides upstream methods.
type EndDeviceClaimingUpstream interface {
	// SupportsJoinEUI returns whether this EndDeviceClaimingServer is configured to a Join Server that supports this Join EUI.
	SupportsJoinEUI(types.EUI64) bool
	// RegisterRoutes registers web routes.
	RegisterRoutes(server *web.Server)

	Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error)
	AuthorizeApplication(ctx context.Context, req *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error)
	UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error)
}

// DeviceClaimingServer is the Device Claiming Server.
type DeviceClaimingServer struct {
	*component.Component
	ctx context.Context

	config Config

	endDeviceClaimingUpstreams    map[string]EndDeviceClaimingUpstream
	gatewayClaimingServerUpstream ttnpb.GatewayClaimingServerServer

	grpc struct {
		endDeviceClaimingServer *endDeviceClaimingServer
		gatewayClaimingServer   *gatewayClaimingServer
	}
}

const (
	defaultType = "default"
)

// New returns a new Device Claiming component.
func New(c *component.Component, conf *Config, opts ...Option) (*DeviceClaimingServer, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "deviceclaimingserver")

	dcs := &DeviceClaimingServer{
		Component:                  c,
		ctx:                        ctx,
		config:                     *conf,
		endDeviceClaimingUpstreams: make(map[string]EndDeviceClaimingUpstream),
	}
	for _, opt := range opts {
		opt(dcs)
	}

	dcs.endDeviceClaimingUpstreams[defaultType] = noopEDCS{}
	dcs.gatewayClaimingServerUpstream = noopGCLS{}

	// TODO: Implement JS Clients (https://github.com/TheThingsNetwork/lorawan-stack/issues/4841#issuecomment-998294988)
	// Switch on the API type defined for a Join EUI and instantiate clients and add to `dcs.endDeviceClaimingUpstreams`.

	dcs.grpc.endDeviceClaimingServer = &endDeviceClaimingServer{
		DCS: dcs,
	}
	dcs.grpc.gatewayClaimingServer = &gatewayClaimingServer{
		DCS: dcs,
	}

	c.RegisterGRPC(dcs)
	c.RegisterWeb(dcs)
	return dcs, nil
}

// Option configures GatewayClaimingServer.
type Option func(*DeviceClaimingServer)

// Context returns the context of the Device Claiming Server.
func (dcs *DeviceClaimingServer) Context() context.Context {
	return dcs.ctx
}

// Roles returns the roles that the Device Claiming Server fulfills.
func (dcs *DeviceClaimingServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER}
}

// RegisterServices registers services provided by dcs at s.
func (dcs *DeviceClaimingServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEndDeviceClaimingServerServer(s, dcs.grpc.endDeviceClaimingServer)
	ttnpb.RegisterGatewayClaimingServerServer(s, dcs.grpc.gatewayClaimingServer)
}

// RegisterHandlers registers gRPC handlers.
func (dcs *DeviceClaimingServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEndDeviceClaimingServerHandler(dcs.Context(), s, conn)
	ttnpb.RegisterGatewayClaimingServerHandler(dcs.Context(), s, conn)
}

// RegisterRoutes implements web.Registerer. It registers the Device Claiming Server to the web server.
func (dcs *DeviceClaimingServer) RegisterRoutes(server *web.Server) {
	for _, edcs := range dcs.endDeviceClaimingUpstreams {
		edcs.RegisterRoutes(server)
	}
}
