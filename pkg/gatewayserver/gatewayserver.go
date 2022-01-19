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
	stdio "io"
	stdlog "log"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	iogrpc "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/grpc"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mqtt"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/ns"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/packetbroker"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GatewayServer implements the Gateway Server component.
//
// The Gateway Server exposes the Gs, GtwGs and NsGs services and MQTT and UDP frontends for gateways.
type GatewayServer struct {
	*component.Component
	ctx context.Context

	config *Config

	requireRegisteredGateways bool
	forward                   map[string][]types.DevAddrPrefix

	entityRegistry EntityRegistry

	upstreamHandlers map[string]upstream.Handler

	connections sync.Map // string to connectionEntry

	statsRegistry                     GatewayConnectionStatsRegistry
	updateConnectionStatsDebounceTime time.Duration
}

// Option configures GatewayServer.
type Option func(*GatewayServer)

// WithRegistry overrides the registry.
func WithRegistry(registry EntityRegistry) Option {
	return func(gs *GatewayServer) {
		gs.entityRegistry = registry
	}
}

// Context returns the context of the Gateway Server.
func (gs *GatewayServer) Context() context.Context {
	return gs.ctx
}

var (
	errListenFrontend = errors.DefineFailedPrecondition(
		"listen_frontend",
		"failed to start frontend listener `{protocol}` on address `{address}`",
	)
	errNotConnected        = errors.DefineNotFound("not_connected", "gateway `{gateway_uid}` not connected")
	errSetupUpstream       = errors.DefineFailedPrecondition("upstream", "failed to setup upstream `{name}`")
	errInvalidUpstreamName = errors.DefineInvalidArgument("invalid_upstream_name", "upstream `{name}` is invalid")

	modelAttribute    = "model"
	firmwareAttribute = "firmware"
)

