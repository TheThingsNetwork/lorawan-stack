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
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth/rights"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/scheduling"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	"github.com/TheThingsNetwork/ttn/pkg/toa"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/validate"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

type connection struct {
	observations   ttnpb.GatewayObservations
	observationsMu sync.RWMutex

	scheduler scheduling.Scheduler

	cancel   context.CancelFunc
	linkSend func(*ttnpb.GatewayDown) error
}

func (c *connection) Send(down *ttnpb.GatewayDown) (err error) {
	span := scheduling.Span{
		Start: scheduling.ConcentratorTime(down.DownlinkMessage.TxMetadata.Timestamp),
	}
	span.Duration, err = toa.Compute(down.DownlinkMessage.RawPayload, down.DownlinkMessage.Settings)
	if err != nil {
		return errors.NewWithCause(err, "Could not compute time-on-air of the downlink")
	}

	err = c.scheduler.ScheduleAt(span, down.DownlinkMessage.Settings.Frequency)
	if err != nil {
		return errors.NewWithCause(err, "Could not schedule downlink")
	}

	err = c.linkSend(down)
	if err != nil {
		return errors.NewWithCause(err, "Could not send downlink")
	}

	return nil
}

// GatewayServer implements the gateway server component.
//
// The gateway server exposes the Gs, GtwGs and NsGs services.
type GatewayServer struct {
	*component.Component

	config Config

	connections   map[string]*connection
	connectionsMu sync.Mutex
}

// New returns new *GatewayServer.
func New(c *component.Component, conf Config) (*GatewayServer, error) {
	gs := &GatewayServer{
		Component: c,

		config: conf,

		connections: map[string]*connection{},
	}

	hooks.RegisterUnaryHook("/ttn.v3.Gs/GetGatewayObservations", rights.HookName, c.RightsHook.UnaryHook())

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

func (conn *connection) addUpstreamObservations(up *ttnpb.GatewayUp) {
	now := time.Now().UTC()

	conn.observationsMu.Lock()

	if up.GatewayStatus != nil {
		conn.observations.LastStatus = up.GatewayStatus
		conn.observations.LastStatusReceivedAt = &now
	}

	if nbUplinks := len(up.UplinkMessages); nbUplinks > 0 {
		conn.observations.LastUplinkReceivedAt = &now
	}

	conn.observationsMu.Unlock()
}

func (conn *connection) addDownstreamObservations(down *ttnpb.GatewayDown) {
	now := time.Now().UTC()

	conn.observationsMu.Lock()
	conn.observations.LastDownlinkReceivedAt = &now
	conn.observationsMu.Unlock()
}

// GetGatewayObservations returns gateway information as observed by the gateway server.
func (gs *GatewayServer) GetGatewayObservations(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error) {
	if !gs.config.DisableAuth && !ttnpb.IncludesRights(rights.FromContext(ctx), ttnpb.RIGHT_GATEWAY_STATUS) {
		return nil, ErrPermissionDenied.New(nil)
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

	connection.observationsMu.RLock()
	observations := connection.observations
	connection.observationsMu.RUnlock()

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
		return ErrPermissionDenied.New(nil)
	}

	return nil
}
