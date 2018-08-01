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

	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
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
			return nil, errUDPSocket.WithCause(err)
		}
		c.Logger().WithField("address", conf.UDPAddress).Info("Listening for UDP connections")

		ctx, cancel := context.WithCancel(c.Context())
		go gs.runUDPEndpoint(ctx, conn)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}

	for _, mqttEndpoint := range []struct {
		address, protocol string
		create            func(component.Listener) (net.Listener, error)
	}{
		{
			address:  conf.MQTT.Listen,
			protocol: "tcp",
			create:   component.Listener.TCP,
		},
		{
			address:  conf.MQTT.ListenTLS,
			protocol: "tls",
			create:   component.Listener.TLS,
		},
	} {
		if mqttEndpoint.address == "" {
			continue
		}
		componentLis, err := c.ListenTCP(mqttEndpoint.address)
		if err != nil {
			return nil, err
		}
		lis, err := mqttEndpoint.create(componentLis)
		if err != nil {
			return nil, err
		}
		mqttLis := mqttnet.NewListener(lis, mqttEndpoint.protocol)
		go gs.runMQTTEndpoint(mqttLis)
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

// Roles returns the roles that the Gateway Server fulfills.
func (gs *GatewayServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_GATEWAY_SERVER}
}

// GetGatewayObservations returns gateway information as observed by the Gateway Server.
func (gs *GatewayServer) GetGatewayObservations(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error) {
	if err := rights.RequireGateway(ctx, *id, ttnpb.RIGHT_GATEWAY_STATUS_READ); err != nil {
		return nil, err
	}

	gtwID := id.GetGatewayID()
	if err := validate.ID(gtwID); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, id)
	gs.connectionsMu.Lock()
	connection, ok := gs.connections[uid]
	gs.connectionsMu.Unlock()

	if !ok {
		return nil, errGatewayNotConnected.WithAttributes("gateway_uid", uid)
	}

	observations := connection.getObservations()

	return &observations, nil
}

func (gs *GatewayServer) getIdentityServer() (ttnpb.IsGatewayClient, error) {
	peer := gs.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, nil, nil)
	if peer == nil {
		return nil, errNoIdentityServerFound
	}
	isConn := peer.Conn()
	if isConn == nil {
		return nil, errNoReadyConnectionToIdentityServer
	}

	return ttnpb.NewIsGatewayClient(isConn), nil
}

func (gs *GatewayServer) getGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*ttnpb.Gateway, error) {
	is, err := gs.getIdentityServer()
	if err != nil {
		return nil, err
	}
	return is.GetGateway(ctx, id)
}