// New returns new *GatewayServer.
func New(c *component.Component, conf *Config, opts ...Option) (gs *GatewayServer, err error) {
	forward, err := conf.ForwardDevAddrPrefixes()
	if err != nil {
		return nil, err
	}
	if len(forward) == 0 {
		forward[""] = []types.DevAddrPrefix{{}}
	}

	ctx := log.NewContextWithField(c.Context(), "namespace", "gatewayserver")

	gs = &GatewayServer{
		Component:                         c,
		ctx:                               ctx,
		config:                            conf,
		requireRegisteredGateways:         conf.RequireRegisteredGateways,
		forward:                           forward,
		upstreamHandlers:                  make(map[string]upstream.Handler),
		statsRegistry:                     conf.Stats,
		updateConnectionStatsDebounceTime: conf.UpdateConnectionStatsDebounceTime,
		entityRegistry:                    NewIS(c),
	}
	for _, opt := range opts {
		opt(gs)
	}

	// Setup forwarding table.
	for name, prefix := range gs.forward {
		if len(prefix) == 0 {
			continue
		}
		if name == "" {
			name = "cluster"
		}
		var handler upstream.Handler
		switch name {
		case "cluster":
			handler = ns.NewHandler(gs.Context(), c, c, prefix)
		case "packetbroker":
			handler = packetbroker.NewHandler(gs.Context(), packetbroker.Config{
				GatewayRegistry: gs.entityRegistry,
				Cluster:         c,
				DevAddrPrefixes: prefix,
				UpdateInterval:  conf.PacketBroker.UpdateGatewayInterval,
				UpdateJitter:    conf.PacketBroker.UpdateGatewayJitter,
				OnlineTTLMargin: conf.PacketBroker.OnlineTTLMargin,
			})
		default:
			return nil, errInvalidUpstreamName.WithAttributes("name", name)
		}
		if err := handler.Setup(gs.Context()); err != nil {
			return nil, errSetupUpstream.WithCause(err).WithAttributes("name", name)
		}
		gs.upstreamHandlers[name] = handler
	}

	// Register gRPC services.
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("gatewayserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", cluster.HookName, c.ClusterAuthUnaryHook())
	c.RegisterGRPC(gs)

	// Start UDP listeners.
	for addr, fallbackFrequencyPlanID := range conf.UDP.Listeners {
		addr := addr
		fallbackFrequencyPlanID := fallbackFrequencyPlanID
		gs.RegisterTask(&task.Config{
			Context: gs.Context(),
			ID:      fmt.Sprintf("serve_udp/%s", addr),
			Func: func(ctx context.Context) error {
				var conn *net.UDPConn
				conn, err = gs.ListenUDP(addr)
				if err != nil {
					return errListenFrontend.WithCause(err).WithAttributes(
						"protocol", "udp",
						"address", addr,
					)
				}
				defer conn.Close()
				lisCtx := ctx
				if fallbackFrequencyPlanID != "" {
					lisCtx = frequencyplans.WithFallbackID(ctx, fallbackFrequencyPlanID)
				}
				return udp.Serve(lisCtx, gs, conn, conf.UDP.Config)
			},
			Restart: task.RestartOnFailure,
			Backoff: task.DefaultBackoffConfig,
		})
	}

	// Start MQTT listeners.
	for _, version := range []struct {
		Format mqtt.Format
		Config config.MQTT
	}{
		{
			Format: mqtt.NewProtobuf(gs.ctx),
			Config: conf.MQTT,
		},
		{
			Format: mqtt.NewProtobufV2(gs.ctx),
			Config: conf.MQTTV2,
		},
	} {
		for _, endpoint := range []component.Endpoint{
			component.NewTCPEndpoint(version.Config.Listen, "MQTT"),
			component.NewTLSEndpoint(version.Config.ListenTLS, "MQTT"),
		} {
			version := version
			endpoint := endpoint
			if endpoint.Address() == "" {
				continue
			}
			gs.RegisterTask(&task.Config{
				Context: gs.Context(),
				ID:      fmt.Sprintf("serve_mqtt/%s", endpoint.Address()),
				Func: func(ctx context.Context) error {
					l, err := gs.ListenTCP(endpoint.Address())
					if err != nil {
						return errListenFrontend.WithCause(err).WithAttributes(
							"address", endpoint.Address(),
							"protocol", endpoint.Protocol(),
						)
					}
					lis, err := endpoint.Listen(l)
					if err != nil {
						return errListenFrontend.WithCause(err).WithAttributes(
							"address", endpoint.Address(),
							"protocol", endpoint.Protocol(),
						)
					}
					defer lis.Close()
					return mqtt.Serve(ctx, gs, lis, version.Format, endpoint.Protocol())
				},
				Restart: task.RestartOnFailure,
				Backoff: task.DefaultBackoffConfig,
			})
		}
	}

	// Start Web Socket listeners.
	type listenerConfig struct {
		fallbackFreqPlanID string
		listen             string
		listenTLS          string
		frontend           ws.Config
	}
	for _, version := range []struct {
		Name           string
		Formatter      ws.Formatter
		listenerConfig listenerConfig
	}{
		{
			Name:      "basicstation",
			Formatter: lbslns.NewFormatter(conf.BasicStation.MaxValidRoundTripDelay),
			listenerConfig: listenerConfig{
				fallbackFreqPlanID: conf.BasicStation.FallbackFrequencyPlanID,
				listen:             conf.BasicStation.Listen,
				listenTLS:          conf.BasicStation.ListenTLS,
				frontend:           conf.BasicStation.Config,
			},
		},
	} {
		ctx := gs.Context()
		if version.listenerConfig.fallbackFreqPlanID != "" {
			ctx = frequencyplans.WithFallbackID(ctx, version.listenerConfig.fallbackFreqPlanID)
		}
		web, err := ws.New(ctx, gs, version.Formatter, version.listenerConfig.frontend)
		if err != nil {
			return nil, err
		}
		for _, endpoint := range []component.Endpoint{
			component.NewTCPEndpoint(version.listenerConfig.listen, version.Name),
			component.NewTLSEndpoint(version.listenerConfig.listenTLS, version.Name, tlsconfig.WithNextProtos("h2", "http/1.1")),
		} {
			endpoint := endpoint
			if endpoint.Address() == "" {
				continue
			}
			gs.RegisterTask(&task.Config{
				Context: gs.Context(),
				ID:      fmt.Sprintf("serve_%s/%s", version.Name, endpoint.Address()),
				Func: func(ctx context.Context) error {
					l, err := gs.ListenTCP(endpoint.Address())
					if err != nil {
						return errListenFrontend.WithCause(err).WithAttributes(
							"address", endpoint.Address(),
							"protocol", endpoint.Protocol(),
						)
					}
					lis, err := endpoint.Listen(l)
					if err != nil {
						return errListenFrontend.WithCause(err).WithAttributes(
							"address", endpoint.Address(),
							"protocol", endpoint.Protocol(),
						)
					}
					defer lis.Close()

					srv := http.Server{
						Handler:           web,
						ReadTimeout:       120 * time.Second,
						ReadHeaderTimeout: 5 * time.Second,
						ErrorLog:          stdlog.New(stdio.Discard, "", 0),
					}
					go func() {
						<-ctx.Done()
						srv.Close()
					}()
					return srv.Serve(lis)
				},
				Restart: task.RestartOnFailure,
				Backoff: task.DefaultBackoffConfig,
			})
		}
	}
	return gs, nil
}

