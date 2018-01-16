// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GatewayServer implements the gateway server component.
//
// The gateway server exposes the Gs, GtwGs and NsGs services.
type GatewayServer struct {
	*component.Component
}

// Config represents the GatewayServer configuration.
type Config struct {
}

// New returns new *GatewayServer.
func New(c *component.Component, conf *Config) *GatewayServer {
	gs := &GatewayServer{
		Component: c,
	}
	c.RegisterGRPC(gs)
	return gs
}

// GetGatewayObservations returns gateway information as observed by the gateway server.
func (gs *GatewayServer) GetGatewayObservations(ctx context.Context, id *ttnpb.GatewayIdentifier) (*ttnpb.GatewayObservations, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// Link is called by the gateway agent to start exchanging traffic.
func (gs *GatewayServer) Link(stream ttnpb.GtwGs_LinkServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// GetFrequencyPlan is called by the gateway agent to retrieve a frequency plan.
func (gs *GatewayServer) GetFrequencyPlan(ctx context.Context, req *ttnpb.FrequencyPlanRequest) (*ttnpb.FrequencyPlan, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// ScheduleDownlink is called by the network server to schedule a downlink message.
func (gs *GatewayServer) ScheduleDownlink(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// RegisterServices registers services provided by gs at s.
func (gs *GatewayServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsServer(s, gs)
	ttnpb.RegisterGtwGsServer(s, gs)
	ttnpb.RegisterNsGsServer(s, gs)
}

// RegisterHandlers registers gRPC handlers.
func (gs *GatewayServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {}

// Roles returns the roles that the gateway server fulfils
func (gs *GatewayServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_GATEWAY_SERVER}
}
