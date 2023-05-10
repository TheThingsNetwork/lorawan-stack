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
	"time"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// GatewayConnectionStatsRegistry implements the GatewayConnectionStatsRegistry interface.
type GatewayConnectionStatsRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

// Init initializes the GatewayConnectionStatsRegistry.
func (r *GatewayConnectionStatsRegistry) Init(ctx context.Context) error {
	return ttnredis.InitMutex(ctx, r.Redis)
}

func (r *GatewayConnectionStatsRegistry) key(uid string) string {
	return r.Redis.Key("uid", uid)
}

// Set sets or clears the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Set(
	ctx context.Context,
	ids *ttnpb.GatewayIdentifiers,
	f func(*ttnpb.GatewayConnectionStats) (*ttnpb.GatewayConnectionStats, []string, error),
	ttl time.Duration,
	gets ...string,
) error {
	uid := unique.ID(ctx, ids)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return err
	}

	defer trace.StartRegion(ctx, "set gateway connection stats").End()

	uk := r.key(uid)
	err = ttnredis.LockedWatch(ctx, r.Redis, uk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		stored := &ttnpb.GatewayConnectionStats{}
		cmd := ttnredis.GetProto(ctx, tx, uk)
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var pb *ttnpb.GatewayConnectionStats
		if stored != nil {
			pb = &ttnpb.GatewayConnectionStats{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			if pb, err = applyGatewayConnectionStatsFieldMask(nil, pb, gets...); err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(pb)
		if err != nil {
			return err
		}
		if stored == nil && pb == nil {
			return nil
		}
		var pipelined func(redis.Pipeliner) error
		if pb == nil {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(ctx, uk)
				return nil
			}
		} else {
			updated := &ttnpb.GatewayConnectionStats{}
			if stored != nil {
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
			}
			if updated, err = applyGatewayConnectionStatsFieldMask(updated, pb, sets...); err != nil {
				return err
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}
			pipelined = func(p redis.Pipeliner) error {
				_, err = ttnredis.SetProto(ctx, p, uk, updated, ttl)
				return err
			}
		}
		_, err = tx.TxPipelined(ctx, pipelined)
		return err
	})
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// Get returns the connection stats for a gateway.
func (r *GatewayConnectionStatsRegistry) Get(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers,
) (*ttnpb.GatewayConnectionStats, error) {
	uid := unique.ID(ctx, ids)
	result := &ttnpb.GatewayConnectionStats{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.key(uid)).ScanProto(result); err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return result, nil
}

func applyGatewayConnectionStatsFieldMask(
	dst, src *ttnpb.GatewayConnectionStats,
	paths ...string,
) (*ttnpb.GatewayConnectionStats, error) {
	if dst == nil {
		dst = &ttnpb.GatewayConnectionStats{}
	}
	return dst, dst.SetFields(src, paths...)
}

// BatchGet returns the connection stats for a batch of gateways.
// NotFound errors indicating that the gateway is either not connected
// or is connected to a different cluster, are ignored.
func (r *GatewayConnectionStatsRegistry) BatchGet(
	ctx context.Context,
	ids []*ttnpb.GatewayIdentifiers,
	paths ...string,
) (map[string]*ttnpb.GatewayConnectionStats, error) {
	ret := make(map[string]*ttnpb.GatewayConnectionStats, len(ids))
	keys := make([]string, 0, len(ids))
	for _, gtwIDs := range ids {
		uid := unique.ID(ctx, gtwIDs)
		keys = append(keys, r.key(uid))
	}
	rawValues, err := r.Redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}

	for i, val := range rawValues {
		switch val := val.(type) {
		case nil:
			continue
		case string:
			stats := &ttnpb.GatewayConnectionStats{}
			if err := ttnredis.UnmarshalProto(val, stats); err != nil {
				log.FromContext(ctx).WithError(err).Warnf("Failed to decode stats payload")
				continue
			}
			// Copy only the requested paths.
			if len(paths) > 0 {
				stats, err = applyGatewayConnectionStatsFieldMask(nil, stats, paths...)
				if err != nil {
					return nil, err
				}
			}

			// The result of MGet is in the same order as the input keys passed to it.
			// MGet inserts "nil" values for keys that don't have values, thereby maintaining the order.
			// So we can use the index of the result to correlate the gateway IDs.
			ret[ids[i].GatewayId] = stats
		default:
			log.FromContext(ctx).WithField("element", val).Warn("Invalid element in stats payloads")
			continue
		}
	}
	return ret, nil
}
