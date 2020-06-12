// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/square/go-jose.v2"
)

const (
	upstreamBufferSize   = 1 << 6
	downstreamBufferSize = 1 << 5

	// publishMessageTimeout defines the timeout for publishing messages.
	publishMessageTimeout = 3 * time.Second
)

// TenantContextFiller fills the parent context based on the tenant ID.
type TenantContextFiller func(parent context.Context, tenantID string) (context.Context, error)

// Agent implements the Packet Broker Agent component, acting as Home Network.
//
// Agent exposes the GsPba and NsPba interfaces for forwarding uplink and subscribing to uplink.
type Agent struct {
	*component.Component
	ctx context.Context

	dataPlaneAddress string
	netID            types.NetID
	tenantID,
	clusterID string
	tlsConfig         TLSConfig
	forwarderConfig   ForwarderConfig
	homeNetworkConfig HomeNetworkConfig
	devAddrPrefixes   []types.DevAddrPrefix

	tenantContextFillers []TenantContextFiller

	upstreamCh   chan *ttnpb.GatewayUplinkMessage
	downstreamCh chan *ttnpb.DownlinkMessage

	grpc struct {
		nsPba ttnpb.NsPbaServer
		gsPba ttnpb.GsPbaServer
	}
}

// Option configures Agent.
type Option func(*Agent)

// WithTenantContextFiller returns an Option that appends the given filler to the end device identifiers
// context fillers.
func WithTenantContextFiller(filler TenantContextFiller) Option {
	return func(a *Agent) {
		a.tenantContextFillers = append(a.tenantContextFillers, filler)
	}
}

var (
	errNetID    = errors.DefineFailedPrecondition("net_id", "invalid NetID `{net_id}`")
	errTokenKey = errors.DefineFailedPrecondition("token_key", "invalid token key", "length")
)

// New returns a new Packet Broker Agent.
func New(c *component.Component, conf *Config, opts ...Option) (*Agent, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "packetbrokeragent")
	logger := log.FromContext(ctx)

	var devAddrPrefixes []types.DevAddrPrefix
	if hn := conf.HomeNetwork; hn.Enable {
		devAddrPrefixes = append(devAddrPrefixes, hn.DevAddrPrefixes...)
		if len(devAddrPrefixes) == 0 {
			devAddr, err := types.NewDevAddr(conf.NetID, nil)
			if err != nil {
				return nil, errNetID.WithAttributes("net_id", conf.NetID).WithCause(err)
			}
			devAddrPrefix := types.DevAddrPrefix{
				DevAddr: devAddr,
				Length:  uint8(conf.NetID.IDBits()),
			}
			devAddrPrefixes = append(devAddrPrefixes, devAddrPrefix)
		}
	}

	a := &Agent{
		Component: c,
		ctx:       ctx,

		dataPlaneAddress:  conf.DataPlaneAddress,
		netID:             conf.NetID,
		tenantID:          conf.TenantID,
		clusterID:         conf.ClusterID,
		tlsConfig:         conf.TLS,
		forwarderConfig:   conf.Forwarder,
		homeNetworkConfig: conf.HomeNetwork,
		devAddrPrefixes:   devAddrPrefixes,
	}
	if a.forwarderConfig.Enable {
		a.upstreamCh = make(chan *ttnpb.GatewayUplinkMessage, upstreamBufferSize)
		if len(a.forwarderConfig.TokenKey) == 0 {
			a.forwarderConfig.TokenKey = random.Bytes(16)
			logger.WithField("token_key", hex.EncodeToString(a.forwarderConfig.TokenKey)).Warn("No token key configured, generated a random one")
		}
		var (
			enc jose.ContentEncryption
			alg jose.KeyAlgorithm
		)
		switch l := len(a.forwarderConfig.TokenKey); l {
		case 16:
			enc, alg = jose.A128GCM, jose.A128GCMKW
		case 32:
			enc, alg = jose.A256GCM, jose.A256GCMKW
		default:
			return nil, errTokenKey.WithAttributes("length", l).New()
		}
		var err error
		a.forwarderConfig.TokenEncrypter, err = jose.NewEncrypter(enc, jose.Recipient{
			Algorithm: alg,
			Key:       a.forwarderConfig.TokenKey,
		}, nil)
		if err != nil {
			return nil, errTokenKey.WithCause(err)
		}
	}
	if a.homeNetworkConfig.Enable {
		a.downstreamCh = make(chan *ttnpb.DownlinkMessage, downstreamBufferSize)
	}
	a.grpc.nsPba = &nsPbaServer{
		downstreamCh: a.downstreamCh,
	}
	a.grpc.gsPba = &gsPbaServer{
		upstreamCh: a.upstreamCh,
	}
	for _, opt := range opts {
		opt(a)
	}

	if a.forwarderConfig.Enable {
		c.RegisterTask(c.Context(), "pb_publish_uplink", a.publishUplink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
		c.RegisterTask(c.Context(), "pb_subscribe_downlink", a.subscribeDownlink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
	}
	if a.homeNetworkConfig.Enable {
		c.RegisterTask(c.Context(), "pb_subscribe_uplink", a.subscribeUplink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
		c.RegisterTask(c.Context(), "pb_publish_downlink", a.publishDownlink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GsPba", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsPba", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(a)
	return a, nil
}

// Context returns the context.
func (a *Agent) Context() context.Context {
	return a.ctx
}

// Roles returns the Packet Broker Agent cluster role.
func (a *Agent) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_PACKET_BROKER_AGENT}
}

// RegisterServices registers services provided by a at s.
func (a *Agent) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsPbaServer(s, a.grpc.nsPba)
	ttnpb.RegisterGsPbaServer(s, a.grpc.gsPba)
}

// RegisterHandlers registers gRPC handlers.
func (a *Agent) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
}

