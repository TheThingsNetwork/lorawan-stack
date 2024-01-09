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

// Package packetbrokeragent contains the implementation of the Packet Broker Agent component.
package packetbrokeragent

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpctracer"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
	"gopkg.in/square/go-jose.v2"
)

const (
	// subscribeStreamCount is the number of subscription streams that are used by the Forwarder and Home Network roles.
	subscribeStreamCount = 4

	// publishMessageTimeout defines the timeout for publishing messages.
	publishMessageTimeout = 3 * time.Second
)

var (
	appendUplinkCorrelationID   = events.RegisterCorrelationIDPrefix("uplink", "pba:uplink")
	appendDownlinkCorrelationID = events.RegisterCorrelationIDPrefix("downlink", "pba:downlink")
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

// PacketBrokerClusterIDBuilder builds a Packet Broker Cluster ID from a The Things Stack Cluster ID.
type PacketBrokerClusterIDBuilder func(clusterID string) (string, error)

func literalClusterID(clusterID string) (string, error) {
	return clusterID, nil
}

// RegistrationInfoExtractor extracts registration information from the context.
type RegistrationInfoExtractor func(ctx context.Context, homeNetworkClusterID string, clusterIDBuilder PacketBrokerClusterIDBuilder) (*RegistrationInfo, error)

type uplinkMessage struct {
	context.Context
	*packetbroker.UplinkMessage
}

type downlinkMessage struct {
	context.Context
	*ttnpb.PacketBrokerAgentUplinkToken
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
	clusterIDBuilder  PacketBrokerClusterIDBuilder
	subscriptionGroup string
	authenticator     authenticator
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
	ctx := tracer.NewContextWithTracer(c.Context(), tracerNamespace)

	ctx = log.NewContextWithField(ctx, "namespace", logNamespace)
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
				Length:  uint8(32 - types.NwkAddrBits(conf.NetID)),
			}
			devAddrPrefixes = append(devAddrPrefixes, devAddrPrefix)
		}
	}

	var authenticator authenticator
	switch mode := conf.AuthenticationMode; mode {
	case "oauth2":
		var tlsConfig tlsConfigurator
		if !conf.Insecure {
			tlsConfig = c
		}
		authenticator = newOAuth2(ctx, conf.OAuth2, tlsConfig,
			conf.IAMAddress,
			conf.ControlPlaneAddress,
			conf.DataPlaneAddress,
			conf.MapperAddress,
		)
	}

	clusterIDBuilder := literalClusterID
	clusterID, err := clusterIDBuilder(conf.ClusterID)
	if err != nil {
		return nil, err
	}
	homeNetworkClusterID := conf.HomeNetworkClusterID
	if homeNetworkClusterID == "" {
		homeNetworkClusterID = conf.ClusterID
	}
	homeNetworkClusterID, err = clusterIDBuilder(homeNetworkClusterID)
	if err != nil {
		return nil, err
	}

	subscriptionGroup := conf.ClusterID
	if subscriptionGroup == "" {
		subscriptionGroup = "default"
	}

	a := &Agent{
		Component:            c,
		ctx:                  ctx,
		dataPlaneAddress:     conf.DataPlaneAddress,
		netID:                conf.NetID,
		subscriptionTenantID: conf.TenantID,
		clusterID:            clusterID,
		homeNetworkClusterID: homeNetworkClusterID,
		clusterIDBuilder:     clusterIDBuilder,
		subscriptionGroup:    subscriptionGroup,
		authenticator:        authenticator,
		forwarderConfig:      conf.Forwarder,
		homeNetworkConfig:    conf.HomeNetwork,
		devAddrPrefixes:      devAddrPrefixes,
		tenantIDExtractor: func(_ context.Context) string {
			return conf.TenantID
		},
		registrationInfoExtractor: func(_ context.Context, homeNetworkClusterID string, _ PacketBrokerClusterIDBuilder) (*RegistrationInfo, error) {
			blocks := make([]*ttnpb.PacketBrokerDevAddrBlock, len(devAddrPrefixes))
			for i, p := range devAddrPrefixes {
				blocks[i] = &ttnpb.PacketBrokerDevAddrBlock{
					DevAddrPrefix: &ttnpb.DevAddrPrefix{
						DevAddr: p.DevAddr.Bytes(),
						Length:  uint32(p.Length),
					},
					HomeNetworkClusterId: homeNetworkClusterID,
				}
			}
			contactInfo := make([]*ttnpb.ContactInfo, 0, 2)
			if adminContact := conf.Registration.AdministrativeContact.ContactInfo(ttnpb.ContactType_CONTACT_TYPE_OTHER); adminContact != nil {
				contactInfo = append(contactInfo, adminContact)
			}
			if techContact := conf.Registration.TechnicalContact.ContactInfo(ttnpb.ContactType_CONTACT_TYPE_TECHNICAL); techContact != nil {
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
		a.upstreamCh = make(chan *uplinkMessage)
		if a.forwarderConfig.WorkerPool.Limit <= 1 {
			a.forwarderConfig.WorkerPool.Limit = 2
		}
		if len(a.forwarderConfig.TokenKey) == 0 {
			a.forwarderConfig.TokenKey = random.Bytes(16)
			logger.WithField("token_key", hex.EncodeToString(a.forwarderConfig.TokenKey)).Warn("No token key configured, generated a random one")
		}
		var (
			legacyEncryption   jose.ContentEncryption
			legacyKeyAlgorithm jose.KeyAlgorithm
		)
		switch l := len(a.forwarderConfig.TokenKey); l {
		case 16:
			legacyEncryption, legacyKeyAlgorithm = jose.A128GCM, jose.A128GCMKW
		case 32:
			legacyEncryption, legacyKeyAlgorithm = jose.A256GCM, jose.A256GCMKW
		default:
			return nil, errTokenKey.WithAttributes("length", l).New()
		}
		var err error
		a.forwarderConfig.LegacyTokenEncrypter, err = jose.NewEncrypter(legacyEncryption, jose.Recipient{
			Algorithm: legacyKeyAlgorithm,
			Key:       a.forwarderConfig.TokenKey,
		}, nil)
		if err != nil {
			return nil, errTokenKey.WithCause(err)
		}
		blockCipher, err := aes.NewCipher(a.forwarderConfig.TokenKey)
		if err != nil {
			return nil, errTokenKey.WithCause(err)
		}
		a.forwarderConfig.TokenAEAD, err = cipher.NewGCM(blockCipher)
		if err != nil {
			return nil, errTokenKey.WithCause(err)
		}
	}

	if a.homeNetworkConfig.Enable {
		a.downstreamCh = make(chan *downlinkMessage)
		if a.homeNetworkConfig.WorkerPool.Limit <= 1 {
			a.homeNetworkConfig.WorkerPool.Limit = 2
		}
	}

	for _, opt := range opts {
		opt(a)
	}
	if a.authenticator == nil {
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

	getFrequencyPlanStore := func(ctx context.Context) (frequencyPlansStore, error) {
		return a.FrequencyPlansStore(ctx)
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
			frequencyPlansStore: getFrequencyPlanStore,
			upstreamCh:          a.upstreamCh,
			mapperConn:          mapperConn,
			entityRegistry:      newIS(c),
		}
	} else {
		a.grpc.gsPba = &disabledServer{}
	}
	if a.homeNetworkConfig.Enable {
		a.grpc.nsPba = &nsPbaServer{
			contextDecoupler: a,
			downstreamCh:     a.downstreamCh,
			frequencyPlans:   getFrequencyPlanStore,
		}
	} else {
		a.grpc.nsPba = &disabledServer{}
	}

	newTaskConfig := func(id string, fn task.Func) *task.Config {
		return &task.Config{
			Context: ctx,
			ID:      id,
			Func:    fn,
			Restart: task.RestartOnFailure,
			Backoff: task.DialBackoffConfig,
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

	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpctracer.TracerHook, rpctracer.UnaryTracerHook(tracerNamespace)},
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook(logNamespace)},
	} {
		for _, filter := range []string{
			"/ttn.lorawan.v3.Pba",
			"/ttn.lorawan.v3.NsPba",
			"/ttn.lorawan.v3.GsPba",
		} {
			c.GRPC.RegisterUnaryHook(filter, hook.name, hook.middleware)
		}
	}
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.NsPba", cluster.HookName, c.ClusterAuthUnaryHook())
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.GsPba", cluster.HookName, c.ClusterAuthUnaryHook())

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
	baseDialOpts, err := a.authenticator.DialOptions(ctx)
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

func (a *Agent) publishUplink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
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

	logger := log.FromContext(ctx)
	logger.Info("Connected as Forwarder")

	wp := workerpool.NewWorkerPool(workerpool.Config[*uplinkMessage]{
		Component:  a,
		Context:    ctx,
		Name:       "pb_forwarder_publish",
		Handler:    a.forwarderPublisher(conn),
		MaxWorkers: a.forwarderConfig.WorkerPool.Limit,
	})
	defer wp.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-a.upstreamCh:
			if err := wp.Publish(ctx, msg); err != nil {
				logger.WithError(err).Warn("Forwarder publisher busy, drop message")
			}
		}
	}
}

