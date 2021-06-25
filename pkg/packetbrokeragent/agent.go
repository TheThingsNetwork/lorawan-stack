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
	routingpb "go.packetbroker.org/api/routing"
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

// TenantIDExtractor extracts the tenant ID from the context.
type TenantIDExtractor func(ctx context.Context) string

// RegistrationInfo contains information about a Packet Broker registration.
type RegistrationInfo struct {
	Name          string
	DevAddrBlocks []*ttnpb.PacketBrokerDevAddrBlock
	ContactInfo   []*ttnpb.ContactInfo
	Listed        bool
}

// RegistrationInfoExtractor extracts registration information from the context.
type RegistrationInfoExtractor func(ctx context.Context, homeNetworkClusterID string) (*RegistrationInfo, error)

type uplinkMessage struct {
	context.Context
	*packetbroker.UplinkMessage
}

type downlinkMessage struct {
	context.Context
	*agentUplinkToken
	*packetbroker.DownlinkMessage
}

// Agent implements the Packet Broker Agent component, acting as Home Network.
//
// Agent exposes the Pba interface for Packet Broker registration and routing policy management.
// Agent also exposes the GsPba and NsPba interfaces for forwarding uplink and subscribing to uplink.
type Agent struct {
	*component.Component
	ctx context.Context

	dataPlaneAddress string
	netID            types.NetID
	subscriptionTenantID,
	clusterID,
	homeNetworkClusterID string
	dialOptions       func(context.Context) ([]grpc.DialOption, error)
	forwarderConfig   ForwarderConfig
	homeNetworkConfig HomeNetworkConfig
	devAddrPrefixes   []types.DevAddrPrefix

	tenantContextFillers      []TenantContextFiller
	tenantIDExtractor         TenantIDExtractor
	registrationInfoExtractor RegistrationInfoExtractor

	upstreamCh   chan *uplinkMessage
	downstreamCh chan *downlinkMessage

	grpc struct {
		pba   ttnpb.PbaServer
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

// WithTenantIDExtractor returns an Option that configures the Agent to use the given tenant ID extractor.
func WithTenantIDExtractor(extractor TenantIDExtractor) Option {
	return func(a *Agent) {
		a.tenantIDExtractor = extractor
	}
}

// WithRegistrationInfo returns an Option that configures the Agent to use the given registration information extractor.
func WithRegistrationInfo(extractor RegistrationInfoExtractor) Option {
	return func(a *Agent) {
		a.registrationInfoExtractor = extractor
	}
}

var (
	errNetID              = errors.DefineFailedPrecondition("net_id", "invalid NetID `{net_id}`")
	errAuthenticationMode = errors.DefineFailedPrecondition("authentication_mode", "invalid authentication mode `{mode}`")
	errTokenKey           = errors.DefineFailedPrecondition("token_key", "invalid token key", "length")
)

// New returns a new Packet Broker Agent.
func New(c *component.Component, conf *Config, opts ...Option) (*Agent, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "packetbrokeragent")
	logger := log.FromContext(ctx)

	var devAddrPrefixes []types.DevAddrPrefix
	if conf.HomeNetwork.Enable {
		devAddrPrefixes = append(devAddrPrefixes, conf.HomeNetwork.DevAddrPrefixes...)
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

	var dialOptions func(context.Context) ([]grpc.DialOption, error)
	switch mode := conf.AuthenticationMode; mode {
	case "tls":
		dialOptions = func(ctx context.Context) ([]grpc.DialOption, error) {
			tlsConfig, err := c.GetTLSClientConfig(ctx)
			if err != nil {
				return nil, err
			}
			if conf.TLS.Source == "key-vault" {
				conf.TLS.KeyVault.KeyVault = c.KeyVault
			}
			if err = conf.TLS.ApplyTo(tlsConfig); err != nil {
				return nil, err
			}
			return []grpc.DialOption{
				grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
			}, nil
		}
	case "oauth2":
		dialOptions = func(ctx context.Context) ([]grpc.DialOption, error) {
			tlsConfig, err := c.GetTLSClientConfig(ctx)
			if err != nil {
				return nil, err
			}
			res := make([]grpc.DialOption, 2)
			res[0] = grpc.WithPerRPCCredentials(rpcclient.OAuth2(
				ctx,
				conf.OAuth2.TokenURL,
				conf.OAuth2.ClientID,
				conf.OAuth2.ClientSecret,
				[]string{"networks"},
				conf.Insecure,
			))
			if conf.Insecure {
				res[1] = grpc.WithInsecure()
			} else {
				res[1] = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
			}
			return res, nil
		}
	}

	homeNetworkClusterID := conf.HomeNetworkClusterID
	if homeNetworkClusterID == "" {
		homeNetworkClusterID = conf.ClusterID
	}
	a := &Agent{
		Component:            c,
		ctx:                  ctx,
		dataPlaneAddress:     conf.DataPlaneAddress,
		netID:                conf.NetID,
		subscriptionTenantID: conf.TenantID,
		clusterID:            conf.ClusterID,
		homeNetworkClusterID: homeNetworkClusterID,
		dialOptions:          dialOptions,
		forwarderConfig:      conf.Forwarder,
		homeNetworkConfig:    conf.HomeNetwork,
		devAddrPrefixes:      devAddrPrefixes,
		tenantIDExtractor: func(_ context.Context) string {
			return conf.TenantID
		},
		registrationInfoExtractor: func(_ context.Context, homeNetworkClusterID string) (*RegistrationInfo, error) {
			blocks := make([]*ttnpb.PacketBrokerDevAddrBlock, len(devAddrPrefixes))
			for i, p := range devAddrPrefixes {
				blocks[i] = &ttnpb.PacketBrokerDevAddrBlock{
					DevAddrPrefix: &ttnpb.DevAddrPrefix{
						DevAddr: &p.DevAddr,
						Length:  uint32(p.Length),
					},
					HomeNetworkClusterID: homeNetworkClusterID,
				}
			}
			contactInfo := make([]*ttnpb.ContactInfo, 0, 2)
			if adminContact := conf.Registration.AdministrativeContact.ContactInfo(ttnpb.CONTACT_TYPE_OTHER); adminContact != nil {
				contactInfo = append(contactInfo, adminContact)
			}
			if techContact := conf.Registration.TechnicalContact.ContactInfo(ttnpb.CONTACT_TYPE_TECHNICAL); techContact != nil {
				contactInfo = append(contactInfo, techContact)
			}
			return &RegistrationInfo{
				Name:          conf.Registration.Name,
				DevAddrBlocks: blocks,
				ContactInfo:   contactInfo,
				Listed:        conf.Registration.Listed,
			}, nil
		},
	}

	if a.forwarderConfig.Enable {
		a.upstreamCh = make(chan *uplinkMessage, upstreamBufferSize)
		if a.forwarderConfig.WorkerPool.Limit <= 1 {
			a.forwarderConfig.WorkerPool.Limit = 2
		}
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
		a.downstreamCh = make(chan *downlinkMessage, downstreamBufferSize)
		if a.homeNetworkConfig.WorkerPool.Limit <= 1 {
			a.homeNetworkConfig.WorkerPool.Limit = 2
		}
	}

	for _, opt := range opts {
		opt(a)
	}
	if a.dialOptions == nil {
		return nil, errAuthenticationMode.WithAttributes("mode", conf.AuthenticationMode)
	}

	if a.forwarderConfig.Enable || a.homeNetworkConfig.Enable {
		iamConn, err := a.dialContext(ctx, conf.IAMAddress)
		if err != nil {
			return nil, err
		}
		cpConn, err := a.dialContext(ctx, conf.ControlPlaneAddress)
		if err != nil {
			return nil, err
		}
		a.grpc.pba = &pbaServer{
			Agent:   a,
			iamConn: iamConn,
			cpConn:  cpConn,
		}
	} else {
		a.grpc.pba = &disabledServer{}
	}
	if a.forwarderConfig.Enable {
		mapperConn, err := a.dialContext(ctx, conf.MapperAddress)
		if err != nil {
			return nil, err
		}
		a.grpc.gsPba = &gsPbaServer{
			netID:               a.netID,
			clusterID:           a.clusterID,
			config:              a.forwarderConfig,
			messageEncrypter:    a,
			contextDecoupler:    a,
			tenantIDExtractor:   a.tenantIDExtractor,
			frequencyPlansStore: a.FrequencyPlans,
			upstreamCh:          a.upstreamCh,
			mapperConn:          mapperConn,
		}
	} else {
		a.grpc.gsPba = &disabledServer{}
	}
	if a.homeNetworkConfig.Enable {
		a.grpc.nsPba = &nsPbaServer{
			contextDecoupler: a,
			downstreamCh:     a.downstreamCh,
		}
	} else {
		a.grpc.nsPba = &disabledServer{}
	}

	newTaskConfig := func(id string, fn component.TaskFunc) *component.TaskConfig {
		return &component.TaskConfig{
			Context: c.Context(),
			ID:      id,
			Func:    fn,
			Restart: component.TaskRestartOnFailure,
			Backoff: component.DialTaskBackoffConfig,
		}
	}
	if a.forwarderConfig.Enable && a.dataPlaneAddress != "" {
		c.RegisterTask(newTaskConfig("pb_publish_uplink", a.publishUplink))
		c.RegisterTask(newTaskConfig("pb_subscribe_downlink", a.subscribeDownlink))
	}
	if a.homeNetworkConfig.Enable && a.dataPlaneAddress != "" {
		c.RegisterTask(newTaskConfig("pb_subscribe_uplink", a.subscribeUplink))
		c.RegisterTask(newTaskConfig("pb_publish_downlink", a.publishDownlink))
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
	ttnpb.RegisterPbaServer(s, a.grpc.pba)
	ttnpb.RegisterNsPbaServer(s, a.grpc.nsPba)
	ttnpb.RegisterGsPbaServer(s, a.grpc.gsPba)
}

// RegisterHandlers registers gRPC handlers.
func (a *Agent) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterPbaHandler(a.Context(), s, conn)
}

func (a *Agent) dialContext(ctx context.Context, target string, dialOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	baseDialOpts, err := a.dialOptions(ctx)
	if err != nil {
		return nil, err
	}
	defaultOpts := rpcclient.DefaultDialOptions(ctx)
	opts := make([]grpc.DialOption, 0, len(defaultOpts)+len(baseDialOpts)+len(dialOpts))
	opts = append(opts, defaultOpts...)
	opts = append(opts, baseDialOpts...)
	opts = append(opts, dialOpts...)
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
		"forwarder_cluster_id", a.clusterID,
	))

	conn, err := a.dialContext(ctx, a.dataPlaneAddress,
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:up:%s", events.NewCorrelationID()))

	logger := log.FromContext(ctx)
	logger.Info("Connected as Forwarder")

	uplinkCh := make(chan *uplinkMessage)
	defer close(uplinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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

func (a *Agent) runForwarderPublisher(ctx context.Context, conn *grpc.ClientConn, uplinkCh <-chan *uplinkMessage) error {
	client := routingpb.NewForwarderDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case up := <-uplinkCh:
			tenantID := a.tenantIDExtractor(up.Context)
			msg := up.UplinkMessage
			ctx := log.NewContextWithFields(ctx, log.Fields(
				"forwarder_tenant_id", tenantID,
			))
			logger := log.FromContext(ctx)
			req := &routingpb.PublishUplinkMessageRequest{
				ForwarderNetId:     a.netID.MarshalNumber(),
				ForwarderTenantId:  tenantID,
				ForwarderClusterId: a.clusterID,
				Message:            msg,
			}
			ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
			res, err := client.Publish(ctx, req)
			if err != nil {
				logger.WithError(err).Warn("Failed to publish uplink message")
			} else {
				logger.WithField("message_id", res.Id).Debug("Published uplink message")
				pbaMetrics.uplinkForwarded.WithLabelValues(ctx,
					a.netID.String(), tenantID, a.clusterID,
				).Inc()
			}
			cancel()
		}
	}
}

func (a *Agent) subscribeDownlink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"forwarder_net_id", a.netID,
		"forwarder_tenant_id", a.subscriptionTenantID,
		"forwarder_cluster_id", a.clusterID,
		"group", a.clusterID,
	))

	conn, err := a.dialContext(ctx, a.dataPlaneAddress,
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:down:%s", events.NewCorrelationID()))

	client := routingpb.NewForwarderDataClient(conn)
	stream, err := client.Subscribe(ctx, &routingpb.SubscribeForwarderRequest{
		ForwarderNetId:     a.netID.MarshalNumber(),
		ForwarderTenantId:  a.subscriptionTenantID,
		ForwarderClusterId: a.clusterID,
		Group:              a.clusterID,
	})
	if err != nil {
		return err
	}
	logger := log.FromContext(ctx)
	logger.Info("Subscribed as Forwarder")

	var (
		reportCh       = make(chan *packetbroker.DownlinkMessageDeliveryStateChange)
		reportWorkerCh = make(chan *packetbroker.DownlinkMessageDeliveryStateChange)
		downlinkCh     = make(chan *packetbroker.RoutedDownlinkMessage)
	)
	defer close(downlinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		var workers int32
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case rep := <-reportCh:
				select {
				case reportWorkerCh <- rep:
				default:
					if int(atomic.LoadInt32(&workers)) < a.forwarderConfig.WorkerPool.Limit/2 {
						wg.Add(1)
						worker := atomic.AddInt32(&workers, 1)
						logger := logger.WithField("worker", worker)
						go func() {
							if err := a.reportDownlinkMessageDeliveryState(log.NewContext(ctx, logger), client, reportWorkerCh); err != nil {
								logger.WithError(err).Warn("Forwarder downlink reporter worker stopped")
							}
							wg.Done()
							atomic.AddInt32(&workers, -1)
						}()
					}
					select {
					case reportWorkerCh <- rep:
					case <-time.After(workerBusyTimeout):
						logger.Warn("Forwarder downlink reporter workers busy, drop message")
					}
				}
			}
		}
	}()

	var workers int32
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case downlinkCh <- msg:
		default:
			if int(atomic.LoadInt32(&workers)) < a.forwarderConfig.WorkerPool.Limit/2 {
				wg.Add(1)
				worker := atomic.AddInt32(&workers, 1)
				logger := logger.WithField("worker", worker)
				go func() {
					if err := a.handleDownlink(log.NewContext(ctx, logger), downlinkCh, reportCh); err != nil {
						logger.WithError(err).Warn("Forwarder downlink handler stopped")
					}
					wg.Done()
					atomic.AddInt32(&workers, -1)
				}()
			}
			select {
			case downlinkCh <- msg:
			case <-time.After(workerBusyTimeout):
				logger.Warn("Forwarder downlink handler busy, drop message")
			}
		}
	}
}