func (a *Agent) dialContext(ctx context.Context, config TLSConfig, target string) (*grpc.ClientConn, error) {
	cert, err := config.loadCertificate(ctx, a.KeyVault)
	if err != nil {
		return nil, err
	}
	tlsConfig, err := a.GetTLSClientConfig(ctx, component.WithTLSCertificates(cert))
	if err != nil {
		return nil, err
	}
	opts := append(rpcclient.DefaultDialOptions(ctx),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	)
	return grpc.DialContext(ctx, target, opts...)
}

const (
	// workerIdleTimeout is the duration after which an idle worker stops to save resources.
	workerIdleTimeout = (1 << 7) * time.Millisecond
	// workerBusyTimeout is the duration after which a message is dropped if all workers are busy.
	workerBusyTimeout = (1 << 6) * time.Millisecond
)

func (a *Agent) publishUplink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"forwarder_net_id", a.netID,
		"forwarder_id", a.clusterID,
		"forwarder_tenant_id", a.tenantID,
	))

	conn, err := a.dialContext(ctx, a.tlsConfig, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:up:%s", events.NewCorrelationID()))

	logger := log.FromContext(ctx)
	logger.Info("Connected as Forwarder")

	uplinkCh := make(chan *ttnpb.GatewayUplinkMessage)
	defer close(uplinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	var workers int32

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-a.upstreamCh:
			select {
			case uplinkCh <- msg:
			default:
				if int(atomic.LoadInt32(&workers)) < a.forwarderConfig.WorkerPool.Limit {
					wg.Add(1)
					atomic.AddInt32(&workers, 1)
					go func() {
						if err := a.runForwarderPublisher(ctx, conn, uplinkCh); err != nil {
							logger.WithError(err).Warn("Forwarder publisher stopped")
						}
						wg.Done()
						atomic.AddInt32(&workers, -1)
					}()
				}
				select {
				case uplinkCh <- msg:
				case <-time.After(workerBusyTimeout):
					logger.Warn("Forwarder publisher busy, drop message")
				}
			}
		}
	}
}

