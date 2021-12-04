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

package metadata

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
)

const (
	subsystem     = "as_metadata"
	metadataLabel = "metadata"
	locationLabel = "location"
)

var metaMetrics = &metadataMetrics{
	cacheHits: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "cache_hits_total",
			Help:      "Total number of metadata cache hits",
		},
		[]string{metadataLabel},
	),
	cacheMisses: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "cache_misses_total",
			Help:      "Total number of metadata cache misses",
		},
		[]string{metadataLabel},
	),
	registryRetrievals: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "registry_retrievals_total",
			Help:      "Total number of metadata registry retrievals",
		},
		[]string{metadataLabel},
	),
	registryUpdates: metrics.NewContextualGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "registry_updates_total",
			Help:      "Total number of metadata registry updates",
		},
		[]string{metadataLabel},
	),
}

func init() {
	metrics.MustRegister(metaMetrics)
}

type metadataMetrics struct {
	cacheHits          *metrics.ContextualGaugeVec
	cacheMisses        *metrics.ContextualGaugeVec
	registryRetrievals *metrics.ContextualGaugeVec
	registryUpdates    *metrics.ContextualGaugeVec
}

func (m metadataMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.cacheHits.Describe(ch)
	m.cacheMisses.Describe(ch)
	m.registryRetrievals.Describe(ch)
	m.registryUpdates.Describe(ch)
}

func (m metadataMetrics) Collect(ch chan<- prometheus.Metric) {
	m.cacheHits.Collect(ch)
	m.cacheMisses.Collect(ch)
	m.registryRetrievals.Collect(ch)
	m.registryUpdates.Collect(ch)
}

func registerMetadataCacheHit(ctx context.Context, metadata string) {
	metaMetrics.cacheHits.WithLabelValues(ctx, metadata).Inc()
}

func registerMetadataCacheMiss(ctx context.Context, metadata string) {
	metaMetrics.cacheMisses.WithLabelValues(ctx, metadata).Inc()
}

func registerMetadataRegistryRetrieval(ctx context.Context, metadata string) {
	metaMetrics.registryRetrievals.WithLabelValues(ctx, metadata).Inc()
}

func registerMetadataRegistryUpdate(ctx context.Context, metadata string) {
	metaMetrics.registryUpdates.WithLabelValues(ctx, metadata).Inc()
}
