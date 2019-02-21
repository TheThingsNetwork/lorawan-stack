// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstationlns"
	iogrpc "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/grpc"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

// connConcurrentUplinks is the number of goroutines per gateway connection to handle upstream messages.
var connConcurrentUplinks = 16

// GatewayServer implements the Gateway Server component.
//
// The Gateway Server exposes the Gs, GtwGs and NsGs services and MQTT and UDP frontends for gateways.
type GatewayServer struct {
	*component.Component
	io.Server

	config *Config

	connections sync.Map
}

var (
	errListenFrontend = errors.DefineFailedPrecondition(
		"listen_frontend",
		"failed to start frontend listener `{protocol}` on address `{address}`",
	)
	errNotConnected = errors.DefineNotFound("not_connected", "gateway `{gateway_uid}` not connected")
)

// New returns new *GatewayServer.
func New(c *component.Component, conf *Config) (gs *GatewayServer, err error) {
	gs = &GatewayServer{
		Component: c,
		config:    conf,
	}

	ctx, cancel := context.WithCancel(c.Context())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	for addr, fallbackFrequencyPlanID := range conf.UDP.Listeners {
		var conn *net.UDPConn
		conn, err = gs.ListenUDP(addr)
		if err != nil {
			return nil, errListenFrontend.WithCause(err).WithAttributes(
				"protocol", "udp",
				"address", addr,
			)
		}
		lisCtx := ctx
		if fallbackFrequencyPlanID != "" {
			lisCtx = frequencyplans.WithFallbackID(ctx, fallbackFrequencyPlanID)
		}
		udp.Start(lisCtx, gs, conn, conf.UDP.Config)
	}

	for _, version := range []struct {
		Format mqtt.Format
		Config MQTTConfig
	}{
		{
			Format: mqtt.Protobuf,
			Config: conf.MQTT,
		},
		{
			Format: mqtt.ProtobufV2,
			Config: conf.MQTTV2,
		},
	} {
		for _, lis := range []struct {
			Listen   string
			Protocol string
			Net      func(component.Listener) (net.Listener, error)
		}{
			{
				Listen:   version.Config.Listen,
				Protocol: "tcp",
				Net:      component.Listener.TCP,
			},
			{
				Listen:   version.Config.ListenTLS,
				Protocol: "tls",
				Net:      component.Listener.TLS,
			},
		} {
			if lis.Listen == "" {
				continue
			}
			var componentLis component.Listener
			var netLis net.Listener
			componentLis, err = gs.ListenTCP(lis.Listen)
			if err == nil {
				netLis, err = lis.Net(componentLis)
			}
			if err != nil {
				return nil, errListenFrontend.WithCause(err).WithAttributes(
					"protocol", lis.Protocol,
					"address", lis.Listen,
				)
			}
			mqtt.Start(ctx, gs, netLis, version.Format, lis.Protocol)
		}
	}

	bsCtx := gs.FillContext(ctx)
	if conf.BasicStation.FallbackFrequencyPlanID != "" {
		bsCtx = frequencyplans.WithFallbackID(bsCtx, conf.BasicStation.FallbackFrequencyPlanID)
	}
	c.RegisterWeb(basicstationlns.New(bsCtx, gs))

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(gs)
	return gs, nil
}

// RegisterServices registers services provided by gs at s.
func (gs *GatewayServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsServer(s, gs)
	ttnpb.RegisterNsGsServer(s, gs)
	ttnpb.RegisterGtwGsServer(s, iogrpc.New(gs))
}

// RegisterHandlers registers gRPC handlers.
func (gs *GatewayServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {}

// Roles returns the roles that the Gateway Server fulfills.
func (gs *GatewayServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_GATEWAY_SERVER}
}

// CustomContextFromIdentifier returns a derived context based on the given identifiers to use for the connection.
var CustomContextFromIdentifier func(context.Context, ttnpb.GatewayIdentifiers) (context.Context, error)

var (
	errEntityRegistryNotFound = errors.DefineNotFound(
		"entity_registry_not_found",
		"Entity Registry not found",
	)
	errGatewayEUINotRegistered = errors.DefineNotFound(
		"gateway_eui_not_registered",
		"gateway EUI `{eui}` is not registered",
	)
	errEmptyIdentifiers = errors.Define("empty_identifiers", "empty identifiers")
)