func (a *Agent) runForwarderPublisher(ctx context.Context, conn *grpc.ClientConn, uplinkCh <-chan *ttnpb.GatewayUplinkMessage) error {
	logger := log.FromContext(ctx)
	client := packetbroker.NewRouterForwarderDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case up := <-uplinkCh:
			msg, err := toPBUplink(ctx, up, a.forwarderConfig)
			if err != nil {
				logger.WithError(err).Warn("Failed to convert outgoing uplink message")
				continue
			}
			if err := a.encryptUplink(ctx, msg); err != nil {
				logger.WithError(err).Warn("Failed to encrypt outgoing uplink message")
				continue
			}
			req := &packetbroker.PublishUplinkMessageRequest{
				ForwarderNetId:    a.netID.MarshalNumber(),
				ForwarderId:       a.clusterID,
				ForwarderTenantId: a.tenantID,
				Message:           msg,
			}
			ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
			res, err := client.Publish(ctx, req)
			if err != nil {
				logger.WithError(err).Warn("Failed to publish uplink message")
			} else {
				logger.WithField("message_id", res.Id).Debug("Published uplink message")
			}
			cancel()
		}
	}
}

func (a *Agent) subscribeDownlink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"forwarder_net_id", a.netID,
		"forwarder_id", a.clusterID,
		"forwarder_tenant_id", a.tenantID,
		"group", a.clusterID,
	))

	conn, err := a.dialContext(ctx, a.tlsConfig, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:down:%s", events.NewCorrelationID()))

	client := packetbroker.NewRouterForwarderDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeForwarderRequest{
		ForwarderNetId:    a.netID.MarshalNumber(),
		ForwarderId:       a.clusterID,
		ForwarderTenantId: a.tenantID,
		Group:             a.clusterID,
	})
	if err != nil {
		return err
	}
	logger := log.FromContext(ctx)
	logger.Info("Subscribed as Forwarder")

	downlinkCh := make(chan *packetbroker.RoutedDownlinkMessage)
	defer close(downlinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	var workers int32

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case downlinkCh <- msg:
		default:
			if int(atomic.LoadInt32(&workers)) < a.forwarderConfig.WorkerPool.Limit {
				wg.Add(1)
				atomic.AddInt32(&workers, 1)
				go func() {
					if err := a.handleDownlink(ctx, downlinkCh); err != nil {
						logger.WithError(err).Warn("Forwarder subscriber stopped")
					}
					wg.Done()
					atomic.AddInt32(&workers, -1)
				}()
			}
			select {
			case downlinkCh <- msg:
			case <-time.After(workerBusyTimeout):
				logger.Warn("Forwarder subscriber busy, drop message")
			}
		}
	}
}

func (a *Agent) handleDownlink(ctx context.Context, downlinkCh <-chan *packetbroker.RoutedDownlinkMessage) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case down := <-downlinkCh:
			if down.Message == nil {
				continue
			}
			ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:downlink:%s", down.Id))
			var homeNetworkNetID types.NetID
			homeNetworkNetID.UnmarshalNumber(down.HomeNetworkNetId)
			ctx = log.NewContextWithFields(ctx, log.Fields(
				"message_id", down.Id,
				"from_home_network_net_id", homeNetworkNetID,
				"from_home_network_tenant_id", down.HomeNetworkNetId,
			))
			if err := a.handleDownlinkMessage(ctx, down); err != nil {
				logger.WithError(err).Debug("Failed to handle incoming downlink message")
			}
		}
	}
}