// RegisterServices registers services provided by gs at s.
func (gs *GatewayServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsServer(s, gs)
	ttnpb.RegisterNsGsServer(s, gs)
	ttnpb.RegisterGtwGsServer(s, iogrpc.New(gs,
		iogrpc.WithMQTTConfigProvider(
			config.MQTTConfigProviderFunc(func(ctx context.Context) (*config.MQTT, error) {
				config, err := gs.GetConfig(ctx)
				if err != nil {
					return nil, err
				}
				return &config.MQTT, nil
			})),
		iogrpc.WithMQTTV2ConfigProvider(
			config.MQTTConfigProviderFunc(func(ctx context.Context) (*config.MQTT, error) {
				config, err := gs.GetConfig(ctx)
				if err != nil {
					return nil, err
				}
				return &config.MQTTV2, nil
			})),
	))
}

// RegisterHandlers registers gRPC handlers.
func (gs *GatewayServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterGsHandler(gs.Context(), s, conn)
	ttnpb.RegisterGtwGsHandler(gs.Context(), s, conn)
}

// Roles returns the roles that the Gateway Server fulfills.
func (gs *GatewayServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_GATEWAY_SERVER}
}

var (
	errGatewayEUINotRegistered = errors.DefineNotFound(
		"gateway_eui_not_registered",
		"gateway EUI `{eui}` is not registered",
	)
	errEmptyIdentifiers = errors.Define("empty_identifiers", "empty identifiers")
)

// FillGatewayContext fills the given context and identifiers.
// This method should only be used for request contexts.
func (gs *GatewayServer) FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error) {
	ctx = gs.FillContext(ctx)
	if ids.IsZero() || ids.Eui != nil && ids.Eui.IsZero() {
		return nil, ttnpb.GatewayIdentifiers{}, errEmptyIdentifiers.New()
	}
	if ids.GatewayId == "" {
		extIDs, err := gs.entityRegistry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			Eui: ids.Eui,
		})
		if err == nil {
			ids = *extIDs
		} else if errors.IsNotFound(err) {
			if gs.requireRegisteredGateways {
				return nil, ttnpb.GatewayIdentifiers{}, errGatewayEUINotRegistered.WithAttributes("eui", *ids.Eui).WithCause(err)
			}
			ids.GatewayId = fmt.Sprintf("eui-%v", strings.ToLower(ids.Eui.String()))
		} else {
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
	errUnauthenticatedGatewayConnection = errors.DefineUnauthenticated(
		"unauthenticated_gateway_connection",
		"gateway requires an authenticated connection",
	)
	errNewConnection = errors.DefineAborted(
		"new_connection",
		"new connection from same gateway",
	)
)

type connectionEntry struct {
	*io.Connection
	tasksDone *sync.WaitGroup
}