func (a *Agent) reportDownlinkMessageDeliveryState(ctx context.Context, client routingpb.ForwarderDataClient, reportCh <-chan *packetbroker.DownlinkMessageDeliveryStateChange) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case rep := <-reportCh:
			var homeNetworkNetID types.NetID
			homeNetworkNetID.UnmarshalNumber(rep.HomeNetworkNetId)
			var forwarderNetID types.NetID
			forwarderNetID.UnmarshalNumber(rep.ForwarderNetId)
			logger := log.FromContext(ctx).WithFields(log.Fields(
				"message_id", rep.Id,
				"from_home_network_net_id", homeNetworkNetID,
				"from_home_network_tenant_id", rep.HomeNetworkTenantId,
				"from_home_network_cluster_id", rep.HomeNetworkClusterId,
			))
			ctx := log.NewContext(ctx, logger)
			logger.Debug("Received downlink message delivery state change")

			pbaMetrics.downlinkStateReported.WithLabelValues(ctx,
				forwarderNetID.String(), rep.ForwarderTenantId, rep.ForwarderClusterId,
			).Inc()

			_, err := client.ReportDownlinkMessageDeliveryState(ctx, &routingpb.DownlinkMessageDeliveryStateChangeRequest{
				StateChange: rep,
			})
			if err != nil {
				logger.WithError(err).Warn("Failed to report downlink message delivery state change")
			}
		}
	}
}