func (a *Agent) handleDownlinkMessage(ctx context.Context, down *packetbroker.RoutedDownlinkMessage) error {
	receivedAt := time.Now()
	logger := log.FromContext(ctx)

	for _, filler := range a.tenantContextFillers {
		var err error
		if ctx, err = filler(ctx, down.ForwarderTenantId); err != nil {
			logger.WithError(err).Warn("Failed to fill context for incoming downlink message")
			return err
		}
	}

	ids, msg, err := fromPBDownlink(ctx, down.Message, receivedAt, a.forwarderConfig)
	if err != nil {
		logger.WithError(err).Warn("Failed to convert incoming uplink message")
		return err
	}

	req := msg.GetRequest()
	pairs := []interface{}{
		"gateway_uid", unique.ID(ctx, ids),
		"attempt_rx1", req.Rx1Frequency != 0,
		"attempt_rx2", req.Rx2Frequency != 0,
		"downlink_class", req.Class,
		"downlink_priority", req.Priority,
		"frequency_plan", req.FrequencyPlanID,
	}
	if req.Rx1Frequency != 0 {
		pairs = append(pairs,
			"rx1_delay", req.Rx1Delay,
			"rx1_data_rate", req.Rx1DataRateIndex,
			"rx1_frequency", req.Rx1Frequency,
		)
	}
	if req.Rx2Frequency != 0 {
		pairs = append(pairs,
			"rx2_data_rate", req.Rx2DataRateIndex,
			"rx2_frequency", req.Rx2Frequency,
		)
	}
	logger = logger.WithFields(log.Fields(pairs...))

	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, ids)
	if err != nil {
		return err
	}
	res, err := ttnpb.NewNsGsClient(conn).ScheduleDownlink(ctx, msg, a.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Failed to schedule downlink")
		return err
	}
	transmitAt := time.Now().Add(res.Delay)
	logger.WithFields(log.Fields(
		"transmission_delay", res.Delay,
		"transmit_at", transmitAt,
	)).Debug("Scheduled downlink")
	return nil
}

func (a *Agent) getSubscriptionFilters() []*packetbroker.RoutingFilter {
	devAddrPrefixes := make([]*packetbroker.DevAddrPrefix, len(a.devAddrPrefixes))
	for i, p := range a.devAddrPrefixes {
		devAddrPrefixes[i] = &packetbroker.DevAddrPrefix{
			Value:  p.DevAddr.MarshalNumber(),
			Length: uint32(p.Length),
		}
	}
	filters := []*packetbroker.RoutingFilter{
		// Subscribe to MAC payload based on DevAddrPrefixes.
		{
			Message: &packetbroker.RoutingFilter_Mac{
				Mac: &packetbroker.RoutingFilter_MACPayload{
					DevAddrPrefixes: devAddrPrefixes,
				},
			},
		},
		// Subscribe to any join-request.
		{
			Message: &packetbroker.RoutingFilter_JoinRequest_{
				JoinRequest: &packetbroker.RoutingFilter_JoinRequest{
					EuiPrefixes: []*packetbroker.RoutingFilter_JoinRequest_EUIPrefixes{{}},
				},
			},
		},
	}
	if a.forwarderConfig.Enable && a.homeNetworkConfig.BlacklistForwarder {
		// Blacklist Forwarder to avoid looping traffic via Packet Broker.
		forwardersBlacklist := &packetbroker.RoutingFilter_ForwarderBlacklist{
			ForwarderBlacklist: &packetbroker.ForwarderIdentifiers{
				List: []*packetbroker.ForwarderIdentifier{
					{
						NetId:       a.netID.MarshalNumber(),
						ForwarderId: a.clusterID,
					},
				},
			},
		}
		for _, f := range filters {
			f.Forwarders = forwardersBlacklist
		}
	}

	return filters
}

