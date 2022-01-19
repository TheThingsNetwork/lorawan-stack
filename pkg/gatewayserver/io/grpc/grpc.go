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

package grpc

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc/peer"
)

// Option represents an option for the gRPC frontend.
type Option interface {
	apply(*impl)
}

type optionFunc func(*impl)

func (f optionFunc) apply(i *impl) { f(i) }

// WithMQTTConfigProvider sets the MQTT configuration provider for the gRPC frontend.
func WithMQTTConfigProvider(provider config.MQTTConfigProvider) Option {
	return optionFunc(func(i *impl) {
		i.mqttConfigProvider = provider
	})
}

// WithMQTTV2ConfigProvider sets the MQTT v2 configuration provider for the gRPC frontend.
func WithMQTTV2ConfigProvider(provider config.MQTTConfigProvider) Option {
	return optionFunc(func(i *impl) {
		i.mqttv2ConfigProvider = provider
	})
}

type impl struct {
	server               io.Server
	mqttConfigProvider   config.MQTTConfigProvider
	mqttv2ConfigProvider config.MQTTConfigProvider

	tokens io.DownlinkTokens
}

// New returns a new gRPC frontend.
func New(server io.Server, opts ...Option) ttnpb.GtwGsServer {
	i := &impl{server: server}
	for _, opt := range opts {
		opt.apply(i)
	}
	return i
}

func (*impl) Protocol() string            { return "grpc" }
func (*impl) SupportsDownlinkClaim() bool { return false }

var errConnect = errors.Define("connect", "failed to connect gateway `{gateway_uid}`")

// LinkGateway links the gateway to the Gateway Server.
func (s *impl) LinkGateway(link ttnpb.GtwGs_LinkGatewayServer) error {
	ctx := log.NewContextWithField(link.Context(), "namespace", "gatewayserver/io/grpc")

	ids := ttnpb.GatewayIdentifiers{
		GatewayId: rpcmetadata.FromIncomingContext(ctx).ID,
	}
	ctx, ids, err := s.server.FillGatewayContext(ctx, ids)
	if err != nil {
		return err
	}
	if err = s.server.ValidateGatewayID(ctx, ids); err != nil {
		return err
	}
	if err = rights.RequireGateway(ctx, ids, ttnpb.Right_RIGHT_GATEWAY_LINK); err != nil {
		return err
	}

	if peer, ok := peer.FromContext(ctx); ok {
		ctx = log.NewContextWithField(ctx, "remote_addr", peer.Addr.String())
	}
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "gateway_uid", uid)
	logger := log.FromContext(ctx)
	conn, err := s.server.Connect(ctx, s, ids)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return errConnect.WithCause(err).WithAttributes("gateway_uid", uid)
	}

	go func() {
		resource := ratelimit.GatewayUpResource(ctx, ids)
		for {
			if err := ratelimit.Require(s.server.RateLimiter(), resource); err != nil {
				logger.WithError(err).Warn("Terminate connection")
				conn.Disconnect(err)
				return
			}
			msg, err := link.Recv()
			if err != nil {
				if !errors.IsCanceled(err) {
					logger.WithError(err).Warn("Link failed")
				}
				conn.Disconnect(err)
				return
			}
			now := time.Now()

			logger.WithFields(log.Fields(
				"has_status", msg.GatewayStatus != nil,
				"uplink_count", len(msg.UplinkMessages),
			)).Debug("Received message")

			for _, up := range io.UniqueUplinkMessagesByRSSI(msg.UplinkMessages) {
				up.ReceivedAt = ttnpb.ProtoTimePtr(now)
				if err := conn.HandleUp(up, nil); err != nil {
					logger.WithError(err).Warn("Failed to handle uplink message")
				}
			}
			if msg.GatewayStatus != nil {
				if err := conn.HandleStatus(msg.GatewayStatus); err != nil {
					logger.WithError(err).Warn("Failed to handle status message")
				}
			}
			if ack := msg.TxAcknowledgment; ack != nil {
				if token, ok := s.tokens.ParseTokenFromCorrelationIDs(ack.GetCorrelationIds()); ok {
					if down, _, ok := s.tokens.Get(token, time.Now()); ok {
						ack.DownlinkMessage = down
					}
				}
				if err := conn.HandleTxAck(ack); err != nil {
					logger.WithError(err).Warn("Failed to handle Tx acknowledgment")
				}
			}
		}
	}()

	for {
		select {
		case <-conn.Context().Done():
			return conn.Context().Err()
		case down := <-conn.Down():
			token := s.tokens.Next(down, time.Now())
			down.CorrelationIds = append(down.CorrelationIds, s.tokens.FormatCorrelationID(token))
			msg := &ttnpb.GatewayDown{
				DownlinkMessage: down,
			}
			logger.Info("Send downlink message")
			if err := link.Send(msg); err != nil {
				logger.WithError(err).Warn("Failed to send message")
				conn.Disconnect(err)
				return err
			}
		}
	}
}

func (s *impl) GetConcentratorConfig(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ConcentratorConfig, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/grpc")

	ids := ttnpb.GatewayIdentifiers{
		GatewayId: rpcmetadata.FromIncomingContext(ctx).ID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := rights.RequireGateway(ctx, ids, ttnpb.Right_RIGHT_GATEWAY_LINK); err != nil {
		return nil, err
	}
	fps, err := s.server.GetFrequencyPlans(ctx, ids)
	if err != nil {
		return nil, err
	}
	// TODO: Support multiple frequency plans (https://github.com/TheThingsNetwork/lorawan-stack/issues/1820)
	var fp *frequencyplans.FrequencyPlan
	for _, v := range fps {
		fp = v
		break
	}
	return fp.ToConcentratorConfig()
}

var errNoMQTTConfigProvider = errors.DefineUnimplemented("no_configuration_provider", "no MQTT configuration provider available")

func getMQTTConnectionProvider(ctx context.Context, ids *ttnpb.GatewayIdentifiers, provider config.MQTTConfigProvider) (*ttnpb.MQTTConnectionInfo, error) {
	if err := rights.RequireGateway(ctx, *ids, ttnpb.Right_RIGHT_GATEWAY_INFO); err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, errNoMQTTConfigProvider.New()
	}
	config, err := provider.GetMQTTConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &ttnpb.MQTTConnectionInfo{
		PublicAddress:    config.PublicAddress,
		PublicTlsAddress: config.PublicTLSAddress,
		Username:         unique.ID(ctx, *ids),
	}, nil
}

func (s *impl) GetMQTTConnectionInfo(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.MQTTConnectionInfo, error) {
	return getMQTTConnectionProvider(ctx, ids, s.mqttConfigProvider)
}

func (s *impl) GetMQTTV2ConnectionInfo(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.MQTTConnectionInfo, error) {
	return getMQTTConnectionProvider(ctx, ids, s.mqttv2ConfigProvider)
}
