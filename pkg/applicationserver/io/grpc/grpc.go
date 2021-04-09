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

// EndDeviceFetcher retrieves end device information from identifiers.
type EndDeviceFetcher interface {
	Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error)
}

type defaultFetcher struct{}

func (f *defaultFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	return &ttnpb.EndDevice{EndDeviceIdentifiers: ids}, nil
}

type defaultMessageProcessor struct{}

func (p *defaultMessageProcessor) EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	return nil
}

func (p *defaultMessageProcessor) DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
	return nil
}

func (p *defaultMessageProcessor) DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	return nil
}

// SkipPayloadCryptoFunc is a function that checks if the end device should skip payload crypto operations.
type SkipPayloadCryptoFunc func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (bool, error)

func defaultSkipPayloadCrypto(context.Context, ttnpb.EndDeviceIdentifiers) (bool, error) {
	return false, nil
}

type impl struct {
	server             io.Server
	fetcher            EndDeviceFetcher
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

// WithEndDeviceFetcher sets the EndDeviceFetcher that will be used by the gRPC frontend.
func WithEndDeviceFetcher(f EndDeviceFetcher) Option {
	return optionFunc(func(i *impl) {
		i.fetcher = f
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
	i := &impl{server: server, fetcher: &defaultFetcher{}, processor: &defaultMessageProcessor{}, skipPayloadCrypto: defaultSkipPayloadCrypto}
	for _, opt := range opts {
		opt.apply(i)
	}
	return i
}

var errConnect = errors.Define("connect", "failed to connect application `{application_uid}`")

func (s *impl) Subscribe(ids *ttnpb.ApplicationIdentifiers, stream ttnpb.AppAs_SubscribeServer) error {
	ctx := log.NewContextWithField(stream.Context(), "namespace", "applicationserver/io/grpc")

	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
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
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE); err != nil {
		return nil, err
	}
	if err := s.server.DownlinkQueuePush(ctx, req.EndDeviceIdentifiers, req.Downlinks); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *impl) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE); err != nil {
		return nil, err
	}
	if err := s.server.DownlinkQueueReplace(ctx, req.EndDeviceIdentifiers, req.Downlinks); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *impl) DownlinkQueueList(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	items, err := s.server.DownlinkQueueList(ctx, *ids)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{
		Downlinks: items,
	}, nil
}

var errNoMQTTConfigProvider = errors.DefineUnimplemented("no_configuration_provider", "no MQTT configuration provider available")

func (s *impl) GetMQTTConnectionInfo(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.MQTTConnectionInfo, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_INFO); err != nil {
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
		PublicTLSAddress: config.PublicTLSAddress,
		Username:         unique.ID(ctx, *ids),
	}, nil
}

var errPayloadCryptoSkipped = errors.DefineFailedPrecondition("payload_crypto_skipped", "payload crypto skipped")

func (s *impl) SimulateUplink(ctx context.Context, up *ttnpb.ApplicationUp) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, up.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_UP_WRITE); err != nil {
		return nil, err
	}
	skip, err := s.skipPayloadCrypto(ctx, up.EndDeviceIdentifiers)
	if err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to determine if the payload crypto should be skipped")
	} else if skip {
		return nil, errPayloadCryptoSkipped.New()
	}
	up.Simulated = true
	dev, err := s.fetcher.Get(ctx, up.EndDeviceIdentifiers, "ids")
	if err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to fetch end device identifiers")
	} else {
		up.EndDeviceIdentifiers = dev.EndDeviceIdentifiers
	}
	if err := s.server.Publish(ctx, up); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (as *impl) EncodeDownlink(ctx context.Context, req *ttnpb.EncodeDownlinkRequest) (*ttnpb.EncodeDownlinkResponse, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	if err := as.processor.EncodeDownlink(ctx, *req.EndDeviceIds, req.VersionIds, req.Downlink, req.Formatter, req.Parameter); err != nil {
		return nil, err
	}
	return &ttnpb.EncodeDownlinkResponse{
		Downlink: req.Downlink,
	}, nil
}

func (as *impl) DecodeUplink(ctx context.Context, req *ttnpb.DecodeUplinkRequest) (*ttnpb.DecodeUplinkResponse, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	if err := as.processor.DecodeUplink(ctx, *req.EndDeviceIds, req.VersionIds, req.Uplink, req.Formatter, req.Parameter); err != nil {
		return nil, err
	}
	return &ttnpb.DecodeUplinkResponse{
		Uplink: req.Uplink,
	}, nil
}

func (as *impl) DecodeDownlink(ctx context.Context, req *ttnpb.DecodeDownlinkRequest) (*ttnpb.DecodeDownlinkResponse, error) {
	if err := rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	if err := as.processor.DecodeDownlink(ctx, *req.EndDeviceIds, req.VersionIds, req.Downlink, req.Formatter, req.Parameter); err != nil {
		return nil, err
	}
	return &ttnpb.DecodeDownlinkResponse{
		Downlink: req.Downlink,
	}, nil
}