func (a *Agent) subscribeUplink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"home_network_net_id", a.netID,
		"home_network_tenant_id", a.tenantID,
		"group", a.clusterID,
	))
	logger := log.FromContext(ctx)

	conn, err := a.dialContext(ctx, a.tlsConfig, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:up:%s", events.NewCorrelationID()))

	filters := a.getSubscriptionFilters()
	for i, f := range filters {
		logger := logger
		var (
			forwardersType string
			forwarderIDs   []*packetbroker.ForwarderIdentifier
		)
		switch fwd := f.Forwarders.(type) {
		case *packetbroker.RoutingFilter_ForwarderBlacklist:
			forwardersType = "blacklist"
			forwarderIDs = fwd.ForwarderBlacklist.List
		case *packetbroker.RoutingFilter_ForwarderWhitelist:
			forwardersType = "whitelist"
			forwarderIDs = fwd.ForwarderWhitelist.List
		}
		if forwarderIDs != nil {
			formatted := make([]string, 0, len(forwarderIDs))
			for _, fwd := range forwarderIDs {
				if fwd.ForwarderId != "" {
					formatted = append(formatted, fmt.Sprintf("%s/%s", packetbroker.NetID(fwd.NetId), fwd.ForwarderId))
				} else {
					formatted = append(formatted, fmt.Sprintf("%s", packetbroker.NetID(fwd.NetId)))
				}
			}
			logger = logger.WithFields(log.Fields(
				"forwarders_type", forwardersType,
				"forwarders", formatted,
			))
		}
		if f.GatewayMetadata != nil {
			logger = logger.WithField("gateway_metadata", f.GatewayMetadata.Value)
		}
		var (
			messageType string
			message     interface{}
		)
		switch msg := f.Message.(type) {
		case *packetbroker.RoutingFilter_JoinRequest_:
			messageType = "join_request"
			formatted := make([]string, 0, len(msg.JoinRequest.EuiPrefixes))
			for _, prefixes := range msg.JoinRequest.EuiPrefixes {
				formatted = append(formatted, fmt.Sprintf("[JoinEUI: %016X/%d DevEUI: %016X/%d]", prefixes.JoinEui, prefixes.JoinEuiLength, prefixes.DevEui, prefixes.DevEuiLength))
			}
			message = formatted
		case *packetbroker.RoutingFilter_Mac:
			messageType = "mac"
			message = msg.Mac.DevAddrPrefixes
		}
		if messageType != "" {
			logger = logger.WithFields(log.Fields(
				"message_type", messageType,
				"message", message,
			))
		}
		logger.WithField("i", i).Debug("Configured filter")
	}

	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeHomeNetworkRequest{
		HomeNetworkNetId:    a.netID.MarshalNumber(),
		HomeNetworkTenantId: a.tenantID,
		Filters:             filters,
		Group:               a.clusterID,
	})
	if err != nil {
		return err
	}
	logger.Info("Subscribed as Home Network")

	uplinkCh := make(chan *packetbroker.RoutedUplinkMessage)
	defer close(uplinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	var workers int32

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case uplinkCh <- msg:
		default:
			if int(atomic.LoadInt32(&workers)) < a.homeNetworkConfig.WorkerPool.Limit {
				wg.Add(1)
				atomic.AddInt32(&workers, 1)
				go func() {
					if err := a.handleUplink(ctx, uplinkCh); err != nil {
						logger.WithError(err).Warn("Home Network subscriber stopped")
					}
					wg.Done()
					atomic.AddInt32(&workers, -1)
				}()
			}
			select {
			case uplinkCh <- msg:
			case <-time.After(workerBusyTimeout):
				logger.Warn("Home Network subscriber busy, drop message")
			}
		}
	}
}

func (a *Agent) handleUplink(ctx context.Context, uplinkCh <-chan *packetbroker.RoutedUplinkMessage) error {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case up := <-uplinkCh:
			if up.Message == nil {
				continue
			}
			ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:uplink:%s", up.Id))
			var forwarderNetID types.NetID
			forwarderNetID.UnmarshalNumber(up.ForwarderNetId)
			ctx = log.NewContextWithFields(ctx, log.Fields(
				"message_id", up.Id,
				"from_forwarder_net_id", forwarderNetID,
				"from_forwarder_id", up.ForwarderId,
				"from_forwarder_tenant_id", up.ForwarderTenantId,
			))
			if err := a.handleUplinkMessage(ctx, up); err != nil {
				logger.WithError(err).Debug("Failed to handle incoming uplink message")
			}
		}
	}
}

