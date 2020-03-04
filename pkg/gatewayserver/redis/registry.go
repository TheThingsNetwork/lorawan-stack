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

	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// GatewayConnectionStatsRegistry implements the GatewayConnectionStatsRegistry interface.
type GatewayConnectionStatsRegistry struct {
	Redis *ttnredis.Client
}

func (r *GatewayConnectionStatsRegistry) key(uid string) string {
	return r.Redis.Key("uid", uid)
}

// Set sets or clears the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Set(ctx context.Context, ids ttnpb.GatewayIdentifiers, stats *ttnpb.GatewayConnectionStats) error {
	uid := unique.ID(ctx, ids)
	if stats == nil {
		return r.Redis.Del(r.key(uid)).Err()
	}

	_, err := ttnredis.SetProto(r.Redis, r.key(uid), stats, 0)
	return err
}

// Get returns the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Get(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error) {
	uid := unique.ID(ctx, ids)
	stats := &ttnpb.GatewayConnectionStats{}
	err := ttnredis.GetProto(r.Redis, r.key(uid)).ScanProto(stats)
	if err != nil {
		stats = nil
	}
	return stats, err
}
