// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package cryptoutil

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

type cryptoMetrics struct {
	cacheHit  *metrics.ContextualCounterVec
	cacheMiss *metrics.ContextualCounterVec
}

func (m cryptoMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.cacheHit.Describe(ch)
	m.cacheMiss.Describe(ch)
}

func (m cryptoMetrics) Collect(ch chan<- prometheus.Metric) {
	m.cacheHit.Collect(ch)
	m.cacheMiss.Collect(ch)
}

const (
	subsystem = "cryptoutil"
)

var cMetrics = &cryptoMetrics{
	cacheHit: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "cache_hit",
			Help:      "Number of cache hits",
		},
		[]string{"cache"},
	),
	cacheMiss: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "cache_miss",
			Help:      "Number of cache misses",
		},
		[]string{"cache"},
	),
}

func init() {
	metrics.MustRegister(cMetrics)
}

type unwrapEntry struct {
	value []byte
	err   error
}

type cachedVault struct {
	crypto.KeyVault
	unwrapCache gcache.Cache
}

func NewCacheKeyVault(main crypto.KeyVault, ttl time.Duration, size int) crypto.KeyVault {
	return &cachedVault{
		KeyVault:    main,
		unwrapCache: gcache.New(size).Expiration(ttl).ARC().Build(),
	}
}

func unwrapCacheKey(ciphertext []byte, kekLabel string) string {
	return fmt.Sprintf("%v:%v", kekLabel, ciphertext)
}

func (c *cachedVault) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	id := unwrapCacheKey(ciphertext, kekLabel)
	if val, err := c.unwrapCache.Get(id); err == nil {
		cMetrics.cacheHit.WithLabelValues(ctx, "unwrap").Inc()
		v := val.(*unwrapEntry)
		return v.value, v.err
	}
	v := &unwrapEntry{}
	c.unwrapCache.Set(id, v)
	cMetrics.cacheMiss.WithLabelValues(ctx, "unwrap").Inc()
	v.value, v.err = c.KeyVault.Unwrap(ctx, ciphertext, kekLabel)
	return v.value, v.err
}
