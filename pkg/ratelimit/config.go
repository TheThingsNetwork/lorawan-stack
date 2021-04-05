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

	"github.com/throttled/throttled"
	"github.com/throttled/throttled/v2/store/memstore"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// defaultMaxSize for the rate limiter store.
const defaultMaxSize = 1 << 12

// Profile represents configuration for a rate limiting class.
type Profile struct {
	Name         string   `name:"name" description:"Rate limiting class name"`
	MaxPerMin    uint     `name:"max-per-min" description:"Maximum allowed rate (per minute)"`
	MaxBurst     uint     `name:"max-burst" description:"Maximum rate allowed for short bursts"`
	Associations []string `name:"associations" description:"List of classes to apply this profile on"`
}

// MemoryConfig represents configuration for the in-memory rate limiting store.
type MemoryConfig struct {
	MaxSize uint `name:"max-size" description:"Maximum store size for the rate limiter"`
}

// Config represents configuration for rate limiting.
type Config struct {
	Memory   MemoryConfig `name:"memory" description:"In-memory rate limiting store configuration"`
	Profiles []Profile    `name:"profiles" description:"Rate limiting profiles"`
}

var errRateLimitExceeded = errors.DefineResourceExhausted("rate_limit_exceeded", "rate limit of `{rate}` accesses per minute exceeded for resource `{key}`")

// New creates a new ratelimit.Interface from configuration.
func (c Config) New(ctx context.Context) (Interface, error) {
	defaultLimiter := &NoopRateLimiter{}
	if len(c.Profiles) == 0 {
		return defaultLimiter, nil
	}

	l := &muxRateLimiter{
		defaultLimiter: defaultLimiter,
		limiters:       make(map[string]Interface, len(c.Profiles)),
	}
	for _, profile := range c.Profiles {
		if len(profile.Associations) == 0 {
			continue
		}
		limiter, err := profile.New(ctx, c.Memory.MaxSize)
		if err != nil {
			return nil, err
		}
		for _, assocName := range profile.Associations {
			l.limiters[assocName] = limiter
		}
	}
	return l, nil
}

// New creates a new ratelimit.Interface from configuration.
func (c Profile) New(ctx context.Context, size uint) (Interface, error) {
	if size == 0 {
		size = defaultMaxSize
	}
	store, err := memstore.New(int(size))
	if err != nil {
		return nil, err
	}
	if c.MaxBurst == 0 {
		c.MaxBurst = c.MaxPerMin
	}
	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(int(c.MaxPerMin)),
		MaxBurst: int(c.MaxBurst - 1),
	}
	limiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		return nil, err
	}
	return &rateLimiter{ctx, limiter}, nil
}