// FillGatewayContext fills the given context and identifiers.
func (gs *GatewayServer) FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error) {
	ctx = gs.FillContext(ctx)
	if ids.IsZero() {
		return nil, ttnpb.GatewayIdentifiers{}, errEmptyIdentifiers
	}
	if ids.GatewayID == "" {
		er := gs.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, nil)
		if er == nil {
			return nil, ttnpb.GatewayIdentifiers{}, errEntityRegistryNotFound
		}
		extIDs, err := ttnpb.NewGatewayRegistryClient(er.Conn()).GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			EUI: *ids.EUI,
		}, gs.WithClusterAuth())
		if err == nil {
			ids = *extIDs
		} else if errors.IsNotFound(err) {
			if gs.config.RequireRegisteredGateways {
				return nil, ttnpb.GatewayIdentifiers{}, errGatewayEUINotRegistered.WithAttributes("eui", *ids.EUI).WithCause(err)
			}
			ids.GatewayID = fmt.Sprintf("eui-%v", strings.ToLower(ids.EUI.String()))
		} else {
			return nil, ttnpb.GatewayIdentifiers{}, err
		}
	}
	if filler := CustomContextFromIdentifier; filler != nil {
		var err error
		if ctx, err = filler(ctx, ids); err != nil {
			return nil, ttnpb.GatewayIdentifiers{}, err
		}
	}
	return ctx, ids, nil
}

var (
	errGatewayNotRegistered = errors.DefineNotFound(
		"gateway_not_registered",
		"gateway `{gateway_uid}` is not registered",
	)
	errNoFallbackFrequencyPlan = errors.DefineNotFound(
		"no_fallback_frequency_plan",
		"gateway `{gateway_uid}` is not registered and no fallback frequency plan defined",
	)
)

// Connect connects a gateway by its identifiers to the Gateway Server, and returns a io.Connection for traffic and
// control.
func (gs *GatewayServer) Connect(ctx context.Context, frontend io.Frontend, ids ttnpb.GatewayIdentifiers) (*io.Connection, error) {
	if err := rights.RequireGateway(ctx, ids, ttnpb.RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"protocol", frontend.Protocol(),
		"gateway_uid", uid,
	))
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gateway_conn:%s", events.NewCorrelationID()))

	er := gs.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, nil)
	if er == nil {
		return nil, errEntityRegistryNotFound
	}
	var err error
	var callOpt grpc.CallOption
	callOpt, err = rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials())
	if errors.IsUnauthenticated(err) {
		callOpt = gs.WithClusterAuth()
	} else if err != nil {
		return nil, err
	}
	gtw, err := ttnpb.NewGatewayRegistryClient(er.Conn()).Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: ids,
		FieldMask: types.FieldMask{
			Paths: []string{
				"frequency_plan_id",
				"schedule_downlink_late",
				"enforce_duty_cycle",
				"downlink_path_constraint",
			},
		},
	}, callOpt)
	if errors.IsNotFound(err) {
		if gs.config.RequireRegisteredGateways {
			return nil, errGatewayNotRegistered.WithAttributes("gateway_uid", uid).WithCause(err)
		}
		fpID, ok := frequencyplans.FallbackIDFromContext(ctx)
		if !ok {
			return nil, errNoFallbackFrequencyPlan.WithAttributes("gateway_uid", uid)
		}
		logger.Warn("Connect unregistered gateway")
		gtw = &ttnpb.Gateway{
			GatewayIdentifiers:     ids,
			FrequencyPlanID:        fpID,
			EnforceDutyCycle:       true,
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
		}
	} else if err != nil {
		return nil, err
	}
	fp, err := gs.FrequencyPlans.GetByID(gtw.FrequencyPlanID)
	if err != nil {
		return nil, err
	}
	var scheduler *scheduling.Scheduler
	if !frontend.HasScheduler() {
		var err error
		scheduler, err = scheduling.NewScheduler(ctx, fp, gtw.EnforceDutyCycle)
		if err != nil {
			return nil, err
		}
	}

	conn := io.NewConnection(ctx, frontend.Protocol(), gtw, fp, scheduler)
	gs.connections.Store(uid, conn)
	registerGatewayConnect(ctx, ids)
	logger.Info("Connected")
	go gs.handleUpstream(conn)
	return conn, nil
}

// GetConnection returns the *io.Connection for the given gateway. If not found, this method returns nil, false.
func (gs *GatewayServer) GetConnection(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*io.Connection, bool) {
	conn, loaded := gs.connections.Load(unique.ID(ctx, ids))
	if !loaded {
		return nil, false
	}
	return conn.(*io.Connection), true
}

