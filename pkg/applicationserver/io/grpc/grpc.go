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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
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

// GetEndDeviceIdentifiersFunc retrieves the end device identifiers including the EUIs and DevAddr.
type GetEndDeviceIdentifiersFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDeviceIdentifiers, error)

type defaultMessageProcessor struct{}

func (p *defaultMessageProcessor) EncodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	return nil
}

func (p *defaultMessageProcessor) DecodeUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
	return nil
}

func (p *defaultMessageProcessor) DecodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	return nil
}

// SkipPayloadCryptoFunc is a function that checks if the end device should skip payload crypto operations.
type SkipPayloadCryptoFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (bool, error)

type impl struct {
	server             io.Server
	getIdentifiers     GetEndDeviceIdentifiersFunc
	mqttConfigProvider config.MQTTConfigProvider
	processor          messageprocessors.PayloadProcessor
	skipPayloadCrypto  SkipPayloadCryptoFunc
}

// WithMQTTConfigProvider sets the MQTT configuration provider for the gRPC frontend.
func WithMQTTConfigProvider(provider config.MQTTConfigProvider) Option {
	return optionFunc(func(i *impl) {
		i.mqttConfigProvider = provider
	})
}

// WithGetEndDeviceIdentifiers sets the end device identifiers retriever that will be used by the gRPC frontend.
func WithGetEndDeviceIdentifiers(f GetEndDeviceIdentifiersFunc) Option {
	return optionFunc(func(i *impl) {
		i.getIdentifiers = f
	})
}

// WithPayloadProcessor sets the PayloadProcessor that will be used by the gRPC frontend.
func WithPayloadProcessor(processor messageprocessors.PayloadProcessor) Option {
	return optionFunc(func(i *impl) {
		i.processor = processor
	})
}

// WithSkipPayloadCrypto sets the skip payload crypto predicate that will be used by the gRPC frontend.
func WithSkipPayloadCrypto(f SkipPayloadCryptoFunc) Option {
	return optionFunc(func(i *impl) {
		i.skipPayloadCrypto = f
	})
}

// New returns a new gRPC frontend.
func New(server io.Server, opts ...Option) ttnpb.AppAsServer {
	i := &impl{
		server: server,
		getIdentifiers: func(_ context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDeviceIdentifiers, error) {
			return ids, nil
		},
		processor: &defaultMessageProcessor{},
		skipPayloadCrypto: func(_ context.Context, _ *ttnpb.EndDeviceIdentifiers) (bool, error) {
			return false, nil
		},
	}
	for _, opt := range opts {
		opt.apply(i)
	}
	return i
}

var errConnect = errors.Define("connect", "failed to connect application `{application_uid}`")

func (s *impl) Subscribe(ids *ttnpb.ApplicationIdentifiers, stream ttnpb.AppAs_SubscribeServer) error {
	ctx := log.NewContextWithField(stream.Context(), "namespace", "applicationserver/io/grpc")

	if err := rights.RequireApplication(ctx, *ids, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return err
	}

	if peer, ok := peer.FromContext(ctx); ok {
		ctx = log.NewContextWithField(ctx, "remote_addr", peer.Addr.String())
	}
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "application_uid", uid)
	logger := log.FromContext(ctx)

	sub, err := s.server.Subscribe(ctx, "grpc", ids, true)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return errConnect.WithCause(err).WithAttributes("application_uid", uid)
	}
	logger.Info("Subscribed")
	defer logger.Info("Unsubscribed")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sub.Context().Done():
			return sub.Context().Err()
		case up := <-sub.Up():
			if err := stream.Send(up.ApplicationUp); err != nil {
				logger.WithError(err).Warn("Failed to send message")
				sub.Disconnect(err)
				return err
			}
		}
	}
}

func (s *impl) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE); err != nil {
		return nil, err
	}
	if err := s.server.DownlinkQueuePush(ctx, req.EndDeviceIds, req.Downlinks); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *impl) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE); err != nil {
		return nil, err
	}
	if err := s.server.DownlinkQueueReplace(ctx, req.EndDeviceIds, req.Downlinks); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *impl) DownlinkQueueList(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	if err := rights.RequireApplication(ctx, *ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	items, err := s.server.DownlinkQueueList(ctx, ids)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{
		Downlinks: items,
	}, nil
}

var errNoMQTTConfigProvider = errors.DefineUnimplemented("no_configuration_provider", "no MQTT configuration provider available")

func (s *impl) GetMQTTConnectionInfo(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.MQTTConnectionInfo, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.Right_RIGHT_APPLICATION_INFO); err != nil {
		return nil, err
	}
	if s.mqttConfigProvider == nil {
		return nil, errNoMQTTConfigProvider.New()
	}
	config, err := s.mqttConfigProvider.GetMQTTConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &ttnpb.MQTTConnectionInfo{
		PublicAddress:    config.PublicAddress,
		PublicTlsAddress: config.PublicTLSAddress,
		Username:         unique.ID(ctx, *ids),
	}, nil
}

var errPayloadCryptoSkipped = errors.DefineFailedPrecondition("payload_crypto_skipped", "payload crypto skipped")

func (s *impl) SimulateUplink(ctx context.Context, up *ttnpb.ApplicationUp) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *up.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_UP_WRITE); err != nil {
		return nil, err
	}
	skip, err := s.skipPayloadCrypto(ctx, up.EndDeviceIds)
	if err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to determine if the payload crypto should be skipped")
	} else if skip {
		return nil, errPayloadCryptoSkipped.New()
	}
	up.Simulated = true
	ids, err := s.getIdentifiers(ctx, up.EndDeviceIds)
	if err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to fetch end device identifiers")
	} else {
		up.EndDeviceIds = ids
	}
	if err := s.server.Publish(ctx, up); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *impl) EncodeDownlink(ctx context.Context, req *ttnpb.EncodeDownlinkRequest) (*ttnpb.EncodeDownlinkResponse, error) {
	if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	if err := s.processor.EncodeDownlink(ctx, req.EndDeviceIds, req.VersionIds, req.Downlink, req.Formatter, req.Parameter); err != nil {
		return nil, err
	}
	return &ttnpb.EncodeDownlinkResponse{
		Downlink: req.Downlink,
	}, nil
}

func (s *impl) DecodeUplink(ctx context.Context, req *ttnpb.DecodeUplinkRequest) (*ttnpb.DecodeUplinkResponse, error) {
	if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	if err := s.processor.DecodeUplink(ctx, req.EndDeviceIds, req.VersionIds, req.Uplink, req.Formatter, req.Parameter); err != nil {
		return nil, err
	}
	return &ttnpb.DecodeUplinkResponse{
		Uplink: req.Uplink,
	}, nil
}

func (s *impl) DecodeDownlink(ctx context.Context, req *ttnpb.DecodeDownlinkRequest) (*ttnpb.DecodeDownlinkResponse, error) {
	if err := rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	if err := s.processor.DecodeDownlink(ctx, req.EndDeviceIds, req.VersionIds, req.Downlink, req.Formatter, req.Parameter); err != nil {
		return nil, err
	}
	return &ttnpb.DecodeDownlinkResponse{
		Downlink: req.Downlink,
	}, nil
}
