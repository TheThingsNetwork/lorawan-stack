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
	"io/ioutil"
	stdlog "log"
	"math"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/entityregistry/is"
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
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
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

	gs = &GatewayServer{
		Component:                         c,
		ctx:                               log.NewContextWithField(c.Context(), "namespace", "gatewayserver"),
		config:                            conf,
		requireRegisteredGateways:         conf.RequireRegisteredGateways,
		forward:                           forward,
		upstreamHandlers:                  make(map[string]upstream.Handler),
		statsRegistry:                     conf.Stats,
		updateConnectionStatsDebounceTime: conf.UpdateConnectionStatsDebounceTime,
		entityRegistry:                    is.New(c),
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
		gs.RegisterTask(&component.TaskConfig{
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
			Restart: component.TaskRestartOnFailure,
			Backoff: component.DefaultTaskBackoffConfig,
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
			gs.RegisterTask(&component.TaskConfig{
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
				Restart: component.TaskRestartOnFailure,
				Backoff: component.DefaultTaskBackoffConfig,
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
				frontend: ws.Config{
					UseTrafficTLSAddress: conf.BasicStation.UseTrafficTLSAddress,
					WSPingInterval:       conf.BasicStation.WSPingInterval,
					AllowUnauthenticated: conf.BasicStation.AllowUnauthenticated,
				},
			},
		},
	} {
		ctx := gs.Context()
		if version.listenerConfig.fallbackFreqPlanID != "" {
			ctx = frequencyplans.WithFallbackID(ctx, version.listenerConfig.fallbackFreqPlanID)
		}
		webServer := ws.New(ctx, gs, version.Formatter, version.listenerConfig.frontend)
		for _, endpoint := range []component.Endpoint{
			component.NewTCPEndpoint(version.listenerConfig.listen, version.Name),
			component.NewTLSEndpoint(version.listenerConfig.listenTLS, version.Name, component.WithNextProtos("h2", "http/1.1")),
		} {
			endpoint := endpoint
			if endpoint.Address() == "" {
				continue
			}
			gs.RegisterTask(&component.TaskConfig{
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
						Handler:           webServer,
						ReadTimeout:       120 * time.Second,
						ReadHeaderTimeout: 5 * time.Second,
						ErrorLog:          stdlog.New(ioutil.Discard, "", 0),
					}
					go func() {
						<-ctx.Done()
						srv.Close()
					}()
					return srv.Serve(lis)
				},
				Restart: component.TaskRestartOnFailure,
				Backoff: component.DefaultTaskBackoffConfig,
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
			Eui: *ids.Eui,
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
	upstreamDone chan struct{}
	tasksDone    *sync.WaitGroup
}

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
	ctx = log.NewContext(ctx, logger)
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:conn:%s", events.NewCorrelationID()))

	var isAuthenticated bool
	if _, err := rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials()); err == nil {
		isAuthenticated = true
	}
	gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: ids,
		FieldMask: pbtypes.FieldMask{
			Paths: []string{
				"antennas",
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
			GatewayIdentifiers:     ids,
			FrequencyPlanID:        fpID,
			FrequencyPlanIDs:       []string{fpID},
			EnforceDutyCycle:       true,
			DownlinkPathConstraint: ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE,
			Antennas:               []ttnpb.GatewayAntenna{},
		}
	} else if err != nil {
		return nil, err
	}
	if gtw.RequireAuthenticatedConnection && !isAuthenticated {
		return nil, errUnauthenticatedGatewayConnection.New()
	}

	ids = gtw.GatewayIdentifiers

	conn, err := io.NewConnection(ctx, frontend, gtw, gs.FrequencyPlans, gtw.EnforceDutyCycle, gtw.ScheduleAnytimeDelay)
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	// The tasks will always start once the entry is stored.
	// As such, we must ensure any new connection waits for
	// all of the upstream tasks to finish.
	wg.Add(len(gs.upstreamHandlers))
	connEntry := connectionEntry{
		Connection:   conn,
		upstreamDone: make(chan struct{}),
		tasksDone:    wg,
	}
	for existing, exists := gs.connections.LoadOrStore(uid, connEntry); exists; existing, exists = gs.connections.LoadOrStore(uid, connEntry) {
		existingConnEntry := existing.(connectionEntry)
		logger.Warn("Disconnect existing connection")
		existingConnEntry.Disconnect(errNewConnection)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-existingConnEntry.upstreamDone:
		}
		existingConnEntry.tasksDone.Wait()
	}

	registerGatewayConnect(ctx, ids, frontend.Protocol())
	logger.Info("Connected")

	go gs.handleUpstream(connEntry)
	if gs.statsRegistry != nil {
		go gs.updateConnStats(connEntry)
	}
	if gtw.UpdateLocationFromStatus {
		go gs.handleLocationUpdates(connEntry)
	}

	for name, handler := range gs.upstreamHandlers {
		connCtx := log.NewContextWithField(conn.Context(), "upstream_handler", name)
		handler := handler
		gs.StartTask(&component.TaskConfig{
			Context: connCtx,
			ID:      fmt.Sprintf("%s_connect_gateway_%s", name, ids.GatewayId),
			Func: func(ctx context.Context) error {
				return handler.ConnectGateway(ctx, ids, conn)
			},
			Done:    wg.Done,
			Restart: component.TaskRestartOnFailure,
			Backoff: component.DialTaskBackoffConfig,
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

var errHostHandle = errors.Define("host_handle", "host `{host}` failed to handle message")

var (
	// maxUpstreamHandlers is the maximum number of goroutines per gateway connection to handle upstream messages.
	maxUpstreamHandlers = int32(1 << 5)
	// upstreamHandlerIdleTimeout is the duration after which an idle upstream handler stops to save resources.
	upstreamHandlerIdleTimeout = (1 << 7) * time.Millisecond
)

type upstreamHost struct {
	name          string
	handler       upstream.Handler
	handlers      int32
	handleWg      sync.WaitGroup
	handleCh      chan upstreamItem
	correlationID string
}

type upstreamItem struct {
	ctx context.Context
	val interface{}
}

func (gs *GatewayServer) handleUpstream(conn connectionEntry) {
	var (
		ctx      = conn.Context()
		gtw      = conn.Gateway()
		protocol = conn.Frontend().Protocol()
		logger   = log.FromContext(ctx)
	)
	defer func() {
		gs.connections.Delete(unique.ID(ctx, gtw.GatewayIdentifiers))
		registerGatewayDisconnect(ctx, gtw.GatewayIdentifiers, protocol)
		logger.Info("Disconnected")
		close(conn.upstreamDone)
	}()

	handleFn := func(ctx context.Context, host *upstreamHost) {
		defer recoverHandler(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(upstreamHandlerIdleTimeout):
				return
			case item := <-host.handleCh:
				ctx := item.ctx
				ctx = events.ContextWithCorrelationID(ctx, host.correlationID)
				switch msg := item.val.(type) {
				case *ttnpb.GatewayUplinkMessage:
					up := *msg.UplinkMessage
					msg = &ttnpb.GatewayUplinkMessage{
						BandID:        msg.BandID,
						UplinkMessage: &up,
					}
					msg.CorrelationIDs = append(make([]string, 0, len(msg.CorrelationIDs)+1), msg.CorrelationIDs...)
					msg.CorrelationIDs = append(msg.CorrelationIDs, host.correlationID)
					drop := func(ids ttnpb.EndDeviceIdentifiers, err error) {
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
						registerDropUplink(ctx, gtw, msg.UplinkMessage, host.name, err)
					}
					ids, err := lorawan.GetUplinkMessageIdentifiers(msg.RawPayload)
					if err != nil {
						drop(ttnpb.EndDeviceIdentifiers{}, err)
						break
					}
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
					switch err := host.handler.HandleUplink(ctx, gtw.GatewayIdentifiers, ids, msg); codes.Code(errors.Code(err)) {
					case codes.Canceled, codes.DeadlineExceeded,
						codes.Unknown, codes.Internal,
						codes.Unimplemented, codes.Unavailable:
						drop(ids, errHostHandle.WithCause(err).WithAttributes("host", host.name))
					default:
						registerForwardUplink(ctx, gtw, msg.UplinkMessage, host.name)
					}
				case *ttnpb.GatewayStatus:
					if err := host.handler.HandleStatus(ctx, gtw.GatewayIdentifiers, msg); err != nil {
						registerDropStatus(ctx, gtw, msg, host.name, err)
					} else {
						registerForwardStatus(ctx, gtw, msg, host.name)
					}
				case *ttnpb.TxAcknowledgment:
					if err := host.handler.HandleTxAck(ctx, gtw.GatewayIdentifiers, msg); err != nil {
						logger.WithField("host", host.name).WithError(err).Debug("Drop Tx acknowledgment")
					}
				}
			}
		}
	}

	hosts := make([]*upstreamHost, 0, len(gs.upstreamHandlers))
	for name, handler := range gs.upstreamHandlers {
		host := &upstreamHost{
			name:          name,
			handler:       handler,
			handleCh:      make(chan upstreamItem),
			correlationID: fmt.Sprintf("gs:up:host:%s", events.NewCorrelationID()),
		}
		hosts = append(hosts, host)
		defer host.handleWg.Wait()
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
			msg.CorrelationIDs = append(msg.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
			if msg.Payload == nil {
				pld := &ttnpb.Message{}
				if err := lorawan.UnmarshalMessage(msg.RawPayload, pld); err != nil {
					log.FromContext(ctx).WithError(err).Debug("Failed to decode message payload")
				} else {
					msg.Payload = pld
				}
			}
			val = msg
			registerReceiveUplink(ctx, gtw, msg.UplinkMessage, protocol)
		case msg := <-conn.Status():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:status:%s", events.NewCorrelationID()))
			val = msg
			registerReceiveStatus(ctx, gtw, msg, protocol)
		case msg := <-conn.TxAck():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:tx_ack:%s", events.NewCorrelationID()))
			if d := msg.DownlinkMessage; d != nil {
				d.CorrelationIDs = append(d.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
			}
			if msg.Result == ttnpb.TxAcknowledgment_SUCCESS {
				registerSuccessDownlink(ctx, gtw, protocol)
			} else {
				registerFailDownlink(ctx, gtw, msg, protocol)
			}
			val = msg
		}
		item := upstreamItem{ctx, val}
		for _, host := range hosts {
			host := host
			select {
			case <-ctx.Done():
				return
			case host.handleCh <- item:
			default:
				if atomic.LoadInt32(&host.handlers) < maxUpstreamHandlers {
					atomic.AddInt32(&host.handlers, 1)
					host.handleWg.Add(1)
					registerUpstreamHandlerStart(ctx, host.name)
					go func() {
						defer host.handleWg.Done()
						defer atomic.AddInt32(&host.handlers, -1)
						defer registerUpstreamHandlerStop(ctx, host.name)
						handleFn(ctx, host)
					}()
					go func() {
						select {
						case <-ctx.Done():
							return
						case host.handleCh <- item:
						}
					}()
					continue
				}
				logger.WithField("name", host.name).Warn("Upstream handler busy, fail message")
				switch msg := val.(type) {
				case *ttnpb.UplinkMessage:
					registerFailUplink(ctx, gtw, msg, host.name)
				case *ttnpb.GatewayStatus:
					registerFailStatus(ctx, gtw, msg, host.name)
				}
			}
		}
	}
}

func (gs *GatewayServer) updateConnStats(conn connectionEntry) {
	ctx := conn.Context()
	logger := log.FromContext(ctx)

	ids := conn.Connection.Gateway().GatewayIdentifiers
	connectTime := conn.Connection.ConnectTime()
	stats := &ttnpb.GatewayConnectionStats{
		ConnectedAt: &connectTime,
		Protocol:    conn.Connection.Frontend().Protocol(),
	}

	// Initial update, so that the gateway appears connected.
	if err := gs.statsRegistry.Set(ctx, ids, stats, ttnpb.GatewayConnectionStatsFieldPathsTopLevel); err != nil {
		logger.WithError(err).Error("Failed to initialize connection stats")
	}

	defer func() {
		logger.Debug("Delete connection stats")
		if err := gs.statsRegistry.Set(gs.FromRequestContext(ctx), ids, nil, nil); err != nil {
			logger.WithError(err).Error("Failed to clear connection stats")
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.StatsChanged():
		}
		stats, paths := conn.Stats()
		if err := gs.statsRegistry.Set(ctx, ids, stats, paths); err != nil {
			logger.WithError(err).Error("Failed to update connection stats")
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

func sameAntennaLocations(a, b []ttnpb.GatewayAntenna) bool {
	if len(a) != len(b) {
		return false
	}
	for _, a := range a {
		for _, b := range b {
			if a.Location != nil && b.Location != nil && !sameLocation(*a.Location, *b.Location) {
				return false
			}
			if (a.Location == nil) != (b.Location == nil) {
				return false
			}
		}
	}
	return true
}

func (gs *GatewayServer) handleLocationUpdates(conn connectionEntry) {
	var (
		ctx          = conn.Context()
		gtw          = conn.Gateway()
		lastAntennas []ttnpb.GatewayAntenna
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
				c := len(gtw.Antennas)
				if cs := len(status.AntennaLocations); cs > c {
					c = cs
				}
				antennas := make([]ttnpb.GatewayAntenna, c)
				for i, ant := range gtw.Antennas {
					antennas[i].Gain = ant.Gain
				}
				for i := range antennas {
					if i < len(status.AntennaLocations) && status.AntennaLocations[i] != nil {
						loc := *status.AntennaLocations[i]
						loc.Source = ttnpb.SOURCE_GPS
						antennas[i].Location = &loc
					}
				}
				if lastAntennas != nil && sameAntennaLocations(lastAntennas, antennas) {
					break
				}

				err := gs.entityRegistry.UpdateAntennas(ctx, gtw.GatewayIdentifiers, antennas)
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

// GetFrequencyPlans gets the frequency plans by the gateway identifiers.
func (gs *GatewayServer) GetFrequencyPlans(ctx context.Context, ids ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error) {
	gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: ids,
		FieldMask:          pbtypes.FieldMask{Paths: []string{"frequency_plan_ids"}},
	})
	var fpIDs []string
	if err == nil {
		fpIDs = gtw.FrequencyPlanIDs
	} else if errors.IsNotFound(err) {
		fpID, ok := frequencyplans.FallbackIDFromContext(ctx)
		if !ok {
			return nil, err
		}
		fpIDs = append(fpIDs, fpID)
	} else {
		return nil, err
	}

	fps := make(map[string]*frequencyplans.FrequencyPlan, len(fpIDs))
	for _, fpID := range fpIDs {
		fp, err := gs.FrequencyPlans.GetByID(fpID)
		if err != nil {
			return nil, err
		}
		fps[fpID] = fp
	}
	return fps, nil
}

// ClaimDownlink claims the downlink path for the given gateway.
func (gs *GatewayServer) ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.ClaimIDs(ctx, &ids)
}

// UnclaimDownlink releases the claim of the downlink path for the given gateway.
func (gs *GatewayServer) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.UnclaimIDs(ctx, &ids)
}

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

var errHandlerRecovered = errors.DefineInternal("handler_recovered", "internal server error")

func recoverHandler(ctx context.Context) error {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, p)
		os.Stderr.Write(debug.Stack())
		var err error
		if pErr, ok := p.(error); ok {
			err = errHandlerRecovered.WithCause(pErr)
		} else {
			err = errHandlerRecovered.WithAttributes("panic", p)
		}
		log.FromContext(ctx).WithError(err).Error("Handler failed")
		return err
	}
	return nil
}
