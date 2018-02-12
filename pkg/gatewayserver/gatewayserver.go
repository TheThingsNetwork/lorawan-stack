// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import (
	"context"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/frequencyplans"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/gwpool"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	sendUplinkTimeout = 5 * time.Minute

	frequencyPlansCacheDuration = 2 * time.Hour
)

// GatewayServer implements the gateway server component.
//
// The gateway server exposes the Gs, GtwGs and NsGs services.
type GatewayServer struct {
	*component.Component

	gateways       gwpool.Pool
	frequencyPlans frequencyplans.Store
}

// New returns new *GatewayServer.
func New(c *component.Component, conf *Config) (*GatewayServer, error) {
	fpStore, err := conf.store()
	if err != nil {
		return nil, err
	}

	gs := &GatewayServer{
		Component: c,

		gateways:       gwpool.NewPool(log.FromContext(c.Context()), sendUplinkTimeout),
		frequencyPlans: fpStore,
	}
	c.RegisterGRPC(gs)
	return gs, nil
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

// GetGatewayObservations returns gateway information as observed by the gateway server.
func (gs *GatewayServer) GetGatewayObservations(ctx context.Context, id *ttnpb.GatewayIdentifier) (*ttnpb.GatewayObservations, error) {
	return gs.gateways.GetGatewayObservations(id)
}
