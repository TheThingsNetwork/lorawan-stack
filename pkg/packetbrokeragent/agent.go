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
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	packetbroker "go.packetbroker.org/api/v1beta2"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// EndDeviceIdentifiersContextFiller fills the parent context based on the end device identifiers.
type EndDeviceIdentifiersContextFiller func(parent context.Context, ids ttnpb.EndDeviceIdentifiers) (context.Context, error)

// Agent implements the Packet Broker Agent component, acting as Home Network.
//
// Agent exposes the NsGs interface for scheduling downlink.
type Agent struct {
	*component.Component
	ctx context.Context

	dataPlaneAddress     string
	netID                types.NetID
	homeNetworkTLSConfig TLSConfig
	subscriptionGroup    string
	devAddrPrefixes      []types.DevAddrPrefix

	contextFillers []EndDeviceIdentifiersContextFiller

	grpc struct {
		nsGs ttnpb.NsGsServer
	}
}

// Option configures Agent.
type Option func(*Agent)

// WithEndDeviceIdentifiersContextFiller returns an Option that appends the given filler to the end device identifiers
// context fillers.
func WithEndDeviceIdentifiersContextFiller(filler EndDeviceIdentifiersContextFiller) Option {
	return func(a *Agent) {
		a.contextFillers = append(a.contextFillers, filler)
	}
}

var errNetID = errors.DefineFailedPrecondition("net_id", "invalid NetID `{net_id}`")

// New returns a new Packet Broker Agent.
func New(c *component.Component, conf *Config, opts ...Option) (*Agent, error) {
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
		ctx:       log.NewContextWithField(c.Context(), "namespace", "packetbroker/agent"),

		dataPlaneAddress:     conf.DataPlaneAddress,
		netID:                conf.NetID,
		homeNetworkTLSConfig: conf.HomeNetwork.TLS,
		subscriptionGroup:    conf.SubscriptionGroup,
		devAddrPrefixes:      devAddrPrefixes,
	}
	a.grpc.nsGs = &ttnpb.UnimplementedNsGsServer{}
	for _, opt := range opts {
		opt(a)
	}

	if conf.HomeNetwork.Enable {
		c.RegisterTask(c.Context(), "pb_subscribe_uplink", a.subscribeUplink, component.TaskRestartOnFailure, component.TaskBackoffDial...)
	}

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
	ttnpb.RegisterNsGsServer(s, a.grpc.nsGs)
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

func (a *Agent) getSubscriptionFilters() []*packetbroker.RoutingFilter {
	devAddrPrefixes := make([]*packetbroker.RoutingFilter_MACPayload_DevAddrPrefix, len(a.devAddrPrefixes))
	for i, p := range a.devAddrPrefixes {
		devAddrPrefixes[i] = &packetbroker.RoutingFilter_MACPayload_DevAddrPrefix{
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
					EuiPrefixes: []*packetbroker.RoutingFilter_JoinRequest_EUIPrefixes{},
				},
			},
		},
	}
}

func (a *Agent) subscribeUplink(ctx context.Context) error {
	ctx = log.NewContextWithField(ctx, "namespace", "packetbroker/agent")

	conn, err := a.dialContext(ctx, a.homeNetworkTLSConfig, a.dataPlaneAddress)
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:conn:%s", events.NewCorrelationID()))

	client := packetbroker.NewRouterHomeNetworkDataClient(conn)
	stream, err := client.Subscribe(ctx, &packetbroker.SubscribeHomeNetworkRequest{
		HomeNetworkNetId: a.netID.MarshalNumber(),
		Filters:          a.getSubscriptionFilters(),
	})
	if err != nil {
		return err
	}
	log.FromContext(ctx).Debug("Subscribed as Home Network")

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		up := msg.Message
		if up == nil {
			continue
		}
		ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("pba:uplink:%s", msg.Id))
		var forwarderNetID types.NetID
		forwarderNetID.UnmarshalNumber(msg.ForwarderNetId)
		ctx = log.NewContextWithFields(ctx, log.Fields(
			"message_id", msg.Id,
			"from_forwarder_net_id", forwarderNetID,
			"from_forwarder_id", msg.ForwarderId,
		))
		if err := a.handleUplink(ctx, up); err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to handle uplink message")
		}
	}
}

var errNoPHYPayload = errors.DefineFailedPrecondition("no_phy_payload", "no PHYPayload in message")

func (a *Agent) decryptUplink(ctx context.Context, msg *packetbroker.UplinkMessage) error {
	// TODO: Obtain KEK, decrypt PHYPayload and gateway metadata (https://github.com/TheThingsIndustries/lorawan-stack/issues/1919).
	if msg.PhyPayload.GetPlain() == nil {
		return errNoPHYPayload
	}
	return nil
}

var errMessageIdentifiers = errors.DefineFailedPrecondition("message_identifiers", "invalid message identifiers")

func (a *Agent) handleUplink(ctx context.Context, msg *packetbroker.UplinkMessage) error {
	receivedAt := time.Now()
	logger := log.FromContext(ctx)

	if err := a.decryptUplink(ctx, msg); err != nil {
		logger.WithError(err).Debug("Failed to decrypt message")
		return err
	}
	logger.Debug("Received uplink message")

	ids, err := lorawan.GetUplinkMessageIdentifiers(msg.PhyPayload.GetPlain())
	if err != nil {
		return errMessageIdentifiers
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

	up, err := fromPBUplink(ctx, msg, receivedAt)
	if err != nil {
		return err
	}

	for _, filler := range a.contextFillers {
		var err error
		if ctx, err = filler(ctx, ids); err != nil {
			return err
		}
	}
	conn, err := a.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return err
	}
	_, err = ttnpb.NewGsNsClient(conn).HandleUplink(ctx, up, a.WithClusterAuth())
	return err
}
