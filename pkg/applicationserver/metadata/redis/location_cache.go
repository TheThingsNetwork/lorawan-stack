// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// EndDeviceLocationCache is a Redis end device location cache.
type EndDeviceLocationCache struct {
	Redis *ttnredis.Client
}

func (r *EndDeviceLocationCache) uidKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

const (
	// cachedMarker signals that we have cached the end device locations even if there
	// are no locations available, as Redis does not make a distinction between empty
	// keys and non existing keys.
	cachedMarker = "_cached"
	// errorMarker is used to store errors.
	errorMarker = "_error"
)

var errCacheMiss = errors.DefineNotFound("cache_miss", "cache miss")

// Get returns the locations by the end device identifiers.
func (r *EndDeviceLocationCache) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, time.Duration, error) {
	uidKey := r.uidKey(unique.ID(ctx, ids))
	var (
		hGetAllCmd *redis.StringStringMapCmd
		ttlCmd     *redis.DurationCmd
	)
	if _, err := r.Redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		hGetAllCmd = p.HGetAll(ctx, uidKey)
		ttlCmd = p.PTTL(ctx, uidKey)
		return nil
	}); err != nil {
		return nil, 0, ttnredis.ConvertError(err)
	}
	m, err := hGetAllCmd.Result()
	if err != nil {
		return nil, 0, ttnredis.ConvertError(err)
	}
	if len(m) == 0 {
		return nil, 0, errCacheMiss.New()
	}
	ttl, err := ttlCmd.Result()
	if err != nil {
		return nil, 0, ttnredis.ConvertError(err)
	}
	if s, ok := m[errorMarker]; ok {
		details := &ttnpb.ErrorDetails{}
		if err := ttnredis.UnmarshalProto(s, details); err != nil {
			return nil, 0, err
		}
		return nil, ttl, ttnpb.ErrorDetailsFromProto(details)
	}
	delete(m, cachedMarker)
	if len(m) == 0 {
		return nil, ttl, nil
	}
	locations := make(map[string]*ttnpb.Location, len(m))
	for k, v := range m {
		loc := new(ttnpb.Location)
		if err := ttnredis.UnmarshalProto(v, loc); err != nil {
			return nil, 0, err
		}
		locations[k] = loc
	}
	return locations, ttl, nil
}

func (r *EndDeviceLocationCache) setPairs(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, pairs []string, ttl time.Duration) error {
	uidKey := r.uidKey(unique.ID(ctx, ids))
	if _, err := r.Redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.Del(ctx, uidKey)
		p.HSet(ctx, uidKey, pairs)
		p.PExpire(ctx, uidKey, ttl)
		return nil
	}); err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// SetLocations updates the locations by the end device identifiers.
func (r *EndDeviceLocationCache) SetLocations(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location, ttl time.Duration) error {
	pairs := append(make([]string, 0, 2*len(update)+2), cachedMarker, cachedMarker)
	for k, v := range update {
		s, err := ttnredis.MarshalProto(v)
		if err != nil {
			return err
		}
		pairs = append(pairs, k, s)
	}
	return r.setPairs(ctx, ids, pairs, ttl)
}

// SetErrorDetails stores the location retrieval error by the end device identifiers.
func (r *EndDeviceLocationCache) SetErrorDetails(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, details *ttnpb.ErrorDetails, ttl time.Duration) error {
	s, err := ttnredis.MarshalProto(details)
	if err != nil {
		return err
	}
	return r.setPairs(ctx, ids, []string{errorMarker, s}, ttl)
}

// Delete deletes the locations by the end device identifiers.
func (r *EndDeviceLocationCache) Delete(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) error {
	uidKey := r.uidKey(unique.ID(ctx, ids))
	if err := r.Redis.Del(ctx, uidKey).Err(); err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}
