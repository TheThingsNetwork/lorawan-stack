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
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/basicstationlns"
	iogrpc "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/grpc"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/upstream"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/upstream/ns"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
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

	registry ttnpb.GatewayRegistryClient

	upstreamHandlers map[string]upstream.Handler

	connections sync.Map
}

func (gs *GatewayServer) getRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	if gs.registry != nil {
		return gs.registry, nil
	}
	cc, err := gs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}

// Option configures GatewayServer.
type Option func(*GatewayServer)

// WithRegistry overrides the registry.
func WithRegistry(registry ttnpb.GatewayRegistryClient) Option {
	return func(gs *GatewayServer) {
		gs.registry = registry
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
	errUpstreamType        = errors.DefineUnimplemented("upstream_type_not_implemented", "upstream `{name}` not implemented")
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
		Component:                 c,
		ctx:                       log.NewContextWithField(c.Context(), "namespace", "gatewayserver"),
		config:                    conf,
		requireRegisteredGateways: conf.RequireRegisteredGateways,
		forward:                   forward,
		upstreamHandlers:          make(map[string]upstream.Handler),
	}
	for _, opt := range opts {
		opt(gs)
	}

	ctx, cancel := context.WithCancel(gs.Context())
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
		Config config.MQTT
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
		for _, endpoint := range []component.Endpoint{
			component.NewTCPEndpoint(version.Config.Listen, "MQTT"),
			component.NewTLSEndpoint(version.Config.ListenTLS, "MQTT"),
		} {
			if endpoint.Address() == "" {
				continue
			}
			l, err := gs.ListenTCP(endpoint.Address())
			var lis net.Listener
			if err == nil {
				lis, err = endpoint.Listen(l)
			}
			if err != nil {
				return nil, errListenFrontend.WithCause(err).WithAttributes(
					"address", endpoint.Address(),
					"protocol", endpoint.Protocol(),
				)
			}
			mqtt.Start(ctx, gs, lis, version.Format, endpoint.Protocol())
		}
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("gatewayserver"))
	bsCtx := ctx
	if conf.BasicStation.FallbackFrequencyPlanID != "" {
		bsCtx = frequencyplans.WithFallbackID(ctx, conf.BasicStation.FallbackFrequencyPlanID)
	}

	bsWebServer := basicstationlns.New(bsCtx, gs, conf.BasicStation.UseTrafficTLSAddress)
	for _, endpoint := range []component.Endpoint{
		component.NewTCPEndpoint(conf.BasicStation.Listen, "Basic Station"),
		component.NewTLSEndpoint(conf.BasicStation.ListenTLS, "Basic Station", component.WithNextProtos("h2", "http/1.1")),
	} {
		if endpoint.Address() == "" {
			continue
		}
		l, err := gs.ListenTCP(endpoint.Address())
		var lis net.Listener
		if err == nil {
			lis, err = endpoint.Listen(l)
		}
		if err != nil {
			return nil, errListenFrontend.WithCause(err).WithAttributes(
				"address", endpoint.Address(),
				"protocol", endpoint.Protocol(),
			)
		}
		go func() error {
			return http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bsWebServer.ServeHTTP(w, r)
			}))
		}()
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsGs", cluster.HookName, c.ClusterAuthUnaryHook())

	for name, prefix := range gs.forward {
		if name == "" {
			name = "cluster"
		}
		var handler upstream.Handler
		switch name {
		case "cluster":
			handler = ns.NewHandler(ctx, c, prefix)
		default:
			return nil, errInvalidUpstreamName.WithAttributes("name", name)
		}
		if err := handler.Setup(ctx); err != nil {
			return nil, errSetupUpstream.WithCause(err).WithAttributes("name", name)
		}
		gs.upstreamHandlers[name] = handler
	}

	c.RegisterGRPC(gs)
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
	if ids.IsZero() {
		return nil, ttnpb.GatewayIdentifiers{}, errEmptyIdentifiers
	}
	if ids.GatewayID == "" {
		registry, err := gs.getRegistry(ctx, nil)
		if err != nil {
			return nil, ttnpb.GatewayIdentifiers{}, err
		}
		extIDs, err := registry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			EUI: *ids.EUI,
		}, gs.WithClusterAuth())
		if err == nil {
			ids = *extIDs
		} else if errors.IsNotFound(err) {
			if gs.requireRegisteredGateways {
				return nil, ttnpb.GatewayIdentifiers{}, errGatewayEUINotRegistered.WithAttributes("eui", *ids.EUI).WithCause(err)
			}
			ids.GatewayID = fmt.Sprintf("eui-%v", strings.ToLower(ids.EUI.String()))
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
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:conn:%s", events.NewCorrelationID()))

	var err error
	var callOpt grpc.CallOption
	callOpt, err = rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials())
	if errors.IsUnauthenticated(err) {
		callOpt = gs.WithClusterAuth()
	} else if err != nil {
		return nil, err
	}
	registry, err := gs.getRegistry(ctx, &ids)
	if err != nil {
		return nil, err
	}
	gtw, err := registry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: ids,
		FieldMask: pbtypes.FieldMask{
			Paths: []string{
				"antennas",
				"downlink_path_constraint",
				"enforce_duty_cycle",
				"frequency_plan_id",
				"frequency_plan_ids",
				"location_public",
				"schedule_anytime_delay",
				"schedule_downlink_late",
			},
		},
	}, callOpt)
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
			LocationPublic:         false,
		}
	} else if err != nil {
		return nil, err
	}

	conn, err := io.NewConnection(ctx, frontend, gtw, gs.FrequencyPlans, gtw.EnforceDutyCycle, gtw.ScheduleAnytimeDelay)
	if err != nil {
		return nil, err
	}
	gs.connections.Store(uid, conn)
	registerGatewayConnect(ctx, ids)
	logger.Info("Connected")
	go gs.handleUpstream(conn)

	for name, handler := range gs.upstreamHandlers {
		go func(name string, handler upstream.Handler) {
			logger := log.FromContext(ctx).WithField("handler", name)
			if err := handler.ConnectGateway(conn.Context(), ids, conn); err != nil {
				logger.WithError(err).Warn("Failed to connect gateway on upstream")
			}
		}(name, handler)
	}
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
	errHostHandle      = errors.Define("host_handle", "host `{host}` failed to handle message")
)