// Connect connects a gateway by its identifiers to the Gateway Server, and returns a io.Connection for traffic and
// control.
func (gs *GatewayServer) Connect(ctx context.Context, frontend io.Frontend, ids ttnpb.GatewayIdentifiers) (*io.Connection, error) {
	if err := gs.entityRegistry.AssertGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"protocol", frontend.Protocol(),
		"gateway_uid", uid,
	))
	ctx = log.NewContext(ctx, logger)
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:conn:%s", events.NewCorrelationID()))

	var isAuthenticated bool
	if _, err := rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials()); err == nil {
		isAuthenticated = true
	}
	gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: &ids,
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{
				"antennas",
				"attributes",
				"disable_packet_broker_forwarding",
				"downlink_path_constraint",
				"enforce_duty_cycle",
				"frequency_plan_id",
				"frequency_plan_ids",
				"location_public",
				"require_authenticated_connection",
				"schedule_anytime_delay",
				"schedule_downlink_late",
				"status_public",
				"update_location_from_status",
			},
		},
	})
	if errors.IsNotFound(err) {
		if gs.requireRegisteredGateways {
			return nil, errGatewayNotRegistered.WithAttributes("gateway_uid", uid).WithCause(err)
		}
		fpID, ok := frequencyplans.FallbackIDFromContext(ctx)
		if !ok {
			return nil, errNoFallbackFrequencyPlan.WithAttributes("gateway_uid", uid)
		}
		logger.Warn("Connect unregistered gateway")
		gtw = &ttnpb.Gateway{
			Ids:                    &ids,
			FrequencyPlanId:        fpID,
			FrequencyPlanIds:       []string{fpID},
			EnforceDutyCycle:       true,
			DownlinkPathConstraint: ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NONE,
			Antennas:               []*ttnpb.GatewayAntenna{},
		}
	} else if err != nil {
		return nil, err
	}
	if gtw.RequireAuthenticatedConnection && !isAuthenticated {
		return nil, errUnauthenticatedGatewayConnection.New()
	}

	ids = *gtw.GetIds()

	fps, err := gs.FrequencyPlansStore(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := io.NewConnection(ctx, frontend, gtw, fps, gtw.EnforceDutyCycle, ttnpb.StdDuration(gtw.ScheduleAnytimeDelay))
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	// The tasks will always start once the entry is stored.
	// As such, we must ensure any new connection waits for
	// all of the upstream tasks to finish.
	wg.Add(len(gs.upstreamHandlers))
	connEntry := connectionEntry{
		Connection: conn,
		tasksDone:  wg,
	}
	for existing, exists := gs.connections.LoadOrStore(uid, connEntry); exists; existing, exists = gs.connections.LoadOrStore(uid, connEntry) {
		existingConnEntry := existing.(connectionEntry)
		logger.Warn("Disconnect existing connection")
		existingConnEntry.Disconnect(errNewConnection.New())
		existingConnEntry.tasksDone.Wait()
	}

	registerGatewayConnect(ctx, ids, frontend.Protocol())
	logger.Info("Connected")

	gs.startDisconnectOnChangeTask(connEntry)
	gs.startHandleUpstreamTask(connEntry)
	gs.startUpdateConnStatsTask(connEntry)
	gs.startHandleLocationUpdatesTask(connEntry)
	gs.startHandleVersionUpdatesTask(connEntry)

	for name, handler := range gs.upstreamHandlers {
		connCtx := log.NewContextWithField(conn.Context(), "upstream_handler", name)
		handler := handler
		gs.StartTask(&task.Config{
			Context: connCtx,
			ID:      fmt.Sprintf("%s_connect_gateway_%s", name, ids.GatewayId),
			Func: func(ctx context.Context) error {
				return handler.ConnectGateway(ctx, ids, conn)
			},
			Done:    wg.Done,
			Restart: task.RestartOnFailure,
			Backoff: task.DialBackoffConfig,
		})
	}
	return conn, nil
}

// GetConnection returns the *io.Connection for the given gateway. If not found, this method returns nil, false.
func (gs *GatewayServer) GetConnection(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*io.Connection, bool) {
	entry, loaded := gs.connections.Load(unique.ID(ctx, ids))
	if !loaded {
		return nil, false
	}
	return entry.(connectionEntry).Connection, true
}

func requireDisconnect(connected, current *ttnpb.Gateway) bool {
	if !sameAntennaLocations(connected.GetAntennas(), current.GetAntennas()) {
		// Gateway Server may update the location from status messages. If the locations aren't the same, but if the new
		// location is a GPS location, do not disconnect the gateway. This is to avoid that updating the location from a
		// gateway status message results in disconnecting the gateway.
		if len(current.Antennas) > 0 && current.Antennas[0].Location != nil && current.Antennas[0].Location.Source != ttnpb.SOURCE_GPS {
			return true
		}
	}
	if connected.DownlinkPathConstraint != current.DownlinkPathConstraint ||
		connected.DisablePacketBrokerForwarding != current.DisablePacketBrokerForwarding ||
		connected.EnforceDutyCycle != current.EnforceDutyCycle ||
		connected.LocationPublic != current.LocationPublic ||
		connected.RequireAuthenticatedConnection != current.RequireAuthenticatedConnection ||
		ttnpb.StdDurationOrZero(connected.ScheduleAnytimeDelay) != ttnpb.StdDurationOrZero(current.ScheduleAnytimeDelay) ||
		connected.ScheduleDownlinkLate != current.ScheduleDownlinkLate ||
		connected.StatusPublic != current.StatusPublic ||
		connected.UpdateLocationFromStatus != current.UpdateLocationFromStatus ||
		connected.FrequencyPlanId != current.FrequencyPlanId ||
		len(connected.FrequencyPlanIds) != len(current.FrequencyPlanIds) {
		return true
	}
	for i := range connected.FrequencyPlanIds {
		if connected.FrequencyPlanIds[i] != current.FrequencyPlanIds[i] {
			return true
		}
	}
	return false
}

var errGatewayChanged = errors.Define("gateway_changed", "gateway changed in registry")

