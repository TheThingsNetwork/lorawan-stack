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

package redis

import (
	"context"
	"runtime/trace"

	"github.com/go-redis/redis/v7"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// GatewayConnectionStatsRegistry implements the GatewayConnectionStatsRegistry interface.
type GatewayConnectionStatsRegistry struct {
	Redis *ttnredis.Client
}

const (
	downKey   = "down"
	upKey     = "up"
	statusKey = "status"
)

var (
	errNotFound     = errors.DefineNotFound("stats_not_found", "gateway stats not found")
	errInvalidStats = errors.DefineCorruption("invalid_stats", "invalid `{type}` stats in store")
)

func (r *GatewayConnectionStatsRegistry) key(which string, uid string) string {
	return r.Redis.Key(which, "uid", uid)
}

// Set sets or clears the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Set(ctx context.Context, ids ttnpb.GatewayIdentifiers, stats *ttnpb.GatewayConnectionStats, updateUp bool, updateDown bool, updateStatus bool) error {
	uid := unique.ID(ctx, ids)

	defer trace.StartRegion(ctx, "set gateway connection stats").End()

	_, err := r.Redis.Pipelined(func(p redis.Pipeliner) error {
		for _, part := range []struct {
			key    string
			update bool
			fields func() *ttnpb.GatewayConnectionStats
		}{
			{
				key:    r.key(upKey, uid),
				update: updateUp,
				fields: func() *ttnpb.GatewayConnectionStats {
					return &ttnpb.GatewayConnectionStats{
						LastUplinkReceivedAt: stats.LastUplinkReceivedAt,
						UplinkCount:          stats.UplinkCount,
					}
				},
			},
			{
				key:    r.key(downKey, uid),
				update: updateDown,
				fields: func() *ttnpb.GatewayConnectionStats {
					return &ttnpb.GatewayConnectionStats{
						LastDownlinkReceivedAt: stats.LastDownlinkReceivedAt,
						DownlinkCount:          stats.DownlinkCount,
						RoundTripTimes:         stats.RoundTripTimes,
					}
				},
			},
			{
				key:    r.key(statusKey, uid),
				update: updateStatus,
				fields: func() *ttnpb.GatewayConnectionStats {
					return &ttnpb.GatewayConnectionStats{
						ConnectedAt:          stats.ConnectedAt,
						Protocol:             stats.Protocol,
						LastStatus:           stats.LastStatus,
						LastStatusReceivedAt: stats.LastStatusReceivedAt,
					}
				},
			},
		} {
			if !part.update {
				continue
			}
			if stats == nil {
				p.Del(part.key)
			} else {
				ttnredis.SetProto(p, part.key, part.fields(), 0)
			}
		}
		return nil
	})

	if err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// Get returns the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Get(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error) {
	uid := unique.ID(ctx, ids)
	result := &ttnpb.GatewayConnectionStats{}
	stats := &ttnpb.GatewayConnectionStats{}

	retrieved, err := r.Redis.MGet(r.key(upKey, uid), r.key(downKey, uid), r.key(statusKey, uid)).Result()
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}

	if retrieved[0] == nil && retrieved[1] == nil && retrieved[2] == nil {
		return nil, errNotFound
	}

	// Retrieve uplink stats.
	if retrieved[0] != nil {
		if err = ttnredis.UnmarshalProto(retrieved[0].(string), stats); err != nil {
			return nil, errInvalidStats.WithAttributes("type", "uplink").WithCause(err)
		}
		result.LastUplinkReceivedAt = stats.LastUplinkReceivedAt
		result.UplinkCount = stats.UplinkCount
	}

	// Retrieve downlink stats.
	if retrieved[1] != nil {
		if err = ttnredis.UnmarshalProto(retrieved[1].(string), stats); err != nil {
			return nil, errInvalidStats.WithAttributes("type", "downlink").WithCause(err)
		}
		result.LastDownlinkReceivedAt = stats.LastDownlinkReceivedAt
		result.DownlinkCount = stats.DownlinkCount
		result.RoundTripTimes = stats.RoundTripTimes
	}

	// Retrieve gateway status.
	if retrieved[2] != nil {
		if err = ttnredis.UnmarshalProto(retrieved[2].(string), stats); err != nil {
			return nil, errInvalidStats.WithAttributes("type", "status").WithCause(err)
		}
		result.ConnectedAt = stats.ConnectedAt
		result.Protocol = stats.Protocol
		result.LastStatus = stats.LastStatus
		result.LastStatusReceivedAt = stats.LastStatusReceivedAt
	}

	return result, nil
}