var (
	// maxUpstreamHandlers is the maximum number of goroutines per gateway connection to handle upstream messages.
	maxUpstreamHandlers = int32(1 << 5)
	// upstreamHandlerIdleTimeout is the duration after which an idle upstream handler stops to save resources.
	upstreamHandlerIdleTimeout = (1 << 7) * time.Millisecond
	// upstreamHandlerBusyTimeout is the duration after traffic gets dropped if all upstream handlers are busy.
	upstreamHandlerBusyTimeout = (1 << 6) * time.Millisecond
)

type upstreamHost struct {
	name     string
	handler  func(ids *ttnpb.EndDeviceIdentifiers) upstream.Handler
	callOpts []grpc.CallOption
	handlers int32
	handleWg sync.WaitGroup
	handleCh chan upstreamItem
}

type upstreamItem struct {
	ctx  context.Context
	val  interface{}
	host *upstreamHost
}

func (gs *GatewayServer) handleUpstream(conn *io.Connection) {
	ctx := conn.Context()
	logger := log.FromContext(ctx)
	defer func() {
		ids := conn.Gateway().GatewayIdentifiers
		gs.connections.Delete(unique.ID(ctx, ids))
		registerGatewayDisconnect(ctx, ids)
		logger.Info("Disconnected")
	}()

	handleFn := func(host *upstreamHost) {
		defer host.handleWg.Done()
		defer atomic.AddInt32(&host.handlers, -1)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(upstreamHandlerIdleTimeout):
				return
			case item := <-host.handleCh:
				ctx := item.ctx
				switch msg := item.val.(type) {
				case *ttnpb.GatewayUplinkMessage:
					registerReceiveUplink(ctx, conn.Gateway(), msg.UplinkMessage, host.name)
					drop := func(ids ttnpb.EndDeviceIdentifiers, err error) {
						logger := logger.WithError(err)
						if ids.JoinEUI != nil {
							logger = logger.WithField("join_eui", *ids.JoinEUI)
						}
						if ids.DevEUI != nil && !ids.DevEUI.IsZero() {
							logger = logger.WithField("dev_eui", *ids.DevEUI)
						}
						if ids.DevAddr != nil && !ids.DevAddr.IsZero() {
							logger = logger.WithField("dev_addr", *ids.DevAddr)
						}
						logger.Debug("Drop message")
						registerDropUplink(ctx, conn.Gateway(), msg.UplinkMessage, host.name, err)
					}
					ids, err := lorawan.GetUplinkMessageIdentifiers(msg.RawPayload)
					if err != nil {
						drop(ttnpb.EndDeviceIdentifiers{}, err)
						break
					}
					handler := item.host.handler(&ids)
					if handler == nil {
						break
					}
					if err := handler.HandleUplink(ctx, conn.Gateway().GatewayIdentifiers, ids, msg); err != nil {
						drop(ids, errHostHandle.WithCause(err).WithAttributes("host", item.host.name))
						break
					}
					registerForwardUplink(ctx, conn.Gateway(), msg.UplinkMessage, item.host.name)
				case *ttnpb.GatewayStatus:
					registerReceiveStatus(ctx, conn.Gateway(), msg)
					handler := item.host.handler(nil)
					if handler == nil {
						break
					}
					if err := handler.HandleStatus(ctx, conn.Gateway().GatewayIdentifiers, msg); err != nil {
						registerForwardStatus(ctx, conn.Gateway(), msg, item.host.name)
					} else {
						registerDropStatus(ctx, conn.Gateway(), msg, item.host.name, err)
					}
				}
			}
		}
	}

	hosts := make([]*upstreamHost, 0, len(gs.upstreamHandlers))
	for name, handler := range gs.upstreamHandlers {
		handler := handler
		passDevAddr := func(prefixes []types.DevAddrPrefix, devAddr types.DevAddr) bool {
			for _, prefix := range prefixes {
				if devAddr.HasPrefix(prefix) {
					return true
				}
			}
			return false
		}
		hosts = append(hosts, &upstreamHost{
			name: name,
			handler: func(ids *ttnpb.EndDeviceIdentifiers) upstream.Handler {
				if ids != nil && ids.DevAddr != nil && !passDevAddr(handler.GetDevAddrPrefixes(), *ids.DevAddr) {
					return nil
				}
				return handler
			},
			handleCh: make(chan upstreamItem),
		})
	}

	for _, host := range hosts {
		defer host.handleWg.Wait()
	}

	for {
		ctx := ctx
		var val interface{}
		select {
		case <-ctx.Done():
			return
		case msg := <-conn.Up():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:uplink:%s", events.NewCorrelationID()))
			msg.CorrelationIDs = append(msg.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
			val = msg
		case msg := <-conn.Status():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:status:%s", events.NewCorrelationID()))
			val = msg
		case msg := <-conn.TxAck():
			ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("gs:tx_ack:%s", events.NewCorrelationID()))
			msg.CorrelationIDs = append(msg.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
			if msg.Result == ttnpb.TxAcknowledgment_SUCCESS {
				registerSuccessDownlink(ctx, conn.Gateway())
			} else {
				registerFailDownlink(ctx, conn.Gateway(), msg)
			}
			// TODO: Send Tx acknowledgement upstream (https://github.com/TheThingsNetwork/lorawan-stack/issues/76)
			continue
		}
		for _, host := range hosts {
			item := upstreamItem{
				ctx:  ctx,
				val:  val,
				host: host,
			}
			select {
			case host.handleCh <- item:
			default:
				if atomic.LoadInt32(&host.handlers) < maxUpstreamHandlers {
					atomic.AddInt32(&host.handlers, 1)
					host.handleWg.Add(1)
					go handleFn(host)
				}
				select {
				case host.handleCh <- item:
				case <-time.After(upstreamHandlerBusyTimeout):
					logger.WithField("name", host.name).Warn("Upstream handler busy, drop message")
					switch msg := val.(type) {
					case *ttnpb.UplinkMessage:
						registerFailUplink(ctx, conn.Gateway(), msg, host.name)
					case *ttnpb.GatewayStatus:
						registerFailStatus(ctx, conn.Gateway(), msg, host.name)
					}
				}
			}
		}
	}
}

// GetFrequencyPlans gets the frequency plans by the gateway identifiers.
func (gs *GatewayServer) GetFrequencyPlans(ctx context.Context, ids ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error) {
	var err error
	var callOpt grpc.CallOption
	callOpt, err = rpcmetadata.WithForwardedAuth(ctx, gs.AllowInsecureForCredentials())
	if errors.IsUnauthenticated(err) {
		callOpt = gs.WithClusterAuth()
	} else if err != nil {
		return nil, err
	}
	registry, err := gs.getRegistry(ctx, &ids)
	if err != nil {
		return nil, err
	}
	gtw, err := registry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIdentifiers: ids,
		FieldMask:          pbtypes.FieldMask{Paths: []string{"frequency_plan_ids"}},
	}, callOpt)
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
	return gs.ClaimIDs(ctx, ids)
}

// UnclaimDownlink releases the claim of the downlink path for the given gateway.
func (gs *GatewayServer) UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error {
	return gs.UnclaimIDs(ctx, ids)
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