func (gs *GatewayServer) startDisconnectOnChangeTask(conn connectionEntry) {
	conn.tasksDone.Add(1)
	gs.StartTask(&task.Config{
		Context: conn.Context(),
		ID:      fmt.Sprintf("disconnect_on_change_%s", unique.ID(conn.Context(), conn.Gateway().GetIds())),
		Func: func(ctx context.Context) error {
			d := random.Jitter(gs.config.FetchGatewayInterval, gs.config.FetchGatewayJitter)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d):
			}

			gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: conn.Gateway().GetIds(),
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"antennas",
						"disable_packet_broker_forwarding",
						"downlink_path_constraint",
						"enforce_duty_cycle",
						"frequency_plan_id",
						"frequency_plan_ids",
						"location_public",
						"require_authenticated_connection",
						"schedule_anytime_delay",
						"schedule_downlink_late",
						"status_public",
						"update_location_from_status",
					},
				},
			})
			if err != nil {
				if errors.IsUnauthenticated(err) || errors.IsPermissionDenied(err) {
					// Since there is an active connection, the `Get` request will not return a `NotFound` error as the gateway existed during the connect, since the rights assertion fails first.
					// Instead,
					// 1. If the gateway is connected with an API key and is deleted, the IS returns an `Unauthenticated`, since the API Key is also deleted.
					// 2. If the gateway is connected without an API key (UDP, LBS in unauthenticated mode) and is deleted the IS returns an `PermissionDenied` as there are no rights for these IDs.
					log.FromContext(ctx).WithError(err).Debug("Gateway was deleted and/or the API key used to link the gateway was invalidated")
					conn.Disconnect(err)
				} else {
					log.FromContext(ctx).WithError(err).Warn("Failed to get gateway")
				}
				return err
			}
			if requireDisconnect(conn.Gateway(), gtw) {
				log.FromContext(ctx).Info("Gateway changed in registry, disconnect")
				conn.Disconnect(errGatewayChanged.New())
			}

			return nil
		},
		Done:    conn.tasksDone.Done,
		Restart: task.RestartAlways,
		Backoff: task.DialBackoffConfig,
	})
}

func (gs *GatewayServer) startHandleUpstreamTask(conn connectionEntry) {
	conn.tasksDone.Add(1)
	gs.StartTask(&task.Config{
		Context: conn.Context(),
		ID:      fmt.Sprintf("handle_upstream_%s", unique.ID(conn.Context(), conn.Gateway().GetIds())),
		Func: func(ctx context.Context) error {
			gs.handleUpstream(ctx, conn)
			return nil
		},
		Done:    conn.tasksDone.Done,
		Restart: task.RestartNever,
		Backoff: task.DialBackoffConfig,
	})
}

func (gs *GatewayServer) startUpdateConnStatsTask(conn connectionEntry) {
	if gs.statsRegistry == nil {
		return
	}
	conn.tasksDone.Add(1)
	gs.StartTask(&task.Config{
		Context: conn.Context(),
		ID:      fmt.Sprintf("update_connection_stats_%s", unique.ID(conn.Context(), conn.Gateway().GetIds())),
		Func: func(ctx context.Context) error {
			gs.updateConnStats(ctx, conn)
			return nil
		},
		Done:    conn.tasksDone.Done,
		Restart: task.RestartNever,
		Backoff: task.DialBackoffConfig,
	})
}

func (gs *GatewayServer) startHandleLocationUpdatesTask(conn connectionEntry) {
	if !conn.Gateway().GetUpdateLocationFromStatus() {
		return
	}
	conn.tasksDone.Add(1)
	gs.StartTask(&task.Config{
		Context: conn.Context(),
		ID:      fmt.Sprintf("handle_location_updates_%s", unique.ID(conn.Context(), conn.Gateway().GetIds())),
		Func: func(ctx context.Context) error {
			gs.handleLocationUpdates(ctx, conn)
			return nil
		},
		Done:    conn.tasksDone.Done,
		Restart: task.RestartNever,
		Backoff: task.DialBackoffConfig,
	})
}

func (gs *GatewayServer) startHandleVersionUpdatesTask(conn connectionEntry) {
	conn.tasksDone.Add(1)
	gs.StartTask(&task.Config{
		Context: conn.Context(),
		ID:      fmt.Sprintf("handle_version_updates_%s", unique.ID(conn.Context(), conn.Gateway().GetIds())),
		Func: func(ctx context.Context) error {
			gs.handleVersionInfoUpdates(ctx, conn)
			return nil
		},
		Done:    conn.tasksDone.Done,
		Restart: task.RestartNever,
		Backoff: task.DialBackoffConfig,
	})
}

var errHostHandle = errors.Define("host_handle", "host `{host}` failed to handle message")

type upstreamHost struct {
	name          string
	handler       upstream.Handler
	pool          workerpool.WorkerPool
	gtw           *ttnpb.Gateway
	correlationID string
}