func (a *Agent) forwarderPublisher(conn *grpc.ClientConn) workerpool.Handler[*uplinkMessage] {
	client := routingpb.NewForwarderDataClient(conn)
	h := func(ctx context.Context, up *uplinkMessage) {
		tenantID := a.tenantIDExtractor(up.Context)
		msg := up.UplinkMessage
		ctx = log.NewContextWithFields(ctx, log.Fields(
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
		defer cancel()
		res, err := client.Publish(ctx, req)
		if err != nil {
			logger.WithError(err).Warn("Failed to publish uplink message")
		} else {
			logger.WithField("message_id", res.Id).Debug("Published uplink message")
			pbaMetrics.uplinkForwarded.WithLabelValues(ctx,
				a.netID.String(), tenantID, a.clusterID,
			).Inc()
		}
	}
	return h
}

func (a *Agent) subscribeDownlink(ctx context.Context) error {
	ctx = log.NewContextWithFields(ctx, log.Fields(
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

	client := routingpb.NewForwarderDataClient(conn)

	reportPool := workerpool.NewWorkerPool(workerpool.Config[*packetbroker.DownlinkMessageDeliveryStateChange]{
		Component:  a,
		Context:    ctx,
		Name:       "pb_forwarder_subscribe_report",
		Handler:    a.handleDownlinkMessageDeliveryState(client),
		MaxWorkers: a.forwarderConfig.WorkerPool.Limit / 2,
	})
	defer reportPool.Wait()

	downlinkPool := workerpool.NewWorkerPool(workerpool.Config[*packetbroker.RoutedDownlinkMessage]{
		Component:  a,
		Context:    ctx,
		Name:       "pb_forwarder_subscribe",
		Handler:    a.handleDownlink(reportPool),
		MaxWorkers: a.forwarderConfig.WorkerPool.Limit / 2,
	})
	defer downlinkPool.Wait()

	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < subscribeStreamCount; i++ {
		wg.Add(1)
		a.StartTask(&task.Config{
			Context: ctx,
			ID:      "pb_forwarder_subscribe_stream",
			Func:    a.subscribeDownlinkStream(client, downlinkPool),
			Done:    wg.Done,
			Restart: task.RestartOnFailure,
			Backoff: task.DialBackoffConfig,
		})
	}

	<-ctx.Done()
	return ctx.Err()
}

func (a *Agent) subscribeDownlinkStream(
	client routingpb.ForwarderDataClient,
	downlinkPool workerpool.WorkerPool[*packetbroker.RoutedDownlinkMessage],
) func(context.Context) error {
	f := func(ctx context.Context) (err error) {
		stream, err := client.Subscribe(ctx, &routingpb.SubscribeForwarderRequest{
			ForwarderNetId:     a.netID.MarshalNumber(),
			ForwarderTenantId:  a.subscriptionTenantID,
			ForwarderClusterId: a.clusterID,
			Group:              a.subscriptionGroup,
		})
		if err != nil {
			return err
		}
		logger := log.FromContext(ctx)
		logger.Info("Subscribed as Forwarder")
		for {
			msg, err := stream.Recv()
			if err != nil {
				return err
			}
			if err := downlinkPool.Publish(ctx, msg); err != nil {
				logger.WithError(err).Warn("Forwarder downlink handler busy, drop message")
			}
		}
	}
	return f
}

func (*Agent) handleDownlinkMessageDeliveryState(
	client routingpb.ForwarderDataClient,
) workerpool.Handler[*packetbroker.DownlinkMessageDeliveryStateChange] {
	h := func(ctx context.Context, rep *packetbroker.DownlinkMessageDeliveryStateChange) {
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
		ctx = log.NewContext(ctx, logger)
		logger.Debug("Received downlink message delivery state change")

		pbaMetrics.downlinkStateReported.WithLabelValues(ctx,
			forwarderNetID.String(), rep.ForwarderTenantId, rep.ForwarderClusterId,
		).Inc()

		ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
		defer cancel()

		_, err := client.ReportDownlinkMessageDeliveryState(ctx, &routingpb.DownlinkMessageDeliveryStateChangeRequest{
			StateChange: rep,
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to report downlink message delivery state change")
		}
	}
	return h
}

func (a *Agent) handleDownlink(
	reportPool workerpool.WorkerPool[*packetbroker.DownlinkMessageDeliveryStateChange],
) workerpool.Handler[*packetbroker.RoutedDownlinkMessage] {
	h := func(ctx context.Context, down *packetbroker.RoutedDownlinkMessage) {
		if down.Message == nil {
			return
		}
		ctx = appendDownlinkCorrelationID(ctx, down.Id)
		var homeNetworkNetID types.NetID
		homeNetworkNetID.UnmarshalNumber(down.HomeNetworkNetId)
		ctx = log.NewContextWithFields(ctx, log.Fields(
			"message_id", down.Id,
			"from_home_network_net_id", homeNetworkNetID,
			"from_home_network_tenant_id", down.HomeNetworkTenantId,
			"from_home_network_cluster_id", down.HomeNetworkClusterId,
		))
		if err := a.handleDownlinkMessage(ctx, down, reportPool); err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to handle incoming downlink message")
		}
	}
	return h
}

func (a *Agent) handleDownlinkMessage(
	ctx context.Context,
	down *packetbroker.RoutedDownlinkMessage,
	reportPool workerpool.WorkerPool[*packetbroker.DownlinkMessageDeliveryStateChange],
) (err error) {
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
	defer func(ctx context.Context) {
		if err != nil && report.Result == nil {
			report.Result = &packetbroker.DownlinkMessageDeliveryStateChange_Error{
				Error: packetbroker.DownlinkMessageProcessingError_DOWNLINK_INTERNAL,
			}
		}
		if err := reportPool.Publish(ctx, report); err != nil {
			logger.WithError(err).Warn("Forwarder downlink reporter enqueuer busy, drop message")
		}
	}(ctx)

	for _, filler := range a.tenantContextFillers {
		var err error
		if ctx, err = filler(ctx, down.ForwarderTenantId); err != nil {
			logger.WithError(err).Warn("Failed to fill context for incoming downlink message")
			return err
		}
	}

	forwarderData := forwarderAdditionalData(down.ForwarderNetId, down.ForwarderTenantId, down.ForwarderClusterId)
	uid, msg, err := fromPBDownlink(ctx, down.Message, forwarderData, receivedAt, a.forwarderConfig)
	if err != nil {
		logger.WithError(err).Warn("Failed to convert incoming downlink message")
		return err
	}
	ids, err := unique.ToGatewayID(uid)
	if err != nil {
		logger.WithField("gateway_uid", uid).WithError(err).Warn("Failed to get gateway identifier")
		return err
	}
	report.ForwarderGatewayId = toPBGatewayIdentifier(ids, a.forwarderConfig)

	req := msg.GetRequest()
	pairs := []any{
		"gateway_uid", uid,
		"attempt_rx1", req.Rx1Frequency != 0,
		"attempt_rx2", req.Rx2Frequency != 0,
		"downlink_class", req.Class,
		"downlink_priority", req.Priority,
		"frequency_plan", req.FrequencyPlanId,
	}
	if req.Rx1Frequency != 0 {
		pairs = append(pairs,
			"rx1_delay", req.Rx1Delay,
			"rx1_data_rate", req.Rx1DataRate,
			"rx1_frequency", req.Rx1Frequency,
		)
	}
	if req.Rx2Frequency != 0 {
		pairs = append(pairs,
			"rx2_data_rate", req.Rx2DataRate,
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

	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to get Gateway Server peer")
		report.Result = &packetbroker.DownlinkMessageDeliveryStateChange_Error{
			Error: packetbroker.DownlinkMessageProcessingError_GATEWAY_NOT_CONNECTED,
		}
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
		"transmit_at", time.Now().Add(ttnpb.StdDurationOrZero(res.Delay)),
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

	filters := a.getSubscriptionFilters()
	for i, f := range filters {
		logger := logger
		if f.GatewayMetadata != nil {
			logger = logger.WithField("gateway_metadata", f.GatewayMetadata.Value)
		}
		var (
			messageType string
			message     any
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

	reportPool := workerpool.NewWorkerPool(workerpool.Config[*packetbroker.UplinkMessageDeliveryStateChange]{
		Component:  a,
		Context:    ctx,
		Name:       "pb_homenetwork_subscribe_report",
		Handler:    a.handleUplinkMessageDeliveryState(client),
		MaxWorkers: a.homeNetworkConfig.WorkerPool.Limit / 2,
	})
	defer reportPool.Wait()

	uplinkPool := workerpool.NewWorkerPool(workerpool.Config[*packetbroker.RoutedUplinkMessage]{
		Component:  a,
		Context:    ctx,
		Name:       "pb_homenetwork_subscribe",
		Handler:    a.handleUplink(reportPool),
		MaxWorkers: a.homeNetworkConfig.WorkerPool.Limit / 2,
	})
	defer uplinkPool.Wait()

	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < subscribeStreamCount; i++ {
		wg.Add(1)
		a.StartTask(&task.Config{
			Context: ctx,
			ID:      "pb_homenetwork_subscribe_stream",
			Func:    a.subscribeUplinkStream(client, uplinkPool, filters),
			Done:    wg.Done,
			Restart: task.RestartOnFailure,
			Backoff: task.DialBackoffConfig,
		})
	}

	<-ctx.Done()
	return ctx.Err()
}

func (a *Agent) subscribeUplinkStream(
	client routingpb.HomeNetworkDataClient,
	uplinkPool workerpool.WorkerPool[*packetbroker.RoutedUplinkMessage],
	filters []*packetbroker.RoutingFilter,
) func(context.Context) error {
	f := func(ctx context.Context) (err error) {
		stream, err := client.Subscribe(ctx, &routingpb.SubscribeHomeNetworkRequest{
			HomeNetworkNetId:     a.netID.MarshalNumber(),
			HomeNetworkTenantId:  a.subscriptionTenantID,
			HomeNetworkClusterId: a.homeNetworkClusterID,
			Filters:              filters,
			Group:                a.subscriptionGroup,
		})
		if err != nil {
			return err
		}
		logger := log.FromContext(ctx)
		logger.Info("Subscribed as Home Network")
		for {
			msg, err := stream.Recv()
			if err != nil {
				return err
			}
			if err := uplinkPool.Publish(ctx, msg); err != nil {
				logger.WithError(err).Warn("Home Network uplink handler busy, drop message")
			}
		}
	}
	return f
}

func (*Agent) handleUplinkMessageDeliveryState(
	client routingpb.HomeNetworkDataClient,
) workerpool.Handler[*packetbroker.UplinkMessageDeliveryStateChange] {
	h := func(ctx context.Context, rep *packetbroker.UplinkMessageDeliveryStateChange) {
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
		ctx = log.NewContext(ctx, logger)
		logger.Debug("Received uplink message delivery state change")

		pbaMetrics.uplinkStateReported.WithLabelValues(ctx,
			homeNetworkNetID.String(), rep.HomeNetworkTenantId, rep.HomeNetworkClusterId,
		).Inc()

		ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
		defer cancel()

		_, err := client.ReportUplinkMessageDeliveryState(ctx, &routingpb.UplinkMessageDeliveryStateChangeRequest{
			StateChange: rep,
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to report uplink message delivery state change")
		}
	}
	return h
}

func (a *Agent) handleUplink(
	reportPool workerpool.WorkerPool[*packetbroker.UplinkMessageDeliveryStateChange],
) workerpool.Handler[*packetbroker.RoutedUplinkMessage] {
	h := func(ctx context.Context, up *packetbroker.RoutedUplinkMessage) {
		if up.Message == nil {
			return
		}
		ctx = appendUplinkCorrelationID(ctx, up.Id)
		var forwarderNetID types.NetID
		forwarderNetID.UnmarshalNumber(up.ForwarderNetId)
		ctx = log.NewContextWithFields(ctx, log.Fields(
			"message_id", up.Id,
			"from_forwarder_net_id", forwarderNetID,
			"from_forwarder_tenant_id", up.ForwarderTenantId,
			"from_forwarder_cluster_id", up.ForwarderClusterId,
		))
		if err := a.handleUplinkMessage(ctx, up, reportPool); err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to handle incoming uplink message")
		}
	}
	return h
}

var errMessageIdentifiers = errors.DefineFailedPrecondition("message_identifiers", "invalid message identifiers")

func (a *Agent) handleUplinkMessage(
	ctx context.Context,
	up *packetbroker.RoutedUplinkMessage,
	reportPool workerpool.WorkerPool[*packetbroker.UplinkMessageDeliveryStateChange],
) (err error) {
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
	defer func(ctx context.Context) {
		if err != nil && report.Error == nil {
			report.Error = &packetbroker.UplinkMessageProcessingErrorValue{
				Value: packetbroker.UplinkMessageProcessingError_UPLINK_INTERNAL,
			}
		}
		if err := reportPool.Publish(ctx, report); err != nil {
			logger.WithError(err).Warn("Home Network uplink reporter enqueuer busy, drop message")
		}
	}(ctx)

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
		logger = logger.WithField("join_eui", types.MustEUI64(ids.JoinEui))
	}
	if devEUI := types.MustEUI64(ids.DevEui).OrZero(); !devEUI.IsZero() {
		logger = logger.WithField("dev_eui", devEUI)
	}
	if devAddr := types.MustDevAddr(ids.DevAddr).OrZero(); !devAddr.IsZero() {
		logger = logger.WithField("dev_addr", devAddr)
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

	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	if err != nil {
		return err
	}

	_, err = ttnpb.NewGsNsClient(conn).HandleUplink(ctx, msg, a.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Debug("Network Server failed to handle uplink message")
		reportError := packetbroker.UplinkMessageProcessingError_UPLINK_UNKNOWN_ERROR
		switch {
		case errors.IsNotFound(err):
			reportError = packetbroker.UplinkMessageProcessingError_NOT_FOUND
		case errors.IsAlreadyExists(err):
			reportError = packetbroker.UplinkMessageProcessingError_DUPLICATE_PAYLOAD
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

	logger := log.FromContext(ctx)
	logger.Info("Connected as Home Network")

	wp := workerpool.NewWorkerPool(workerpool.Config[*downlinkMessage]{
		Component:  a,
		Context:    ctx,
		Name:       "pb_homenetwork_publish",
		Handler:    a.homeNetworkPublisher(conn),
		MaxWorkers: a.homeNetworkConfig.WorkerPool.Limit,
	})
	defer wp.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-a.downstreamCh:
			if err := wp.Publish(ctx, msg); err != nil {
				logger.WithError(err).Warn("Home Network publisher busy, drop message")
			}
		}
	}
}

func (a *Agent) homeNetworkPublisher(conn *grpc.ClientConn) workerpool.Handler[*downlinkMessage] {
	client := routingpb.NewHomeNetworkDataClient(conn)
	h := func(ctx context.Context, down *downlinkMessage) {
		tenantID := a.tenantIDExtractor(down.Context)
		msg, token := down.DownlinkMessage, down.PacketBrokerAgentUplinkToken
		forwarderNetID := types.MustNetID(token.ForwarderNetId)
		ctx = log.NewContextWithFields(ctx, log.Fields(
			"forwarder_net_id", forwarderNetID,
			"forwarder_tenant_id", token.ForwarderTenantId,
			"forwarder_cluster_id", token.ForwarderClusterId,
			"home_network_tenant_id", tenantID,
		))
		logger := log.FromContext(ctx)
		req := &routingpb.PublishDownlinkMessageRequest{
			HomeNetworkNetId:     a.netID.MarshalNumber(),
			HomeNetworkTenantId:  tenantID,
			HomeNetworkClusterId: a.homeNetworkClusterID,
			ForwarderNetId:       forwarderNetID.MarshalNumber(),
			ForwarderTenantId:    token.ForwarderTenantId,
			ForwarderClusterId:   token.ForwarderClusterId,
			Message:              msg,
		}
		ctx, cancel := context.WithTimeout(ctx, publishMessageTimeout)
		defer cancel()
		res, err := client.Publish(ctx, req)
		if err != nil {
			logger.WithError(err).Warn("Failed to publish downlink message")
		} else {
			logger.WithField("message_id", res.Id).Debug("Published downlink message")
			pbaMetrics.downlinkForwarded.WithLabelValues(ctx,
				a.netID.String(), tenantID, a.homeNetworkClusterID,
				forwarderNetID.String(), token.ForwarderTenantId, token.ForwarderClusterId,
			).Inc()
		}
	}
	return h
}