var errMessageIdentifiers = errors.DefineFailedPrecondition("message_identifiers", "invalid message identifiers")

func (a *Agent) handleUplinkMessage(ctx context.Context, up *packetbroker.RoutedUplinkMessage) error {
	receivedAt := time.Now()
	logger := log.FromContext(ctx)

	if err := a.decryptUplink(ctx, up.Message); err != nil {
		logger.WithError(err).Warn("Failed to decrypt message")
		return err
	}
	logger.Debug("Received uplink message")

	ids, err := lorawan.GetUplinkMessageIdentifiers(up.Message.PhyPayload.GetPlain())
	if err != nil {
		return errMessageIdentifiers.New()
	}

	if ids.JoinEUI != nil {
		logger = logger.WithField("join_eui", *ids.JoinEUI)
	}
	if ids.DevEUI != nil && !ids.DevEUI.IsZero() {
		logger = logger.WithField("dev_eui", *ids.DevEUI)
	}
	if ids.DevAddr != nil && !ids.DevAddr.IsZero() {
		logger = logger.WithField("dev_addr", *ids.DevAddr)
	}

	msg, err := fromPBUplink(ctx, up, receivedAt)
	if err != nil {
		logger.WithError(err).Warn("Failed to convert incoming uplink message")
		return err
	}

	for _, filler := range a.tenantContextFillers {
		var err error
		if ctx, err = filler(ctx, up.HomeNetworkTenantId); err != nil {
			logger.WithError(err).Warn("Failed to fill context for incoming uplink message")
			return err
		}
	}
	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return err
	}
	_, err = ttnpb.NewGsNsClient(conn).HandleUplink(ctx, msg, a.WithClusterAuth())
	return err
}

func (a *Agent) publishDownlink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"home_network_net_id", a.netID,
		"home_network_tenant_id", a.tenantID,
	))

	conn, err := a.dialContext(ctx, a.tlsConfig, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:down:%s", events.NewCorrelationID()))

	logger := log.FromContext(ctx)
	logger.Info("Connected as Home Network")

	downlinkCh := make(chan *ttnpb.DownlinkMessage)
	defer close(downlinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	var workers int32

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-a.downstreamCh:
			select {
			case downlinkCh <- msg:
			default:
				if int(atomic.LoadInt32(&workers)) < a.homeNetworkConfig.WorkerPool.Limit {
					wg.Add(1)
					atomic.AddInt32(&workers, 1)
					go func() {
						if err := a.runHomeNetworkPublisher(ctx, conn, downlinkCh); err != nil {
							logger.WithError(err).Warn("Home Network publisher stopped")
						}
						wg.Done()
						atomic.AddInt32(&workers, -1)
					}()
				}
				select {
				case downlinkCh <- msg:
				case <-time.After(workerBusyTimeout):
					logger.Warn("Home Network publisher busy, drop message")
				}
			}
		}
	}
}

func (a *Agent) runHomeNetworkPublisher(ctx context.Context, conn *grpc.ClientConn, downlinkCh <-chan *ttnpb.DownlinkMessage) error {
	logger := log.FromContext(ctx)
	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case down := <-downlinkCh:
			msg, token, err := toPBDownlink(ctx, down)
			if err != nil {
				logger.WithError(err).Warn("Failed to convert outgoing downlink message")
				continue
			}
			req := &packetbroker.PublishDownlinkMessageRequest{
				HomeNetworkNetId:    a.netID.MarshalNumber(),
				HomeNetworkTenantId: a.tenantID,
				ForwarderNetId:      token.ForwarderNetID.MarshalNumber(),
				ForwarderTenantId:   token.ForwarderTenantID,
				ForwarderId:         token.ForwarderID,
				Message:             msg,
			}
			ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
			res, err := client.Publish(ctx, req)
			if err != nil {
				logger.WithError(err).Warn("Failed to publish downlink message")
			} else {
				logger.WithField("message_id", res.Id).Debug("Published downlink message")
			}
			cancel()
		}
	}
}