func (host *upstreamHost) handlePacket(ctx context.Context, item interface{}) {
	ctx = events.ContextWithCorrelationID(ctx, host.correlationID)
	logger := log.FromContext(ctx)
	gtw := host.gtw
	switch msg := item.(type) {
	case *ttnpb.GatewayUplinkMessage:
		up := *msg.Message
		msg = &ttnpb.GatewayUplinkMessage{
			BandId:  msg.BandId,
			Message: &up,
		}
		msg.Message.CorrelationIds = append(make([]string, 0, len(msg.Message.CorrelationIds)+1), msg.Message.CorrelationIds...)
		msg.Message.CorrelationIds = append(msg.Message.CorrelationIds, host.correlationID)
		drop := func(ids *ttnpb.EndDeviceIdentifiers, err error) {
			logger := logger.WithError(err)
			if ids.JoinEui != nil {
				logger = logger.WithField("join_eui", *ids.JoinEui)
			}
			if ids.DevEui != nil && !ids.DevEui.IsZero() {
				logger = logger.WithField("dev_eui", *ids.DevEui)
			}
			if ids.DevAddr != nil && !ids.DevAddr.IsZero() {
				logger = logger.WithField("dev_addr", *ids.DevAddr)
			}
			logger.Debug("Drop message")
			registerDropUplink(ctx, gtw, msg, host.name, err)
		}
		ids := up.Payload.EndDeviceIdentifiers()
		var pass bool
		switch {
		case ids.DevAddr != nil:
			for _, prefix := range host.handler.DevAddrPrefixes() {
				if ids.DevAddr.HasPrefix(prefix) {
					pass = true
					break
				}
			}
		default:
			pass = true
		}
		if !pass {
			break
		}
		switch err := host.handler.HandleUplink(ctx, *gtw.Ids, ids, msg); codes.Code(errors.Code(err)) {
		case codes.Canceled, codes.DeadlineExceeded,
			codes.Unknown, codes.Internal,
			codes.Unimplemented, codes.Unavailable:
			drop(ids, errHostHandle.WithCause(err).WithAttributes("host", host.name))
		default:
			registerForwardUplink(ctx, gtw, msg.Message, host.name)
		}
	case *ttnpb.GatewayStatus:
		if err := host.handler.HandleStatus(ctx, *gtw.Ids, msg); err != nil {
			registerDropStatus(ctx, gtw, msg, host.name, err)
		} else {
			registerForwardStatus(ctx, gtw, msg, host.name)
		}
	case *ttnpb.TxAcknowledgment:
		if err := host.handler.HandleTxAck(ctx, *gtw.Ids, msg); err != nil {
			registerDropTxAck(ctx, gtw, msg, host.name, err)
		} else {
			registerForwardTxAck(ctx, gtw, msg, host.name)
		}
	}
}

func (gs *GatewayServer) handleUpstream(ctx context.Context, conn connectionEntry) {
	var (
		gtw      = conn.Gateway()
		protocol = conn.Frontend().Protocol()
		logger   = log.FromContext(ctx)
	)
	defer func() {
		gs.connections.Delete(unique.ID(ctx, gtw.GetIds()))
		registerGatewayDisconnect(ctx, *gtw.GetIds(), protocol, ctx.Err())
		logger.Info("Disconnected")
	}()

	hosts := make([]*upstreamHost, 0, len(gs.upstreamHandlers))
	for name, handler := range gs.upstreamHandlers {
		if name == "packetbroker" && gtw.DisablePacketBrokerForwarding {
			continue
		}
		host := &upstreamHost{
			name:          name,
			handler:       handler,
			gtw:           gtw,
			correlationID: fmt.Sprintf("gs:up:host:%s", events.NewCorrelationID()),
		}
		wp := workerpool.NewWorkerPool(workerpool.Config{
			Component:  gs,
			Context:    ctx,
			Name:       fmt.Sprintf("upstream_handlers_%v", name),
			Handler:    host.handlePacket,
			MinWorkers: -1,
			MaxWorkers: 32,
			QueueSize:  -1,
		})
		defer wp.Wait()
		host.pool = wp
		hosts = append(hosts, host)
	}

	for {
		var (
			ctx = ctx
			val interface{}
		)
		select {
		case <-ctx.Done():
			return
		case msg := <-conn.Up():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:uplink:%s", events.NewCorrelationID()))
			msg.Message.CorrelationIds = append(msg.Message.CorrelationIds, events.CorrelationIDsFromContext(ctx)...)
			if msg.Message.Payload == nil {
				msg.Message.Payload = &ttnpb.Message{}
				if err := lorawan.UnmarshalMessage(msg.Message.RawPayload, msg.Message.Payload); err != nil {
					registerDropUplink(ctx, gtw, msg, "validation", err)
					continue
				}
			}
			val = msg
			registerReceiveUplink(ctx, gtw, msg.Message, protocol)
		case msg := <-conn.Status():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:status:%s", events.NewCorrelationID()))
			val = msg
			registerReceiveStatus(ctx, gtw, msg, protocol)
		case msg := <-conn.TxAck():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:tx_ack:%s", events.NewCorrelationID()))
			if d := msg.DownlinkMessage; d != nil {
				d.CorrelationIds = append(d.CorrelationIds, events.CorrelationIDsFromContext(ctx)...)
			}
			if msg.Result == ttnpb.TxAcknowledgment_SUCCESS {
				registerSuccessDownlink(ctx, gtw, protocol)
			} else {
				registerFailDownlink(ctx, gtw, msg, protocol)
			}
			val = msg
			registerReceiveTxAck(ctx, gtw, msg, protocol)
		}
		for _, host := range hosts {
			err := host.pool.Publish(ctx, val)
			if err == nil {
				continue
			}
			logger.WithField("name", host.name).WithError(err).Warn("Upstream handler publish failed")
			switch msg := val.(type) {
			case *ttnpb.GatewayUplinkMessage:
				registerDropUplink(ctx, gtw, msg, host.name, err)
			case *ttnpb.GatewayStatus:
				registerDropStatus(ctx, gtw, msg, host.name, err)
			case *ttnpb.TxAcknowledgment:
				registerDropTxAck(ctx, gtw, msg, host.name, err)
			default:
				panic("unreachable")
			}
		}
	}
}

