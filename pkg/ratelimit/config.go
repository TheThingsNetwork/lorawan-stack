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

package ratelimit

import (
	"context"

	"github.com/throttled/throttled/v2"
	redisstore "github.com/throttled/throttled/v2/store/goredisstore.v9"
	"github.com/throttled/throttled/v2/store/memstore"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"gopkg.in/yaml.v2"
)

// defaultMaxSize for the rate limiter store.
const defaultMaxSize = 1 << 12

var errRateLimitExceeded = errors.DefineResourceExhausted("rate_limit_exceeded", "rate limit of `{rate}` accesses per minute exceeded for resource `{key}`")

// StoreConfig represents configuration for rate limiting stores.
type StoreConfig struct {
	Provider string
	Memory   config.RateLimitingMemory
	Redis    *ttnredis.Client
}

// New creates a new ratelimit.Interface from configuration.
func New(ctx context.Context, conf config.RateLimiting, blobConf config.BlobConfig, httpClientProvider httpclient.Provider) (Interface, error) {
	defaultLimiter := &NoopRateLimiter{}
	profiles := conf.Profiles

	fetcher, err := conf.Fetcher(ctx, blobConf, httpClientProvider)
	if err != nil {
		return nil, err
	}
	if fetcher != nil {
		b, err := fetcher.File("rate-limiting.yml")
		if err != nil {
			return nil, err
		}
		var overrideProfiles struct {
			Profiles []config.RateLimitingProfile `yaml:"profiles"`
		}
		if err := yaml.Unmarshal(b, &overrideProfiles); err != nil {
			return nil, err
		}
		profiles = append(profiles, overrideProfiles.Profiles...)
	}
	if len(profiles) == 0 {
		return defaultLimiter, nil
	}

	l := &muxRateLimiter{
		defaultLimiter: defaultLimiter,
		limiters:       make(map[string]Interface, len(profiles)),
	}
	for _, profile := range profiles {
		if len(profile.Associations) == 0 {
			continue
		}
		limiter, err := NewProfile(ctx, profile, StoreConfig{
			Provider: conf.Provider,
			Memory:   conf.Memory,
			Redis:    conf.Redis.Client,
		})
		if err != nil {
			return nil, err
		}
		for _, assocName := range profile.Associations {
			l.limiters[assocName] = limiter
		}
	}
	return l, nil
}

var errInvalidRate = errors.DefineInvalidArgument("invalid_rate", "invalid rate `{rate}` for profile `{name}`")

func newStore(conf StoreConfig) (store throttled.GCRAStoreCtx, err error) {
	switch conf.Provider {
	case "redis":
		return redisstore.NewCtx(conf.Redis, conf.Redis.Key(""))

	default:
		return memstore.NewCtx(int(conf.Memory.MaxSize))
	}
}

// NewProfile returns a new ratelimit.Interface from profile configuration.
func NewProfile(ctx context.Context, conf config.RateLimitingProfile, storeConf StoreConfig) (Interface, error) {
	if s := &storeConf.Memory.MaxSize; *s == 0 {
		*s = defaultMaxSize
	}
	if conf.MaxPerMin == 0 {
		return nil, errInvalidRate.WithAttributes("rate", conf.MaxPerMin, "name", conf.Name)
	}
	store, err := newStore(storeConf)
	if err != nil {
		return nil, err
	}
	if conf.MaxBurst == 0 {
		conf.MaxBurst = conf.MaxPerMin
	}
	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(int(conf.MaxPerMin)),
		MaxBurst: int(conf.MaxBurst - 1),
	}
	limiter, err := throttled.NewGCRARateLimiterCtx(store, quota)
	if err != nil {
		return nil, err
	}
	return &rateLimiter{ctx, limiter}, nil
}