var (
	errNoNetworkServer = errors.DefineNotFound("no_network_server", "no Network Server found to handle message")
)

func (gs *GatewayServer) handleUpstream(conn *io.Connection) {
	ctx := conn.Context()
	logger := log.FromContext(ctx)
	defer func() {
		ids := conn.Gateway().GatewayIdentifiers
		gs.connections.Delete(unique.ID(ctx, ids))
		gs.UnclaimDownlink(ctx, ids)
		registerGatewayDisconnect(ctx, ids)
		logger.Info("Disconnected")
	}()
	wg := &sync.WaitGroup{}
	for i := 0; i < connConcurrentUplinks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-gs.Context().Done():
					return
				case <-ctx.Done():
					return
				case msg := <-conn.Up():
					ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:uplink:%s", events.NewCorrelationID()))
					msg.CorrelationIDs = append(msg.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
					registerReceiveUplink(ctx, conn.Gateway(), msg)
					drop := func(ids ttnpb.EndDeviceIdentifiers, err error) {
						logger := logger.WithError(err)
						if ids.JoinEUI != nil && !ids.JoinEUI.IsZero() {
							logger = logger.WithField("join_eui", *ids.JoinEUI)
						}
						if ids.DevEUI != nil && !ids.DevEUI.IsZero() {
							logger = logger.WithField("dev_eui", *ids.DevEUI)
						}
						if ids.DevAddr != nil && !ids.DevAddr.IsZero() {
							logger = logger.WithField("dev_addr", *ids.DevAddr)
						}
						logger.Debug("Drop message")
						registerDropUplink(ctx, ids, conn.Gateway(), msg, err)
					}
					ids, err := lorawan.GetUplinkMessageIdentifiers(msg)
					if err != nil {
						drop(ttnpb.EndDeviceIdentifiers{}, err)
						break
					}
					ns := gs.GetPeer(ctx, ttnpb.PeerInfo_NETWORK_SERVER, ids)
					if ns == nil {
						drop(ids, errNoNetworkServer)
						break
					}
					if _, err := ttnpb.NewGsNsClient(ns.Conn()).HandleUplink(ctx, msg, gs.WithClusterAuth()); err != nil {
						drop(ids, err)
						break
					}
					registerForwardUplink(ctx, ids, conn.Gateway(), msg, ns.Name())
				case status := <-conn.Status():
					ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:status:%s", events.NewCorrelationID()))
					registerReceiveStatus(ctx, conn.Gateway(), status)
				case ack := <-conn.TxAck():
					ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:tx_ack:%s", events.NewCorrelationID()))
					ack.CorrelationIDs = append(ack.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
					if ack.Result == ttnpb.TxAcknowledgment_SUCCESS {
						registerSuccessDownlink(ctx, conn.Gateway())
					} else {
						registerFailDownlink(ctx, conn.Gateway(), ack)
					}
				}
			}
		}()
	}
	wg.Wait()
}

// GetFrequencyPlan gets the specified frequency plan by the gateway identifiers.
func (gs *GatewayServer) GetFrequencyPlan(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*frequencyplans.FrequencyPlan, error) {
	er := gs.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, nil)
	if er == nil {
		return nil, errEntityRegistryNotFound
	}
	var err error
	var callOpt grpc.CallOption
	callOpt, err = rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials())
	if errors.IsUnauthenticated(err) {
		callOpt = gs.WithClusterAuth()
	} else if err != nil {
		return nil, err
	}
	gtw, err := ttnpb.NewGatewayRegistryClient(er.Conn()).Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: ids,
		FieldMask:          types.FieldMask{Paths: []string{"frequency_plan_id"}},
	}, callOpt)
	var fpID string
	if err == nil {
		fpID = gtw.FrequencyPlanID
	} else if errors.IsNotFound(err) {
		var ok bool
		fpID, ok = frequencyplans.FallbackIDFromContext(ctx)
		if !ok {
			return nil, err
		}
	} else {
		return nil, err
	}
	return gs.FrequencyPlans.GetByID(fpID)
}

// ClaimDownlink claims the downlink path for the given gateway.
func (gs *GatewayServer) ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.ClaimIDs(ctx, ids)
}

// UnclaimDownlink releases the claim of the downlink path for the given gateway.
func (gs *GatewayServer) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.UnclaimIDs(ctx, ids)
}