func (gs *GatewayServer) updateConnStats(ctx context.Context, conn connectionEntry) {
	decoupledCtx := gs.FromRequestContext(ctx)
	logger := log.FromContext(ctx)

	ids := conn.Connection.Gateway().GetIds()
	connectTime := conn.Connection.ConnectTime()
	stats := &ttnpb.GatewayConnectionStats{
		ConnectedAt: ttnpb.ProtoTimePtr(connectTime),
		Protocol:    conn.Connection.Frontend().Protocol(),
	}

	// Initial update, so that the gateway appears connected.
	if err := gs.statsRegistry.Set(decoupledCtx, *ids, stats, ttnpb.GatewayConnectionStatsFieldPathsTopLevel, 0); err != nil {
		logger.WithError(err).Warn("Failed to initialize connection stats")
	}

	defer func() {
		logger.Debug("Delete connection stats")
		stats.ConnectedAt = nil
		stats.DisconnectedAt = ttnpb.ProtoTimePtr(time.Now())

		if err := gs.statsRegistry.Set(
			decoupledCtx, *ids, stats,
			[]string{"connected_at", "disconnected_at"},
			gs.config.ConnectionStatsDisconnectTTL,
		); err != nil {
			logger.WithError(err).Warn("Failed to clear connection stats")
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.StatsChanged():
		}
		stats, paths := conn.Stats()
		if err := gs.statsRegistry.Set(decoupledCtx, *ids, stats, paths, 0); err != nil {
			logger.WithError(err).Warn("Failed to update connection stats")
		}
		timeout := time.After(gs.updateConnectionStatsDebounceTime)
		select {
		case <-ctx.Done():
			return
		case <-timeout:
		}
	}
}

const (
	allowedLocationDelta = 0.00001
)

func sameLocation(a, b ttnpb.Location) bool {
	return a.Altitude == b.Altitude && a.Accuracy == b.Accuracy &&
		math.Abs(a.Latitude-b.Latitude) <= allowedLocationDelta &&
		math.Abs(a.Longitude-b.Longitude) <= allowedLocationDelta
}

func sameAntennaLocations(a, b []*ttnpb.GatewayAntenna) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		a, b := a[i], b[i]
		if a.Location != nil && b.Location != nil && !sameLocation(*a.Location, *b.Location) {
			return false
		}
		if (a.Location == nil) != (b.Location == nil) {
			return false
		}
	}
	return true
}

var statusLocationFields = ttnpb.ExcludeFields(ttnpb.LocationFieldPathsNested, "source")

