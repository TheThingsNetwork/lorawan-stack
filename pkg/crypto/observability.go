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

package crypto

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
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
	subsystem = "crypto"
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

// RegisterCacheHit registers a cache hit for the provided cache.
func RegisterCacheHit(ctx context.Context, cache string) {
	cMetrics.cacheHit.WithLabelValues(ctx, cache).Inc()
}

// RegisterCacheMiss registers a cache miss for the provided cache.
func RegisterCacheMiss(ctx context.Context, cache string) {
	cMetrics.cacheMiss.WithLabelValues(ctx, cache).Inc()
}
