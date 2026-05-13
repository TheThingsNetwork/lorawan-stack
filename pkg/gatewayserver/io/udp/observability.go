// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package udp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
)

const subsystem = "gs_io_udp"

var udpMetrics = &messageMetrics{
	messageReceived: metrics.NewCounter(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "message_received_total",
			Help:      "Total number of received UDP messages",
		},
	),
	messageForwarded: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "message_forwarded_total",
			Help:      "Total number of forwarded UDP messages",
		},
		[]string{"type"},
	),
	messageDropped: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "message_dropped_total",
			Help:      "Total number of dropped UDP messages",
		},
		[]string{"error"},
	),

	unmarshalTypeErrors: metrics.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "unmarshal_type_errors_total",
			Help:      "Total number of unmarshal type errors",
		},
		[]string{"type", "struct", "field"},
	),
	schedulingDelay: metrics.NewHistogram(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "scheduling_delay_seconds",
			Help:      "Histogram of the delay (seconds) introduced when scheduling downlink messages late",
			// 1s, 1m, 5m, 10m, 30m, 1h, 1h30m, 2h, 3h, 4h, 6h.
			Buckets: []float64{1, 60, 300, 600, 1800, 3600, 5400, 7200, 10800, 14400, 21600},
		},
	),
	pendingDownlinks: metrics.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "pending_downlinks",
			Help:      "Number of late-scheduled downlink messages waiting to be written",
		},
	),
}

func init() {
	metrics.MustRegister(udpMetrics)
}

type messageMetrics struct {
	messageReceived  prometheus.Counter
	messageForwarded *prometheus.CounterVec
	messageDropped   *prometheus.CounterVec

	unmarshalTypeErrors *prometheus.CounterVec
	schedulingDelay     prometheus.Histogram
	pendingDownlinks    prometheus.Gauge
}

// Describe implements prometheus.Collector.
func (m messageMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.messageReceived.Describe(ch)
	m.messageForwarded.Describe(ch)
	m.messageDropped.Describe(ch)

	m.unmarshalTypeErrors.Describe(ch)
	m.schedulingDelay.Describe(ch)
	m.pendingDownlinks.Describe(ch)
}

// Collect implements prometheus.Collector.
func (m messageMetrics) Collect(ch chan<- prometheus.Metric) {
	m.messageReceived.Collect(ch)
	m.messageForwarded.Collect(ch)
	m.messageDropped.Collect(ch)

	m.unmarshalTypeErrors.Collect(ch)
	m.schedulingDelay.Collect(ch)
	m.pendingDownlinks.Collect(ch)
}

func registerMessageReceived(_ context.Context) {
	udpMetrics.messageReceived.Inc()
}

func registerMessageForwarded(_ context.Context, tp encoding.PacketType) {
	udpMetrics.messageForwarded.WithLabelValues(tp.String()).Inc()
}

func registerSchedulingDelay(_ context.Context, delay time.Duration) {
	udpMetrics.schedulingDelay.Observe(delay.Seconds())
}

func registerPendingDownlinkAdded(_ context.Context) {
	udpMetrics.pendingDownlinks.Inc()
}

func registerPendingDownlinkRemoved(_ context.Context) {
	udpMetrics.pendingDownlinks.Dec()
}

func registerMessageDropped(_ context.Context, err error) {
	errorLabel := "unknown"
	if ttnErr, ok := errors.From(err); ok {
		errorLabel = ttnErr.FullName()
	} else if jsonErr := (&json.SyntaxError{}); errors.As(err, &jsonErr) {
		errorLabel = "encoding/json:syntax"
	} else if jsonErr := (&json.UnmarshalTypeError{}); errors.As(err, &jsonErr) {
		errorLabel = "encoding/json:unmarshal_type"
		udpMetrics.unmarshalTypeErrors.WithLabelValues(jsonErr.Type.Name(), jsonErr.Struct, jsonErr.Field).Inc()
	}
	udpMetrics.messageDropped.WithLabelValues(errorLabel).Inc()
}
