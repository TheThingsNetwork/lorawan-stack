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

package gatewayserver

import (
	"context"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth/rights"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/frequencyplans"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/pool"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	sendUplinkTimeout   = 5 * time.Minute
	rightsCacheDuration = time.Hour
)

// GatewayServer implements the gateway server component.
//
// The gateway server exposes the Gs, GtwGs and NsGs services.
type GatewayServer struct {
	*component.Component

	Config

	gateways       pool.Pool
	frequencyPlans frequencyplans.Store
}

type isConnector struct {
	*component.Component
}

func (c isConnector) Get(ctx context.Context) *grpc.ClientConn {
	peer := c.Component.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, nil, nil)
	if peer == nil {
		return nil
	}

	return peer.Conn()
}

// New returns new *GatewayServer.
func New(c *component.Component, conf Config) (*GatewayServer, error) {
	gs := &GatewayServer{
		Component: c,

		gateways:       pool.NewPool(c.Logger(), sendUplinkTimeout),
		frequencyPlans: conf.store(),

		Config: conf,
	}

	hook, err := rights.New(c.Context(), isConnector{Component: c}, rights.Config{TTL: rightsCacheDuration})
	if err != nil {
		return nil, err
	}
	hooks.RegisterUnaryHook("/ttn.v3.Gs/GetGatewayObservations", "claims-builder", hook.UnaryHook())

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
func (gs *GatewayServer) GetGatewayObservations(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error) {
	if !gs.DisableAuth && !ttnpb.IncludesRights(rights.FromContext(ctx), ttnpb.RIGHT_GATEWAY_STATUS) {
		return nil, ErrPermissionDenied.New(nil)
	}

	return gs.gateways.GetGatewayObservations(id)
}

func checkAuthorization(ctx context.Context, is ttnpb.IsGatewayClient, right ttnpb.Right) error {
	md := rpcmetadata.FromIncomingContext(ctx)

	if md.AuthType == "" || md.AuthValue == "" {
		return ErrUnauthorized.New(nil)
	}

	if md.AuthType != "Bearer" {
		return errors.Errorf("Expected authentication type to be `Bearer` but got `%s` instead", md.AuthType)
	}

	res, err := is.ListGatewayRights(ctx, &ttnpb.GatewayIdentifiers{GatewayID: md.ID}, grpc.PerRPCCredentials(&md))
	if err != nil {
		return errors.NewWithCause(err, "Could not fetch gateway rights for the credentials passed")
	}

	if !ttnpb.IncludesRights(res.Rights, right) {
		return ErrPermissionDenied.New(nil)
	}

	return nil
}