func (a *Agent) handleDownlink(ctx context.Context, downlinkCh <-chan *packetbroker.RoutedDownlinkMessage, reportCh chan<- *packetbroker.DownlinkMessageDeliveryStateChange) error {
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
				"from_home_network_tenant_id", down.HomeNetworkTenantId,
				"from_home_network_cluster_id", down.HomeNetworkClusterId,
			))
			if err := a.handleDownlinkMessage(ctx, down, reportCh); err != nil {
				log.FromContext(ctx).WithError(err).Debug("Failed to handle incoming downlink message")
			}
		}
	}
}

func (a *Agent) handleDownlinkMessage(ctx context.Context, down *packetbroker.RoutedDownlinkMessage, reportCh chan<- *packetbroker.DownlinkMessageDeliveryStateChange) (err error) {
	receivedAt := time.Now()
	logger := log.FromContext(ctx)

	report := &packetbroker.DownlinkMessageDeliveryStateChange{
		Id:                   down.Id,
		HomeNetworkNetId:     down.HomeNetworkNetId,
		HomeNetworkTenantId:  down.HomeNetworkTenantId,
		HomeNetworkClusterId: down.HomeNetworkClusterId,
		ForwarderNetId:       down.ForwarderNetId,
		ForwarderTenantId:    down.ForwarderTenantId,
		ForwarderClusterId:   down.ForwarderClusterId,
		DownlinkToken:        down.Message.DownlinkToken,
		State:                packetbroker.MessageDeliveryState_PROCESSED,
	}
	defer func() {
		if err != nil && report.Result == nil {
			report.Result = &packetbroker.DownlinkMessageDeliveryStateChange_Error{
				Error: packetbroker.DownlinkMessageProcessingError_DOWNLINK_INTERNAL,
			}
		}
		select {
		case <-ctx.Done():
		case reportCh <- report:
		case <-time.After(workerBusyTimeout):
			logger.Warn("Forwarder downlink reporter enqueuer busy, drop message")
		}
	}()

	for _, filler := range a.tenantContextFillers {
		var err error
		if ctx, err = filler(ctx, down.ForwarderTenantId); err != nil {
			logger.WithError(err).Warn("Failed to fill context for incoming downlink message")
			return err
		}
	}

	uid, msg, err := fromPBDownlink(ctx, down.Message, receivedAt, a.forwarderConfig)
	if err != nil {
		logger.WithError(err).Warn("Failed to convert incoming downlink message")
		return err
	}
	ids, err := unique.ToGatewayID(uid)
	if err != nil {
		logger.WithField("gateway_uid", uid).WithError(err).Warn("Failed to get gateway identifier")
		return err
	}

	req := msg.GetRequest()
	pairs := []interface{}{
		"gateway_uid", uid,
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

	var homeNetworkNetID types.NetID
	homeNetworkNetID.UnmarshalNumber(down.HomeNetworkNetId)
	var forwarderNetID types.NetID
	forwarderNetID.UnmarshalNumber(down.ForwarderNetId)
	pbaMetrics.downlinkReceived.WithLabelValues(ctx,
		homeNetworkNetID.String(), down.HomeNetworkTenantId, down.HomeNetworkClusterId,
		forwarderNetID.String(), down.ForwarderTenantId, down.ForwarderClusterId,
	).Inc()

	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, &ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to get Gateway Server peer")
		return err
	}

	res, err := ttnpb.NewNsGsClient(conn).ScheduleDownlink(ctx, msg, a.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Failed to schedule downlink")
		report.Result = &packetbroker.DownlinkMessageDeliveryStateChange_Error{
			Error: packetbroker.DownlinkMessageProcessingError_DOWNLINK_UNKNOWN_ERROR,
		}
		return err
	}
	report.Result = &packetbroker.DownlinkMessageDeliveryStateChange_Success{
		Success: &packetbroker.DownlinkMessageDeliveryStateChange_TransmissionResult{
			Rx1: res.Rx1,
			Rx2: res.Rx2,
		},
	}

	logger.WithFields(log.Fields(
		"transmission_delay", res.Delay,
		"transmit_at", time.Now().Add(res.Delay),
		"rx1", res.Rx1,
		"rx2", res.Rx2,
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
	return []*packetbroker.RoutingFilter{
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
}

func (a *Agent) subscribeUplink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"home_network_net_id", a.netID,
		"home_network_tenant_id", a.subscriptionTenantID,
		"home_network_cluster_id", a.homeNetworkClusterID,
		"group", a.clusterID,
	))
	logger := log.FromContext(ctx)

	conn, err := a.dialContext(ctx, a.dataPlaneAddress,
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:up:%s", events.NewCorrelationID()))

	filters := a.getSubscriptionFilters()
	for i, f := range filters {
		logger := logger
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
			formatted := make([]string, 0, len(msg.Mac.DevAddrPrefixes))
			for _, prefix := range msg.Mac.DevAddrPrefixes {
				formatted = append(formatted, fmt.Sprintf("DevAddr: %08X/%d", prefix.Value, prefix.Length))
			}
			message = formatted
		}
		if messageType != "" {
			logger = logger.WithFields(log.Fields(
				"message_type", messageType,
				"message", message,
			))
		}
		logger.WithField("i", i).Debug("Configured filter")
	}

	client := routingpb.NewHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &routingpb.SubscribeHomeNetworkRequest{
		HomeNetworkNetId:     a.netID.MarshalNumber(),
		HomeNetworkTenantId:  a.subscriptionTenantID,
		HomeNetworkClusterId: a.homeNetworkClusterID,
		Filters:              filters,
		Group:                a.clusterID,
	})
	if err != nil {
		return err
	}
	logger.Info("Subscribed as Home Network")

	var (
		reportCh       = make(chan *packetbroker.UplinkMessageDeliveryStateChange)
		reportWorkerCh = make(chan *packetbroker.UplinkMessageDeliveryStateChange)
		uplinkCh       = make(chan *packetbroker.RoutedUplinkMessage)
	)
	defer close(uplinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var workers int32
		for {
			select {
			case <-ctx.Done():
				return
			case rep := <-reportCh:
				select {
				case reportWorkerCh <- rep:
				default:
					if int(atomic.LoadInt32(&workers)) < a.homeNetworkConfig.WorkerPool.Limit/2 {
						wg.Add(1)
						worker := atomic.AddInt32(&workers, 1)
						logger := logger.WithField("worker", worker)
						go func() {
							if err := a.reportUplinkMessageDeliveryState(log.NewContext(ctx, logger), client, reportWorkerCh); err != nil {
								logger.WithError(err).Warn("Home Network uplink reporter worker stopped")
							}
							wg.Done()
							atomic.AddInt32(&workers, -1)
						}()
					}
					select {
					case reportWorkerCh <- rep:
					case <-time.After(workerBusyTimeout):
						logger.Warn("Home Network uplink reporter workers busy, drop message")
					}
				}
			}
		}
	}()

	var workers int32
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case uplinkCh <- msg:
		default:
			if int(atomic.LoadInt32(&workers)) < a.homeNetworkConfig.WorkerPool.Limit/2 {
				wg.Add(1)
				worker := atomic.AddInt32(&workers, 1)
				logger := logger.WithField("worker", worker)
				go func() {
					if err := a.handleUplink(log.NewContext(ctx, logger), uplinkCh, reportCh); err != nil {
						logger.WithError(err).Warn("Home Network uplink handler stopped")
					}
					wg.Done()
					atomic.AddInt32(&workers, -1)
				}()
			}
			select {
			case uplinkCh <- msg:
			case <-time.After(workerBusyTimeout):
				logger.Warn("Home Network uplink handler busy, drop message")
			}
		}
	}
}

