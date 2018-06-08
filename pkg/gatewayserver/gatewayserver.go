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

// Package gatewayserver contains the structs and methods necessary to start a gRPC Gateway Server
package gatewayserver

import (
	"context"
	"net"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
	"google.golang.org/grpc"
)

// GatewayServer implements the Gateway Server component.
//
// The Gateway Server exposes the Gs, GtwGs and NsGs services.
type GatewayServer struct {
	*component.Component

	config Config

	connections   map[string]connection
	connectionsMu sync.Mutex
}

// New returns new *GatewayServer.
func New(c *component.Component, conf Config) (gs *GatewayServer, err error) {
	gs = &GatewayServer{
		Component: c,

		config: conf,

		connections: map[string]connection{},
	}

	if conf.UDPAddress != "" {
		var conn *net.UDPConn
		conn, err = gs.ListenUDP(conf.UDPAddress)
		if err != nil {
			return nil, errors.NewWithCause(err, "Could not open UDP socket")
		}

		ctx, cancel := context.WithCancel(c.Context())
		go gs.runUDPBridge(ctx, conn)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}

	rightsHook, err := c.RightsHook()
	if err != nil {
		return nil, err
	}
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Gs/GetGatewayObservations", rights.HookName, rightsHook.UnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", cluster.HookName, c.UnaryHook())

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

// Roles returns the roles that the Gateway Server fulfills.
func (gs *GatewayServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_GATEWAY_SERVER}
}

// GetGatewayObservations returns gateway information as observed by the Gateway Server.
func (gs *GatewayServer) GetGatewayObservations(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error) {
	if !gs.config.DisableAuth && !ttnpb.IncludesRights(rights.FromContext(ctx), ttnpb.RIGHT_GATEWAY_STATUS_READ) {
		return nil, common.ErrPermissionDenied.New(nil)
	}

	gtwID := id.GetGatewayID()
	if err := validate.ID(gtwID); err != nil {
		return nil, err
	}

	gs.connectionsMu.Lock()
	connection, ok := gs.connections[id.UniqueID(ctx)]
	gs.connectionsMu.Unlock()

	if !ok {
		return nil, ErrGatewayNotConnected.New(errors.Attributes{"gateway_id": id.GatewayID})
	}

	observations := connection.getObservations()

	return &observations, nil
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
		return common.ErrPermissionDenied.New(nil)
	}

	return nil
}
