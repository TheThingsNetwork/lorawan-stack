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

package devicerepository

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
)

type codecType string

const (
	downlinkEncoder codecType = "downlinkEncoder"
	downlinkDecoder codecType = "downlinkDecoder"
	uplinkDecoder   codecType = "uplinkDecoder"

	// cacheTTL is the TTL for cached payload formatters.
	cacheTTL = time.Hour
	// cacheErrorTTL is the TTL for cached payload formatters, when there was an error retrieving them.
	cacheErrorTTL = 5 * time.Minute
	// cacheSize is the cache size.
	cacheSize = 4096
)

// Cluster represents the interface the cluster.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
}

// PayloadEncoderDecoderProvider provides a messageprocessors.PayloadEncoderDecoder
// for the provided formatter.
type PayloadEncoderDecoderProvider interface {
	GetPayloadEncoderDecoder(ctx context.Context, formatter ttnpb.PayloadFormatter) (messageprocessors.PayloadEncoderDecoder, error)
}

// cacheProcessors stores the payload processors.
type cacheProcessors struct {
	uplinkProcessor   func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationUplink) error
	downlinkProcessor func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error
}

// cacheItem stores the payload processors as well as the error response.
type cacheItem struct {
	processors *cacheProcessors
	err        error
}

type host struct {
	ctx context.Context

	cluster  Cluster
	provider PayloadEncoderDecoderProvider

	singleflight singleflight.Group
	cache        gcache.Cache
}

// New creates a new PayloadEncoderDecoder that retrieves codecs from the Device Repository
// and uses an underlying PayloadEncoderDecoder to execute them.
func New(provider PayloadEncoderDecoderProvider, cluster Cluster) messageprocessors.PayloadEncoderDecoder {
	return &host{
		cluster:  cluster,
		provider: provider,

		cache: gcache.New(cacheSize).LFU().Build(),
	}
}

func cacheKey(codec codecType, version *ttnpb.EndDeviceVersionIdentifiers) string {
	return fmt.Sprintf("%s:%s:%s:%s:%v", version.BrandId, version.ModelId, version.FirmwareVersion, version.BandId, codec)
}

var errNoVersionIdentifiers = errors.DefineInvalidArgument("no_version_identifiers", "no version identifiers for device")

func (h *host) retrieve(ctx context.Context, codec codecType, ids *ttnpb.ApplicationIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers) (*cacheProcessors, error) {
	if version == nil {
		return nil, errNoVersionIdentifiers.New()
	}
	key := cacheKey(codec, version)
	if cachedInterface, err := h.cache.Get(key); err == nil {
		cached := cachedInterface.(*cacheItem)
		return cached.processors, cached.err
	}
	cc, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_DEVICE_REPOSITORY, nil)
	if err != nil {
		return nil, err
	}
	result, err, _ := h.singleflight.Do(key, func() (any, error) {
		var (
			formatter interface {
				GetFormatter() ttnpb.PayloadFormatter
				GetFormatterParameter() string
			}
			err error
		)
		req := &ttnpb.GetPayloadFormatterRequest{
			ApplicationIds: ids,
			VersionIds:     version,
		}
		client := ttnpb.NewDeviceRepositoryClient(cc)
		switch codec {
		case downlinkDecoder:
			formatter, err = client.GetDownlinkDecoder(ctx, req, h.cluster.WithClusterAuth())
		case downlinkEncoder:
			formatter, err = client.GetDownlinkEncoder(ctx, req, h.cluster.WithClusterAuth())
		case uplinkDecoder:
			formatter, err = client.GetUplinkDecoder(ctx, req, h.cluster.WithClusterAuth())
		default:
			panic(fmt.Sprintf("Invalid codec type: %v", codec))
		}
		var cachedProcessors *cacheProcessors
		if err == nil {
			cachedProcessors, err = h.compileProcessor(ctx, codec, formatter.GetFormatter(), formatter.GetFormatterParameter())
		}
		expire := cacheTTL
		if err != nil {
			expire = cacheErrorTTL
		}
		if err := h.cache.SetWithExpire(key, &cacheItem{cachedProcessors, err}, expire); err != nil {
			log.FromContext(h.ctx).WithError(err).Error("Failed to cache payload formatter")
		}
		return cachedProcessors, err
	})
	if err != nil {
		return nil, err
	}
	return result.(*cacheProcessors), nil
}

func (h *host) compileProcessor(ctx context.Context, codec codecType, formatter ttnpb.PayloadFormatter, parameter string) (*cacheProcessors, error) {
	encoderDecoder, err := h.provider.GetPayloadEncoderDecoder(ctx, formatter)
	if err != nil {
		return nil, err
	}

	if compilableEncoderDecoder, canCompile := encoderDecoder.(messageprocessors.CompilablePayloadEncoderDecoder); canCompile {
		switch codec {
		case downlinkDecoder:
			run, err := compilableEncoderDecoder.CompileDownlinkDecoder(ctx, parameter)
			if err != nil {
				return nil, err
			}
			return &cacheProcessors{
				downlinkProcessor: run,
			}, nil
		case downlinkEncoder:
			run, err := compilableEncoderDecoder.CompileDownlinkEncoder(ctx, parameter)
			if err != nil {
				return nil, err
			}
			return &cacheProcessors{
				downlinkProcessor: run,
			}, nil
		case uplinkDecoder:
			run, err := compilableEncoderDecoder.CompileUplinkDecoder(ctx, parameter)
			if err != nil {
				return nil, err
			}
			return &cacheProcessors{
				uplinkProcessor: run,
			}, nil
		default:
			panic(fmt.Sprintf("invalid codec type: %v", codec))
		}
	}

	switch codec {
	case downlinkDecoder:
		return &cacheProcessors{
			downlinkProcessor: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink) error {
				return encoderDecoder.DecodeDownlink(ctx, ids, version, message, parameter)
			},
		}, nil
	case downlinkEncoder:
		return &cacheProcessors{
			downlinkProcessor: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink) error {
				return encoderDecoder.EncodeDownlink(ctx, ids, version, message, parameter)
			},
		}, nil
	case uplinkDecoder:
		return &cacheProcessors{
			uplinkProcessor: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink) error {
				return encoderDecoder.DecodeUplink(ctx, ids, version, message, parameter)
			},
		}, nil
	default:
		panic(fmt.Sprintf("invalid codec type: %v", codec))
	}
}

// EncodeDownlink encodes a downlink message.
func (h *host) EncodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
	res, err := h.retrieve(ctx, downlinkEncoder, ids.ApplicationIds, version)
	if err != nil {
		return err
	}
	return res.downlinkProcessor(ctx, ids, version, message)
}

// DecodeUplink decodes an uplink message.
func (h *host) DecodeUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error {
	res, err := h.retrieve(ctx, uplinkDecoder, ids.ApplicationIds, version)
	if err != nil {
		return err
	}
	return res.uplinkProcessor(ctx, ids, version, message)
}

// DecodeDownlink decodes a downlink message.
func (h *host) DecodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
	res, err := h.retrieve(ctx, downlinkDecoder, ids.ApplicationIds, version)
	if err != nil {
		return err
	}
	return res.downlinkProcessor(ctx, ids, version, message)
}
