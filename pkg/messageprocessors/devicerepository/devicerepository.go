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
	downlinkEncoder = codecType("downlinkEncoder")
	downlinkDecoder = codecType("downlinkDecoder")
	uplinkDecoder   = codecType("uplinkDecoder")

	// cacheTTL is the TTL for cached payload formatters.
	cacheTTL = time.Hour
	// cacheErrorTTL is the TTL for cached payload formatters, when there was an error retrieving them.
	cacheErrorTTL = 5 * time.Minute
	// cacheSize is the cache size.
	cacheSize = 500
)

type PayloadFormatter interface {
	GetFormatter() ttnpb.PayloadFormatter
	GetFormatterParameter() string
}

// Cluster represents the interface the cluster.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
}

// cacheItem stores the payload formatter as well as the error response.
type cacheItem struct {
	formatter PayloadFormatter
	err       error
}

type host struct {
	ctx context.Context

	cluster   Cluster
	processor messageprocessors.PayloadProcessor

	singleflight singleflight.Group
	cache        gcache.Cache
}

// New creates a new PayloadEncodeDecoder that retrieves codecs from the Device Repository
// and uses an underlying PayloadEncodeDecoder to execute them.
func New(processor messageprocessors.PayloadProcessor, cluster Cluster) messageprocessors.PayloadEncodeDecoder {
	return &host{
		cluster:   cluster,
		processor: processor,

		cache: gcache.New(cacheSize).LFU().Build(),
	}
}

func cacheKey(codec codecType, version *ttnpb.EndDeviceVersionIdentifiers) string {
	return fmt.Sprintf("%s:%s:%s:%s:%v", version.BrandID, version.ModelID, version.FirmwareVersion, version.BandID, codec)
}

var errNoVersionIdentifiers = errors.DefineInvalidArgument("no_version_identifiers", "no version identifiers for device")

func (h *host) retrieve(ctx context.Context, codec codecType, ids ttnpb.ApplicationIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers) (PayloadFormatter, error) {
	if version == nil {
		return nil, errNoVersionIdentifiers.New()
	}
	key := cacheKey(codec, version)
	if cachedInterface, err := h.cache.Get(key); err == nil {
		cached := cachedInterface.(cacheItem)
		return cached.formatter, cached.err
	}
	cc, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_DEVICE_REPOSITORY, nil)
	if err != nil {
		return nil, err
	}
	result, err, _ := h.singleflight.Do(key, func() (interface{}, error) {
		var (
			formatter PayloadFormatter
			err       error
		)
		req := &ttnpb.GetPayloadFormatterRequest{
			ApplicationIds: ids,
			VersionIDs:     version,
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

		expire := cacheTTL
		if err != nil {
			expire = cacheErrorTTL
		}
		if err := h.cache.SetWithExpire(key, cacheItem{formatter, err}, expire); err != nil {
			log.FromContext(h.ctx).WithError(err).Error("Failed to cache payload formatter")
		}
		return formatter, err
	})
	if err != nil {
		return nil, err
	}
	return result.(PayloadFormatter), nil
}

// EncodeDownlink encodes a downlink message.
func (h *host) EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
	res, err := h.retrieve(ctx, downlinkEncoder, ids.ApplicationIdentifiers, version)
	if err != nil {
		return err
	}
	return h.processor.EncodeDownlink(ctx, ids, version, message, res.GetFormatter(), res.GetFormatterParameter())
}

// DecodeUplink decodes an uplink message.
func (h *host) DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error {
	res, err := h.retrieve(ctx, uplinkDecoder, ids.ApplicationIdentifiers, version)
	if err != nil {
		return err
	}
	return h.processor.DecodeUplink(ctx, ids, version, message, res.GetFormatter(), res.GetFormatterParameter())
}

// DecodeDownlink decodes a downlink message.
func (h *host) DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
	res, err := h.retrieve(ctx, downlinkDecoder, ids.ApplicationIdentifiers, version)
	if err != nil {
		return err
	}
	return h.processor.DecodeDownlink(ctx, ids, version, message, res.GetFormatter(), res.GetFormatterParameter())
}
