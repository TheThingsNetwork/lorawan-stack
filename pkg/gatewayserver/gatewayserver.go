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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	stdio "io"
	stdlog "log"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/mtls"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
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
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ttigw"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/ns"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/packetbroker"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewaytokens"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpctracer"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	appendUplinkCorrelationID = events.RegisterCorrelationIDPrefix("uplink", "gs:uplink")
	appendTxAckCorrelationID  = events.RegisterCorrelationIDPrefix("tx_ack", "gs:tx_ack")
)

// GatewayServer implements the Gateway Server component.
//
// The Gateway Server exposes the Gs, GtwGs and NsGs services and MQTT and UDP frontends for gateways.
type GatewayServer struct {
	ttnpb.UnimplementedGsServer
	ttnpb.UnimplementedNsGsServer

	*component.Component
	ctx context.Context

	config *Config

	requireRegisteredGateways bool
	forward                   map[string][]types.DevAddrPrefix

	entityRegistry EntityRegistry

	upstreamHandlers map[string]upstream.Handler

	connections sync.Map // string to connectionEntry

	statsRegistry GatewayConnectionStatsRegistry

	certVerifier CertificateVerifier
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

// CertificateVerifier abstracts certificate verification functions.
type CertificateVerifier interface {
	Verify(ctx context.Context, clientType mtls.ClientType, cn string, cert *x509.Certificate) error
}

var (
	errListenFrontend = errors.DefineFailedPrecondition(
		"listen_frontend",
		"start frontend listener `{protocol}` on address `{address}`",
	)
	errNotConnected        = errors.DefineNotFound("not_connected", "gateway `{gateway_uid}` not connected")
	errSetupUpstream       = errors.DefineFailedPrecondition("upstream", "setup upstream `{name}`")
	errInvalidUpstreamName = errors.DefineInvalidArgument("invalid_upstream_name", "upstream `{name}` is invalid")

	modelAttribute    = "model"
	firmwareAttribute = "firmware"
)

// New returns new *GatewayServer.
func New(c *component.Component, conf *Config, opts ...Option) (gs *GatewayServer, err error) {
	ctx := tracer.NewContextWithTracer(c.Context(), tracerNamespace)

	forward, err := conf.ForwardDevAddrPrefixes()
	if err != nil {
		return nil, err
	}
	if len(forward) == 0 {
		forward[""] = []types.DevAddrPrefix{{}}
	}

	ctx = log.NewContextWithField(ctx, "namespace", logNamespace)

	gs = &GatewayServer{
		Component:                 c,
		ctx:                       ctx,
		config:                    conf,
		requireRegisteredGateways: conf.RequireRegisteredGateways,
		forward:                   forward,
		upstreamHandlers:          make(map[string]upstream.Handler),
		statsRegistry:             conf.Stats,
		entityRegistry:            NewIS(c),
		certVerifier:              c.CAStore(),
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
	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpctracer.TracerHook, rpctracer.UnaryTracerHook(tracerNamespace)},
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook(logNamespace)},
	} {
		for _, filter := range []string{
			"/ttn.lorawan.v3.Ns",
			"/ttn.lorawan.v3.NsGs",
			"/ttn.lorawan.v3.GtwGs",
		} {
			c.GRPC.RegisterUnaryHook(filter, hook.name, hook.middleware)
		}
	}
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", cluster.HookName, c.ClusterAuthUnaryHook())

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

	// Start Semtech web socket listeners.
	for _, version := range []struct {
		Name      string
		Formatter semtechws.Formatter
		FallbackFreqPlanID,
		Listen,
		ListenTLS string
		Frontend semtechws.Config
	}{
		{
			Name:               "basicstation",
			Formatter:          lbslns.NewFormatter(conf.BasicStation.MaxValidRoundTripDelay),
			FallbackFreqPlanID: conf.BasicStation.FallbackFrequencyPlanID,
			Listen:             conf.BasicStation.Listen,
			ListenTLS:          conf.BasicStation.ListenTLS,
			Frontend:           conf.BasicStation.Config,
		},
	} {
		ctx := gs.Context()
		if version.FallbackFreqPlanID != "" {
			ctx = frequencyplans.WithFallbackID(ctx, version.FallbackFreqPlanID)
		}
		web, err := semtechws.New(ctx, gs, version.Formatter, version.Frontend)
		if err != nil {
			return nil, err
		}
		for _, endpoint := range []component.Endpoint{
			component.NewTCPEndpoint(version.Listen, version.Name),
			component.NewTLSEndpoint(version.ListenTLS, version.Name, tlsconfig.WithNextProtos("h2", "http/1.1")),
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
						Handler:     web,
						ReadTimeout: 120 * time.Second,
						// The ReadHeaderTimeout should be sufficiently long for embedded devices to perform the TLS handshake
						// before headers can be sent.
						// For example, The Things Indoor Gateway connecting via HTTPS to The Things Stack presenting a TLS server
						// certificate using ECDSA, the TLS handshake typically takes up to 10 seconds because there is limited
						// hardware acceleration as compared to RSA.
						ReadHeaderTimeout: 30 * time.Second,
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

	// Start The Things Industries gateway web socket listeners.
	ttiGWHandler, err := ttigw.New(ctx, gs, conf.TheThingsIndustriesGateway.Config)
	if err != nil {
		return nil, err
	}
	for _, endpoint := range []component.Endpoint{
		component.NewTCPEndpoint(conf.TheThingsIndustriesGateway.Listen, "ttigw"),
		component.NewTLSEndpoint(conf.TheThingsIndustriesGateway.ListenTLS, "ttigw",
			tlsconfig.WithTLSClientAuth(tls.RequestClientCert, nil, nil),
			tlsconfig.WithNextProtos("h2", "http/1.1"),
		),
	} {
		endpoint := endpoint
		if endpoint.Address() == "" {
			continue
		}
		gs.RegisterTask(&task.Config{
			Context: gs.Context(),
			ID:      fmt.Sprintf("serve_ttigw/%s", endpoint.Address()),
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
					Handler:           ttiGWHandler,
					ReadTimeout:       120 * time.Second,
					ReadHeaderTimeout: 30 * time.Second,
					ErrorLog:          stdlog.New(stdio.Discard, "", 0),
					BaseContext: func(net.Listener) context.Context {
						ctx := context.Background()
						if fallbackID := conf.TheThingsIndustriesGateway.FallbackFrequencyPlanID; fallbackID != "" {
							ctx = frequencyplans.WithFallbackID(ctx, fallbackID)
						}
						return ctx
					},
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
func (gs *GatewayServer) FillGatewayContext(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (context.Context, *ttnpb.GatewayIdentifiers, error) {
	ctx = gs.FillContext(ctx)
	if ids.IsZero() {
		return nil, nil, errEmptyIdentifiers.New()
	}
	var (
		linkRightsInContext bool
		linkRights          = &ttnpb.Rights{
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_GATEWAY_LINK,
			},
		}
	)
	if ids.GatewayId == "" {
		eui := types.MustEUI64(ids.Eui)
		if eui.OrZero().IsZero() {
			return nil, nil, errEmptyIdentifiers.New()
		}
		extIDs, err := gs.entityRegistry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			Eui: eui.Bytes(),
		})
		if err == nil {
			ids = extIDs
		} else if errors.IsNotFound(err) {
			if gs.requireRegisteredGateways {
				return nil, nil, errGatewayEUINotRegistered.WithAttributes("eui", eui).WithCause(err)
			}
			ids.GatewayId = fmt.Sprintf("eui-%v", strings.ToLower(eui.String()))
			ctx = rights.NewContext(ctx, &rights.Rights{
				GatewayRights: *rights.NewMap(map[string]*ttnpb.Rights{
					unique.ID(ctx, ids): linkRights,
				}),
			})
			linkRightsInContext = true
		} else {
			return nil, nil, err
		}
	}
	if cert := mtls.ClientCertificateFromContext(ctx); cert != nil {
		// Verify the client certificate.
		err := gs.certVerifier.Verify(ctx, mtls.ClientTypeGateway, types.MustEUI64(ids.Eui).String(), cert)
		if err != nil {
			return nil, nil, errUnauthenticatedGatewayConnection.WithCause(err)
		}
		if !linkRightsInContext {
			token := gatewaytokens.New(gs.config.GatewayTokenHashKeyID, ids, linkRights, gs.KeyService())
			ctx, err = gatewaytokens.AuthenticatedContext(gatewaytokens.NewContext(ctx, token))
			if err != nil {
				return nil, nil, err
			}
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

// AssertGatewayRights checks that the caller has the required rights over the provided gateway identifiers.
func (gs *GatewayServer) AssertGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers, rights ...ttnpb.Right) error {
	return gs.entityRegistry.AssertGatewayRights(ctx, ids, rights...)
}

// Connect connects a gateway by its identifiers to the Gateway Server, and returns a io.Connection for traffic and
// control.
func (gs *GatewayServer) Connect(
	ctx context.Context,
	frontend io.Frontend,
	ids *ttnpb.GatewayIdentifiers,
	addr *ttnpb.GatewayRemoteAddress,
	opts ...io.ConnectionOption,
) (*io.Connection, error) {
	if err := gs.AssertGatewayRights(ctx, ids, ttnpb.Right_RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"protocol", frontend.Protocol(),
		"gateway_uid", uid,
		"gateway_ip_address", addr.Ip,
	))
	ctx = log.NewContext(ctx, logger)

	var isAuthenticated bool
	if _, err := rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials()); err == nil {
		isAuthenticated = true
	}
	gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: ids,
		FieldMask: ttnpb.FieldMask(
			"administrative_contact",
			"antennas",
			"attributes",
			"disable_packet_broker_forwarding",
			"downlink_path_constraint",
			"enforce_duty_cycle",
			"frequency_plan_id",
			"frequency_plan_ids",
			"gateway_server_address",
			"location_public",
			"require_authenticated_connection",
			"schedule_anytime_delay",
			"schedule_downlink_late",
			"status_public",
			"technical_contact",
			"update_location_from_status",
		),
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
			Ids:                    ids,
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

	ids = gtw.GetIds()

	fps, err := gs.FrequencyPlansStore(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := io.NewConnection(
		ctx, frontend, gtw, fps, gtw.EnforceDutyCycle, ttnpb.StdDuration(gtw.ScheduleAnytimeDelay), addr, opts...,
	)
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

	registerGatewayConnect(ctx, ids, &ttnpb.GatewayConnectionStats{
		ConnectedAt:          timestamppb.New(conn.ConnectTime()),
		Protocol:             conn.Frontend().Protocol(),
		GatewayRemoteAddress: conn.GatewayRemoteAddress(),
	})
	logger.Info("Connected")

	gs.startDisconnectOnChangeTask(connEntry)
	gs.startHandleUpstreamTask(connEntry)
	gs.startUpdateConnStatsTask(connEntry)
	// Unauthenticated connections cannot update the gateway entity.
	// As such, there is no reason to start these tasks, since they
	// will perpetually fail.
	if isAuthenticated {
		gs.startHandleLocationUpdatesTask(connEntry)
		gs.startHandleVersionUpdatesTask(connEntry)
	}

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
func (gs *GatewayServer) GetConnection(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*io.Connection, bool) {
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
		if len(current.Antennas) > 0 && current.Antennas[0].Location != nil && current.Antennas[0].Location.Source != ttnpb.LocationSource_SOURCE_GPS {
			return true
		}
	}
	if !sameAntennaGain(connected.GetAntennas(), current.GetAntennas()) {
		return true
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
		len(connected.FrequencyPlanIds) != len(current.FrequencyPlanIds) ||
		connected.GatewayServerAddress != current.GatewayServerAddress {
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
			for {
				d := random.Jitter(gs.config.FetchGatewayInterval, gs.config.FetchGatewayJitter)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(d):
				}

				gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
					GatewayIds: conn.Gateway().GetIds(),
					FieldMask: ttnpb.FieldMask(
						"antennas",
						"disable_packet_broker_forwarding",
						"downlink_path_constraint",
						"enforce_duty_cycle",
						"frequency_plan_id",
						"frequency_plan_ids",
						"gateway_server_address",
						"location_public",
						"require_authenticated_connection",
						"schedule_anytime_delay",
						"schedule_downlink_late",
						"status_public",
						"update_location_from_status",
					),
				})
				if err != nil {
					if errors.IsUnauthenticated(err) || errors.IsPermissionDenied(err) {
						// Since there is an active connection, the `Get` request will not return a `NotFound` error as the
						// gateway existed during the connect, since the rights assertion fails first.
						// Instead,
						// 1. If the gateway is connected with an API key and is deleted, the IS returns an `Unauthenticated`,
						// since the API Key is also deleted.
						// 2. If the gateway is connected without an API key (UDP, LBS in unauthenticated mode) and is deleted
						// the IS returns an `PermissionDenied` as there are no rights for these IDs.
						log.FromContext(ctx).WithError(err).Debug("Gateway was deleted and/or the API key used to link the gateway was invalidated") // nolint: lll
						conn.Disconnect(err)
					} else {
						log.FromContext(ctx).WithError(err).Warn("Failed to get gateway")
					}
					return err
				}
				if requireDisconnect(conn.Gateway(), gtw) {
					log.FromContext(ctx).Info("Gateway changed in registry, disconnect")
					conn.Disconnect(errGatewayChanged.New())
					return nil
				}
			}
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

var errHostHandle = errors.Define("host_handle", "host `{host}` handle message")

type upstreamHost struct {
	name    string
	handler upstream.Handler
	pool    workerpool.WorkerPool[any]
	gtw     *ttnpb.Gateway
}

func (host *upstreamHost) handlePacket(ctx context.Context, item any) {
	logger := log.FromContext(ctx)
	gtw := host.gtw
	// Each concurrent upstream host will receive the message and edit it in order
	// to append the correlation IDs. This would be a concurrent write, so we are
	// creating a shallow message copy in order to safely edit the correlation IDs.
	switch msg := item.(type) {
	case *ttnpb.GatewayUplinkMessage:
		ctx := events.ContextWithCorrelationID(ctx, msg.Message.CorrelationIds...)
		msg = &ttnpb.GatewayUplinkMessage{
			BandId: msg.BandId,
			Message: &ttnpb.UplinkMessage{
				RawPayload:         msg.Message.RawPayload,
				Payload:            msg.Message.Payload,
				Settings:           msg.Message.Settings,
				RxMetadata:         msg.Message.RxMetadata,
				ReceivedAt:         msg.Message.ReceivedAt,
				CorrelationIds:     events.CorrelationIDsFromContext(ctx),
				DeviceChannelIndex: msg.Message.DeviceChannelIndex,
				ConsumedAirtime:    msg.Message.ConsumedAirtime,
			},
		}
		drop := func(ids *ttnpb.EndDeviceIdentifiers, err error) {
			logger := logger.WithError(err)
			if joinEUI := types.MustEUI64(ids.JoinEui).OrZero(); !joinEUI.IsZero() {
				logger = logger.WithField("join_eui", joinEUI)
			}
			if devEUI := types.MustEUI64(ids.DevEui).OrZero(); !devEUI.IsZero() {
				logger = logger.WithField("dev_eui", devEUI)
			}
			if devAddr := types.MustDevAddr(ids.DevAddr).OrZero(); !devAddr.IsZero() {
				logger = logger.WithField("dev_addr", devAddr)
			}
			logger.Debug("Drop message")
			registerDropUplink(ctx, gtw, msg, host.name, err)
		}
		ids := msg.Message.Payload.EndDeviceIdentifiers()
		var pass bool
		switch {
		case ids.DevAddr != nil:
			for _, prefix := range host.handler.DevAddrPrefixes() {
				if types.MustDevAddr(ids.DevAddr).HasPrefix(prefix) {
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
		switch err := host.handler.HandleUplink(ctx, gtw.Ids, ids, msg); codes.Code(errors.Code(err)) {
		case codes.Canceled, codes.DeadlineExceeded,
			codes.Unknown, codes.Internal,
			codes.Unimplemented, codes.Unavailable:
			drop(ids, errHostHandle.WithCause(err).WithAttributes("host", host.name))
		default:
			registerForwardUplink(ctx, gtw, msg, host.name)
		}
	case *ttnpb.GatewayStatus:
		if err := host.handler.HandleStatus(ctx, gtw.Ids, msg); err != nil {
			registerDropStatus(ctx, gtw, msg, host.name, err)
		} else {
			registerForwardStatus(ctx, gtw, msg, host.name)
		}
	case *ttnpb.TxAcknowledgment:
		correlationIDs := make([]string, 0, len(msg.CorrelationIds)+len(msg.DownlinkMessage.GetCorrelationIds()))
		correlationIDs = append(correlationIDs, msg.CorrelationIds...)
		correlationIDs = append(correlationIDs, msg.DownlinkMessage.GetCorrelationIds()...)
		ctx := events.ContextWithCorrelationID(ctx, correlationIDs...)
		msg = &ttnpb.TxAcknowledgment{
			CorrelationIds:  events.CorrelationIDsFromContext(ctx),
			Result:          msg.Result,
			DownlinkMessage: msg.DownlinkMessage,
		}
		if down := msg.DownlinkMessage; down != nil {
			msg.DownlinkMessage = &ttnpb.DownlinkMessage{
				RawPayload:     down.RawPayload,
				Payload:        down.Payload,
				EndDeviceIds:   down.EndDeviceIds,
				Settings:       down.Settings,
				CorrelationIds: events.CorrelationIDsFromContext(ctx),
				SessionKeyId:   down.SessionKeyId,
			}
		}
		if err := host.handler.HandleTxAck(ctx, gtw.Ids, msg); err != nil {
			registerDropTxAck(ctx, gtw, msg, host.name, err)
		} else {
			registerForwardTxAck(ctx, gtw, msg, host.name)
		}
	default:
		panic(fmt.Sprintf("unknown type %T", msg))
	}
}

var errMessageCRC = errors.DefineInvalidArgument("message_crc", "message CRC failed")

func (gs *GatewayServer) handleUpstream(ctx context.Context, conn connectionEntry) {
	var (
		gtw      = conn.Gateway()
		protocol = conn.Frontend().Protocol()
		logger   = log.FromContext(ctx)
	)
	defer func() {
		gs.connections.Delete(unique.ID(ctx, gtw.GetIds()))
		registerGatewayDisconnect(ctx, gtw.GetIds(), protocol, ctx.Err())
		logger.Info("Disconnected")
	}()

	hosts := make([]*upstreamHost, 0, len(gs.upstreamHandlers))
	for name, handler := range gs.upstreamHandlers {
		if name == "packetbroker" && gtw.DisablePacketBrokerForwarding {
			continue
		}
		host := &upstreamHost{
			name:    name,
			handler: handler,
			gtw:     gtw,
		}
		wp := workerpool.NewWorkerPool(workerpool.Config[any]{
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
			val any
		)
		select {
		case <-ctx.Done():
			return
		case msg := <-conn.Up():
			ctx = events.ContextWithCorrelationID(ctx, msg.Message.CorrelationIds...)
			ctx = appendUplinkCorrelationID(ctx)
			msg.Message.CorrelationIds = events.CorrelationIDsFromContext(ctx)
			if msg.Message.Payload == nil {
				msg.Message.Payload = &ttnpb.Message{}
				if err := lorawan.UnmarshalMessage(msg.Message.RawPayload, msg.Message.Payload); err != nil {
					continue
				}
			}
			registerReceiveUplink(ctx, gtw, msg, protocol)
			if crcStatus := msg.Message.CrcStatus; crcStatus != nil && !crcStatus.Value {
				registerDropUplink(ctx, gtw, msg, "", errMessageCRC.New())
				continue
			}
			if err := msg.Message.Payload.ValidateFields(); err != nil {
				registerDropUplink(ctx, gtw, msg, "", err)
				continue
			}
			val = msg
		case msg := <-conn.Status():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:status:%s", events.NewCorrelationID()))
			val = msg
			registerReceiveStatus(ctx, gtw, msg, protocol)
		case msg := <-conn.TxAck():
			correlationIDs := make([]string, 0, len(msg.CorrelationIds)+len(msg.DownlinkMessage.GetCorrelationIds()))
			correlationIDs = append(correlationIDs, msg.CorrelationIds...)
			correlationIDs = append(correlationIDs, msg.DownlinkMessage.GetCorrelationIds()...)
			ctx = events.ContextWithCorrelationID(ctx, correlationIDs...)
			ctx = appendTxAckCorrelationID(ctx)
			msg.CorrelationIds = events.CorrelationIDsFromContext(ctx)
			if d := msg.DownlinkMessage; d != nil {
				d.CorrelationIds = events.CorrelationIDsFromContext(ctx)
			}
			registerReceiveTxAck(ctx, gtw, msg, protocol)
			if msg.Result == ttnpb.TxAcknowledgment_SUCCESS {
				registerSuccessDownlink(ctx, gtw, protocol)
			} else {
				registerFailDownlink(ctx, gtw, msg, protocol)
			}
			val = msg
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

func earliestTimestamp(a, b *timestamppb.Timestamp) *timestamppb.Timestamp {
	switch {
	case a == nil && b == nil:
		return nil
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		if aT, bT := a.AsTime(), b.AsTime(); aT.Before(bT) {
			return a
		}
		return b
	}
}

func (gs *GatewayServer) updateConnStats(ctx context.Context, conn connectionEntry) {
	decoupledCtx := gs.FromRequestContext(ctx)
	logger := log.FromContext(ctx)

	ids := conn.Connection.Gateway().GetIds()

	// Initial update, so that the gateway appears connected.
	stats := &ttnpb.GatewayConnectionStats{
		ConnectedAt:          timestamppb.New(conn.Connection.ConnectTime()),
		Protocol:             conn.Connection.Frontend().Protocol(),
		GatewayRemoteAddress: conn.Connection.GatewayRemoteAddress(),
	}
	registerGatewayConnectionStats(ctx, ids, ttnpb.Clone(stats))
	if gs.statsRegistry != nil {
		if err := gs.statsRegistry.Set(
			decoupledCtx,
			ids,
			func(*ttnpb.GatewayConnectionStats) (*ttnpb.GatewayConnectionStats, []string, error) {
				return stats, ttnpb.GatewayConnectionStatsFieldPathsTopLevel, nil
			},
			gs.config.ConnectionStatsTTL,
		); err != nil {
			logger.WithError(err).Warn("Failed to initialize connection stats")
		}
	}

	defer func() {
		logger.Debug("Delete connection stats")
		stats := &ttnpb.GatewayConnectionStats{
			ConnectedAt:    nil,
			DisconnectedAt: timestamppb.Now(),
		}
		registerGatewayConnectionStats(decoupledCtx, ids, ttnpb.Clone(stats))
		if gs.statsRegistry == nil {
			return
		}
		if err := gs.statsRegistry.Set(
			decoupledCtx,
			ids,
			func(*ttnpb.GatewayConnectionStats) (*ttnpb.GatewayConnectionStats, []string, error) {
				return stats, []string{"connected_at", "disconnected_at"}, nil
			},
			gs.config.ConnectionStatsDisconnectTTL,
		); err != nil {
			logger.WithError(err).Warn("Failed to clear connection stats")
		}
	}()

	var (
		nextStats  = time.NewTimer(gs.config.UpdateConnectionStatsInterval)
		lastUpdate = time.Now() // Start with a debounce, the initial update has already been sent.
	)
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.StatsChanged():
			if !nextStats.Stop() {
				<-nextStats.C
			}
		case <-nextStats.C:
		}
		nextStats.Reset(gs.config.UpdateConnectionStatsInterval)

		// Debounce the updates with jitter to spread event publishes and store updates over time.
		// If the time since the last update is longer than the debounce time, the update happens immediately.
		if wait := gs.config.UpdateConnectionStatsDebounceTime - time.Since(lastUpdate); random.CanJitter(wait, debounceJitter) {
			duration := random.Jitter(wait, debounceJitter)
			select {
			case <-ctx.Done():
				return
			case <-time.After(duration):
			}
		}
		lastUpdate = time.Now()

		stats, paths := conn.Stats()
		// The stats must be cloned, they are passed to the event publisher and must not be modified
		// by another goroutine during proto marshalling, otherwise it will trigger a panic.
		registerGatewayConnectionStats(decoupledCtx, ids, ttnpb.Clone(stats))
		if gs.statsRegistry == nil {
			continue
		}
		if err := gs.statsRegistry.Set(
			decoupledCtx,
			ids,
			func(pb *ttnpb.GatewayConnectionStats) (*ttnpb.GatewayConnectionStats, []string, error) {
				if pb != nil {
					stats.ConnectedAt = earliestTimestamp(stats.ConnectedAt, pb.ConnectedAt)
				}
				return stats, paths, nil
			},
			gs.config.ConnectionStatsTTL,
			"connected_at",
		); err != nil {
			logger.WithError(err).Warn("Failed to update connection stats")
		}
	}
}

const (
	allowedLocationDelta = 0.00001
	debounceJitter       = 0.25
)

func sameLocation(a, b *ttnpb.Location) bool {
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
		if a.Location != nil && b.Location != nil && !sameLocation(a.Location, b.Location) {
			return false
		}
		if (a.Location == nil) != (b.Location == nil) {
			return false
		}
	}
	return true
}

func sameAntennaGain(a, b []*ttnpb.GatewayAntenna) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		a, b := a[i], b[i]
		if a.Gain != b.Gain {
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
							Source: ttnpb.LocationSource_SOURCE_GPS,
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

				err := gs.entityRegistry.UpdateAntennas(ctx, gtw.GetIds(), antennas)
				if err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to update antennas")
				} else {
					lastAntennas = antennas
				}
			}

			if wait := gs.config.UpdateGatewayLocationDebounceTime; random.CanJitter(wait, debounceJitter) {
				duration := random.Jitter(wait, debounceJitter)
				select {
				case <-ctx.Done():
					return
				case <-time.After(duration):
				}
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
		err := gs.entityRegistry.UpdateAttributes(conn.Context(), conn.Gateway().Ids, gtwAttributes, attributes)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to update version information")
		}
	}
}

// GetFrequencyPlans gets the frequency plans by the gateway identifiers.
func (gs *GatewayServer) GetFrequencyPlans(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error) {
	gtw, err := gs.entityRegistry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: ids,
		FieldMask:  ttnpb.FieldMask("frequency_plan_ids"),
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
func (gs *GatewayServer) ClaimDownlink(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	return gs.ClaimIDs(ctx, ids)
}

// UnclaimDownlink releases the claim of the downlink path for the given gateway.
func (gs *GatewayServer) UnclaimDownlink(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	return gs.UnclaimIDs(ctx, ids)
}

// ValidateGatewayID implements io.Server.
func (gs *GatewayServer) ValidateGatewayID(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
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
