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
	packetbroker "go.packetbroker.org/api/v2"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/square/go-jose.v2"
)

const upstreamBufferSize = 64

// TenantContextFiller fills the parent context based on the tenant ID.
type TenantContextFiller func(parent context.Context, tenantID string) (context.Context, error)

// Agent implements the Packet Broker Agent component, acting as Home Network.
//
// Agent exposes the GsPba and NsPba interfaces for forwarding uplink and subscribing to uplink.
type Agent struct {
	*component.Component
	ctx context.Context

	dataPlaneAddress  string
	netID             types.NetID
	tenantID          string
	forwarderConfig   ForwarderConfig
	homeNetworkConfig HomeNetworkConfig
	subscriptionGroup string
	devAddrPrefixes   []types.DevAddrPrefix

	tenantContextFillers []TenantContextFiller

	upstreamCh chan *ttnpb.GatewayUplinkMessage

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
	ctx := log.NewContextWithField(c.Context(), "namespace", "packetbroker/agent")
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
		forwarderConfig:   conf.Forwarder,
		homeNetworkConfig: conf.HomeNetwork,
		subscriptionGroup: conf.SubscriptionGroup,
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
	a.grpc.nsPba = &ttnpb.UnimplementedNsPbaServer{}
	a.grpc.gsPba = &gsPbaServer{
		upstreamCh: a.upstreamCh,
	}
	for _, opt := range opts {
		opt(a)
	}

	if a.forwarderConfig.Enable {
		c.RegisterTask(c.Context(), "pb_forward_uplink", a.forwardUplink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
	}
	if a.homeNetworkConfig.Enable {
		c.RegisterTask(c.Context(), "pb_subscribe_uplink", a.subscribeUplink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
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

func (a *Agent) forwardUplink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbroker/agent",
		"forwarder_net_id", a.netID,
		"forwarder_id", a.forwarderConfig.ID,
		"forwarder_tenant_id", a.tenantID,
	))

	conn, err := a.dialContext(ctx, a.forwarderConfig.TLS, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:%s", events.NewCorrelationID()))

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
				if atomic.LoadInt32(&workers) < a.forwarderConfig.WorkerPool.MaximumWorkerCount {
					wg.Add(1)
					atomic.AddInt32(&workers, 1)
					go func() {
						if err := a.runForwarder(ctx, conn, uplinkCh); err != nil {
							logger.WithError(err).Warn("Forwarder stopped")
						}
						wg.Done()
						atomic.AddInt32(&workers, -1)
					}()
				}
				select {
				case uplinkCh <- msg:
				case <-time.After(a.forwarderConfig.WorkerPool.BusyTimeout):
					logger.Warn("Forwarder busy, drop message")
				}
			}
		}
	}
}

func (a *Agent) runForwarder(ctx context.Context, conn *grpc.ClientConn, uplinkCh <-chan *ttnpb.GatewayUplinkMessage) error {
	logger := log.FromContext(ctx)
	client := packetbroker.NewRouterForwarderDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(a.forwarderConfig.WorkerPool.IdleTimeout):
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
				ForwarderId:       a.forwarderConfig.ID,
				ForwarderTenantId: a.tenantID,
				Message:           msg,
			}
			ctx, cancel := context.WithCancel(ctx)
			progress, err := client.Publish(ctx, req)
			if err != nil {
				logger.WithError(err).Warn("Failed to publish uplink message")
				cancel()
				continue
			}
			status, err := progress.Recv()
			if err != nil {
				logger.WithError(err).Warn("Failed to receive published uplink message status")
				cancel()
				continue
			}
			logger.WithFields(log.Fields(
				"message_id", status.Id,
				"state", status.State,
			)).Debug("Publish uplink message state changed")
			cancel()
		}
	}
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
				JoinRequest: &packetbroker.RoutingFilter_JoinRequest{},
			},
		},
	}
	if a.forwarderConfig.Enable {
		// Add self to blacklist to avoid looping traffic via Packet Broker.
		forwardersBlacklist := &packetbroker.RoutingFilter_ForwarderBlacklist{
			ForwarderBlacklist: &packetbroker.ForwarderIdentifiers{
				List: []*packetbroker.ForwarderIdentifier{
					{
						NetId:       a.netID.MarshalNumber(),
						ForwarderId: a.forwarderConfig.ID,
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
		"namespace", "packetbroker/agent",
		"home_network_net_id", a.netID,
		"home_network_tenant_id", a.tenantID,
	))

	conn, err := a.dialContext(ctx, a.homeNetworkConfig.TLS, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:%s", events.NewCorrelationID()))

	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeHomeNetworkRequest{
		HomeNetworkNetId:    a.netID.MarshalNumber(),
		HomeNetworkTenantId: a.tenantID,
		Filters:             a.getSubscriptionFilters(),
		Group:               a.subscriptionGroup,
	})
	if err != nil {
		return err
	}
	logger := log.FromContext(ctx)
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
			if atomic.LoadInt32(&workers) < a.homeNetworkConfig.WorkerPool.MaximumWorkerCount {
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
			case <-time.After(a.homeNetworkConfig.WorkerPool.BusyTimeout):
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
		case <-time.After(a.homeNetworkConfig.WorkerPool.IdleTimeout):
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

	msg, err := fromPBUplink(ctx, up.Message, receivedAt)
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
