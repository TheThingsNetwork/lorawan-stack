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
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
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
	// storedAtMarker is used to store the timestamp of the last Set operation.
	storedAtMarker = "_stored_at"
	// errorMarker is used to store errors.
	errorMarker = "_error"
)

var errCacheMiss = errors.DefineNotFound("cache_miss", "cache miss")

// Get returns the locations by the end device identifiers.
func (r *EndDeviceLocationCache) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, *time.Time, error) {
	uidKey := r.uidKey(unique.ID(ctx, ids))
	m, err := r.Redis.HGetAll(ctx, uidKey).Result()
	if err != nil {
		return nil, nil, ttnredis.ConvertError(err)
	}
	if len(m) == 0 {
		return nil, nil, errCacheMiss.New()
	}
	var storedAt time.Time
	if s, ok := m[storedAtMarker]; ok {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		storedAt = time.Unix(0, n)
		delete(m, storedAtMarker)
	}
	if s, ok := m[errorMarker]; ok {
		details := &ttnpb.ErrorDetails{}
		if err := ttnredis.UnmarshalProto(s, details); err != nil {
			return nil, nil, err
		}
		return nil, &storedAt, ttnpb.ErrorDetailsFromProto(details)
	}
	if len(m) == 0 {
		return nil, &storedAt, nil
	}
	locations := make(map[string]*ttnpb.Location, len(m))
	for k, v := range m {
		loc := new(ttnpb.Location)
		if err := ttnredis.UnmarshalProto(v, loc); err != nil {
			return nil, nil, err
		}
		locations[k] = loc
	}
	return locations, &storedAt, nil
}

// Set updates the locations by the end device identifiers.
func (r *EndDeviceLocationCache) Set(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location, ttl time.Duration) error {
	pairs := append(make([]string, 0, 2*len(update)+2), storedAtMarker, fmt.Sprintf("%v", time.Now().UnixNano()))
	for k, v := range update {
		s, err := ttnredis.MarshalProto(v)
		if err != nil {
			return err
		}
		pairs = append(pairs, k, s)
	}
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

// Delete deletes the locations by the end device identifiers.
func (r *EndDeviceLocationCache) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) error {
	uidKey := r.uidKey(unique.ID(ctx, ids))
	if err := r.Redis.Del(ctx, uidKey).Err(); err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}