func (a *Agent) reportUplinkMessageDeliveryState(ctx context.Context, client routingpb.HomeNetworkDataClient, reportCh <-chan *packetbroker.UplinkMessageDeliveryStateChange) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case rep := <-reportCh:
			var forwarderNetID types.NetID
			forwarderNetID.UnmarshalNumber(rep.ForwarderNetId)
			var homeNetworkNetID types.NetID
			homeNetworkNetID.UnmarshalNumber(rep.HomeNetworkNetId)
			logger := log.FromContext(ctx).WithFields(log.Fields(
				"message_id", rep.Id,
				"from_forwarder_net_id", forwarderNetID,
				"from_forwarder_tenant_id", rep.ForwarderTenantId,
				"from_forwarder_cluster_id", rep.ForwarderClusterId,
			))
			ctx := log.NewContext(ctx, logger)
			logger.Debug("Received uplink message delivery state change")

			pbaMetrics.uplinkStateReported.WithLabelValues(ctx,
				homeNetworkNetID.String(), rep.HomeNetworkTenantId, rep.HomeNetworkClusterId,
			).Inc()

			_, err := client.ReportUplinkMessageDeliveryState(ctx, &routingpb.UplinkMessageDeliveryStateChangeRequest{
				StateChange: rep,
			})
			if err != nil {
				logger.WithError(err).Warn("Failed to report uplink message delivery state change")
			}
		}
	}
}

