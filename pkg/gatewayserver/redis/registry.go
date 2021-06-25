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

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// GatewayConnectionStatsRegistry implements the GatewayConnectionStatsRegistry interface.
type GatewayConnectionStatsRegistry struct {
	Redis *ttnredis.Client
}

func (r *GatewayConnectionStatsRegistry) key(uid string) string {
	return r.Redis.Key("uid", uid)
}

// Set sets or clears the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Set(ctx context.Context, ids ttnpb.GatewayIdentifiers, stats *ttnpb.GatewayConnectionStats, paths []string) error {
	uid := unique.ID(ctx, ids)

	defer trace.StartRegion(ctx, "set gateway connection stats").End()

	uk := r.key(uid)
	var err error
	if stats == nil {
		err = r.Redis.Del(ctx, uk).Err()
	} else {
		err = r.Redis.Watch(ctx, func(tx *redis.Tx) error {
			pb := &ttnpb.GatewayConnectionStats{}
			if err := ttnredis.GetProto(ctx, tx, uk).ScanProto(pb); err != nil && !errors.IsNotFound(err) {
				return err
			}

			if err := pb.SetFields(stats, paths...); err != nil {
				return err
			}
			_, err := ttnredis.SetProto(ctx, tx, uk, pb, 0)
			return err
		}, uk)
	}
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// Get returns the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Get(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error) {
	uid := unique.ID(ctx, ids)
	result := &ttnpb.GatewayConnectionStats{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.key(uid)).ScanProto(result); err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return result, nil
}
