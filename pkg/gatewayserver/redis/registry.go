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

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// GatewayConnectionStatsRegistry implements the GatewayConnectionStatsRegistry interface.
type GatewayConnectionStatsRegistry struct {
	Redis *ttnredis.Client
}

var (
	down            = "down"
	up              = "up"
	status          = "status"
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
		for _, this := range []struct {
			key    string
			update bool
		}{
			{r.key(up, uid), updateUp},
			{r.key(down, uid), updateDown},
			{r.key(status, uid), updateStatus},
		} {
			if this.update {
				if stats == nil {
					p.Del(this.key)
				} else {
					ttnredis.SetProto(p, this.key, stats, 0)
				}
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

	retrieved, err := r.Redis.MGet(r.key(up, uid), r.key(down, uid), r.key(status, uid)).Result()
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
		result.RoundTripTimes = stats.RoundTripTimes
	}

	// Retrieve downlink stats.
	if retrieved[1] != nil {
		if err = ttnredis.UnmarshalProto(retrieved[1].(string), stats); err != nil {
			return nil, errInvalidStats.WithAttributes("type", "downlink").WithCause(err)
		}
		result.LastDownlinkReceivedAt = stats.LastDownlinkReceivedAt
		result.DownlinkCount = stats.DownlinkCount
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