func (a *Agent) handleUplink(ctx context.Context, uplinkCh <-chan *packetbroker.RoutedUplinkMessage, reportCh chan<- *packetbroker.UplinkMessageDeliveryStateChange) error {
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
				"from_forwarder_tenant_id", up.ForwarderTenantId,
				"from_forwarder_cluster_id", up.ForwarderClusterId,
			))
			if err := a.handleUplinkMessage(ctx, up, reportCh); err != nil {
				log.FromContext(ctx).WithError(err).Debug("Failed to handle incoming uplink message")
			}
		}
	}
}

var errMessageIdentifiers = errors.DefineFailedPrecondition("message_identifiers", "invalid message identifiers")

func (a *Agent) handleUplinkMessage(ctx context.Context, up *packetbroker.RoutedUplinkMessage, reportCh chan<- *packetbroker.UplinkMessageDeliveryStateChange) (err error) {
	receivedAt := time.Now()
	logger := log.FromContext(ctx)

	report := &packetbroker.UplinkMessageDeliveryStateChange{
		Id:                   up.Id,
		HomeNetworkNetId:     up.HomeNetworkNetId,
		HomeNetworkTenantId:  up.HomeNetworkTenantId,
		HomeNetworkClusterId: up.HomeNetworkClusterId,
		ForwarderNetId:       up.ForwarderNetId,
		ForwarderTenantId:    up.ForwarderTenantId,
		ForwarderClusterId:   up.ForwarderClusterId,
		ForwarderUplinkToken: up.Message.ForwarderUplinkToken,
		State:                packetbroker.MessageDeliveryState_PROCESSED,
	}
	defer func() {
		if err != nil && report.Error == nil {
			report.Error = &packetbroker.UplinkMessageProcessingErrorValue{
				Value: packetbroker.UplinkMessageProcessingError_UPLINK_INTERNAL,
			}
		}
		select {
		case <-ctx.Done():
		case reportCh <- report:
		case <-time.After(workerBusyTimeout):
			logger.Warn("Home Network uplink reporter enqueuer busy, drop message")
		}
	}()

	if err := a.decryptUplink(ctx, up.Message); err != nil {
		logger.WithError(err).Warn("Failed to decrypt message")
		return err
	}
	logger.Debug("Received uplink message")

	ids, err := lorawan.GetUplinkMessageIdentifiers(up.Message.PhyPayload.GetPlain())
	if err != nil {
		return errMessageIdentifiers.New()
	}

	if ids.JoinEui != nil {
		logger = logger.WithField("join_eui", *ids.JoinEui)
	}
	if ids.DevEui != nil && !ids.DevEui.IsZero() {
		logger = logger.WithField("dev_eui", *ids.DevEui)
	}
	if ids.DevAddr != nil && !ids.DevAddr.IsZero() {
		logger = logger.WithField("dev_addr", *ids.DevAddr)
	}

	msg, err := fromPBUplink(ctx, up, receivedAt, a.homeNetworkConfig.IncludeHops)
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

	var forwarderNetID types.NetID
	forwarderNetID.UnmarshalNumber(up.ForwarderNetId)
	var homeNetworkNetID types.NetID
	homeNetworkNetID.UnmarshalNumber(up.HomeNetworkNetId)
	pbaMetrics.uplinkReceived.WithLabelValues(ctx,
		forwarderNetID.String(), up.ForwarderTenantId, up.ForwarderClusterId,
		homeNetworkNetID.String(), up.HomeNetworkTenantId, up.HomeNetworkClusterId,
	).Inc()

	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, &ids)
	if err != nil {
		return err
	}

	_, err = ttnpb.NewGsNsClient(conn).HandleUplink(ctx, msg, a.WithClusterAuth())
	if err != nil {
		reportError := packetbroker.UplinkMessageProcessingError_UPLINK_UNKNOWN_ERROR
		if errors.IsNotFound(err) {
			reportError = packetbroker.UplinkMessageProcessingError_MATCH_SESSION
		}
		report.Error = &packetbroker.UplinkMessageProcessingErrorValue{
			Value: reportError,
		}
		return err
	}

	return nil
}

