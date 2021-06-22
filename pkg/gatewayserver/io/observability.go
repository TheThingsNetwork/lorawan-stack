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

package io

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type messageMetrics struct {
	repeatedUplinks *metrics.ContextualCounterVec
}

// Describe implements prometheus.Collector.
func (m *messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.repeatedUplinks.Describe(ch)
}

// Collect implements prometheus.Collector.
func (m *messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.repeatedUplinks.Collect(ch)
}

var ioMetrics = &messageMetrics{
	repeatedUplinks: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: "gs",
			Name:      "uplink_repeated_total",
			Help:      "Total number of repeated gateway uplinks",
		},
		[]string{"protocol"},
	),
}

var evtRepeatUp = events.Define(
	"gs.up.repeat", "received repeated uplink message from gateway",
	events.WithVisibility(ttnpb.RIGHT_GATEWAY_TRAFFIC_READ),
	events.WithDataType(&ttnpb.GatewayIdentifiers{}),
)

func registerRepeatUp(ctx context.Context, emitEvent bool, gtw *ttnpb.Gateway, protocol string) {
	ioMetrics.repeatedUplinks.WithLabelValues(ctx, protocol).Inc()
	if emitEvent {
		events.Publish(evtRepeatUp.NewWithIdentifiersAndData(ctx, gtw, nil))
	}
}

func init() {
	metrics.MustRegister(ioMetrics)
}