func (gs *GatewayServer) handleLocationUpdates(ctx context.Context, conn connectionEntry) {
	var (
		gtw          = conn.Gateway()
		lastAntennas []*ttnpb.GatewayAntenna
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.LocationChanged():
			status, _, ok := conn.StatusStats()
			if ok && len(status.AntennaLocations) > 0 {
				// Construct the union of antennas that are in the gateway fetched from the entity registry with the antennas
				// that are in the status message.
				gtwAntennas := gtw.GetAntennas()
				c := len(gtwAntennas)
				if cs := len(status.AntennaLocations); cs > c {
					c = cs
				}
				antennas := make([]*ttnpb.GatewayAntenna, c)
				for i := range antennas {
					antennas[i] = &ttnpb.GatewayAntenna{}

					if i < len(gtwAntennas) {
						if err := antennas[i].SetFields(
							gtwAntennas[i],
							ttnpb.GatewayAntennaFieldPathsNested...,
						); err != nil {
							log.FromContext(ctx).WithError(err).Warn("Failed to clone antenna")
						}
					}

					if i < len(status.AntennaLocations) && status.AntennaLocations[i] != nil {
						antennas[i].Location = &ttnpb.Location{
							Source: ttnpb.SOURCE_GPS,
						}
						if err := antennas[i].Location.SetFields(
							status.AntennaLocations[i],
							statusLocationFields...,
						); err != nil {
							log.FromContext(ctx).WithError(err).Warn("Failed to clone antenna location")
						}
					}
				}
				if lastAntennas != nil && sameAntennaLocations(lastAntennas, antennas) {
					break
				}

				err := gs.entityRegistry.UpdateAntennas(ctx, *gtw.GetIds(), antennas)
				if err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to update antennas")
				} else {
					lastAntennas = antennas
				}
			}

			timeout := time.After(gs.config.UpdateGatewayLocationDebounceTime)
			select {
			case <-ctx.Done():
				return
			case <-timeout:
			}
		}
	}
}

// handleVersionInfoUpdates updates gateway attributes with version info.
// This function runs exactly once; only for the first status message of each connection, since version information should not change within the same connection.
func (gs *GatewayServer) handleVersionInfoUpdates(ctx context.Context, conn connectionEntry) {
	select {
	case <-ctx.Done():
		return
	case <-conn.VersionInfoChanged():
		status, _, ok := conn.StatusStats()
		versionsFromStatus := status.Versions
		if !ok || versionsFromStatus["model"] == "" || versionsFromStatus["firmware"] == "" {
			return
		}
		gtwAttributes := conn.Gateway().Attributes
		if versionsFromStatus[modelAttribute] == gtwAttributes[modelAttribute] && versionsFromStatus[firmwareAttribute] == gtwAttributes[firmwareAttribute] {
			return
		}
		attributes := map[string]string{
			modelAttribute:    versionsFromStatus[modelAttribute],
			firmwareAttribute: versionsFromStatus[firmwareAttribute],
		}
		d := random.Jitter(gs.config.UpdateVersionInfoDelay, 0.25)
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
		}
		err := gs.entityRegistry.UpdateAttributes(conn.Context(), *conn.Gateway().Ids, gtwAttributes, attributes)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to update version information")
		}
	}
}

// GetFrequencyPlans gets the frequency plans by the gateway identifiers.
func (gs *GatewayServer) GetFrequencyPlans(ctx context.Context, ids ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error) {
	gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: &ids,
		FieldMask:  &pbtypes.FieldMask{Paths: []string{"frequency_plan_ids"}},
	})
	var fpIDs []string
	if err == nil {
		fpIDs = gtw.FrequencyPlanIds
	} else if errors.IsNotFound(err) {
		fpID, ok := frequencyplans.FallbackIDFromContext(ctx)
		if !ok {
			return nil, err
		}
		fpIDs = append(fpIDs, fpID)
	} else {
		return nil, err
	}

	fps, err := gs.FrequencyPlansStore(ctx)
	if err != nil {
		return nil, err
	}

	fpGroup := make(map[string]*frequencyplans.FrequencyPlan, len(fpIDs))
	for _, fpID := range fpIDs {
		fp, err := fps.GetByID(fpID)
		if err != nil {
			return nil, err
		}
		fpGroup[fpID] = fp
	}
	return fpGroup, nil
}

// ClaimDownlink claims the downlink path for the given gateway.
func (gs *GatewayServer) ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.ClaimIDs(ctx, &ids)
}

// UnclaimDownlink releases the claim of the downlink path for the given gateway.
func (gs *GatewayServer) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.UnclaimIDs(ctx, &ids)
}

// ValidateGatewayID implements io.Server.
func (gs *GatewayServer) ValidateGatewayID(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.entityRegistry.ValidateGatewayID(ctx, ids)
}

type ctxConfigKeyType struct{}

// GetConfig returns the Gateway Server config based on the context.
func (gs *GatewayServer) GetConfig(ctx context.Context) (*Config, error) {
	if val, ok := ctx.Value(&ctxConfigKeyType{}).(*Config); ok {
		return val, nil
	}
	return gs.config, nil
}

// GetMQTTConfig returns the MQTT frontend configuration based on the context.
func (gs *GatewayServer) GetMQTTConfig(ctx context.Context) (*config.MQTT, error) {
	config, err := gs.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &config.MQTT, nil
}