func (a *Agent) publishDownlink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "packetbrokeragent",
		"home_network_net_id", a.netID,
		"home_network_cluster_id", a.homeNetworkClusterID,
	))

	conn, err := a.dialContext(ctx, a.dataPlaneAddress,
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:down:%s", events.NewCorrelationID()))

	logger := log.FromContext(ctx)
	logger.Info("Connected as Home Network")

	downlinkCh := make(chan *downlinkMessage)
	defer close(downlinkCh)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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

func (a *Agent) runHomeNetworkPublisher(ctx context.Context, conn *grpc.ClientConn, downlinkCh <-chan *downlinkMessage) error {
	client := routingpb.NewHomeNetworkDataClient(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workerIdleTimeout):
			return nil
		case down := <-downlinkCh:
			tenantID := a.tenantIDExtractor(down.Context)
			msg, token := down.DownlinkMessage, down.agentUplinkToken
			ctx := log.NewContextWithFields(ctx, log.Fields(
				"forwarder_net_id", token.ForwarderNetID,
				"forwarder_tenant_id", token.ForwarderTenantID,
				"forwarder_cluster_id", token.ForwarderClusterID,
				"home_network_tenant_id", tenantID,
			))
			logger := log.FromContext(ctx)
			req := &routingpb.PublishDownlinkMessageRequest{
				HomeNetworkNetId:     a.netID.MarshalNumber(),
				HomeNetworkTenantId:  tenantID,
				HomeNetworkClusterId: a.homeNetworkClusterID,
				ForwarderNetId:       token.ForwarderNetID.MarshalNumber(),
				ForwarderTenantId:    token.ForwarderTenantID,
				ForwarderClusterId:   token.ForwarderClusterID,
				Message:              msg,
			}
			ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
			res, err := client.Publish(ctx, req)
			if err != nil {
				logger.WithError(err).Warn("Failed to publish downlink message")
			} else {
				logger.WithField("message_id", res.Id).Debug("Published downlink message")
				pbaMetrics.downlinkForwarded.WithLabelValues(ctx,
					a.netID.String(), tenantID, a.homeNetworkClusterID,
					token.ForwarderNetID.String(), token.ForwarderTenantID, token.ForwarderClusterID,
				).Inc()
			}
			cancel()
		}
	}
}
