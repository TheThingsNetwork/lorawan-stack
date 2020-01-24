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

	pbtypes "github.com/gogo/protobuf/types"
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

// New returns a new Packet Broker Agent.
func New(c *component.Component, conf *Config, opts ...Option) (*Agent, error) {
	a := &Agent{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "packetbroker/agent"),

		dataPlaneAddress:     conf.DataPlaneAddress,
		netID:                conf.NetID,
		homeNetworkTLSConfig: conf.HomeNetwork.TLS,
		subscriptionGroup:    conf.SubscriptionGroup,
		devAddrPrefixes:      conf.DevAddrPrefixes,
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

var (
	errNoPHYPayload       = errors.DefineFailedPrecondition("no_phy_payload", "no PHYPayload in message")
	errMessageIdentifiers = errors.DefineFailedPrecondition("message_identifiers", "invalid message identifiers")
	errDataRate           = errors.DefineFailedPrecondition("data_rate", "invalid data rate index `{index}` in region `{region}`")
	errWrapUplinkTokens   = errors.DefineAborted("wrap_uplink_tokens", "wrap uplink tokens")
)

func (a *Agent) handleUplink(ctx context.Context, msg *packetbroker.UplinkMessage) error {
	receivedAt := time.Now()
	logger := log.FromContext(ctx)

	phyPayload := msg.GetPhyPayload().GetPlain()
	if len(phyPayload) == 0 {
		return errNoPHYPayload
	}

	ids, err := lorawan.GetUplinkMessageIdentifiers(phyPayload)
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
	logger.Debug("Received plain message")

	dataRate, ok := fromPBDataRate(msg.GatewayRegion, int(msg.DataRateIndex))
	if !ok {
		return errDataRate.WithAttributes(
			"index", msg.DataRateIndex,
			"region", msg.GatewayRegion,
		)
	}

	uplinkToken, err := wrapUplinkTokens(msg.ForwarderUplinkToken, msg.GatewayUplinkToken)
	if err != nil {
		return errWrapUplinkTokens.WithCause(err)
	}

	up := &ttnpb.UplinkMessage{
		RawPayload: phyPayload,
		Settings: ttnpb.TxSettings{
			DataRate:      dataRate,
			DataRateIndex: ttnpb.DataRateIndex(msg.DataRateIndex),
			Frequency:     msg.Frequency,
		},
		ReceivedAt:     receivedAt,
		CorrelationIDs: events.CorrelationIDsFromContext(ctx),
	}

	var receiveTime *time.Time
	if t, err := pbtypes.TimestampFromProto(msg.GatewayReceiveTime); err == nil {
		receiveTime = &t
	}
	if gtwMd := msg.GatewayMetadata; gtwMd != nil {
		if md := gtwMd.GetPlainLocalization().GetTerrestrial(); md != nil {
			for _, ant := range md.Antennas {
				up.RxMetadata = append(up.RxMetadata, &ttnpb.RxMetadata{
					GatewayIdentifiers:    cluster.PacketBrokerGatewayID,
					AntennaIndex:          ant.Index,
					Time:                  receiveTime,
					FineTimestamp:         ant.FineTimestamp.GetValue(),
					RSSI:                  ant.SignalQuality.GetChannelRssi(),
					SignalRSSI:            ant.SignalQuality.GetSignalRssi(),
					RSSIStandardDeviation: ant.SignalQuality.GetRssiStandardDeviation().GetValue(),
					SNR:                   ant.SignalQuality.GetSnr(),
					FrequencyOffset:       ant.SignalQuality.GetFrequencyOffset(),
					Location:              fromPBLocation(ant.Location),
					UplinkToken:           uplinkToken,
				})
			}
		} else if md := gtwMd.GetPlainSignalQuality().GetTerrestrial(); md != nil {
			for _, ant := range md.Antennas {
				up.RxMetadata = append(up.RxMetadata, &ttnpb.RxMetadata{
					GatewayIdentifiers:    cluster.PacketBrokerGatewayID,
					AntennaIndex:          ant.Index,
					Time:                  receiveTime,
					RSSI:                  ant.Value.GetChannelRssi(),
					SignalRSSI:            ant.Value.GetSignalRssi(),
					RSSIStandardDeviation: ant.Value.GetRssiStandardDeviation().GetValue(),
					SNR:                   ant.Value.GetSnr(),
					FrequencyOffset:       ant.Value.GetFrequencyOffset(),
					UplinkToken:           uplinkToken,
				})
			}
		}
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
